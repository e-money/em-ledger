// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney_test

import (
	"fmt"
	"github.com/e-money/em-ledger/networktest"
	. "github.com/onsi/ginkgo"
)

// Setup a testnet and leave it running for local experimentation
var _ = Describe("Local Testnet", func() {

	tearDownAfterTests = false

	Describe("Authority manages issuers", func() {
		It("creates a new testnet", createNewTestnet)

		It("Prints information to the user", func() {
			var (
				keystore        = testnet.Keystore.GetPath()
				chainid         = testnet.ChainID()
				acc1            = testnet.Keystore.Key1
				node            = networktest.DefaultNode
				defaultPassword = networktest.KeyPwd
			)

			fmt.Println("\nKeystore location", keystore)
			fmt.Println("Key store passwords", defaultPassword)
			fmt.Println("Local net chain-id", chainid)
			fmt.Println("Node address", node)
			// todo (reviewer) : rest server must be enabled in config/app.toml
			//fmt.Println("Lite client interface available at http://localhost:1317/swagger-ui/")

			fmt.Println("Command-line flags for testnet:")
			fmt.Printf("--home %v --node %v --chain-id %v\n", keystore, node, chainid)

			fmt.Println("\n -- Example commands:")
			fmt.Printf("./build/emd  keys list --home %v --keyring-backend test\n\n", keystore)
			fmt.Printf("./build/emd  q staking validators --home %v --node %v --chain-id %v\n\n", keystore, node, chainid)
			fmt.Printf("./build/emd  q account %v --node %v\n\n", acc1.GetAddress(), node)

			fmt.Printf("./build/emd  tx market add-limit 50000eeur 45000echf orderid1 --from %v --node %v --chain-id %v --home %v --yes --keyring-backend test\n\n", acc1.GetAddress(), node, chainid, keystore)

			fmt.Printf("./build/emd  q market instrument eeur echf --node %v\n\n", node)

			fmt.Println(" -- Run this command for a pre-configured local environment:")
			fmt.Printf("EM_NODE=%v EM_HOME=%v EM_KEYRING_BACKEND=test EM_CHAIN_ID=%v sh\n", node, keystore, chainid)

		})
	})
})
