// +build bdd

package emoney

import (
	"testing"
	"time"

	nt "emoney/networktest"
	apptypes "emoney/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"
)

var (
	testnet = nt.NewTestnet()
)

func init() {
	apptypes.ConfigureSDK()
}

func createNewTestnet() {
	awaitReady, err := testnet.Restart()
	Expect(err).ShouldNot(HaveOccurred())
	Expect(awaitReady()).To(BeTrue())
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

	config.DefaultReporterConfig.SlowSpecThreshold = time.Hour.Seconds()

	RunSpecs(t, "em-ledger integration tests")
}
