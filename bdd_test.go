// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/version"

	apptypes "github.com/e-money/em-ledger/types"

	nt "github.com/e-money/em-ledger/networktest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"
)

var (
	testnet = func() nt.Testnet {
		version.Name = "e-money" // Used by the keyring library.
		version.ClientName = "emcli"
		version.ServerName = "emd"

		apptypes.ConfigureSDK()
		return nt.NewTestnet()
	}()

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
