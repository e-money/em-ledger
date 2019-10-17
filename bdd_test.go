// +build bdd

package emoney

import (
	"testing"

	nt "emoney/networktest"
	apptypes "emoney/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	testnet = nt.NewTestnet()
)

func init() {
	apptypes.ConfigureSDK()
}

func TestSuite(t *testing.T) {
	BeforeSuite(func() {
		err := testnet.Setup()
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterSuite(func() {
		err := testnet.Teardown()
		Expect(err).ShouldNot(HaveOccurred())
	})

	RegisterFailHandler(Fail)

	RunSpecs(t, "em-ledger integration tests")
}
