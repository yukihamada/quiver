const { ethers } = require('ethers');

// Generate a new wallet
const wallet = ethers.Wallet.createRandom();

console.log('\n🔑 New Deployment Wallet Generated:\n');
console.log('Address:', wallet.address);
console.log('Private Key:', wallet.privateKey.slice(2)); // Remove 0x prefix
console.log('\n⚠️  IMPORTANT: Save this private key securely!');
console.log('⚠️  Add it to your .env file as PRIVATE_KEY=<key>');
console.log('\n💰 Fund this address with test MATIC on Polygon Amoy:');
console.log('   https://faucet.polygon.technology/');
console.log('\n');