module ecommerce/backend

go 1.22

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
	golang.org/x/crypto v0.21.0
)

// Pinned to the official read-only GitHub mirror of golang.org/x/crypto.
// This is functionally identical to the canonical module and avoids
// requiring direct access to golang.org in network-restricted build
// environments (e.g. CI runners with an egress allowlist). Safe to
// remove this line if your build environment can reach golang.org/proxy.golang.org.
replace golang.org/x/crypto => github.com/golang/crypto v0.21.0
