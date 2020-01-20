// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney

import (
	"testing"
	"time"

	nt "github.com/e-money/em-ledger/networktest"
	apptypes "github.com/e-money/em-ledger/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"
)

func init() {
	apptypes.ConfigureSDK()
}

var (
	testnet = nt.NewTestnet()
	// If set to false, the tests will not clean up the docker containers that are started during the tests.
	tearDownAfterTests = true
)

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
		if tearDownAfterTests {
			err := testnet.Teardown()
			Expect(err).ShouldNot(HaveOccurred())
		}
	})

	RegisterFailHandler(Fail)

	config.DefaultReporterConfig.SlowSpecThreshold = time.Hour.Seconds()

	RunSpecs(t, "em-ledger integration tests")
}
