import { duration, increaseTimeTo } from './helpers/increaseTime';
import latestTime from './helpers/latestTime';
import expectThrow from './helpers/expectThrow';
const KeepToken = artifacts.require('./KeepToken.sol');
const TokenStaking = artifacts.require('./TokenStaking.sol');

contract('TestTokenStake', function(accounts) {

  let token, stakingContract,
    account_one = accounts[0],
    account_one_operator = accounts[1],
    account_one_magpie = accounts[2],
    account_two = accounts[3];

  before(async () => {
    token = await KeepToken.new();
    stakingContract = await TokenStaking.new(token.address, duration.days(30));
  });

  it("should send tokens correctly", async function() {
    let amount = web3.utils.toBN(1000000000);

    // Starting balances
    let account_one_starting_balance = await token.balanceOf.call(account_one);
    let account_two_starting_balance = await token.balanceOf.call(account_two);

    // Send tokens
    await token.transfer(account_two, amount, {from: account_one});

    // Ending balances
    let account_one_ending_balance = await token.balanceOf.call(account_one);
    let account_two_ending_balance = await token.balanceOf.call(account_two);

    assert.equal(account_one_ending_balance.eq(account_one_starting_balance.sub(amount)), true, "Amount wasn't correctly taken from the sender");
    assert.equal(account_two_ending_balance.eq(account_two_starting_balance.add(amount)), true, "Amount wasn't correctly sent to the receiver");

  });

  it("should stake and unstake tokens correctly", async function() {

    let stakingAmount = web3.utils.toBN(10000000);

    // Starting balances
    let account_one_starting_balance = await token.balanceOf.call(account_one);

    let signature = Buffer.from((await web3.eth.sign(web3.utils.soliditySha3(account_one), account_one_operator)).substr(2), 'hex');
    let data = Buffer.concat([Buffer.from(account_one_magpie.substr(2), 'hex'), signature]);

    // Stake tokens using approveAndCall pattern
    await token.approveAndCall(stakingContract.address, stakingAmount, '0x' + data.toString('hex'), {from: account_one});

    // Ending balances
    let account_one_ending_balance = await token.balanceOf.call(account_one);
    let account_one_operator_stake_balance = await stakingContract.balanceOf.call(account_one_operator);

    assert.equal(account_one_ending_balance.eq(account_one_starting_balance.sub(stakingAmount)), true, "Staking amount should be transfered from sender balance");
    assert.equal(account_one_operator_stake_balance.eq(stakingAmount), true, "Staking amount should be added to the sender staking balance");

    // Initiate unstake tokens as token owner
    await stakingContract.initiateUnstake(stakingAmount/2, account_one_operator, {from: account_one});

    // Initiate unstake tokens as operator
    await stakingContract.initiateUnstake(stakingAmount/2, account_one_operator, {from: account_one_operator});

    // should not be able to finish unstake
    await expectThrow(stakingContract.finishUnstake(account_one_operator));

    // jump in time, full withdrawal delay
    await increaseTimeTo(await latestTime()+duration.days(30));

    // should be able to finish unstake
    await stakingContract.finishUnstake(account_one_operator);

    // should fail cause there is no stake to unstake
    await expectThrow(stakingContract.finishUnstake(account_one_operator));

    // check balances
    account_one_ending_balance = await token.balanceOf.call(account_one);
    account_one_operator_stake_balance = await stakingContract.balanceOf.call(account_one_operator);

    assert.equal(account_one_ending_balance.eq(account_one_starting_balance), true, "Staking amount should be transfered to sender balance");
    assert.equal(account_one_operator_stake_balance.isZero(), true, "Staking amount should be removed from sender staking balance");

    // Starting balances
    account_one_starting_balance = await token.balanceOf.call(account_one);

    signature = Buffer.from((await web3.eth.sign(web3.utils.soliditySha3(account_one), account_one_operator)).substr(2), 'hex');
    data = Buffer.concat([Buffer.from(account_one_magpie.substr(2), 'hex'), signature]);

    // Stake tokens using approveAndCall pattern
    await token.approveAndCall(stakingContract.address, stakingAmount, '0x' + data.toString('hex'), {from: account_one});

    // Ending balances
    account_one_ending_balance = await token.balanceOf.call(account_one);
    account_one_operator_stake_balance = await stakingContract.balanceOf.call(account_one_operator);

    assert.equal(account_one_ending_balance.eq(account_one_starting_balance.sub(stakingAmount)), true, "Staking amount should be transfered from sender balance for the second time");
    assert.equal(account_one_operator_stake_balance.eq(stakingAmount), true, "Staking amount should be added to the sender staking balance for the second time");
  });

});