// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tidwall/gjson"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/sjson"
)

var _ = Describe("Buyback", func() {
	var (
		emcli = testnet.NewEmcli()
		key1  = testnet.Keystore.Key1
		key2  = testnet.Keystore.Key2
	)

	It("starts a new testnet", func() {
		awaitReady, err := testnet.RestartWithModifications(
			func(bz []byte) []byte {
				genesisTime := time.Now().Add(-365 * 24 * time.Hour).UTC()
				bz, _ = sjson.SetBytes(bz, "genesis_time", genesisTime.Format(time.RFC3339))
				return bz
			})

		Expect(err).ShouldNot(HaveOccurred())
		Expect(awaitReady()).To(BeTrue())
	})

	// todo (Alex) : balance at the end does not match expectations
	XIt("queries the buyback balance", func() {
		awaitReady, err := testnet.RestartWithModifications(
			func(bz []byte) []byte {
				genesisTime := time.Now().Add(-365 * 24 * time.Hour).UTC()
				bz, _ = sjson.SetBytes(bz, "genesis_time", genesisTime.Format(time.RFC3339))
				return bz
			})

		Expect(err).ShouldNot(HaveOccurred())
		Expect(awaitReady()).To(BeTrue())

		var js []gjson.Result
		var bz []byte
		for i := 0; i < 20; i++ { // await
			time.Sleep(500 * time.Millisecond)
			var err error
			bz, err = emcli.QueryBuybackBalance()
			Expect(err).ToNot(HaveOccurred())

			js = gjson.GetBytes(bz, "balance").Array()
			if len(js) == 3 {
				break
			}
		}
		Expect(js).To(HaveLen(3), "Buyback module does not appear to have a balance %v", string(bz))

		time.Sleep(4 * time.Second)

		// Generate some trades to set a market price for ungm
		_, success, err := emcli.MarketAddLimitOrder(key1, "1000eeur", "4000ungm", tmrand.Str(10))
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(BeTrue())

		_, success, err = emcli.MarketAddLimitOrder(key2, "4000ungm", "1000eeur", tmrand.Str(10))
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(BeTrue())

		time.Sleep(4 * time.Second)

		supplyBefore, err := emcli.QueryTotalSupply()
		Expect(err).ToNot(HaveOccurred())

		// Sell some NGM tokens to the buyback module and verify that they are burned.
		_, success, err = emcli.MarketAddMarketOrder(key1, "ungm", "1000eeur", tmrand.Str(10), sdk.NewDecWithPrec(1, 2))
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(BeTrue())

		time.Sleep(4 * time.Second)

		supplyAfter, err := emcli.QueryTotalSupply()
		Expect(err).ToNot(HaveOccurred())

		ngmSupplyBefore, _ := sdk.NewIntFromString(gjson.GetBytes(supplyBefore, "supply.#(denom==\"ungm\").amount").Str)
		ngmSupplyAfter, _ := sdk.NewIntFromString(gjson.GetBytes(supplyAfter, "supply.#(denom==\"ungm\").amount").Str)

		Expect(ngmSupplyBefore.Sub(ngmSupplyAfter)).To(Equal(sdk.NewInt(4000)))
	})
})
