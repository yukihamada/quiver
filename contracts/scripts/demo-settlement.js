const { ethers } = require("ethers");
const fs = require("fs");

// Contract ABIs
const TOKEN_ABI = [
  "function balanceOf(address) view returns (uint256)",
  "function transfer(address to, uint256 amount) returns (bool)"
];

const SETTLEMENT_ABI = [
  "function processReceipt(bytes32 receiptId, address provider, uint256 amount)",
  "function withdraw()",
  "function providerBalances(address) view returns (uint256)",
  "event ReceiptProcessed(bytes32 indexed receiptId, address indexed provider, uint256 amount)",
  "event Withdrawn(address indexed provider, uint256 amount)"
];

async function main() {
  // Connect to local hardhat node
  const provider = new ethers.JsonRpcProvider("http://127.0.0.1:8545");
  
  // Use hardhat account 0 as admin
  const adminWallet = new ethers.Wallet(
    "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
    provider
  );
  
  // Use account 1 as provider
  const providerWallet = new ethers.Wallet(
    "0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d",
    provider
  );
  
  // Contract addresses from deployment
  const tokenAddress = "0x5FbDB2315678afecb367f032d93F642f64180aa3";
  const settlementAddress = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512";
  
  // Create contract instances
  const token = new ethers.Contract(tokenAddress, TOKEN_ABI, adminWallet);
  const settlement = new ethers.Contract(settlementAddress, SETTLEMENT_ABI, adminWallet);
  
  console.log("=== QUIVer Settlement Demo ===\n");
  
  // Check initial balances
  const providerBalance = await token.balanceOf(providerWallet.address);
  console.log(`Provider initial QUIV balance: ${ethers.formatEther(providerBalance)} QUIV`);
  
  // Simulate processing a receipt from the inference we did earlier
  const receiptId = ethers.id("JU9fUkkqbn9FVyiUie3mYX"); // Receipt ID from demo
  const rewardAmount = ethers.parseEther("10"); // 10 QUIV tokens as reward
  
  console.log("\n1. Processing inference receipt...");
  const tx1 = await settlement.processReceipt(
    receiptId,
    providerWallet.address,
    rewardAmount
  );
  await tx1.wait();
  console.log(`   ✓ Receipt processed, 10 QUIV allocated to provider`);
  
  // Check provider's pending balance
  const pendingBalance = await settlement.providerBalances(providerWallet.address);
  console.log(`   Provider pending balance: ${ethers.formatEther(pendingBalance)} QUIV`);
  
  // Provider withdraws their rewards
  console.log("\n2. Provider withdrawing rewards...");
  const settlementProvider = settlement.connect(providerWallet);
  const tx2 = await settlementProvider.withdraw();
  await tx2.wait();
  console.log(`   ✓ Withdrawal successful`);
  
  // Check final balance
  const finalBalance = await token.balanceOf(providerWallet.address);
  console.log(`   Provider final QUIV balance: ${ethers.formatEther(finalBalance)} QUIV`);
  
  console.log("\n=== Demo Complete ===");
  console.log("Provider earned 10 QUIV tokens for processing the inference request!");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });