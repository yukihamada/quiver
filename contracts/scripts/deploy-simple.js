const hre = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  console.log("Starting deployment...");
  
  const [deployer] = await hre.ethers.getSigners();
  console.log("Deploying contracts with account:", deployer.address);
  
  const balance = await hre.ethers.provider.getBalance(deployer.address);
  console.log("Account balance:", hre.ethers.formatEther(balance));

  // Deploy SimpleQUIVToken
  console.log("\n1. Deploying SimpleQUIVToken...");
  const SimpleQUIVToken = await hre.ethers.getContractFactory("SimpleQUIVToken");
  const token = await SimpleQUIVToken.deploy();
  await token.waitForDeployment();
  const tokenAddress = await token.getAddress();
  console.log("SimpleQUIVToken deployed to:", tokenAddress);

  // Save deployment addresses
  const deploymentInfo = {
    network: hre.network.name,
    chainId: hre.network.config.chainId,
    deployer: deployer.address,
    timestamp: new Date().toISOString(),
    contracts: {
      SimpleQUIVToken: tokenAddress
    }
  };

  const deploymentsDir = path.join(__dirname, "../deployments");
  if (!fs.existsSync(deploymentsDir)) {
    fs.mkdirSync(deploymentsDir);
  }

  const filename = path.join(deploymentsDir, `${hre.network.name}.json`);
  fs.writeFileSync(filename, JSON.stringify(deploymentInfo, null, 2));
  console.log(`\nDeployment info saved to ${filename}`);

  console.log("\nDeployment complete!");
  console.log("================================");
  console.log("Contract Address:");
  console.log("SimpleQUIVToken:", tokenAddress);
  console.log("================================");
  
  // Update .env file with the deployed address
  const envPath = path.join(__dirname, "../.env");
  let envContent = fs.readFileSync(envPath, 'utf8');
  envContent = envContent.replace(/QUIV_TOKEN_ADDRESS=.*/, `QUIV_TOKEN_ADDRESS=${tokenAddress}`);
  fs.writeFileSync(envPath, envContent);
  console.log("\n✅ Updated .env with token address");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });