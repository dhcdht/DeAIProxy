// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
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
        uint256 unstakeRequestedAt;
        uint256 pendingUnstakeAmount;
    }

    mapping(address => NodeInfo) public nodes;
    mapping(address => bool) public authorizedSlashers;

    uint256 public minStake = 1000 * 10**18;
    uint256 public cooldownPeriod = 7 days;

    event NodeStaked(address indexed node, uint256 amount);
    event UnstakeRequested(address indexed node, uint256 amount);
    event NodeUnstaked(address indexed node, uint256 amount);
    event NodeSlashed(address indexed node, uint256 amount, address beneficiary);
    event SlasherAdded(address indexed slasher);
    event SlasherRemoved(address indexed slasher);

    constructor(address _proxyToken) Ownable(msg.sender) {
        proxyToken = IERC20(_proxyToken);
    }

    modifier onlySlasher() {
        require(authorizedSlashers[msg.sender], "Not an authorized slasher");
        _;
    }

    function addSlasher(address slasher) external onlyOwner {
        authorizedSlashers[slasher] = true;
        emit SlasherAdded(slasher);
    }

    function removeSlasher(address slasher) external onlyOwner {
        authorizedSlashers[slasher] = false;
        emit SlasherRemoved(slasher);
    }

    function stake(uint256 amount) external {
        proxyToken.transferFrom(msg.sender, address(this), amount);
        
        nodes[msg.sender].stakedAmount += amount;
        if (nodes[msg.sender].stakedAmount >= minStake) {
            nodes[msg.sender].isActive = true;
        }
        
        emit NodeStaked(msg.sender, amount);
    }

    function requestUnstake(uint256 amount) external {
        require(nodes[msg.sender].stakedAmount >= amount, "Insufficient stake");
        nodes[msg.sender].pendingUnstakeAmount = amount;
        nodes[msg.sender].unstakeRequestedAt = block.timestamp;
        
        emit UnstakeRequested(msg.sender, amount);
    }

    function executeUnstake() external {
        NodeInfo storage node = nodes[msg.sender];
        require(node.pendingUnstakeAmount > 0, "No pending unstake");
        require(block.timestamp >= node.unstakeRequestedAt + cooldownPeriod, "Cooldown not elapsed");
        
        uint256 amount = node.pendingUnstakeAmount;
        node.pendingUnstakeAmount = 0;
        node.stakedAmount -= amount;
        
        if (node.stakedAmount < minStake) {
            node.isActive = false;
        }
        
        proxyToken.transfer(msg.sender, amount);
        emit NodeUnstaked(msg.sender, amount);
    }

    function slash(address nodeAddr, uint256 amount, address beneficiary) external onlySlasher {
        NodeInfo storage node = nodes[nodeAddr];
        uint256 actualSlash = amount > node.stakedAmount ? node.stakedAmount : amount;
        
        node.stakedAmount -= actualSlash;
        if (node.stakedAmount < minStake) {
            node.isActive = false;
        }
        
        proxyToken.transfer(beneficiary, actualSlash);
        emit NodeSlashed(nodeAddr, actualSlash, beneficiary);
    }

    function isNodeEligible(address node) external view returns (bool) {
        return nodes[node].isActive;
    }
}
