# QUIVer Token Economics (QUIV)

## Overview
The QUIV token is the native utility and governance token of the QUIVer network, designed to incentivize participation, ensure quality of service, and enable decentralized governance of the AI inference marketplace.

## Token Utility

### 1. **Payment for AI Inference** 
- Primary medium of exchange for LLM inference requests
- Providers receive QUIV for processing prompts
- Consumers pay QUIV for AI services

### 2. **Staking & Collateral**
- Providers stake QUIV to participate in the network
- Higher stakes unlock priority tiers and better rewards
- Stake acts as collateral against misbehavior

### 3. **Quality Assurance**
- Canary system uses staked tokens as security
- Slashing mechanism for incorrect or malicious responses
- Reputation scores affect earning potential

### 4. **Governance**
- Token holders vote on protocol upgrades
- Proposal submission requires minimum QUIV holdings
- Voting power proportional to staked tokens

## Token Distribution

### Initial Supply: 1 Billion QUIV

| Allocation | Amount | Percentage | Vesting |
|------------|--------|------------|---------|
| Provider Rewards | 350M | 35% | 10-year emission |
| Community Treasury | 200M | 20% | DAO-controlled |
| Team & Advisors | 150M | 15% | 4-year vesting, 1-year cliff |
| Private Sale | 100M | 10% | 1-year vesting |
| Public Sale | 100M | 10% | No vesting |
| Ecosystem Fund | 50M | 5% | 5-year release |
| Liquidity Provision | 50M | 5% | Immediate |

## Emission Schedule

### Provider Rewards (350M QUIV)
- **Year 1-3**: 50M/year (high growth phase)
- **Year 4-6**: 35M/year (maturation phase)
- **Year 7-10**: 25M/year (stability phase)
- **Post-10 years**: Halving every 4 years

### Inflation Control
- Maximum annual inflation: 5% after year 10
- Burn mechanisms from protocol fees
- Target long-term equilibrium: 2-3% net inflation

## Staking Tiers & Benefits

### Tier Requirements
| Tier | Minimum Stake | Priority Multiplier | APY Base |
|------|---------------|-------------------|----------|
| Bronze | 10,000 QUIV | 1.0x | 8% |
| Silver | 50,000 QUIV | 1.1x | 10% |
| Gold | 250,000 QUIV | 1.25x | 12% |
| Platinum | 1,000,000 QUIV | 1.5x | 15% |

### Additional Benefits
- **Request Priority**: Higher tiers get preferential routing
- **Fee Discounts**: 5-20% reduction in protocol fees
- **Governance Weight**: 1.5x-3x voting power multipliers
- **Slashing Protection**: Higher tiers have lower slashing rates

## Fee Structure

### Transaction Fees
- **Base Fee**: 0.5% of transaction value
- **Network Fee**: 0.1% (burned)
- **Treasury Fee**: 0.2% (DAO treasury)
- **Referral Fee**: 0.2% (if applicable)

### Provider Economics
- **Gross Revenue**: 100% of consumer payments
- **Protocol Fee**: 5% (from gross)
- **Net Revenue**: 95% to provider
- **Staking Rewards**: Additional 8-15% APY

## Quality Metrics & Slashing

### Performance Metrics
1. **Response Time** (weight: 25%)
   - Target: <2 seconds
   - Penalty: -0.1% stake per violation

2. **Accuracy Score** (weight: 40%)
   - Measured via canary responses
   - Penalty: -0.5% stake for failures

3. **Uptime** (weight: 20%)
   - Target: >99.9%
   - Penalty: -0.2% stake per hour downtime

4. **Throughput** (weight: 15%)
   - Requests handled per hour
   - Affects priority score only

### Slashing Rules
- **Minor Infractions**: 0.1-1% of stake
- **Major Violations**: 1-5% of stake
- **Malicious Behavior**: 10-20% of stake
- **Cooldown Period**: 24 hours between slashes

## Governance

### Proposal Types
1. **Technical Upgrades**: 60% approval required
2. **Economic Parameters**: 55% approval required
3. **Treasury Allocations**: 50% approval required
4. **Emergency Actions**: 75% approval required

### Voting Power
```
Voting Power = QUIV Staked × Tier Multiplier × Time Multiplier
```

### Time Multipliers
- 0-3 months: 1.0x
- 3-6 months: 1.1x
- 6-12 months: 1.25x
- 12+ months: 1.5x

## Revenue Model

### Sources of Protocol Revenue
1. **Transaction Fees**: 5% of all inference payments
2. **Channel Opening Fees**: Fixed 10 QUIV
3. **Premium Features**: Subscription tiers
4. **Data Marketplace**: 10% of data sales

### Revenue Distribution
- **Burn**: 50% (deflationary pressure)
- **Staking Rewards**: 30%
- **Treasury**: 15%
- **Development**: 5%

## Economic Sustainability

### Burn Mechanisms
1. **Network Fees**: 100% burned
2. **Protocol Revenue**: 50% burned
3. **Slashed Tokens**: 100% burned
4. **Inactive Stakes**: Partial burn after 1 year

### Supply/Demand Dynamics
- **Demand Drivers**:
  - Inference payments (primary)
  - Staking requirements
  - Governance participation
  - Speculative holding

- **Supply Constraints**:
  - Fixed maximum supply
  - Significant staking lock-up
  - Burn mechanisms
  - Vesting schedules

## Token Metrics & Projections

### Key Metrics (Target Year 5)
- **Circulating Supply**: 600M QUIV
- **Staked Percentage**: 40-50%
- **Daily Transaction Volume**: $10M+
- **Active Providers**: 10,000+
- **Annual Burn Rate**: 2-3%

### Price Discovery Factors
1. **Network Usage**: Direct correlation with inference demand
2. **Provider Growth**: Supply-side expansion
3. **Technology Advances**: Model efficiency improvements
4. **Market Competition**: Relative positioning
5. **Macro Conditions**: Crypto market cycles

## Risk Mitigation

### Economic Risks
- **Hyperinflation**: Capped emission + burn mechanisms
- **Stake Centralization**: Tiered benefits with diminishing returns
- **Price Volatility**: Treasury reserves for stability
- **Fork Protection**: Time-locked governance changes

### Technical Risks
- **Smart Contract Bugs**: Audits + bug bounties
- **Oracle Failures**: Multiple data sources
- **Network Attacks**: Slashing + reputation system

## Implementation Timeline

### Phase 1: Foundation (Months 1-3)
- Token contract deployment
- Basic staking mechanism
- Initial liquidity provision

### Phase 2: Growth (Months 4-9)
- Full staking tiers
- Governance activation
- Exchange listings

### Phase 3: Maturity (Months 10-12)
- Advanced features
- Cross-chain bridges
- Institutional integration

## Conclusion

The QUIV tokenomics create a sustainable ecosystem that:
- Incentivizes high-quality AI inference provision
- Ensures fair compensation for all participants
- Maintains long-term value through careful supply management
- Enables true decentralization through governance

The model is designed to capture value from the growing AI inference market while providing the economic security necessary for a mission-critical infrastructure service.