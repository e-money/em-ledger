// +build bdd

package emoney

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"

	nt "emoney/networktest"
	apptypes "emoney/types"
	"emoney/x/issuer/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

const (
	// gjson paths
	QGetInflationEUR = "assets.#(denom==\"x2eur\").inflation"
)

func init() {
	apptypes.ConfigureSDK()
}

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Authority Test Suite")
}

var _ = Describe("Authority", func() {
	testnet := nt.NewTestnet()
	emcli := nt.NewEmcli(testnet.Keystore)

	var (
		Authority         = testnet.Keystore.Authority
		Issuer            = testnet.Keystore.Key1
		LiquidityProvider = testnet.Keystore.Key2
		OtherKey          = testnet.Keystore.Key3
	)

	BeforeSuite(func() {
		err := testnet.Setup()
		Expect(err).ShouldNot(HaveOccurred())

		awaitReady, err := testnet.Start()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(awaitReady()).To(BeTrue())
	})

	AfterSuite(func() {
		err := testnet.Teardown()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("Authority manages issuers", func() {
		Context("", func() {
			It("creates an issuer", func() {
				_, err := emcli.AuthorityCreateIssuer(Authority, Issuer, "x2eur", "x0jpy")
				Expect(err).ShouldNot(HaveOccurred())

				bz, err := emcli.QueryIssuers()
				Expect(err).ShouldNot(HaveOccurred())

				var issuers types.Issuers
				json.Unmarshal(bz, &issuers)

				Expect(issuers).To(HaveLen(1))
				Expect(issuers[0].Denoms).To(ConsistOf("x2eur", "x0jpy"))
			})

			It("imposter attempts to act as authority", func() {
				_, err := emcli.AuthorityCreateIssuer(Issuer, LiquidityProvider, "x2chf", "x2dkk")
				Expect(err).To(HaveOccurred())
			})

			It("authority assigns a second issuer to same denomination", func() {
				_, err := emcli.AuthorityCreateIssuer(Authority, OtherKey, "x2dkk", "x0jpy")
				Expect(err).To(HaveOccurred())
			})

			It("creates a liquidity provider", func() {
				// The issuer makes a liquidity provider of EUR
				_, err := emcli.IssuerIncreaseCredit(Issuer, LiquidityProvider, "50000x2eur")
				Expect(err).ToNot(HaveOccurred())

				bz, err := emcli.QueryAccountJson(LiquidityProvider.GetAddress())
				Expect(err).ToNot(HaveOccurred())

				lpaccount := gjson.ParseBytes(bz)
				credit := lpaccount.Get("value.Credit").Array()
				Expect(credit).To(HaveLen(1))
				Expect(credit[0].Get("denom").Str).To(Equal("x2eur"))
				Expect(credit[0].Get("amount").Str).To(Equal("50000"))
			})

			It("changes inflation of a denomination", func() {
				bz, err := emcli.QueryInflation()
				Expect(err).ToNot(HaveOccurred())

				s := gjson.ParseBytes(bz).Get(QGetInflationEUR).Str
				inflationBefore, _ := sdk.NewDecFromStr(s)

				_, err = emcli.IssuerSetInflation(Issuer, "x2eur", "0.1")
				Expect(err).ToNot(HaveOccurred())

				bz, err = emcli.QueryInflation()
				Expect(err).ToNot(HaveOccurred())

				s = gjson.ParseBytes(bz).Get(QGetInflationEUR).Str
				inflationAfter, _ := sdk.NewDecFromStr(s)

				Expect(inflationAfter).ToNot(Equal(inflationBefore))
				Expect(inflationAfter).To(Equal(sdk.MustNewDecFromStr("0.100")))
			})

			It("liquidity provider draws on credit", func() {
				balanceBefore, creditBefore, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
				Expect(err).ShouldNot(HaveOccurred())

				_, err = emcli.LiquidityProviderMint(LiquidityProvider, "20000x2eur")
				Expect(err).ShouldNot(HaveOccurred())

				balanceAfter, creditAfter, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
				Expect(err).ShouldNot(HaveOccurred())

				Expect(balanceAfter).To(Equal(balanceBefore + 20000))
				Expect(creditAfter).To(Equal(creditBefore - 20000))
			})

			It("liquidity provider attempts to overdraw its credit", func() {
				balanceBefore, creditBefore, err := emcli.QueryAccount(LiquidityProvider.GetAddress())

				_, err = emcli.LiquidityProviderMint(LiquidityProvider, "500000x2eur")
				Expect(err).To(HaveOccurred())

				balanceAfter, creditAfter, err := emcli.QueryAccount(LiquidityProvider.GetAddress())

				Expect(balanceAfter).To(Equal(balanceBefore))
				Expect(creditAfter).To(Equal(creditBefore))
			})

			It("liquidity provider burns some tokens", func() {
				balanceBefore, creditBefore, err := emcli.QueryAccount(LiquidityProvider.GetAddress())

				_, err = emcli.LiquidityProviderBurn(LiquidityProvider, "500000x2eur")
				Expect(err).ToNot(HaveOccurred())

				balanceAfter, creditAfter, err := emcli.QueryAccount(LiquidityProvider.GetAddress())

				Expect(balanceAfter).To(Equal(balanceBefore - 500000))
				Expect(creditAfter).To(Equal(creditBefore + 500000))
			})

			It("liquidity provider gets credit reduced", func() {
				_, creditBefore, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
				Expect(err).ToNot(HaveOccurred())

				_, err = emcli.IssuerDecreaseCredit(Issuer, LiquidityProvider, "10000x2eur")
				Expect(err).ShouldNot(HaveOccurred())

				_, creditAfter, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
				Expect(err).ToNot(HaveOccurred())

				Expect(creditAfter).To(Equal(creditBefore - 10000))
			})

			It("liquidity provider gets revoked", func() {
				_, err := emcli.IssuerRevokeCredit(Issuer, LiquidityProvider)
				Expect(err).ShouldNot(HaveOccurred())

				bz, err := emcli.QueryAccountJson(LiquidityProvider.GetAddress())
				credit := gjson.ParseBytes(bz).Get("value.Credit")
				Expect(credit.Exists()).To(BeFalse())
			})

			It("former liquidity provider attempts to draw on credit", func() {
				balanceBefore, _, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
				Expect(err).ToNot(HaveOccurred())

				_, err = emcli.LiquidityProviderMint(LiquidityProvider, "10000x2eur")
				Expect(err).To(HaveOccurred())

				balanceAfter, _, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
				Expect(err).ToNot(HaveOccurred())

				Expect(balanceBefore).To(Equal(balanceAfter))
			})
		})
	})
})
