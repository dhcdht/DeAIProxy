package protocol

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

type TLSProver interface {
	RecordSession(connID string, data []byte) error
	GenerateProof(connID string) ([]byte, error)
}

type TLSVerifier interface {
	VerifyProof(proof []byte, expectedHash []byte) (bool, error)
}

// StubTLSProver is a placeholder for future real zkTLS integration
type StubTLSProver struct {
	sessions map[string][]byte
}

func NewStubTLSProver() *StubTLSProver {
	return &StubTLSProver{sessions: make(map[string][]byte)}
}

func (s *StubTLSProver) RecordSession(connID string, data []byte) error {
	s.sessions[connID] = data
	return nil
}

func (s *StubTLSProver) GenerateProof(connID string) ([]byte, error) {
	data, ok := s.sessions[connID]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	// Simulate a "proof" which is just an HMAC in this prototype
	h := hmac.New(sha256.New, []byte("zktls-secret-stub"))
	h.Write(data)
	return h.Sum(nil), nil
}

type StubTLSVerifier struct{}

func (s *StubTLSVerifier) VerifyProof(proof []byte, expectedData []byte) (bool, error) {
	h := hmac.New(sha256.New, []byte("zktls-secret-stub"))
	h.Write(expectedData)
	expectedProof := h.Sum(nil)
	return hmac.Equal(proof, expectedProof), nil
}
