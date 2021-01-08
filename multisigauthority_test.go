// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var _ = Describe("Authority", func() {
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

		It("Same key signs twice", func() {
			time.Sleep(8 * time.Second) // Avoid querying while block height is 1

			jsonPath, err := ioutil.TempDir("", "")
			Expect(err).To(BeNil())
			defer os.RemoveAll(jsonPath)

			authorityaddress := sdk.AccAddress(Authority.GetPublicKey().Address()).String()

			newMinGasPrices, _ := sdk.ParseDecCoins("0.0006eeur")
			tx, err := emcli.CustomCommand("tx", "authority", "set-gas-prices", authorityaddress, newMinGasPrices.String(), "--generate-only", "--from", authorityaddress, "--trust-node")
			Expect(err).To(BeNil())

			transactionPath := fmt.Sprintf("%v/transaction.json", jsonPath)
			ioutil.WriteFile(transactionPath, []byte(tx), 0777)

			sigPaths := make([]string, 0)
			// Sign twice with key1. Signature count is above threshold, but ...
			for i := 0; i < 2; i++ {
				tx, err = emcli.SignTranscation(transactionPath, key1.GetAddress(), authorityaddress)
				signaturePath := fmt.Sprintf("%v/sign%v.json", jsonPath, i)
				ioutil.WriteFile(signaturePath, []byte(tx), 0777)
				sigPaths = append(sigPaths, signaturePath)
			}

			// Combine the two signatures
			tx, err = emcli.CustomCommand("tx", "multisign", transactionPath, "multikey", sigPaths[0], sigPaths[1])
			Expect(err).To(BeNil())
			ioutil.WriteFile(transactionPath, []byte(tx), 0777)

			tx, err = emcli.CustomCommand("tx", "broadcast", transactionPath)
			Expect(err).To(BeNil())
			Expect(gjson.Parse(tx).Get("logs.0.success").Exists()).To(Equal(false))
		})

		It("set global gas prices", func() {
			jsonPath, err := ioutil.TempDir("", "")
			Expect(err).To(BeNil())
			defer os.RemoveAll(jsonPath)

			authorityaddress := sdk.AccAddress(Authority.GetPublicKey().Address()).String()

			newMinGasPrices, _ := sdk.ParseDecCoins("0.0006eeur")
			tx, err := emcli.CustomCommand("tx", "authority", "set-gas-prices", authorityaddress, newMinGasPrices.String(), "--generate-only", "--from", authorityaddress, "--trust-node")
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

			// Manipulate threshold!
			// val := gjson.Parse(tx).Get("value.signatures.0.pub_key.value.threshold").Raw
			// fmt.Println("threshold :", val)

			//{
			//	bz, _ := sjson.SetBytes([]byte(tx), "value.signatures.0.pub_key.value.threshold", "1")
			//	tx = string(bz)
			//}

			Expect(err).To(BeNil())
			ioutil.WriteFile(transactionPath, []byte(tx), 0777)

			fmt.Println("Ready for broadcast:\n", tx)

			tx, err = emcli.CustomCommand("tx", "broadcast", transactionPath)
			fmt.Println("Output:\n", tx)
			Expect(err).To(BeNil())
			Expect(gjson.Parse(tx).Get("logs").Array()).To(Not(BeEmpty()))

			bz, err := emcli.QueryMinGasPrices()
			Expect(err).To(BeNil())

			jsonGP := gjson.GetBytes(bz, "min_gas_prices")
			Expect(jsonGP.IsArray()).To(BeTrue())
			Expect(jsonGP.Array()).To(HaveLen(1))

			jsonGP = jsonGP.Get("0")
			Expect(jsonGP.Get("denom").Str).To(Equal("eeur"))
			amount, err := sdk.NewDecFromStr(jsonGP.Get("amount").Str)
			Expect(err).To(BeNil())
			Expect(amount).To(Equal(sdk.NewDecWithPrec(6, 4)))
		})
	})
})
