const hre = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  console.log("Starting deployment...");
  
  const [deployer] = await hre.ethers.getSigners();
  console.log("Deploying contracts with account:", deployer.address);
  
  const balance = await hre.ethers.provider.getBalance(deployer.address);
  console.log("Account balance:", hre.ethers.formatEther(balance));

  // Deploy QUIVToken
  console.log("\n1. Deploying QUIVToken...");
  const QUIVToken = await hre.ethers.getContractFactory("QUIVToken");
  const token = await QUIVToken.deploy(deployer.address);
  await token.waitForDeployment();
  const tokenAddress = await token.getAddress();
  console.log("QUIVToken deployed to:", tokenAddress);

  // Deploy QUIVerStaking
  console.log("\n2. Deploying QUIVerStaking...");
  const QUIVerStaking = await hre.ethers.getContractFactory("QUIVerStaking");
  const staking = await QUIVerStaking.deploy(tokenAddress);
  await staking.waitForDeployment();
  const stakingAddress = await staking.getAddress();
  console.log("QUIVerStaking deployed to:", stakingAddress);

  // Deploy QUIVerSettlement
  console.log("\n3. Deploying QUIVerSettlement...");
  const QUIVerSettlement = await hre.ethers.getContractFactory("QUIVerSettlement");
  const settlement = await QUIVerSettlement.deploy(tokenAddress);
  await settlement.waitForDeployment();
  const settlementAddress = await settlement.getAddress();
  console.log("QUIVerSettlement deployed to:", settlementAddress);

  // Grant roles
  console.log("\n4. Setting up roles...");
  
  // Grant MINTER_ROLE to staking contract
  const MINTER_ROLE = await token.MINTER_ROLE();
  await token.grantRole(MINTER_ROLE, stakingAddress);
  console.log("Granted MINTER_ROLE to staking contract");

  // Set token address in staking contract
  await staking.setTokenAddress(tokenAddress);
  console.log("Set token address in staking contract");

  // Save deployment addresses
  const deploymentInfo = {
    network: hre.network.name,
    chainId: hre.network.config.chainId,
    deployer: deployer.address,
    timestamp: new Date().toISOString(),
    contracts: {
      QUIVToken: tokenAddress,
      QUIVerStaking: stakingAddress,
      QUIVerSettlement: settlementAddress
    }
  };

  const deploymentsDir = path.join(__dirname, "../deployments");
  if (!fs.existsSync(deploymentsDir)) {
    fs.mkdirSync(deploymentsDir);
  }

  const filename = path.join(deploymentsDir, `${hre.network.name}.json`);
  fs.writeFileSync(filename, JSON.stringify(deploymentInfo, null, 2));
  console.log(`\nDeployment info saved to ${filename}`);

  // Verify contracts if not on hardhat network
  if (hre.network.name !== "hardhat" && hre.network.name !== "localhost") {
    console.log("\n5. Waiting for block confirmations...");
    await new Promise(resolve => setTimeout(resolve, 30000)); // Wait 30 seconds

    console.log("\n6. Verifying contracts...");
    try {
      await hre.run("verify:verify", {
        address: tokenAddress,
        constructorArguments: [deployer.address]
      });
      console.log("QUIVToken verified");
    } catch (error) {
      console.log("QUIVToken verification failed:", error.message);
    }

    try {
      await hre.run("verify:verify", {
        address: stakingAddress,
        constructorArguments: [tokenAddress]
      });
      console.log("QUIVerStaking verified");
    } catch (error) {
      console.log("QUIVerStaking verification failed:", error.message);
    }

    try {
      await hre.run("verify:verify", {
        address: settlementAddress,
        constructorArguments: [tokenAddress]
      });
      console.log("QUIVerSettlement verified");
    } catch (error) {
      console.log("QUIVerSettlement verification failed:", error.message);
    }
  }

  console.log("\nDeployment complete!");
  console.log("================================");
  console.log("Contract Addresses:");
  console.log("QUIVToken:", tokenAddress);
  console.log("QUIVerStaking:", stakingAddress);
  console.log("QUIVerSettlement:", settlementAddress);
  console.log("================================");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });