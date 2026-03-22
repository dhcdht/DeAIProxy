package protocol

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type StateChannel struct {
	Buyer       common.Address
	Seller      common.Address
	TotalAmount *big.Int
	Nonce       uint64
	ChainID     uint64
	LastSig     []byte
	mu          sync.RWMutex
}

func NewStateChannel(buyer, seller common.Address, chainID uint64) *StateChannel {
	return &StateChannel{
		Buyer:       buyer,
		Seller:      seller,
		TotalAmount: big.NewInt(0),
		ChainID:     chainID,
	}
}

func (s *StateChannel) Update(newTotal *big.Int, nonce uint64, sig []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if nonce <= s.Nonce {
		return fmt.Errorf("nonce must be increasing")
	}
	if newTotal.Cmp(s.TotalAmount) < 0 {
		return fmt.Errorf("total amount cannot decrease")
	}

	// Verify EIP-712 Signature
	bill := Bill{
		Buyer:       s.Buyer,
		Seller:      s.Seller,
		TotalAmount: newTotal,
		Nonce:       nonce,
		ChainID:     s.ChainID,
	}
	hash, err := HashBill(bill)
	if err != nil {
		return err
	}

	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return err
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	if recoveredAddr != s.Buyer {
		return fmt.Errorf("invalid signature")
	}

	s.TotalAmount = newTotal
	s.Nonce = nonce
	s.LastSig = sig
	return nil
}
