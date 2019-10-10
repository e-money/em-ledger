// +build bdd

package emoney

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
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
	QGetCreditEUR    = "value.Credit.#(denom==\"x2eur\").amount"
	QGetCredit       = "value.Credit"
	QGetBalanceEUR   = "value.Account.value.coins.#(denom==\"x2eur\").amount"
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
		Issuer            = testnet.Keystore.Key1
		LiquidityProvider = testnet.Keystore.Key2
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
				txhash, err := emcli.AuthorityCreateIssuer(Issuer, "x2eur", "x0jpy")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(txhash).To(Not(BeEmpty()))

				bz, err := emcli.QueryIssuers()
				Expect(err).ShouldNot(HaveOccurred())

				var issuers types.Issuers
				json.Unmarshal(bz, &issuers)

				Expect(issuers).To(HaveLen(1))
				Expect(issuers[0].Denoms).To(ConsistOf("x2eur", "x0jpy"))
			})

			It("creates a liquidity provider", func() {
				// The issuer makes a liquidity provider of EUR
				_, err := emcli.IssuerIncreaseCredit(Issuer, LiquidityProvider, "50000x2eur")
				Expect(err).ToNot(HaveOccurred())

				bz, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
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
				bz, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
				Expect(err).ShouldNot(HaveOccurred())

				queryresponse := gjson.ParseBytes(bz)
				s := queryresponse.Get(QGetBalanceEUR).Str
				balanceBefore, _ := strconv.Atoi(s)

				s = queryresponse.Get(QGetCreditEUR).Str
				creditBefore, _ := strconv.Atoi(s)

				_, err = emcli.LiquidityProviderMint(LiquidityProvider, "20000x2eur")
				Expect(err).ShouldNot(HaveOccurred())

				bz, err = emcli.QueryAccount(LiquidityProvider.GetAddress())
				Expect(err).ShouldNot(HaveOccurred())

				queryresponse = gjson.ParseBytes(bz)
				s = queryresponse.Get(QGetBalanceEUR).Str
				balanceAfter, _ := strconv.Atoi(s)

				s = queryresponse.Get(QGetCreditEUR).Str
				creditAfter, _ := strconv.Atoi(s)

				Expect(balanceAfter).To(Equal(balanceBefore + 20000))
				Expect(creditAfter).To(Equal(creditBefore - 20000))
			})

			It("Liquidity provider gets credit reduced", func() {
				bz, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
				s := gjson.ParseBytes(bz).Get(QGetCreditEUR).Str
				creditBefore, _ := strconv.Atoi(s)

				_, err = emcli.IssuerDecreaseCredit(Issuer, LiquidityProvider, "10000x2eur")
				Expect(err).ShouldNot(HaveOccurred())

				bz, err = emcli.QueryAccount(LiquidityProvider.GetAddress())
				s = gjson.ParseBytes(bz).Get(QGetCreditEUR).Str
				creditAfter, _ := strconv.Atoi(s)

				Expect(creditAfter).To(Equal(creditBefore - 10000))
			})

			It("Liquidity provider gets revoked", func() {
				_, err := emcli.IssuerRevokeCredit(Issuer, LiquidityProvider)
				Expect(err).ShouldNot(HaveOccurred())

				bz, err := emcli.QueryAccount(LiquidityProvider.GetAddress())
				credit := gjson.ParseBytes(bz).Get(QGetCredit)
				Expect(credit.Exists()).To(BeFalse())
			})
		})
	})
})
