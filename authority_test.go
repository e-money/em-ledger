// +build bdd

package emoney

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"time"

	nt "emoney/networktest"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Authority Test Suite")
}

var _ = Describe("Authority", func() {
	testnet := nt.NewTestnet()
	emcli := nt.NewEmcli()

	BeforeSuite(func() {
		fmt.Println("Beforesuite!")
		err := testnet.Setup()
		Expect(err).ShouldNot(HaveOccurred())

		err = testnet.Start()
		Expect(err).ShouldNot(HaveOccurred())

		time.Sleep(5 * time.Second)
	})

	AfterSuite(func() {
		fmt.Println("Aftersuite!")
		err := testnet.Teardown()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("Authority can manage issuers", func() {
		Context("Create an issuer", func() {
			It("Must be possible", func() {
				json, err := emcli.QueryInflation()
				if err != nil {
					Fail(" *** Error")
					return
				}
				fmt.Println("Inflation:\n", string(json))

				fmt.Println(" *** Inside test. Sleeping for a while!")
				time.Sleep(30 * time.Second)
				fmt.Println(" *** Done sleeping in test!")
			})
		})
	})
})
