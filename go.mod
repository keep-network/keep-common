module github.com/keep-network/keep-common

go 1.12

replace (
	github.com/ethereum/go-ethereum => github.com/keep-network/go-ethereum v1.8.27
	github.com/urfave/cli => github.com/keep-network/cli v1.20.0
)

require (
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/aristanetworks/goarista v0.0.0-20190924011532-60b7b74727fd // indirect
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/ethereum/go-ethereum v0.0.0-00010101000000-000000000000
	github.com/ipfs/go-log v0.0.1
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/rs/cors v1.7.0 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/urfave/cli v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20190926114937-fa1a29108794
	golang.org/x/net v0.0.0-20190926025831-c00fd9afed17 // indirect
	golang.org/x/tools v0.0.0-20190925230517-ea99b82c7b93
)
