// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

interface IDeAPNStaking {
    function slash(address node, uint256 amount, address beneficiary) external;
}

interface IDeAPNSettlement {
    function balances(address user) external view returns (uint256);
}

contract DeAPNArbiter is Ownable {
    enum DisputeState { None, Challenged, Resolved }

    struct Dispute {
        address challenger;
        address defendant;
        bytes32 requestHash;
        uint256 escrowAmount;
        uint256 challengedAt;
        bytes32 proofCommitment;
        DisputeState state;
        bool isNodeInnocent;
    }

    IDeAPNStaking public staking;
    IDeAPNSettlement public settlement;
    IERC20 public paymentToken;

    mapping(uint256 => Dispute) public disputes;
    uint256 public disputeCount;
    
    mapping(address => bool) public resolvers;
    
    uint256 public challengeWindow = 180 seconds;
    uint256 public slashMultiplier = 10;

    event DisputeOpened(uint256 indexed id, address indexed challenger, address indexed defendant, uint256 amount);
    event ProofSubmitted(uint256 indexed id, bytes32 commitment);
    event DisputeResolved(uint256 indexed id, bool nodeInnocent);
    event ResolverAdded(address indexed resolver);
    event ResolverRemoved(address indexed resolver);

    constructor(address _staking, address _settlement, address _paymentToken) Ownable(msg.sender) {
        staking = IDeAPNStaking(_staking);
        settlement = IDeAPNSettlement(_settlement);
        paymentToken = IERC20(_paymentToken);
        resolvers[msg.sender] = true;
    }

    modifier onlyResolver() {
        require(resolvers[msg.sender], "Not an authorized resolver");
        _;
    }

    function addResolver(address resolver) external onlyOwner {
        resolvers[resolver] = true;
        emit ResolverAdded(resolver);
    }

    function removeResolver(address resolver) external onlyOwner {
        resolvers[resolver] = false;
        emit ResolverRemoved(resolver);
    }

    function openDispute(address node, bytes32 requestHash, uint256 amount) external {
        // In a real implementation, we would need to verify the bill exists in settlement or use a more complex flow.
        // For MVP Phase 2, we assume the router/buyer triggers this and funds are handled accordingly.
        
        disputeCount++;
        disputes[disputeCount] = Dispute({
            challenger: msg.sender,
            defendant: node,
            requestHash: requestHash,
            escrowAmount: amount,
            challengedAt: block.timestamp,
            proofCommitment: 0,
            state: DisputeState.Challenged,
            isNodeInnocent: false
        });

        emit DisputeOpened(disputeCount, msg.sender, node, amount);
    }

    function submitProof(uint256 id, bytes32 commitment) external {
        Dispute storage d = disputes[id];
        require(msg.sender == d.defendant, "Only defendant can submit proof");
        require(d.state == DisputeState.Challenged, "Dispute not in challenged state");
        require(block.timestamp <= d.challengedAt + challengeWindow, "Challenge window closed");

        d.proofCommitment = commitment;
        emit ProofSubmitted(id, commitment);
    }

    function resolveDispute(uint256 id, bool nodeInnocent) external onlyResolver {
        Dispute storage d = disputes[id];
        require(d.state == DisputeState.Challenged, "Dispute already resolved");
        
        d.state = DisputeState.Resolved;
        d.isNodeInnocent = nodeInnocent;

        if (!nodeInnocent) {
            // Slash node and pay beneficiary (challenger)
            uint256 penalty = d.escrowAmount * slashMultiplier;
            staking.slash(d.defendant, penalty, d.challenger);
        }

        emit DisputeResolved(id, nodeInnocent);
    }

    function setSlashMultiplier(uint256 multiplier) external onlyOwner {
        slashMultiplier = multiplier;
    }
}
