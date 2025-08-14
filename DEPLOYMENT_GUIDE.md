# Smart Contract Deployment Guide

## Prerequisites

1. **Test MATIC**: You need test MATIC tokens to deploy contracts
   - Visit: https://faucet.polygon.technology/
   - Enter address: `0xe3668B9b6EeD1e1ce699418D6316348b52d65171`
   - Request test MATIC for Polygon Amoy testnet

2. **Environment Setup**: The `.env` file is already configured with:
   - Deployment private key
   - RPC endpoints
   - Contract placeholders

## Deployment Steps

Once you have test MATIC:

```bash
cd contracts

# 1. Install dependencies
npm install

# 2. Compile contracts
npm run compile

# 3. Deploy to Polygon Amoy testnet
npm run deploy:amoy
```

## Expected Output

After successful deployment, you'll see:
```
Deploying contracts with account: 0xe3668B9b6EeD1e1ce699418D6316348b52d65171

Deploying QUIVToken...
QUIVToken deployed to: 0x...

Deploying SettlementContract...
SettlementContract deployed to: 0x...

Deploying StakingContract...
StakingContract deployed to: 0x...

✅ Deployment complete!
```

## Contract Addresses

After deployment, update these addresses in your configuration:
- QUIV Token: `<deployed address>`
- Settlement: `<deployed address>`
- Staking: `<deployed address>`

## Verification

To verify contracts on Polygonscan:
```bash
npx hardhat verify --network polygon_amoy <CONTRACT_ADDRESS> <CONSTRUCTOR_ARGS>
```

## Mainnet Deployment

For mainnet deployment:
1. Use real MATIC tokens
2. Update RPC URL to mainnet
3. Double-check all parameters
4. Consider using a multisig wallet