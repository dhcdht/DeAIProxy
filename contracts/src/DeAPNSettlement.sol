// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract DeAPNSettlement {
    using ECDSA for bytes32;

    IERC20 public paymentToken;
    mapping(address => uint256) public balances;

    struct Bill {
        address buyer;
        address seller;
        uint256 amount;
        uint256 nonce;
        bytes buyerSignature;
    }

    event SettlementClosed(address indexed buyer, address indexed seller, uint256 amount);

    constructor(address _paymentToken) {
        paymentToken = IERC20(_paymentToken);
    }

    function deposit(uint256 amount) external {
        paymentToken.transferFrom(msg.sender, address(this), amount);
        balances[msg.sender] += amount;
    }

    /**
     * @dev Close a state channel and settle funds.
     * The seller submits the last signed bill from the buyer.
     */
    function settle(
        address seller,
        uint256 amount,
        uint256 nonce,
        bytes calldata signature
    ) external {
        bytes32 messageHash = keccak256(abi.encodePacked(msg.sender, seller, amount, nonce));
        bytes32 ethSignedMessageHash = messageHash.toEthSignedMessageHash();
        
        address buyer = ethSignedMessageHash.recover(signature);
        
        require(balances[buyer] >= amount, "Insufficient buyer balance");
        
        balances[buyer] -= amount;
        paymentToken.transfer(seller, amount);
        
        emit SettlementClosed(buyer, seller, amount);
    }
}
