// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney_test

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tidwall/gjson"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Buyback", func() {
	var (
		emcli = testnet.NewEmcli()
		key1  = testnet.Keystore.Key1
		key2  = testnet.Keystore.Key2
	)

	queryBuybackBalance := func() (balance sdk.Coins) {
		bz, err := emcli.QueryBuybackBalance()
		Expect(err).ToNot(HaveOccurred())
		js := gjson.GetBytes(bz, "balance")
		json.Unmarshal([]byte(js.Raw), &balance)
		return
	}

	It("starts a new testnet", func() {
		awaitReady, err := testnet.RestartWithModifications(
			func(bz []byte) []byte {
				// Allow for stablecoin inflation to create a buyback balance
				genesisTime := time.Now().Add(-365 * 24 * time.Hour).UTC()
				bz = setGenesisTime(bz, genesisTime)

				// Disable inflation for NGM token to better detect burn events.
				bz = setInflation(bz, "ungm", sdk.ZeroDec())

				// Disable ejpy inflation to be able to accurately detect fee distributions to the buyback module
				bz = setInflation(bz, "ejpy", sdk.ZeroDec())

				return bz
			})

		Expect(err).ShouldNot(HaveOccurred())
		Expect(awaitReady()).To(BeTrue())
	})

	It("Executes a buyback and checks supply", func() {
		var buybackBalance sdk.Coins
		var bz []byte

		for i := 0; i < 20; i++ { // await
			time.Sleep(500 * time.Millisecond)
			buybackBalance = queryBuybackBalance()
			if len(buybackBalance) > 0 {
				break
			}
		}

		// eeur and echf are inflated
		Expect(buybackBalance).To(HaveLen(2), "Buyback module does not appear to have a balance %v", string(bz))

		supplyBefore, err := emcli.QueryTotalSupply()
		Expect(err).ToNot(HaveOccurred())

		// Sell some NGM tokens to the buyback module and verify that they are burned.
		_, success, err := emcli.MarketAddLimitOrder(key1, "4000ungm", "1000eeur", tmrand.Str(10))
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(BeTrue())

		time.Sleep(4 * time.Second)

		supplyAfter, err := emcli.QueryTotalSupply()
		Expect(err).ToNot(HaveOccurred())

		ngmSupplyBefore, _ := sdk.NewIntFromString(gjson.GetBytes(supplyBefore, "supply.#(denom==\"ungm\").amount").Str)
		ngmSupplyAfter, _ := sdk.NewIntFromString(gjson.GetBytes(supplyAfter, "supply.#(denom==\"ungm\").amount").Str)

		Expect(ngmSupplyBefore.Sub(ngmSupplyAfter)).To(Equal(sdk.NewInt(4000)))

	})

	It("pays a fee using ejpy", func() {
		// Check that the buyback module doesn't have an ejpy balance prior to the test
		balance := queryBuybackBalance()
		Expect(balance.AmountOf("ejpy")).To(Equal(sdk.ZeroInt()))

		_, err := emcli.CustomCommand("tx", "bank", "send", key1.GetName(), key2.GetAddress(), "5000eeur", "--fees", "1000ejpy")
		Expect(err).ToNot(HaveOccurred())

		// Verify that ejpy fee was sent to buyback module
		balance = queryBuybackBalance()
		Expect(balance.AmountOf("ejpy")).To(Equal(sdk.NewInt(1000)))
	})

})
