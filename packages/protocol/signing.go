package protocol

import (
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func SignChunk(privKey *ecdsa.PrivateKey, reqID string, seqNum uint64, data []byte) (*SignedChunk, error) {
	hash := crypto.Keccak256Hash(data)

	payload := []byte(reqID)
	buf8 := make([]byte, 8)
	binary.BigEndian.PutUint64(buf8, seqNum)
	payload = append(payload, buf8...)
	payload = append(payload, hash.Bytes()...)
	ts := time.Now().Unix()
	binary.BigEndian.PutUint64(buf8, uint64(ts))
	payload = append(payload, buf8...)

	sigHash := crypto.Keccak256Hash(payload)
	sig, err := crypto.Sign(sigHash.Bytes(), privKey)
	if err != nil {
		return nil, err
	}

	return &SignedChunk{
		RequestID: reqID,
		SeqNum:    seqNum,
		DataHash:  hash.Bytes(),
		Timestamp: ts,
		Signature: sig,
	}, nil
}

func VerifyChunk(walletAddr string, chunk *SignedChunk, data []byte) (bool, error) {
	hash := crypto.Keccak256Hash(data)
	if common.BytesToHash(chunk.DataHash).Hex() != hash.Hex() {
		return false, fmt.Errorf("data hash mismatch")
	}

	payload := []byte(chunk.RequestID)
	buf8 := make([]byte, 8)
	binary.BigEndian.PutUint64(buf8, chunk.SeqNum)
	payload = append(payload, buf8...)
	payload = append(payload, chunk.DataHash...)
	binary.BigEndian.PutUint64(buf8, uint64(chunk.Timestamp))
	payload = append(payload, buf8...)

	sigHash := crypto.Keccak256Hash(payload)

	pubKey, err := crypto.SigToPub(sigHash.Bytes(), chunk.Signature)
	if err != nil {
		return false, err
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return strings.EqualFold(recoveredAddr.Hex(), walletAddr), nil
}
