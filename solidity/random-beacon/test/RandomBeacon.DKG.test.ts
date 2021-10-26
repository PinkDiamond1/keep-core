/* eslint-disable no-await-in-loop */
import { ethers, waffle, helpers, getUnnamedAccounts } from "hardhat"
import { expect } from "chai"
import type { BigNumber, ContractTransaction, Signer } from "ethers"
import type { Address } from "hardhat-deploy/types"
import blsData from "./data/bls"
import { constants, params, testDeployment } from "./fixtures"

import type { RandomBeaconStub } from "../typechain"

const { mineBlocks, mineBlocksTo } = helpers.time

const dkgState = {
  IDLE: 0,
  KEY_GENERATION: 1,
  AWAITING_RESULT: 2,
  CHALLENGE: 3,
}

describe("RandomBeacon", () => {
  const dkgTimeout: number =
    constants.offchainDkgTime +
    constants.groupSize * params.dkgResultSubmissionEligibilityDelay

  const groupPublicKey: string = ethers.utils.hexValue(blsData.groupPubKey)

  let thirdParty: Signer
  let signers: DkgGroupSigners

  let randomBeacon: RandomBeaconStub

  before(async () => {
    thirdParty = await ethers.getSigner((await getUnnamedAccounts())[1])

    // Accounts offset provided to getDkgGroupSigners have to include number of
    // unnamed accounts that were already used.
    signers = await getDkgGroupSigners(constants.groupSize, 1)
  })

  beforeEach("load test fixture", async () => {
    const contracts = await waffle.loadFixture(testDeployment)

    randomBeacon = contracts.randomBeacon as RandomBeaconStub
  })

  describe("genesis", async () => {
    context("when called by a third party", async () => {
      it("should succeed", async () => {
        await randomBeacon.connect(thirdParty).genesis()
      })
    })

    context("with initial contract state", async () => {
      let tx: ContractTransaction
      let expectedSeed: BigNumber

      beforeEach("run genesis", async () => {
        // eslint-disable-next-line @typescript-eslint/no-extra-semi
        ;[tx, expectedSeed] = await genesis()
      })

      it("should emit DkgStarted event", async () => {
        await expect(tx)
          .to.emit(randomBeacon, "DkgStarted")
          .withArgs(expectedSeed)
      })
    })

    context("with genesis in progress", async () => {
      let startBlock: number

      beforeEach("run genesis", async () => {
        const [genesisTx] = await genesis()
        startBlock = genesisTx.blockNumber
      })

      context("with dkg result not submitted", async () => {
        it("should revert with 'current state is not IDLE' error", async () => {
          await expect(randomBeacon.genesis()).to.be.revertedWith(
            "current state is not IDLE"
          )
        })
      })

      context("with dkg result submitted", async () => {
        beforeEach(async () => {
          await mineBlocks(constants.offchainDkgTime)
          await signAndSubmitDkgResult(signers, startBlock)
        })

        // TODO: Add test cases to cover results that are approved, challenged or
        // pending.

        context("with dkg result not approved", async () => {
          it("should revert with 'current state is not IDLE' error", async () => {
            await expect(randomBeacon.genesis()).to.be.revertedWith(
              "current state is not IDLE"
            )
          })
        })
      })
    })
  })

  describe("getGroupCreationState", async () => {
    context("with initial contract state", async () => {
      it("should return IDLE state", async () => {
        expect(await randomBeacon.getGroupCreationState()).to.be.equal(
          dkgState.IDLE
        )
      })
    })

    context("when genesis dkg started", async () => {
      let startBlock: number

      beforeEach("run genesis", async () => {
        const [genesisTx] = await genesis()
        startBlock = genesisTx.blockNumber
      })

      context("at the start of off-chain dkg period", async () => {
        it("should return KEY_GENERATION state", async () => {
          expect(await randomBeacon.getGroupCreationState()).to.be.equal(
            dkgState.KEY_GENERATION
          )
        })
      })

      context("at the end of off-chain dkg period", async () => {
        beforeEach(async () => {
          await mineBlocksTo(startBlock + constants.offchainDkgTime)
        })

        it("should return KEY_GENERATION state", async () => {
          expect(await randomBeacon.getGroupCreationState()).to.be.equal(
            dkgState.KEY_GENERATION
          )
        })
      })

      context("after off-chain dkg period", async () => {
        beforeEach(async () => {
          await mineBlocksTo(startBlock + constants.offchainDkgTime + 1)
        })

        context("when dkg result was not submitted", async () => {
          it("should return AWAITING_RESULT state", async () => {
            expect(await randomBeacon.getGroupCreationState()).to.be.equal(
              dkgState.AWAITING_RESULT
            )
          })

          context("after the dkg timeout period", async () => {
            beforeEach(async () => {
              await mineBlocksTo(startBlock + dkgTimeout + 1)
            })

            it("should return AWAITING_RESULT state", async () => {
              expect(await randomBeacon.getGroupCreationState()).to.be.equal(
                dkgState.AWAITING_RESULT
              )
            })
          })
        })

        context("when dkg result was submitted", async () => {
          beforeEach(async () => {
            await signAndSubmitDkgResult(signers, startBlock)
          })

          context("when dkg result was not approved", async () => {
            it("should return CHALLENGE state", async () => {
              expect(await randomBeacon.getGroupCreationState()).to.be.equal(
                dkgState.CHALLENGE
              )
            })
          })

          context("when dkg result was approved", async () => {
            beforeEach(async () => {
              await mineBlocks(params.dkgResultChallengePeriodLength)
              await randomBeacon.approveDkgResult()
            })

            it("should return IDLE state", async () => {
              expect(await randomBeacon.getGroupCreationState()).to.be.equal(
                dkgState.IDLE
              )
            })
          })

          context("when dkg result was challenged", async () => {
            beforeEach(async () => {
              await randomBeacon.challengeDkgResult()
            })

            it("should return AWAITING_RESULT state", async () => {
              expect(await randomBeacon.getGroupCreationState()).to.be.equal(
                dkgState.AWAITING_RESULT
              )
            })
          })
        })
      })
    })
  })

  describe("hasDkgTimedOut", async () => {
    context("with initial contract state", async () => {
      it("should return false", async () => {
        await expect(await randomBeacon.hasDkgTimedOut()).to.be.false
      })
    })

    context("when genesis dkg started", async () => {
      let startBlock: number

      beforeEach("run genesis", async () => {
        const [genesisTx] = await genesis()
        startBlock = genesisTx.blockNumber
      })

      context("within off-chain dkg period", async () => {
        it("should return false", async () => {
          await expect(await randomBeacon.hasDkgTimedOut()).to.be.false
        })
      })

      context("after off-chain dkg period", async () => {
        beforeEach(async () => {
          await mineBlocksTo(startBlock + constants.offchainDkgTime + 1)
        })

        context("when dkg result was not submitted", async () => {
          it("should return false", async () => {
            await expect(await randomBeacon.hasDkgTimedOut()).to.be.false
          })

          context("at the end of the dkg timeout period", async () => {
            beforeEach(async () => {
              await mineBlocksTo(startBlock + dkgTimeout)
            })

            it("should return false", async () => {
              await expect(await randomBeacon.hasDkgTimedOut()).to.be.false
            })
          })

          context("after the dkg timeout period", async () => {
            beforeEach(async () => {
              await mineBlocksTo(startBlock + dkgTimeout + 1)
            })

            it("should return true", async () => {
              await expect(await randomBeacon.hasDkgTimedOut()).to.be.true
            })
          })
        })

        context("when dkg result was submitted", async () => {
          let resultSubmissionBlock: number

          beforeEach(async () => {
            const { transaction } = await signAndSubmitDkgResult(
              signers,
              startBlock
            )

            resultSubmissionBlock = transaction.blockNumber
          })

          context("when dkg result was not approved", async () => {
            context("at the end of the dkg timeout period", async () => {
              beforeEach(async () => {
                await mineBlocksTo(startBlock + dkgTimeout)
              })

              it("should return false", async () => {
                await expect(await randomBeacon.hasDkgTimedOut()).to.be.false
              })
            })

            context("after the dkg timeout period", async () => {
              beforeEach(async () => {
                await mineBlocksTo(startBlock + dkgTimeout + 1)
              })

              it("should return false", async () => {
                await expect(await randomBeacon.hasDkgTimedOut()).to.be.false
              })
            })

            context("at the end of the challenge period", async () => {
              beforeEach(async () => {
                await mineBlocksTo(
                  resultSubmissionBlock + params.dkgResultChallengePeriodLength
                )
              })

              it("should return false", async () => {
                await expect(await randomBeacon.hasDkgTimedOut()).to.be.false
              })
            })

            context("after the challenge period", async () => {
              beforeEach(async () => {
                await mineBlocksTo(
                  resultSubmissionBlock +
                    params.dkgResultChallengePeriodLength +
                    1
                )
              })

              it("should return false", async () => {
                await expect(await randomBeacon.hasDkgTimedOut()).to.be.false
              })
            })
          })

          context("when dkg result was approved", async () => {
            beforeEach(async () => {
              await mineBlocksTo(
                resultSubmissionBlock + params.dkgResultChallengePeriodLength
              )

              await randomBeacon.approveDkgResult()
            })

            it("should return false", async () => {
              await expect(await randomBeacon.hasDkgTimedOut()).to.be.false
            })
          })

          context("when dkg result was challenged", async () => {
            let challengeBlockNumber: number

            beforeEach(async () => {
              const tx = await randomBeacon.challengeDkgResult()
              challengeBlockNumber = tx.blockNumber
            })

            context("at the end of dkg result submission period", async () => {
              beforeEach(async () => {
                await mineBlocksTo(
                  challengeBlockNumber +
                    constants.groupSize *
                      params.dkgResultSubmissionEligibilityDelay
                )
              })

              it("should return false", async () => {
                await expect(await randomBeacon.hasDkgTimedOut()).to.be.false
              })
            })

            context("after dkg result submission period", async () => {
              beforeEach(async () => {
                await mineBlocksTo(
                  challengeBlockNumber +
                    constants.groupSize *
                      params.dkgResultSubmissionEligibilityDelay +
                    1
                )
              })

              it("should return true", async () => {
                await expect(await randomBeacon.hasDkgTimedOut()).to.be.true
              })
            })
          })
        })
      })
    })
  })

  describe("submitDkgResult", async () => {
    // TODO: Add more tests to cover the DKG result verification function thoroughly.

    context("with initial contract state", async () => {
      it("should revert with 'current state is not AWAITING_RESULT' error", async () => {
        await expect(signAndSubmitDkgResult(signers, 1)).to.be.revertedWith(
          "current state is not AWAITING_RESULT"
        )
      })
    })

    context("with group creation in progress", async () => {
      let startBlock: number

      beforeEach("run genesis", async () => {
        const [genesisTx] = await genesis()

        startBlock = genesisTx.blockNumber
      })

      context("with group creation not timed out", async () => {
        context("with off-chain dkg time not passed", async () => {
          beforeEach(async () => {
            await mineBlocksTo(startBlock + constants.offchainDkgTime - 1)
          })

          it("should revert with 'current state is not AWAITING_RESULT' error", async () => {
            await expect(
              signAndSubmitDkgResult(signers, startBlock)
            ).to.be.revertedWith("current state is not AWAITING_RESULT")
          })
        })

        context("with off-chain dkg time passed", async () => {
          beforeEach(async () => {
            await mineBlocksTo(startBlock + constants.offchainDkgTime)
          })

          it("should revert with less than threshold signers", async () => {
            const filteredSigners = new Map(
              Array.from(signers).filter(
                ([index]) => index < constants.signatureThreshold
              )
            )

            await expect(
              signAndSubmitDkgResult(filteredSigners, startBlock)
            ).to.be.revertedWith("Too few signatures")
          })

          it("should succeed with threshold signers", async () => {
            const filteredSigners = new Map(
              Array.from(signers).filter(
                ([index]) => index <= constants.signatureThreshold
              )
            )

            const {
              transaction: tx,
              dkgResult,
              dkgResultHash,
            } = await signAndSubmitDkgResult(filteredSigners, startBlock)

            await expect(tx)
              .to.emit(randomBeacon, "DkgResultSubmitted")
              .withArgs(dkgResultHash, dkgResult.groupPubKey, signers.get(1))
          })

          it("should succeed for the first submitter", async () => {
            const {
              transaction: tx,
              dkgResult,
              dkgResultHash,
            } = await signAndSubmitDkgResult(signers, startBlock, 1)
            await expect(tx)
              .to.emit(randomBeacon, "DkgResultSubmitted")
              .withArgs(dkgResultHash, dkgResult.groupPubKey, signers.get(1))
          })

          it("should revert for the second submitter", async () => {
            await expect(
              signAndSubmitDkgResult(signers, startBlock, 2)
            ).to.be.revertedWith("Submitter not eligible")
          })

          context(
            "with first submitter eligibility delay period almost ended",
            async () => {
              beforeEach(async () => {
                await mineBlocksTo(
                  startBlock +
                    constants.offchainDkgTime +
                    params.dkgResultSubmissionEligibilityDelay -
                    2
                )
              })

              it("should succeed for the first submitter", async () => {
                const {
                  transaction: tx,
                  dkgResult,
                  dkgResultHash,
                } = await signAndSubmitDkgResult(signers, startBlock, 1)

                await expect(tx)
                  .to.emit(randomBeacon, "DkgResultSubmitted")
                  .withArgs(
                    dkgResultHash,
                    dkgResult.groupPubKey,
                    signers.get(1)
                  )
              })

              it("should revert for the second submitter", async () => {
                await expect(
                  signAndSubmitDkgResult(signers, startBlock, 2)
                ).to.be.revertedWith("Submitter not eligible")
              })
            }
          )

          context(
            "with first submitter eligibility delay period ended",
            async () => {
              beforeEach(async () => {
                await mineBlocksTo(
                  startBlock +
                    constants.offchainDkgTime +
                    params.dkgResultSubmissionEligibilityDelay -
                    1
                )
              })

              it("should succeed for the first submitter", async () => {
                const {
                  transaction: tx,
                  dkgResult,
                  dkgResultHash,
                } = await signAndSubmitDkgResult(signers, startBlock, 1)

                await expect(tx)
                  .to.emit(randomBeacon, "DkgResultSubmitted")
                  .withArgs(
                    dkgResultHash,
                    dkgResult.groupPubKey,
                    signers.get(1)
                  )
              })

              it("should succeed for the second submitter", async () => {
                const {
                  transaction: tx,
                  dkgResult,
                  dkgResultHash,
                } = await signAndSubmitDkgResult(signers, startBlock, 2)

                await expect(tx)
                  .to.emit(randomBeacon, "DkgResultSubmitted")
                  .withArgs(
                    dkgResultHash,
                    dkgResult.groupPubKey,
                    signers.get(2)
                  )
              })

              it("should revert for the third submitter", async () => {
                await expect(
                  signAndSubmitDkgResult(signers, startBlock, 3)
                ).to.be.revertedWith("Submitter not eligible")
              })
            }
          )

          context(
            "with the last submitter eligibility delay period almost ended",
            async () => {
              beforeEach(async () => {
                await mineBlocksTo(startBlock + dkgTimeout - 1)
              })

              it("should succeed for the first submitter", async () => {
                const {
                  transaction: tx,
                  dkgResult,
                  dkgResultHash,
                } = await signAndSubmitDkgResult(signers, startBlock, 1)

                await expect(tx)
                  .to.emit(randomBeacon, "DkgResultSubmitted")
                  .withArgs(
                    dkgResultHash,
                    dkgResult.groupPubKey,
                    signers.get(1)
                  )
              })

              it("should succeed for the last submitter", async () => {
                const {
                  transaction: tx,
                  dkgResult,
                  dkgResultHash,
                } = await signAndSubmitDkgResult(
                  signers,
                  startBlock,
                  constants.groupSize
                )

                await expect(tx)
                  .to.emit(randomBeacon, "DkgResultSubmitted")
                  .withArgs(
                    dkgResultHash,
                    dkgResult.groupPubKey,
                    signers.get(constants.groupSize)
                  )
              })
            }
          )

          context("with dkg result approved", async () => {
            beforeEach(async () => {
              await mineBlocksTo(startBlock + constants.offchainDkgTime)

              await signAndSubmitDkgResult(signers, startBlock)
            })

            it("should revert 'current state is not AWAITING_RESULT' error", async () => {
              await expect(
                signAndSubmitDkgResult(signers, startBlock)
              ).to.be.revertedWith("current state is not AWAITING_RESULT")
            })
          })

          context("with dkg result challenged", async () => {
            beforeEach(async () => {
              await mineBlocksTo(startBlock + constants.offchainDkgTime)

              await signAndSubmitDkgResult(signers, startBlock)

              await randomBeacon.challengeDkgResult()
            })

            it("should allow first member to submit", async () => {
              const {
                transaction: tx,
                dkgResult,
                dkgResultHash,
              } = await signAndSubmitDkgResult(signers, startBlock)

              await expect(tx)
                .to.emit(randomBeacon, "DkgResultSubmitted")
                .withArgs(dkgResultHash, dkgResult.groupPubKey, signers.get(1))
            })
          })
        })
      })

      // TODO: Check challenge adjust start block calculation for eligibility
      // TODO: Check that challenges add up the delay

      context("with group creation timed out", async () => {
        beforeEach("increase time", async () => {
          await mineBlocksTo(startBlock + dkgTimeout)
        })

        context("with timeout not notified", async () => {
          it("should revert with dkg timeout already passed error", async () => {
            await expect(
              signAndSubmitDkgResult(signers, startBlock)
            ).to.be.revertedWith("dkg timeout already passed")
          })
        })
      })
    })
  })

  describe("approveDkgResult", async () => {
    context("with initial contract state", async () => {
      it("should revert with 'current state is not CHALLENGE' error", async () => {
        await expect(randomBeacon.approveDkgResult()).to.be.revertedWith(
          "current state is not CHALLENGE"
        )
      })
    })

    context("with group creation in progress", async () => {
      let startBlock: number

      beforeEach("run genesis", async () => {
        const [genesisTx] = await genesis()

        startBlock = genesisTx.blockNumber
      })

      it("should revert with 'current state is not CHALLENGE' error", async () => {
        await expect(randomBeacon.approveDkgResult()).to.be.revertedWith(
          "current state is not CHALLENGE"
        )
      })

      context("with off-chain dkg time passed", async () => {
        beforeEach(async () => {
          await mineBlocksTo(startBlock + constants.offchainDkgTime)
        })

        context("with dkg result not submitted", async () => {
          it("should revert with 'current state is not CHALLENGE' error", async () => {
            await expect(randomBeacon.approveDkgResult()).to.be.revertedWith(
              "current state is not CHALLENGE"
            )
          })
        })

        context("with dkg result submitted", async () => {
          let resultSubmissionBlock: number
          let dkgResultHash: string

          beforeEach(async () => {
            let tx: ContractTransaction
              // eslint-disable-next-line @typescript-eslint/no-extra-semi
            ;({ transaction: tx, dkgResultHash } = await signAndSubmitDkgResult(
              signers,
              startBlock
            ))

            resultSubmissionBlock = tx.blockNumber
          })

          context("with challenge period not passed", async () => {
            beforeEach(async () => {
              await mineBlocksTo(
                resultSubmissionBlock +
                  params.dkgResultChallengePeriodLength -
                  1
              )
            })

            it("should revert with 'challenge period has not passed yet' error", async () => {
              await expect(randomBeacon.approveDkgResult()).to.be.revertedWith(
                "challenge period has not passed yet"
              )
            })
          })

          context("with challenge period passed", async () => {
            beforeEach(async () => {
              await mineBlocksTo(
                resultSubmissionBlock + params.dkgResultChallengePeriodLength
              )
            })

            context("called by a third party", async () => {
              let tx: ContractTransaction

              beforeEach(async () => {
                tx = await randomBeacon.connect(thirdParty).approveDkgResult()
              })

              it("should emit DkgResultApproved event", async () => {
                await expect(tx)
                  .to.emit(randomBeacon, "DkgResultApproved")
                  .withArgs(dkgResultHash, await thirdParty.getAddress())
              })

              it("should clean dkg data", async () => {
                await assertDkgResultCleanData(randomBeacon)
              })
            })
          })

          context("when there was a challenged result before", async () => {
            beforeEach(async () => {
              await randomBeacon.challengeDkgResult()

              // Submit a second result by another submitter
              const submitterIndex = 5

              await mineBlocks(
                params.dkgResultSubmissionEligibilityDelay * submitterIndex
              )

              let tx: ContractTransaction
              ;({ transaction: tx, dkgResultHash } =
                await signAndSubmitDkgResult(
                  signers,
                  startBlock,
                  submitterIndex
                ))

              resultSubmissionBlock = tx.blockNumber
            })

            context("with challenge period not passed", async () => {
              beforeEach(async () => {
                await mineBlocksTo(
                  resultSubmissionBlock +
                    params.dkgResultChallengePeriodLength -
                    1
                )
              })

              it("should revert with 'challenge period has not passed yet' error", async () => {
                await expect(
                  randomBeacon.approveDkgResult()
                ).to.be.revertedWith("challenge period has not passed yet")
              })
            })

            context("with challenge period passed", async () => {
              beforeEach(async () => {
                await mineBlocksTo(
                  resultSubmissionBlock + params.dkgResultChallengePeriodLength
                )
              })

              it("should succeed", async () => {
                const tx = await randomBeacon
                  .connect(thirdParty)
                  .approveDkgResult()

                await expect(tx)
                  .to.emit(randomBeacon, "DkgResultApproved")
                  .withArgs(dkgResultHash, await thirdParty.getAddress())
              })
            })
          })
        })
      })

      context("with max periods duration", async () => {
        it("should succeed", async () => {
          await mineBlocksTo(startBlock + dkgTimeout - 1)

          await signAndSubmitDkgResult(signers, startBlock)

          await mineBlocks(params.dkgResultChallengePeriodLength)

          await randomBeacon.approveDkgResult()
        })
      })
    })
  })

  describe("notifyDkgTimeout", async () => {
    context("with initial contract state", async () => {
      it("should revert with 'dkg has not timed out' error", async () => {
        await expect(randomBeacon.notifyDkgTimeout()).to.be.revertedWith(
          "dkg has not timed out"
        )
      })
    })

    context("with group creation in progress", async () => {
      let startBlock: number

      beforeEach("run genesis", async () => {
        const [genesisTx] = await genesis()

        startBlock = genesisTx.blockNumber
      })

      context("with dkg not timed out", async () => {
        context("with off-chain dkg time not passed", async () => {
          beforeEach(async () => {
            await mineBlocksTo(startBlock + constants.offchainDkgTime - 1)
          })

          it("should revert with 'dkg has not timed out' error", async () => {
            await expect(randomBeacon.notifyDkgTimeout()).to.be.revertedWith(
              "dkg has not timed out"
            )
          })
        })

        context("with off-chain dkg time passed", async () => {
          beforeEach(async () => {
            await mineBlocksTo(startBlock + constants.offchainDkgTime)
          })

          it("should revert with 'dkg has not timed out' error", async () => {
            await expect(randomBeacon.notifyDkgTimeout()).to.be.revertedWith(
              "dkg has not timed out"
            )
          })
        })

        context("with result submission period almost ended", async () => {
          beforeEach(async () => {
            await mineBlocksTo(startBlock + dkgTimeout - 1)
          })

          it("should revert with 'dkg has not timed out' error", async () => {
            await expect(randomBeacon.notifyDkgTimeout()).to.be.revertedWith(
              "dkg has not timed out"
            )
          })
        })
      })

      context("with dkg timed out", async () => {
        beforeEach(async () => {
          await mineBlocksTo(startBlock + dkgTimeout)
        })

        context("called by a third party", async () => {
          let tx: ContractTransaction

          beforeEach(async () => {
            tx = await randomBeacon.connect(thirdParty).notifyDkgTimeout()
          })

          it("should emit DkgTimedOut event", async () => {
            await expect(tx).to.emit(randomBeacon, "DkgTimedOut")
          })

          it("should clean dkg data", async () => {
            await assertDkgResultCleanData(randomBeacon)
          })
        })
      })
    })
  })

  describe("challengeDkgResult", async () => {
    context("with initial contract state", async () => {
      it("should revert with 'current state is not CHALLENGE' error", async () => {
        await expect(randomBeacon.challengeDkgResult()).to.be.revertedWith(
          "current state is not CHALLENGE"
        )
      })
    })

    context("with group creation in progress", async () => {
      let startBlock: number

      beforeEach("run genesis", async () => {
        const [genesisTx] = await genesis()

        startBlock = genesisTx.blockNumber
      })

      it("should revert with 'current state is not CHALLENGE' error", async () => {
        await expect(randomBeacon.challengeDkgResult()).to.be.revertedWith(
          "current state is not CHALLENGE"
        )
      })

      context("with off-chain dkg time passed", async () => {
        beforeEach(async () => {
          await mineBlocksTo(startBlock + constants.offchainDkgTime)
        })

        context("with dkg result not submitted", async () => {
          it("should revert with 'current state is not CHALLENGE' error", async () => {
            await expect(randomBeacon.challengeDkgResult()).to.be.revertedWith(
              "current state is not CHALLENGE"
            )
          })
        })

        context("with dkg result submitted", async () => {
          let resultSubmissionBlock: number
          let dkgResultHash: string

          beforeEach(async () => {
            let tx: ContractTransaction
              // eslint-disable-next-line @typescript-eslint/no-extra-semi
            ;({ transaction: tx, dkgResultHash } = await signAndSubmitDkgResult(
              signers,
              startBlock
            ))

            resultSubmissionBlock = tx.blockNumber
          })

          context("at the beginning of challenge period", async () => {
            it("should be callable by a third party", async () => {
              await randomBeacon.connect(thirdParty).challengeDkgResult()
            })

            it("should emit DkgResultChallenged event", async () => {
              const tx = await randomBeacon
                .connect(thirdParty)
                .challengeDkgResult()

              await expect(tx)
                .to.emit(randomBeacon, "DkgResultChallenged")
                .withArgs(dkgResultHash, await thirdParty.getAddress())
            })
          })

          context("at the end of challenge period", async () => {
            beforeEach(async () => {
              await mineBlocksTo(
                resultSubmissionBlock +
                  params.dkgResultChallengePeriodLength -
                  1
              )
            })

            it("should be callable by a third party", async () => {
              await randomBeacon.connect(thirdParty).challengeDkgResult()
            })

            it("should emit DkgResultChallenged event", async () => {
              const tx = await randomBeacon
                .connect(thirdParty)
                .challengeDkgResult()

              await expect(tx)
                .to.emit(randomBeacon, "DkgResultChallenged")
                .withArgs(dkgResultHash, await thirdParty.getAddress())
            })
          })

          context("with challenge period passed", async () => {
            beforeEach(async () => {
              await mineBlocksTo(
                resultSubmissionBlock + params.dkgResultChallengePeriodLength
              )
            })

            it("should revert with 'challenge period has already passed' error", async () => {
              await expect(
                randomBeacon.challengeDkgResult()
              ).to.be.revertedWith("challenge period has already passed")
            })
          })
        })
      })
    })

    // This test checks that dkg timeout is adjusted in case of result challenges
    // to include the offset blocks that were mined until the invalid result
    // was challenged.
    it("should enforce submission start offset", async () => {
      const [genesisTx] = await genesis()
      const startBlock = genesisTx.blockNumber

      await mineBlocks(constants.offchainDkgTime)

      // Submit result 1 at the beginning of the submission period
      await signAndSubmitDkgResult(signers, startBlock)

      await expect(
        (
          await randomBeacon.getDkgData()
        ).resultSubmissionStartBlockOffset,
        "invalid resultSubmissionStartBlockOffset for result 1 after submission"
      ).to.equal(0)

      // Challenge result 1 at the beginning of the challenge period
      await randomBeacon.challengeDkgResult()
      let expectedSubmissionOffset = 2 // 1 block for dkg result submission tx + 1 block for challenge tx

      await expect(
        (
          await randomBeacon.getDkgData()
        ).resultSubmissionStartBlockOffset,
        "invalid resultSubmissionStartBlockOffset for result 1 after challenge"
      ).to.equal(expectedSubmissionOffset)

      // Submit result 2 in the middle of the submission period
      let blocksToMine =
        (constants.groupSize * params.dkgResultSubmissionEligibilityDelay) / 2
      await mineBlocks(blocksToMine)
      await signAndSubmitDkgResult(signers, startBlock, constants.groupSize / 2)

      await expect(
        (
          await randomBeacon.getDkgData()
        ).resultSubmissionStartBlockOffset,
        "invalid resultSubmissionStartBlockOffset for result 2 after submission"
      ).to.equal(expectedSubmissionOffset) // same as before

      expectedSubmissionOffset += blocksToMine

      // Challenge result 2 in the middle of the challenge period
      await mineBlocks(params.dkgResultChallengePeriodLength / 2)
      expectedSubmissionOffset += params.dkgResultChallengePeriodLength / 2
      await randomBeacon.challengeDkgResult()
      expectedSubmissionOffset += 2 // 1 block for dkg result submission tx + 1 block for challenge tx

      await expect(
        (
          await randomBeacon.getDkgData()
        ).resultSubmissionStartBlockOffset,
        "invalid resultSubmissionStartBlockOffset for result 2 after challenge"
      ).to.equal(expectedSubmissionOffset)

      // Submit result 3 at the end of the submission period
      blocksToMine =
        constants.groupSize * params.dkgResultSubmissionEligibilityDelay - 1
      await mineBlocks(blocksToMine)
      await signAndSubmitDkgResult(signers, startBlock, constants.groupSize)

      await expect(
        (
          await randomBeacon.getDkgData()
        ).resultSubmissionStartBlockOffset,
        "invalid resultSubmissionStartBlockOffset for result 3 after submission"
      ).to.equal(expectedSubmissionOffset) // same as before

      expectedSubmissionOffset += blocksToMine

      // Challenge result 3 at the end of the challenge period
      blocksToMine = params.dkgResultChallengePeriodLength - 1
      await mineBlocks(blocksToMine)
      expectedSubmissionOffset += blocksToMine

      await expect(
        randomBeacon.callStatic.notifyDkgTimeout()
      ).to.be.revertedWith("dkg has not timed out")

      await randomBeacon.challengeDkgResult()
      expectedSubmissionOffset += 2 // 1 block for dkg result submission tx + 1 block for challenge tx

      await expect(
        (
          await randomBeacon.getDkgData()
        ).resultSubmissionStartBlockOffset,
        "invalid resultSubmissionStartBlockOffset for result 3 after challenge"
      ).to.equal(expectedSubmissionOffset)

      // Submit result 4 after the submission period
      blocksToMine =
        constants.groupSize * params.dkgResultSubmissionEligibilityDelay
      await mineBlocks(blocksToMine)
      await expect(
        signAndSubmitDkgResult(signers, startBlock, constants.groupSize)
      ).to.be.revertedWith("dkg timeout already passed")

      await randomBeacon.notifyDkgTimeout()
    })
  })

  async function genesis(): Promise<[ContractTransaction, BigNumber]> {
    const tx = await randomBeacon.genesis()

    const expectedSeed = ethers.BigNumber.from(
      ethers.utils.keccak256(
        ethers.utils.solidityPack(
          ["uint256", "uint256"],
          [await randomBeacon.genesisSeed(), tx.blockNumber]
        )
      )
    )

    return [tx, expectedSeed]
  }

  interface DkgResult {
    submitterMemberIndex: number
    groupPubKey: string
    misbehaved: string
    signatures: string
    signingMemberIndices: number[]
    members: string[]
  }

  async function signAndSubmitDkgResult(
    // eslint-disable-next-line @typescript-eslint/no-shadow
    signers: DkgGroupSigners,
    startBlock: number,
    submitterIndex = 1
  ): Promise<{
    transaction: ContractTransaction
    dkgResult: DkgResult
    dkgResultHash: string
  }> {
    const noMisbehaved = "0x"

    const { members, signingMemberIndices, signaturesBytes } =
      await signDkgResult(signers, groupPublicKey, noMisbehaved, startBlock)

    const dkgResult = {
      submitterMemberIndex: submitterIndex,
      groupPubKey: blsData.groupPubKey,
      misbehaved: noMisbehaved,
      signatures: signaturesBytes,
      signingMemberIndices,
      members,
    }

    const dkgResultHash = ethers.utils.keccak256(
      ethers.utils.defaultAbiCoder.encode(
        [
          "(uint256 submitterMemberIndex, bytes groupPubKey, bytes misbehaved, bytes signatures, uint256[] signingMemberIndices, address[] members)",
        ],
        [dkgResult]
      )
    )

    const transaction = await randomBeacon
      .connect(await ethers.getSigner(signers.get(submitterIndex)))
      .submitDkgResult(dkgResult)

    return { transaction, dkgResult, dkgResultHash }
  }
})

