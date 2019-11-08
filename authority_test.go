// +build bdd

package emoney

import (
	"emoney/x/issuer/types"
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

const (
	// gjson paths
	QGetInflationEUR = "assets.#(denom==\"x2eur\").inflation"
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
			_, success, err := emcli.AuthorityCreateIssuer(Authority, Issuer, "x2eur", "x0jpy")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue())

			bz, err := emcli.QueryIssuers()
			Expect(err).ShouldNot(HaveOccurred())

			var issuers types.Issuers
			json.Unmarshal(bz, &issuers)

			Expect(issuers).To(HaveLen(1))
			Expect(issuers[0].Denoms).To(ConsistOf("x2eur", "x0jpy"))
		})

		It("imposter attempts to act as authority", func() {
			_, success, err := emcli.AuthorityCreateIssuer(Issuer, LiquidityProvider, "x2chf", "x2dkk")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeFalse())
		})

		It("authority assigns a second issuer to same denomination", func() {
			_, success, err := emcli.AuthorityCreateIssuer(Authority, OtherIssuer, "x2dkk", "x0jpy")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeFalse())
		})

		It("authority creates a second issuer", func() {
			_, success, err := emcli.AuthorityCreateIssuer(Authority, OtherIssuer, "x2dkk")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
		})

		It("creates a liquidity provider", func() {
			// The issuer makes a liquidity provider of EUR
			_, success, err := emcli.IssuerIncreaseCredit(Issuer, LiquidityProvider, "50000x2eur")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			bz, err := emcli.QueryAccountJson(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			lpaccount := gjson.ParseBytes(bz)
			credit := lpaccount.Get("value.credit").Array()
			Expect(credit).To(HaveLen(1))
			Expect(credit[0].Get("denom").Str).To(Equal("x2eur"))
			Expect(credit[0].Get("amount").Str).To(Equal("50000"))
		})

		It("changes inflation of a denomination", func() {
			bz, err := emcli.QueryInflation()
			Expect(err).ToNot(HaveOccurred())

			s := gjson.ParseBytes(bz).Get(QGetInflationEUR).Str
			inflationBefore, _ := sdk.NewDecFromStr(s)

			_, success, err := emcli.IssuerSetInflation(Issuer, "x2eur", "0.1")
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
			_, success, err := emcli.IssuerSetInflation(OtherIssuer, "x2eur", "0.5")

			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeFalse())
		})

		It("liquidity provider draws on credit", func() {
			balanceBefore, creditBefore, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			_, success, err := emcli.LiquidityProviderMint(LiquidityProvider, "20000x2eur")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue())

			balanceAfter, creditAfter, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			Expect(balanceAfter).To(Equal(balanceBefore + 20000))
			Expect(creditAfter).To(Equal(creditBefore - 20000))
		})

		It("liquidity provider attempts to overdraw its credit", func() {
			balanceBefore, creditBefore, err := emcli.QueryAccount(LiquidityProvider.GetAddress())

			_, success, err := emcli.LiquidityProviderMint(LiquidityProvider, "500000x2eur")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeFalse())

			balanceAfter, creditAfter, err := emcli.QueryAccount(LiquidityProvider.GetAddress())

			Expect(balanceAfter).To(Equal(balanceBefore))
			Expect(creditAfter).To(Equal(creditBefore))
		})

		It("liquidity provider burns some tokens", func() {
			balanceBefore, creditBefore, err := emcli.QueryAccount(LiquidityProvider.GetAddress())

			_, success, err := emcli.LiquidityProviderBurn(LiquidityProvider, "500000x2eur")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			balanceAfter, creditAfter, err := emcli.QueryAccount(LiquidityProvider.GetAddress())

			Expect(balanceAfter).To(Equal(balanceBefore - 500000))
			Expect(creditAfter).To(Equal(creditBefore + 500000))
		})

		It("liquidity provider gets credit reduced", func() {
			_, creditBefore, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			_, success, err := emcli.IssuerDecreaseCredit(Issuer, LiquidityProvider, "10000x2eur")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue())

			_, creditAfter, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			Expect(creditAfter).To(Equal(creditBefore - 10000))
		})

		It("liquidity provider gets revoked", func() {
			_, success, err := emcli.IssuerRevokeCredit(Issuer, LiquidityProvider)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue())

			bz, err := emcli.QueryAccountJson(LiquidityProvider.GetAddress())
			credit := gjson.ParseBytes(bz).Get("value.Credit")
			Expect(credit.Exists()).To(BeFalse())
		})

		It("former liquidity provider attempts to draw on credit", func() {
			balanceBefore, _, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			_, success, err := emcli.LiquidityProviderMint(LiquidityProvider, "10000x2eur")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeFalse())

			balanceAfter, _, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			Expect(balanceBefore).To(Equal(balanceAfter))
		})

		It("issuer gets revoked", func() {
			_, success, err := emcli.AuthorityDestroyIssuer(Authority, Issuer)
			Expect(success).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())

			_, success, err = emcli.IssuerSetInflation(Issuer, "x2eur", "0.5")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeFalse())
		})
	})
})
