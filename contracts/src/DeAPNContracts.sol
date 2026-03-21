// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract DeAPNToken is ERC20, Ownable {
    constructor() ERC20("DeAI Proxy Network", "PROXY") Ownable(msg.sender) {
        _mint(msg.sender, 100_000_000 * 10**decimals());
    }
}

contract DeAPNStaking is Ownable {
    IERC20 public proxyToken;
    
    struct NodeInfo {
        uint256 stakedAmount;
        bool isActive;
        string endpoint;
    }

    mapping(address => NodeInfo) public nodes;
    uint256 public minStake = 1000 * 10**18;

    event NodeStaked(address indexed node, uint256 amount);
    event NodeUnstaked(address indexed node, uint256 amount);

    constructor(address _proxyToken) Ownable(msg.sender) {
        proxyToken = IERC20(_proxyToken);
    }

    function stake(uint256 amount) external {
        require(amount >= minStake, "Below minimum stake");
        proxyToken.transferFrom(msg.sender, address(this), amount);
        
        nodes[msg.sender].stakedAmount += amount;
        nodes[msg.sender].isActive = true;
        
        emit NodeStaked(msg.sender, amount);
    }

    function unstake(uint256 amount) external {
        require(nodes[msg.sender].stakedAmount >= amount, "Insufficient stake");
        nodes[msg.sender].stakedAmount -= amount;
        if (nodes[msg.sender].stakedAmount < minStake) {
            nodes[msg.sender].isActive = false;
        }
        proxyToken.transfer(msg.sender, amount);
        
        emit NodeUnstaked(msg.sender, amount);
    }

    function isNodeEligible(address node) external view returns (bool) {
        return nodes[node].isActive;
    }
}
