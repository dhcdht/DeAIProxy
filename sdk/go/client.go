package sdk

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"sync/atomic"

	"github.com/deapn/protocol"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Client struct {
	baseURL string
	privKey *ecdsa.PrivateKey
	wallet  common.Address
	seller  common.Address
	chainID uint64
	nonce   uint64
	amount  *big.Int
}

func NewClient(baseURL string, privKey *ecdsa.PrivateKey, seller common.Address, chainID uint64) *Client {
	return &Client{
		baseURL: baseURL,
		privKey: privKey,
		wallet:  crypto.PubkeyToAddress(privKey.PublicKey),
		seller:  seller,
		chainID: chainID,
		amount:  big.NewInt(0),
	}
}

func (c *Client) PostWithPayment(path string, body []byte, tokenCost int64) ([]byte, error) {
	newAmount := new(big.Int).Add(c.amount, big.NewInt(tokenCost))
	nonce := atomic.AddUint64(&c.nonce, 1)

	bill := protocol.Bill{
		Buyer:       c.wallet,
		Seller:      c.seller,
		TotalAmount: newAmount,
		Nonce:       nonce,
		ChainID:     c.chainID,
	}

	hash, err := protocol.HashBill(bill)
	if err != nil {
		return nil, err
	}

	sig, err := crypto.Sign(hash, c.privKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Attach payment headers
	req.Header.Set("X-DeAPN-Bill-Amount", newAmount.String())
	req.Header.Set("X-DeAPN-Bill-Nonce", string(nonce))
	req.Header.Set("X-DeAPN-Bill-Sig", base64.StdEncoding.EncodeToString(sig))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	c.amount = newAmount // Update on success
	return io.ReadAll(resp.Body)
}
