const { ethers } = require("hardhat");

async function main() {
  const [owner, buyer, node, beneficiary] = await ethers.getSigners();

  console.log("Deploying DeAPNToken...");
  const Token = await ethers.getContractFactory("DeAPNToken");
  const token = await Token.deploy();
  await token.waitForDeployment();
  console.log("Token deployed to:", await token.getAddress());

  console.log("Deploying DeAPNStaking...");
  const Staking = await ethers.getContractFactory("DeAPNStaking");
  const staking = await Staking.deploy(await token.getAddress());
  await staking.waitForDeployment();
  console.log("Staking deployed to:", await staking.getAddress());

  console.log("Deploying DeAPNSettlement...");
  const Settlement = await ethers.getContractFactory("DeAPNSettlement");
  const settlement = await Settlement.deploy(await token.getAddress());
  await settlement.waitForDeployment();
  console.log("Settlement deployed to:", await settlement.getAddress());

  console.log("Deploying DeAPNArbiter...");
  const Arbiter = await ethers.getContractFactory("DeAPNArbiter");
  const arbiter = await Arbiter.deploy(await staking.getAddress(), await settlement.getAddress(), await token.getAddress());
  await arbiter.waitForDeployment();
  console.log("Arbiter deployed to:", await arbiter.getAddress());

  console.log("\nSetup: Node staking...");
  const stakeAmount = ethers.parseEther("10000"); // 10k PROXY
  await token.transfer(node.address, stakeAmount);
  await token.connect(node).approve(await staking.getAddress(), stakeAmount);
  await staking.connect(node).stake(stakeAmount);
  console.log("Node staked 10k PROXY");

  console.log("\nSetup: Buyer depositing...");
  const depositAmount = ethers.parseEther("100"); // 100 USDC (proxy)
  await token.transfer(buyer.address, depositAmount);
  await token.connect(buyer).approve(await settlement.getAddress(), depositAmount);
  await settlement.connect(buyer).deposit(depositAmount);
  console.log("Buyer deposited 100 tokens");

  console.log("\nScenario: Opening a dispute...");
  // Register arbiter as slasher
  await staking.addSlasher(await arbiter.getAddress());
  
  const requestHash = ethers.id("req_123");
  const billAmount = ethers.parseEther("1");
  await arbiter.connect(buyer).openDispute(node.address, requestHash, billAmount);
  console.log("Dispute opened for req_123");

  console.log("\nScenario: Resolving dispute (Node GUILTY)...");
  const initialNodeStake = (await staking.nodes(node.address)).stakedAmount;
  const initialBuyerBalance = await token.balanceOf(buyer.address);

  // Router detects fraud or timeout, resolves against node
  await arbiter.resolveDispute(1, false);
  console.log("Dispute resolved: Node Guilty!");

  const finalNodeStake = (await staking.nodes(node.address)).stakedAmount;
  const finalBuyerBalance = await token.balanceOf(buyer.address);

  console.log("\nResults:");
  console.log("Node Initial Stake:", ethers.formatEther(initialNodeStake));
  console.log("Node Final Stake:", ethers.formatEther(finalNodeStake));
  console.log("Tokens Slashed:", ethers.formatEther(initialNodeStake - finalNodeStake));
  console.log("Buyer Compensation:", ethers.formatEther(finalBuyerBalance - initialBuyerBalance));

  if (finalNodeStake < initialNodeStake) {
    console.log("\nSUCCESS: Slash executed on-chain!");
  } else {
    console.log("\nFAILURE: Slash failed.");
  }
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
