import { expect } from "chai";
import pkg from "hardhat";
const { ethers } = pkg;

describe("DeAPN Phase 2 Integration", function () {
  it("Should execute the full dispute and slash loop correctly", async function () {
    const [owner, buyer, node] = await ethers.getSigners();

    // 1. Deploy
    const Token = await ethers.getContractFactory("DeAPNToken");
    const token = await Token.deploy();
    await token.waitForDeployment();

    const Staking = await ethers.getContractFactory("DeAPNStaking");
    const staking = await Staking.deploy(await token.getAddress());
    await staking.waitForDeployment();

    const Settlement = await ethers.getContractFactory("DeAPNSettlement");
    const settlement = await Settlement.deploy(await token.getAddress());
    await settlement.waitForDeployment();

    const Arbiter = await ethers.getContractFactory("DeAPNArbiter");
    const arbiter = await Arbiter.deploy(await staking.getAddress(), await settlement.getAddress(), await token.getAddress());
    await arbiter.waitForDeployment();

    // 2. Setup Stake
    const stakeAmount = ethers.parseUnits("10000", 18);
    await token.transfer(node.address, stakeAmount);
    await token.connect(node).approve(await staking.getAddress(), stakeAmount);
    await staking.connect(node).stake(stakeAmount);

    // 3. Setup Deposit
    const depositAmount = ethers.parseUnits("100", 18);
    await token.transfer(buyer.address, depositAmount);
    await token.connect(buyer).approve(await settlement.getAddress(), depositAmount);
    await settlement.connect(buyer).deposit(depositAmount);

    // 4. Scenario
    await staking.addSlasher(await arbiter.getAddress());
    const requestHash = ethers.id("req_123");
    const billAmount = ethers.parseUnits("1", 18);
    
    await arbiter.connect(buyer).openDispute(node.address, requestHash, billAmount);
    
    const initialNodeStake = (await staking.nodes(node.address)).stakedAmount;
    const initialBuyerBalance = await token.balanceOf(buyer.address);

    // Resolve Guilty
    await arbiter.resolveDispute(1, false);

    const finalNodeStake = (await staking.nodes(node.address)).stakedAmount;
    const finalBuyerBalance = await token.balanceOf(buyer.address);

    // 5. Assertions
    expect(finalNodeStake).to.be.lessThan(initialNodeStake);
    expect(finalBuyerBalance).to.be.greaterThan(initialBuyerBalance);
    
    const slashAmount = initialNodeStake - finalNodeStake;
    expect(slashAmount).to.equal(billAmount * 10n); // 10x multiplier
    
    console.log("Success: Slashed amount is", ethers.formatEther(slashAmount));
  });
});
