// +build bdd

package emoney

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	nt "emoney/networktest"
	apptypes "emoney/types"
	"emoney/x/issuer/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

const (
	// gjson paths
	QGetCreditEUR  = "value.Credit.#(denom==\"x2eur\").amount"
	QGetBalanceEUR = "value.Account.value.coins.#(denom==\"x2eur\").amount"
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
	keystore := testnet.Keystore
	emcli := nt.NewEmcli(keystore)

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
				bz, err := emcli.AuthorityCreateIssuer(keystore.Key1, "x2eur", "x0jpy")
				Expect(err).ShouldNot(HaveOccurred())

				txhash := gjson.ParseBytes(bz).Get("txhash").String()
				Expect(txhash).To(Not(BeEmpty()))

				// TODO Create a better way to detect chain events and wait for them.
				time.Sleep(time.Second)

				success, err := emcli.QueryTransactionSucessful(txhash)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(success).To(BeTrue())

				bz, err = emcli.QueryIssuers()
				Expect(err).ShouldNot(HaveOccurred())

				var issuers types.Issuers
				json.Unmarshal(bz, &issuers)

				Expect(issuers).To(HaveLen(1))
				Expect(issuers[0].Denoms).To(ConsistOf("x2eur", "x0jpy"))
			})

			It("creates a liquidity provider", func() {
				// The issuer Key1 makes itself a liquidity provider of EUR
				_, err := emcli.IssuerIncreaseCredit(keystore.Key1, keystore.Key1, "50000x2eur")
				Expect(err).ShouldNot(HaveOccurred())

				// TODO Wait on accepted transactions instead.
				time.Sleep(time.Second)

				bz, err := emcli.QueryAccount(keystore.Key1.GetAddress())
				Expect(err).ShouldNot(HaveOccurred())

				lpaccount := gjson.ParseBytes(bz)
				credit := lpaccount.Get("value.Credit").Array()
				Expect(credit).To(HaveLen(1))
				Expect(credit[0].Get("denom").Str).To(Equal("x2eur"))
				Expect(credit[0].Get("amount").Str).To(Equal("50000"))
			})

			It("liquidity provider draws on credit", func() {
				bz, err := emcli.QueryAccount(keystore.Key1.GetAddress())
				Expect(err).ShouldNot(HaveOccurred())

				queryresponse := gjson.ParseBytes(bz)
				s := queryresponse.Get(QGetBalanceEUR).Str
				balanceBefore, _ := strconv.Atoi(s)

				s = queryresponse.Get(QGetCreditEUR).Str
				creditBefore, _ := strconv.Atoi(s)

				_, err = emcli.LiquidityProviderMint(keystore.Key1, "20000x2eur")
				Expect(err).ShouldNot(HaveOccurred())

				time.Sleep(time.Second)
				bz, err = emcli.QueryAccount(keystore.Key1.GetAddress())
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
				bz, err := emcli.QueryAccount(keystore.Key1.GetAddress())
				s := gjson.ParseBytes(bz).Get(QGetCreditEUR).Str
				creditBefore, _ := strconv.Atoi(s)

				_, err = emcli.IssuerDecreaseCredit(keystore.Key1, keystore.Key1, "10000x2eur")
				Expect(err).ShouldNot(HaveOccurred())

				time.Sleep(time.Second)

				bz, err = emcli.QueryAccount(keystore.Key1.GetAddress())
				s = gjson.ParseBytes(bz).Get(QGetCreditEUR).Str
				creditAfter, _ := strconv.Atoi(s)

				Expect(creditAfter).To(Equal(creditBefore - 10000))
			})
		})
	})
})
