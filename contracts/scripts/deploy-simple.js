const hre = require("hardhat");

async function main() {
  console.log("Deploying simplified contracts...");
  
  const [deployer] = await hre.ethers.getSigners();
  console.log("Deploying with account:", deployer.address);
  
  // Deploy token
  const Token = await hre.ethers.getContractFactory("SimpleQUIVToken");
  const token = await Token.deploy();
  await token.waitForDeployment();
  console.log("Token deployed to:", await token.getAddress());
  
  // Deploy settlement
  const Settlement = await hre.ethers.getContractFactory("SimpleSettlement");
  const settlement = await Settlement.deploy(await token.getAddress());
  await settlement.waitForDeployment();
  console.log("Settlement deployed to:", await settlement.getAddress());
  
  // Transfer some tokens to settlement contract for rewards
  const transferAmount = hre.ethers.parseEther("1000000"); // 1M tokens
  await token.transfer(await settlement.getAddress(), transferAmount);
  console.log("Transferred 1M tokens to settlement contract");
  
  console.log("\nDeployment complete!");
  console.log("Token:", await token.getAddress());
  console.log("Settlement:", await settlement.getAddress());
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });