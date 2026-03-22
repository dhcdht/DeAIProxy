module github.com/deapn/node

go 1.26.1

replace github.com/deapn/protocol => ../../packages/protocol

replace github.com/deapn/logger => ../../packages/logger

require (
	github.com/deapn/logger v0.0.0-00010101000000-000000000000
	github.com/deapn/protocol v0.0.0-00010101000000-000000000000
	github.com/ethereum/go-ethereum v1.17.1
	github.com/gorilla/websocket v1.5.3
)

require (
	github.com/ProjectZKM/Ziren/crates/go-runtime/zkvm_runtime v0.0.0-20251001021608-1fe7b43fc4d6 // indirect
	github.com/bits-and-blooms/bitset v1.20.0 // indirect
	github.com/consensys/gnark-crypto v0.18.1 // indirect
	github.com/crate-crypto/go-eth-kzg v1.4.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/ethereum/c-kzg-4844/v2 v2.1.6 // indirect
	github.com/holiman/uint256 v1.3.2 // indirect
	github.com/supranational/blst v0.3.16 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
)
