/* eslint-disable no-await-in-loop */

import {
  ethers,
  waffle,
  helpers,
  getUnnamedAccounts,
  getNamedAccounts,
} from "hardhat"
import { expect } from "chai"
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers"
import {
  constants,
  dkgState,
  params,
  randomBeaconDeployment,
} from "../fixtures"
import type { RandomBeacon, RandomBeaconStub, TestToken } from "../../typechain"
import {
  genesis,
  signAndSubmitCorrectDkgResult,
  noMisbehaved,
} from "../utils/dkg"
import blsData from "../data/bls"
import { registerOperators } from "../utils/operators"

const ZERO_ADDRESS = ethers.constants.AddressZero

const { to1e18 } = helpers.number
const { mineBlocks, mineBlocksTo } = helpers.time
const { keccak256 } = ethers.utils

const fixture = async () => {
  const contracts = await randomBeaconDeployment()

  await registerOperators(
    contracts.randomBeacon as RandomBeacon,
    (await getUnnamedAccounts()).slice(1, 1 + constants.groupSize)
  )

  const randomBeacon = contracts.randomBeacon as RandomBeaconStub & RandomBeacon
  const testToken = contracts.testToken as TestToken

  return {
    randomBeacon,
    testToken,
  }
}

describe.only("System -- e2e", () => {
  // same as in RandomBeacon constructor
  const relayRequestFee = to1e18(200)
  const relayEntryHardTimeout = 5760
  const relayEntrySubmissionEligibilityDelay = 20
  const callbackGasLimit = 50000
  const groupCreationFrequency = 5
  const groupLifetime = 403200
  const groupPubKeys = [
    blsData.groupPubKey,
    blsData.groupPubKey2,
    blsData.groupPubKey3,
  ]

  let randomBeacon: RandomBeaconStub & RandomBeacon
  let testToken: TestToken
  let requester: SignerWithAddress
  let owner: SignerWithAddress

  before(async () => {
    const contracts = await waffle.loadFixture(fixture)

    owner = await ethers.getSigner((await getNamedAccounts()).deployer)
    requester = await ethers.getSigner((await getUnnamedAccounts())[1])
    randomBeacon = contracts.randomBeacon
    testToken = contracts.testToken

    await randomBeacon
      .connect(owner)
      .updateRelayEntryParameters(
        relayRequestFee,
        relayEntrySubmissionEligibilityDelay,
        relayEntryHardTimeout,
        callbackGasLimit
      )

    await randomBeacon
      .connect(owner)
      .updateGroupCreationParameters(groupCreationFrequency, groupLifetime)
  })

  context("when testing a happy path with 15 relay requests", () => {
    let groupPubKeyCounter = 0

    it("should create 3 new groups", async () => {
      expect(await randomBeacon.getGroupCreationState()).to.be.equal(
        dkgState.IDLE
      )

      const [genesisTx, genesisSeed] = await genesis(randomBeacon)

      expect(await randomBeacon.getGroupCreationState()).to.be.equal(
        dkgState.KEY_GENERATION
      )

      // pass key generation state and transition to awaiting result state
      await mineBlocksTo(genesisTx.blockNumber + constants.offchainDkgTime + 1)

      expect(await randomBeacon.getGroupCreationState()).to.be.equal(
        dkgState.AWAITING_RESULT
      )

      const genesisDkgResult = await signAndSubmitCorrectDkgResult(
        randomBeacon,
        groupPubKeys[groupPubKeyCounter],
        genesisSeed,
        genesisTx.blockNumber,
        noMisbehaved
      )

      await mineBlocks(params.dkgResultChallengePeriodLength)

      expect(await randomBeacon.getGroupCreationState()).to.be.equal(
        dkgState.CHALLENGE
      )

      await randomBeacon
        .connect(genesisDkgResult.submitter)
        .approveDkgResult(genesisDkgResult.dkgResult)

      for (let i = 1; i <= 14; i++) {
        await approveTestToken(requester)
        await randomBeacon.connect(requester).requestRelayEntry(ZERO_ADDRESS)

        const txSubmitRelayEntry = await randomBeacon
          .connect(genesisDkgResult.submitter)
          .submitRelayEntry(blsData.groupSignatures[i - 1])

        // every 5th relay request triggers a new dkg
        if (i % groupCreationFrequency === 0) {
          groupPubKeyCounter += 1
          expect(await randomBeacon.getGroupCreationState()).to.be.equal(
            dkgState.KEY_GENERATION
          )

          await mineBlocksTo(
            txSubmitRelayEntry.blockNumber + constants.offchainDkgTime + 1
          )

          expect(await randomBeacon.getGroupCreationState()).to.be.equal(
            dkgState.AWAITING_RESULT
          )

          const dkgResult = await signAndSubmitCorrectDkgResult(
            randomBeacon,
            groupPubKeys[groupPubKeyCounter],
            ethers.BigNumber.from(
              ethers.utils.keccak256(blsData.groupSignatures[i - 1])
            ),
            txSubmitRelayEntry.blockNumber,
            noMisbehaved
          )

          await mineBlocks(params.dkgResultChallengePeriodLength)

          expect(await randomBeacon.getGroupCreationState()).to.be.equal(
            dkgState.CHALLENGE
          )

          await randomBeacon
            .connect(dkgResult.submitter)
            .approveDkgResult(dkgResult.dkgResult)

          expect(await randomBeacon.getGroupCreationState()).to.be.equal(
            dkgState.IDLE
          )
        }
      }

      const groupsRegistry = await randomBeacon.getGroupsRegistry()
      expect(groupsRegistry).to.be.lengthOf(3)
      expect(groupsRegistry[0]).to.deep.equal(keccak256(groupPubKeys[0]))
      expect(groupsRegistry[1]).to.deep.equal(keccak256(groupPubKeys[1]))
      expect(groupsRegistry[2]).to.deep.equal(keccak256(groupPubKeys[2]))
    })
  })

  async function approveTestToken(_requester) {
    await testToken.mint(_requester.address, relayRequestFee)
    await testToken
      .connect(_requester)
      .approve(randomBeacon.address, relayRequestFee)
  }
})
