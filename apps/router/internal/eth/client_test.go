package eth

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

type mockEthCaller struct {
	res []byte
	err error
}

func (m *mockEthCaller) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return m.res, m.err
}

func TestIsNodeEligible(t *testing.T) {
	stakingAddr := "0x1234567890123456789012345678901234567890"
	validWallet := "0x8888888888888888888888888888888888888888"

	tests := []struct {
		name       string
		wallet     string
		mockRes    []byte
		mockErr    error
		wantResult bool
		wantErr    bool
	}{
		{
			name:       "Eligible Node",
			wallet:     validWallet,
			mockRes:    common.LeftPadBytes([]byte{1}, 32),
			wantResult: true,
		},
		{
			name:       "Ineligible Node",
			wallet:     validWallet,
			mockRes:    common.LeftPadBytes([]byte{0}, 32),
			wantResult: false,
		},
		{
			name:    "Invalid Address",
			wallet:  "invalid",
			wantErr: true,
		},
		{
			name:    "RPC Error",
			wallet:  validWallet,
			mockErr: errors.New("rpc error"),
			wantErr: true,
		},
		{
			name:    "Empty Response",
			wallet:  validWallet,
			mockRes: []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockEthCaller{res: tt.mockRes, err: tt.mockErr}
			client := &Client{Eth: mock, StakingAddr: common.HexToAddress(stakingAddr)}

			got, err := client.IsNodeEligible(context.Background(), tt.wallet)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsNodeEligible() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantResult {
				t.Errorf("IsNodeEligible() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestIsValidAddress(t *testing.T) {
	tests := []struct {
		addr string
		want bool
	}{
		{"0x8888888888888888888888888888888888888888", true},
		{"0x123", false},
		{"8888888888888888888888888888888888888888", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		if got := IsValidAddress(tt.addr); got != tt.want {
			t.Errorf("IsValidAddress(%s) = %v, want %v", tt.addr, got, tt.want)
		}
	}
}
