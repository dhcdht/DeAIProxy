module github.com/deapn/router

go 1.26.1

require (
	github.com/deapn/logger v0.0.0-00010101000000-000000000000
	github.com/deapn/protocol v0.0.0-00010101000000-000000000000
	github.com/gorilla/websocket v1.5.3
)

require github.com/google/uuid v1.6.0

replace github.com/deapn/protocol => ../../packages/protocol

replace github.com/deapn/logger => ../../packages/logger
