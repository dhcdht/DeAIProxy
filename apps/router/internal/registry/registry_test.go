package registry

import (
	"context"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/deapn/router/internal/eth"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/websocket"
)

type mockEthCaller struct {
	res []byte
	err error
}

func (m *mockEthCaller) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return m.res, m.err
}

func TestRegistry_Register(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validWallet := "0x8888888888888888888888888888888888888888"

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")
	
	tests := []struct {
		name    string
		wallet  string
		mockRes []byte
		wantErr bool
	}{
		{
			name:    "Successful Registration",
			wallet:  validWallet,
			mockRes: common.LeftPadBytes([]byte{1}, 32),
			wantErr: false,
		},
		{
			name:    "Failed Staking Check",
			wallet:  validWallet,
			mockRes: common.LeftPadBytes([]byte{0}, 32),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientConn, _, err := websocket.DefaultDialer.Dial(u, nil)
			if err != nil {
				t.Fatalf("failed to dial: %v", err)
			}
			defer clientConn.Close()

			mock := &mockEthCaller{res: tt.mockRes}
			ethClient := &eth.Client{Eth: mock, StakingAddr: common.HexToAddress("0x1234567890123456789012345678901234567890")}
			reg := New(logger, ethClient)

			err = reg.Register("node-1", tt.wallet, clientConn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
