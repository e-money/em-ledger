// +build bdd

package emoney

import (
	nt "emoney/networktest"
	apptypes "emoney/types"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"time"
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

	Describe("Authority can manage issuers", func() {
		Context("Create an issuer", func() {
			It("Must be possible", func() {
				bz, err := emcli.AuthorityCreateIssuer(testnet.Keystore.Key1.GetAddress(), "x2eur", "x0jpy")
				Expect(err).ShouldNot(HaveOccurred())

				// TODO Create a better way to detect chain events and wait for them.
				time.Sleep(2 * time.Second)

				bz, err = emcli.QueryIssuers()
				Expect(err).ShouldNot(HaveOccurred())
				fmt.Println(" *** Issuers:\n", string(bz))

			})
		})
	})
})
