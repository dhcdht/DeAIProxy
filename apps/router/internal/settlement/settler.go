package settlement

import (
	"context"
	"log/slog"
	"math/big"
	"sync"

	"github.com/deapn/protocol"
	"github.com/deapn/router/internal/eth"
	"github.com/ethereum/go-ethereum/common"
)

type Settler struct {
	eth      *eth.Client
	log      *slog.Logger
	channels sync.Map // Map[BuyerAddr]protocol.StateChannel
}

func NewSettler(ethClient *eth.Client, logger *slog.Logger) *Settler {
	return &Settler{
		eth: ethClient,
		log: logger,
	}
}

func (s *Settler) ProcessBill(buyer common.Address, seller common.Address, amount *big.Int, nonce uint64, sig []byte, chainID uint64) error {
	val, ok := s.channels.Load(buyer)
	var ch *protocol.StateChannel
	if !ok {
		ch = protocol.NewStateChannel(buyer, seller, chainID)
		s.channels.Store(buyer, ch)
	} else {
		ch = val.(*protocol.StateChannel)
	}

	return ch.Update(amount, nonce, sig)
}

func (s *Settler) SettleOnChain(ctx context.Context, buyer common.Address) error {
	val, ok := s.channels.Load(buyer)
	if !ok {
		return nil
	}
	ch := val.(*protocol.StateChannel)

	s.log.Info("Settling bill on-chain", "buyer", buyer, "amount", ch.TotalAmount)

	// TODO: Call DeAPNSettlement.settle() using raw ABI encoding (Step 1 pattern)
	return nil
}
