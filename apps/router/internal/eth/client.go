package eth

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthCaller defines the minimal interface needed for contract calls.
type EthCaller interface {
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
}

type Client struct {
	Eth         EthCaller
	StakingAddr common.Address
}

func New(rpcURL, stakingAddr string) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		Eth:         client,
		StakingAddr: common.HexToAddress(stakingAddr),
	}, nil
}

// IsNodeEligible checks if a wallet has enough stake by calling the smart contract.
func (c *Client) IsNodeEligible(ctx context.Context, walletAddr string) (bool, error) {
	if !IsValidAddress(walletAddr) {
		return false, fmt.Errorf("invalid wallet address: %s", walletAddr)
	}

	wallet := common.HexToAddress(walletAddr)

	// Method ID = bytes4(keccak256("isNodeEligible(address)")) = 0x8b30e38a
	methodID := []byte{0x8b, 0x30, 0xe3, 0x8a}
	data := append(methodID, common.LeftPadBytes(wallet.Bytes(), 32)...)

	msg := ethereum.CallMsg{
		To:   &c.StakingAddr,
		Data: data,
	}

	res, err := c.Eth.CallContract(ctx, msg, nil)
	if err != nil {
		return false, fmt.Errorf("contract call failed: %w", err)
	}

	if len(res) == 0 {
		return false, fmt.Errorf("empty response from contract")
	}

	// Result is a 32-byte boolean (0 or 1)
	return big.NewInt(0).SetBytes(res).Cmp(big.NewInt(0)) > 0, nil
}

func IsValidAddress(address string) bool {
	return common.IsHexAddress(address) && strings.HasPrefix(address, "0x")
}
