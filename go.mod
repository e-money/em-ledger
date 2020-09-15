module github.com/e-money/em-ledger

go 1.12

require (
	github.com/coreos/go-etcd v2.0.0+incompatible // indirect
	github.com/cosmos/cosmos-sdk v0.39.0
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/emirpasic/gods v1.12.0
	github.com/etcd-io/bbolt v1.3.3 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/onsi/ginkgo v1.7.0
	github.com/onsi/gomega v1.4.3
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.6.3
	github.com/stretchr/testify v1.6.1
	github.com/stumble/gorocksdb v0.0.3 // indirect
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.6
	github.com/tendermint/tm-db v0.5.1
	github.com/tidwall/gjson v1.3.2
	github.com/tidwall/sjson v1.0.4
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8 // indirect
)

// replace github.com/cosmos/cosmos-sdk => ./tmpvendor/cosmos-sdk

// replace github.com/tendermint/tendermint => ./tmpvendor/tendermint
