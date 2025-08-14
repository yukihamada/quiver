// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract SimpleSettlement is Ownable {
    IERC20 public token;
    
    mapping(bytes32 => bool) public processedReceipts;
    mapping(address => uint256) public providerBalances;
    
    event ReceiptProcessed(bytes32 indexed receiptId, address indexed provider, uint256 amount);
    event Withdrawn(address indexed provider, uint256 amount);
    
    constructor(address _token) Ownable(msg.sender) {
        token = IERC20(_token);
    }
    
    function processReceipt(
        bytes32 receiptId,
        address provider,
        uint256 amount
    ) external onlyOwner {
        require(!processedReceipts[receiptId], "Receipt already processed");
        processedReceipts[receiptId] = true;
        providerBalances[provider] += amount;
        emit ReceiptProcessed(receiptId, provider, amount);
    }
    
    function withdraw() external {
        uint256 balance = providerBalances[msg.sender];
        require(balance > 0, "No balance");
        providerBalances[msg.sender] = 0;
        require(token.transfer(msg.sender, balance), "Transfer failed");
        emit Withdrawn(msg.sender, balance);
    }
}