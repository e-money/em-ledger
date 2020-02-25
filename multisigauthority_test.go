// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io/ioutil"
	"os"
)

var _ = Describe("Market", func() {

	var (
		keystore  = testnet.Keystore
		emcli     = testnet.NewEmcli()
		key1      = testnet.Keystore.Key1
		key2      = testnet.Keystore.Key2
		Authority = keystore.MultiKey
	)

	Describe("Authority is a multisig account", func() {
		It("starts a new testnet", func() {
			awaitReady, err := testnet.RestartWithModifications(
				func(bz []byte) []byte {
					bz, _ = sjson.SetBytes(bz, "app_state.authority.key", Authority.GetAddress())

					return bz
				})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(awaitReady()).To(BeTrue())
		})

		It("set global gas prices", func() {
			jsonPath, err := ioutil.TempDir("", "")
			Expect(err).To(BeNil())
			defer os.RemoveAll(jsonPath)

			authorityaddress := sdk.AccAddress(Authority.GetPublicKey().Address()).String()

			newMinGasPrices, _ := sdk.ParseDecCoins("0.0006eeur")
			tx, err := emcli.CustomCommand("tx", "authority", "set-gas-prices", authorityaddress, newMinGasPrices.String(), "--generate-only", "--from", authorityaddress)
			Expect(err).To(BeNil())

			transactionPath := fmt.Sprintf("%v/transaction.json", jsonPath)
			ioutil.WriteFile(transactionPath, []byte(tx), 0777)

			tx, err = emcli.SignTranscation(transactionPath, key1.GetAddress(), authorityaddress)
			signature1Path := fmt.Sprintf("%v/sign1.json", jsonPath)
			ioutil.WriteFile(signature1Path, []byte(tx), 0777)

			tx, err = emcli.SignTranscation(transactionPath, key2.GetAddress(), authorityaddress)
			Expect(err).To(BeNil())
			signature2Path := fmt.Sprintf("%v/sign2.json", jsonPath)
			ioutil.WriteFile(signature2Path, []byte(tx), 0777)

			// Combine the two signatures
			tx, err = emcli.CustomCommand("tx", "multisign", transactionPath, "multikey", signature1Path, signature2Path)
			Expect(err).To(BeNil())
			ioutil.WriteFile(transactionPath, []byte(tx), 0777)

			tx, err = emcli.CustomCommand("tx", "broadcast", transactionPath)
			Expect(err).To(BeNil())
			Expect(gjson.Parse(tx).Get("logs.0.success").Type).To(Equal(gjson.True))

			bz, err := emcli.QueryMinGasPrices()
			Expect(err).To(BeNil())
			minGasPricesStr := gjson.GetBytes(bz, "min_gas_prices").Str
			minGasPrices, err := sdk.ParseDecCoins(minGasPricesStr)
			Expect(err).To(BeNil())
			Expect(minGasPrices).To(Equal(newMinGasPrices))
		})
	})
})
