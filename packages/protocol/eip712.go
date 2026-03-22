package protocol

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type Bill struct {
	Buyer       common.Address \`json:"buyer"\`
	Seller      common.Address \`json:"seller"\`
	TotalAmount *big.Int       \`json:"totalAmount"\`
	Nonce       uint64         \`json:"nonce"\`
	ChainID     uint64         \`json:"chainId"\`
}

func HashBill(bill Bill) ([]byte, error) {
	data := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Bill": []apitypes.Type{
				{Name: "buyer", Type: "address"},
				{Name: "seller", Type: "address"},
				{Name: "totalAmount", Type: "uint256"},
				{Name: "nonce", Type: "uint256"},
			},
		},
		PrimaryType: "Bill",
		Domain: apitypes.TypedDataDomain{
			Name:              "DeAPN",
			Version:           "1",
			ChainId:           (*math.HexOrDecimal256)(big.NewInt(int64(bill.ChainID))),
			VerifyingContract: "0x0000000000000000000000000000000000000000",
		},
		Message: apitypes.TypedDataMessage{
			"buyer":       bill.Buyer.Hex(),
			"seller":      bill.Seller.Hex(),
			"totalAmount": bill.TotalAmount.String(),
			"nonce":       fmt.Sprintf("%d", bill.Nonce),
		},
	}

	typedDataHash, err := data.HashStruct(data.PrimaryType, data.Message)
	if err != nil {
		return nil, err
	}

	domainSeparator, err := data.HashStruct("EIP712Domain", data.Domain.Map())
	if err != nil {
		return nil, err
	}

	rawData := []byte("\x19\x01")
	rawData = append(rawData, domainSeparator...)
	rawData = append(rawData, typedDataHash...)
	
	return crypto.Keccak256(rawData), nil
}
