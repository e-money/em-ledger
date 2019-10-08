// +build bdd

package emoney

import (
	nt "emoney/networktest"
	apptypes "emoney/types"
	"emoney/x/issuer/types"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"time"

	"github.com/tidwall/gjson"
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
			It("Create an issuer", func() {
				bz, err := emcli.AuthorityCreateIssuer(testnet.Keystore.Key1.GetAddress(), "x2eur", "x0jpy")
				Expect(err).ShouldNot(HaveOccurred())

				txhash := gjson.ParseBytes(bz).Get("txhash").String()
				Expect(txhash).To(Not(BeEmpty()))

				// TODO Create a better way to detect chain events and wait for them.
				time.Sleep(2 * time.Second)

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

		})
	})
})
