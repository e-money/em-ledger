// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney_test

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/issuer/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

const (
	// gjson paths
	QGetInflationEUR = "assets.#(denom==\"eeur\").inflation"
)

var _ = Describe("Authority", func() {
	emcli := testnet.NewEmcli()

	var (
		Authority         = testnet.Keystore.Authority
		Issuer            = testnet.Keystore.Key1
		LiquidityProvider = testnet.Keystore.Key2
		OtherIssuer       = testnet.Keystore.Key3
	)

	Describe("Authority manages issuers", func() {
		It("creates a new testnet", createNewTestnet)

		It("creates an issuer", func() {
			time.Sleep(5 * time.Second)

			_, success, err := emcli.AuthorityCreateIssuer(Authority, Issuer, "eeur", "ejpy")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue())

			bz, err := emcli.QueryIssuers()
			Expect(err).ShouldNot(HaveOccurred())

			var issuers types.Issuers
			json.Unmarshal(bz, &issuers)

			Expect(issuers).To(HaveLen(1))
			Expect(issuers[0].Denoms).To(ConsistOf("eeur", "ejpy"))
		})

		It("imposter attempts to act as authority", func() {
			_, success, err := emcli.AuthorityCreateIssuer(Issuer, LiquidityProvider, "echf", "edkk")
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())
		})

		It("authority assigns a second issuer to same denomination", func() {
			_, success, err := emcli.AuthorityCreateIssuer(Authority, OtherIssuer, "edkk", "ejpy")
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())
		})

		It("authority creates a second issuer", func() {
			_, success, err := emcli.AuthorityCreateIssuer(Authority, OtherIssuer, "edkk")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
		})

		It("creates a liquidity provider", func() {
			// The issuer makes a liquidity provider of EUR
			_, success, err := emcli.IssuerIncreaseMintableAmount(Issuer, LiquidityProvider, "50000eeur")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			bz, err := emcli.QueryAccountJson(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			lpaccount := gjson.ParseBytes(bz)
			mintableAmount := lpaccount.Get("value.mintable").Array()
			Expect(mintableAmount).To(HaveLen(1))
			Expect(mintableAmount[0].Get("denom").Str).To(Equal("eeur"))
			Expect(mintableAmount[0].Get("amount").Str).To(Equal("50000"))
		})

		It("changes inflation of a denomination", func() {
			bz, err := emcli.QueryInflation()
			Expect(err).ToNot(HaveOccurred())

			s := gjson.ParseBytes(bz).Get(QGetInflationEUR).Str
			inflationBefore, _ := sdk.NewDecFromStr(s)

			_, success, err := emcli.IssuerSetInflation(Issuer, "eeur", "0.1")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			bz, err = emcli.QueryInflation()
			Expect(err).ToNot(HaveOccurred())

			s = gjson.ParseBytes(bz).Get(QGetInflationEUR).Str
			inflationAfter, _ := sdk.NewDecFromStr(s)

			Expect(inflationAfter).ToNot(Equal(inflationBefore))
			Expect(inflationAfter).To(Equal(sdk.MustNewDecFromStr("0.100")))
		})

		It("attempts to change inflation of denomination not under its control", func() {
			_, success, err := emcli.IssuerSetInflation(OtherIssuer, "eeur", "0.5")

			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())
		})

		It("creates an issuer of a completely new denomination", func() {
			_, success, err := emcli.AuthorityCreateIssuer(Authority, OtherIssuer, "caps")
			Expect(err).To(BeNil())
			Expect(success).To(BeTrue())

			bz, err := emcli.QueryInflation()
			Expect(err).To(BeNil())

			fmt.Println(string(bz))

			s := gjson.ParseBytes(bz).Get("assets.#(denom==\"caps\").inflation").Str
			inflationCaps, _ := sdk.NewDecFromStr(s)
			Expect(inflationCaps).To(Equal(sdk.ZeroDec()))
		})

		It("liquidity provider draws on its mintable amount", func() {
			balanceBefore, mintableBefore, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			_, success, err := emcli.LiquidityProviderMint(LiquidityProvider, "20000eeur")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue())

			balanceAfter, mintableAfter, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			Expect(balanceAfter).To(Equal(balanceBefore + 20000))
			Expect(mintableAfter).To(Equal(mintableBefore - 20000))
		})

		It("liquidity provider attempts to overdraw its mintable balance", func() {
			balanceBefore, mintableBefore, err := emcli.QueryAccount(LiquidityProvider.GetAddress())

			_, success, err := emcli.LiquidityProviderMint(LiquidityProvider, "500000eeur")
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())

			balanceAfter, mintableAfter, err := emcli.QueryAccount(LiquidityProvider.GetAddress())

			Expect(balanceAfter).To(Equal(balanceBefore))
			Expect(mintableAfter).To(Equal(mintableBefore))
		})

		It("liquidity provider burns some tokens", func() {
			balanceBefore, mintableBefore, err := emcli.QueryAccount(LiquidityProvider.GetAddress())

			_, success, err := emcli.LiquidityProviderBurn(LiquidityProvider, "500000eeur")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			balanceAfter, mintableAfter, err := emcli.QueryAccount(LiquidityProvider.GetAddress())

			Expect(balanceAfter).To(Equal(balanceBefore - 500000))
			Expect(mintableAfter).To(Equal(mintableBefore + 500000))
		})

		It("liquidity provider gets mintable amount reduced", func() {
			_, mintableBefore, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			_, success, err := emcli.IssuerDecreaseMintableAmount(Issuer, LiquidityProvider, "10000eeur")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue())

			_, mintableAfter, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			Expect(mintableAfter).To(Equal(mintableBefore - 10000))
		})

		It("liquidity provider gets revoked", func() {
			_, success, err := emcli.IssuerRevokeMinting(Issuer, LiquidityProvider)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue())

			bz, err := emcli.QueryAccountJson(LiquidityProvider.GetAddress())
			mintable := gjson.ParseBytes(bz).Get("value.mintable")
			Expect(mintable.Exists()).To(BeFalse())
		})

		It("former liquidity provider attempts to mint", func() {
			balanceBefore, _, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			_, success, err := emcli.LiquidityProviderMint(LiquidityProvider, "10000eeur")
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())

			balanceAfter, _, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			Expect(balanceBefore).To(Equal(balanceAfter))
		})

		It("issuer gets revoked", func() {
			_, success, err := emcli.AuthorityDestroyIssuer(Authority, Issuer)
			Expect(success).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())

			_, success, err = emcli.IssuerSetInflation(Issuer, "eeur", "0.5")
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())
		})

		It("Authority sets new gas prices", func() {
			prices, err := sdk.ParseDecCoins("0.00005eeur")
			Expect(err).ToNot(HaveOccurred())

			_, success, err := emcli.AuthoritySetMinGasPrices(Authority, prices.String())
			Expect(success).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())

			bz, err := emcli.QueryMinGasPrices()
			Expect(err).ToNot(HaveOccurred())

			_, success, err = emcli.AuthoritySetMinGasPrices(Authority, prices.String(), "--fees", "50eeur")
			Expect(success).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())

			gasPrices := gjson.ParseBytes(bz).Get("min_gas_prices")

			queriedPrices := sdk.DecCoins{}
			for _, price := range gasPrices.Array() {
				gasPrice := sdk.NewDecCoinFromDec(price.Get("denom").Str, sdk.MustNewDecFromStr(price.Get("amount").Str))
				queriedPrices = append(queriedPrices, gasPrice)
			}

			Expect(queriedPrices).To(Equal(prices))

			// A non-authority attempts to set gas prices
			_, success, err = emcli.AuthoritySetMinGasPrices(Issuer, prices.String(), "--fees", "50eeur")
			Expect(success).To(BeFalse())
			Expect(err).To(HaveOccurred())
		})
	})
})
