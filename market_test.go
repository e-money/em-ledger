// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"time"

	cmn "github.com/tendermint/tendermint/libs/common"
)

var _ = Describe("Market", func() {
	emcli := testnet.NewEmcli()

	var (
		acc1 = testnet.Keystore.Key1
		acc2 = testnet.Keystore.Key2
		acc3 = testnet.Keystore.Key3
	)

	Describe("Authority manages issuers", func() {
		It("creates a new testnet", createNewTestnet)

		It("Basic creation of simple orders", func() {
			time.Sleep(5 * time.Second)
			//bz, err := emcli.QueryAccountJson(acc1.GetAddress())
			//fmt.Println(string(bz))
			//Expect(err).ShouldNot(HaveOccurred())

			for i := 0; i < 10; i++ {
				output, success, err := emcli.MarketAddOrder(acc1, "120000x2eur", fmt.Sprintf("%dx2chf", 90000-i*100), cmn.RandStr(10))
				Expect(err).ToNot(HaveOccurred(), "Error output: %v", output)
				Expect(success).To(BeTrue())
			}

			bz, err := emcli.QueryMarketInstrument("x2eur", "x2chf")
			Expect(err).ToNot(HaveOccurred())
			ir := gjson.ParseBytes(bz)
			Expect(ir.Get("orders").Array()).To(HaveLen(10))
		})

		It("Crashing validator can catch up", func() {
			// Kill a validator
			_, err := testnet.KillValidator(2)
			Expect(err).ToNot(HaveOccurred())

			// Create and execute a couple of orders
			for i := 0; i < 7; i++ {
				fmt.Println(" *** ", i)
				_, success, err := emcli.MarketAddOrder(acc2, "90500x2chf", fmt.Sprintf("%dx2eur", 11000-i*400), cmn.RandStr(10))
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			}

			for i := 0; i < 15; i++ {
				_, success, err := emcli.MarketAddOrder(acc3, "440000x0jpy", fmt.Sprintf("%dx2eur", 100000-i*100), cmn.RandStr(10))
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			}

			bz, err := emcli.QueryMarketInstruments()
			Expect(err).ToNot(HaveOccurred())
			fmt.Println("Order book:\n", string(bz))

			time.Sleep(5 * time.Second)
		})
	})
})