type DkgGroupSigners = Map<number, Address>

async function getDkgGroupSigners(
  groupSize: number,
  startAccountsOffset: number
): Promise<DkgGroupSigners> {
  const signers = new Map<number, Address>()

  for (let i = 1; i <= groupSize; i++) {
    const signer = (await getUnnamedAccounts())[startAccountsOffset + i]

    await expect(
      signer,
      `signer [${i}] is not defined; check hardhat network configuration`
    ).is.not.empty

    signers.set(i, signer)
  }

  return signers
}

async function signDkgResult(
  signers: DkgGroupSigners,
  groupPublicKey: string,
  misbehaved: string,
  startBlock: number
) {
  const resultHash = ethers.utils.solidityKeccak256(
    ["bytes", "bytes", "uint256"],
    [groupPublicKey, misbehaved, startBlock]
  )

  const members: string[] = []
  const signingMemberIndices: number[] = []
  const signatures: string[] = []

  // eslint-disable-next-line no-restricted-syntax
  for (const [memberIndex, signer] of signers) {
    members.push(signer)

    signingMemberIndices.push(memberIndex)

    const ethersSigner = await ethers.getSigner(signer)

    const signature = await ethersSigner.signMessage(
      ethers.utils.arrayify(resultHash)
    )

    signatures.push(signature)
  }

  const signaturesBytes: string = ethers.utils.hexConcat(signatures)

  return { members, signingMemberIndices, signaturesBytes }
}

async function assertDkgResultCleanData(randomBeacon: RandomBeaconStub) {
  const dkgData = await randomBeacon.getDkgData()

  expect(
    dkgData.parameters.resultChallengePeriodLength,
    "unexpected resultChallengePeriodLength"
  ).to.eq(params.dkgResultChallengePeriodLength)

  expect(
    dkgData.parameters.resultSubmissionEligibilityDelay,
    "unexpected resultSubmissionEligibilityDelay"
  ).to.eq(params.dkgResultSubmissionEligibilityDelay)

  expect(dkgData.startBlock, "unexpected startBlock").to.eq(0)

  expect(
    dkgData.resultSubmissionStartBlockOffset,
    "unexpected resultSubmissionStartBlockOffset"
  ).to.eq(0)

  expect(dkgData.submittedResultHash, "unexpected submittedResultHash").to.eq(
    ethers.constants.HashZero
  )

  expect(dkgData.submittedResultBlock, "unexpected submittedResultBlock").to.eq(
    0
  )
}