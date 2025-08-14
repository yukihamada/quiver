const { ethers } = require("ethers");
require("dotenv").config();

const privateKey = process.env.PRIVATE_KEY;
const wallet = new ethers.Wallet(privateKey);

console.log("\n🔑 Wallet Information:");
console.log("====================");
console.log("Address:", wallet.address);
console.log("Private Key:", privateKey.substring(0, 6) + "..." + privateKey.substring(58));
console.log("\n📋 To get testnet tokens:");
console.log("1. Copy the address above");
console.log("2. Visit: https://faucet.polygon.technology/");
console.log("3. Select 'Polygon Amoy' network");
console.log("4. Paste your address and request tokens");
console.log("\n⚠️  IMPORTANT: This is a TEST wallet. Never use this private key for real funds!");