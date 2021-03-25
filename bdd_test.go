// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney_test

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	nt "github.com/e-money/em-ledger/networktest"
	apptypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/inflation"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
)

var (
	testnet = func() nt.Testnet {
		version.Name = "e-money" // Used by the keyring library.
		version.AppName = "emd"

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

func setGenesisTime(genesis []byte, genesisTime time.Time) []byte {
	bz, err := sjson.SetBytes(genesis, "genesis_time", genesisTime.Format(time.RFC3339))
	if err != nil {
		panic(err)
	}

	return bz
}

func setInflation(genesis []byte, denom string, newInflation sdk.Dec) []byte {
	inflationJsonGen := gjson.GetBytes(genesis, "app_state.inflation")
	inflationGen := new(inflation.GenesisState)
	err := json.Unmarshal([]byte(inflationJsonGen.String()), inflationGen)
	if err != nil {
		panic(err)
	}

	for index, asset := range inflationGen.InflationState.InflationAssets {
		if asset.Denom == denom {
			inflationGen.InflationState.InflationAssets[index].Inflation = newInflation
		}
	}

	updatedInflation, err := json.Marshal(inflationGen)
	if err != nil {
		panic(err)
	}

	bz, err := sjson.SetRawBytes(genesis, "app_state.inflation", updatedInflation)
	if err != nil {
		panic(err)
	}

	return bz
}
