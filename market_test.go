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
	"strings"
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
			_, success, err := emcli.MarketAddOrder(acc2, "5000x2eur", "100000x0jpy", "acc2cid1")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			// Kill a validator
			_, err = testnet.KillValidator(2)
			Expect(err).ToNot(HaveOccurred())

			// Create and execute a couple of orders
			for i := 0; i < 7; i++ {
				_, success, err := emcli.MarketAddOrder(acc2, "90500x2chf", fmt.Sprintf("%dx2eur", 11000-i*400), cmn.RandStr(10))
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			}

			for i := 0; i < 15; i++ {
				_, success, err := emcli.MarketAddOrder(acc3, "440000x0jpy", fmt.Sprintf("%dx2eur", 100000-i*100), cmn.RandStr(10))
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			}

			_, success, err = emcli.MarketCancelOrder(acc2, "acc2cid1")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			//bz, err := emcli.QueryMarketInstruments()
			//Expect(err).ToNot(HaveOccurred())
			//fmt.Println("Order book:\n", string(bz))

			time.Sleep(4 * time.Second)

			_, err = testnet.ResurrectValidator(2)
			Expect(err).ToNot(HaveOccurred())

			for i := 0; i < 10; i++ {
				_, success, err := emcli.MarketAddOrder(acc3, "90000x2eur", fmt.Sprintf("%dx2chf", 50000-i*100), cmn.RandStr(10))
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			}

			// Wait and while and attempt to discover consensus failure in the logs of the resurrected validator
			time.Sleep(8 * time.Second)

			log, err := testnet.GetValidatorLogs(2)
			Expect(err).ToNot(HaveOccurred())
			if strings.Contains(log, "Wrong Block.Header.AppHash") ||
				strings.Contains(log, "panic") {
				Fail(fmt.Sprintf("Validator 2 does not appear to have re-established consensus:\n%v", log))
			}
		})
	})
})
