// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract DeAPNSettlement {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;

    IERC20 public paymentToken;
    mapping(address => uint256) public balances;
    mapping(bytes32 => bool) public usedBills;

    event SettlementClosed(address indexed buyer, address indexed seller, uint256 amount);
    event Withdrawn(address indexed user, uint256 amount);

    constructor(address _paymentToken) {
        paymentToken = IERC20(_paymentToken);
    }

    function deposit(uint256 amount) external {
        require(paymentToken.transferFrom(msg.sender, address(this), amount), "Transfer failed");
        balances[msg.sender] += amount;
    }

    function withdraw(uint256 amount) external {
        require(balances[msg.sender] >= amount, "Insufficient balance");
        balances[msg.sender] -= amount;
        require(paymentToken.transfer(msg.sender, amount), "Transfer failed");
        emit Withdrawn(msg.sender, amount);
    }

    /**
     * @dev Close a state channel and settle funds.
     * The seller/router submits the last signed bill from the buyer.
     */
    function settle(
        address buyer,
        address seller,
        uint256 amount,
        uint256 nonce,
        bytes calldata signature
    ) external {
        bytes32 billHash = keccak256(abi.encodePacked(buyer, seller, amount, nonce));
        require(!usedBills[billHash], "Bill already used");

        bytes32 ethSignedMessageHash = billHash.toEthSignedMessageHash();
        address recoveredBuyer = ethSignedMessageHash.recover(signature);
        
        require(recoveredBuyer == buyer, "Invalid signature");
        require(balances[buyer] >= amount, "Insufficient buyer balance");
        
        usedBills[billHash] = true;
        balances[buyer] -= amount;
        
        require(paymentToken.transfer(seller, amount), "Transfer failed");
        
        emit SettlementClosed(buyer, seller, amount);
    }
}
