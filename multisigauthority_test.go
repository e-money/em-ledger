// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney_test

import (
	"fmt"
	"io/ioutil"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	nt "github.com/e-money/em-ledger/networktest"
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
			_, _ = nt.IncChain(1) // Avoid querying while block height is 1

			jsonPath, err := ioutil.TempDir("", "")
			Expect(err).To(BeNil())
			defer os.RemoveAll(jsonPath)

			authorityaddress := sdk.AccAddress(Authority.GetPublicKey().Address()).String()

			newMinGasPrices, _ := sdk.ParseDecCoins("0.0006eeur")
			tx, err := emcli.CustomCommand("tx", "authority", "set-gas-prices", authorityaddress, newMinGasPrices.String(), "--generate-only", "--from", authorityaddress)
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

		It("Replace Authority", func() {
			// authority switches from multisig to singlesig
			authorityAddress := sdk.AccAddress(Authority.GetPublicKey().Address()).String()
			newAuthorityAddress := sdk.AccAddress(keystore.Authority.GetPublicKey().Address()).String()

			execAuthMSigTx(authorityAddress, []nt.Key{key1, key2}, "tx", "authority", "replace", authorityAddress,
				newAuthorityAddress, "--generate-only", "--from",
				authorityAddress)

			// create/revoke issuer with the new authority
			ok := nt.AuthCreatesIssuer(emcli, keystore.Authority, key1)
			Expect(ok).To(BeTrue())
			_, success, err := emcli.AuthorityDestroyIssuer(keystore.Authority, key1)
			Expect(success).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())

			// authority switches from singlesig back to original multisig
			authorityAddress, newAuthorityAddress = newAuthorityAddress, authorityAddress

			_, err = emcli.CustomCommand(
				"tx", "authority", "replace", authorityAddress,
				newAuthorityAddress, "--from", authorityAddress,
			)
			Expect(err).To(BeNil())

			// create/revoke issuer with both the former and the new authority
			// former singlesig first
			ok = nt.AuthCreatesIssuer(emcli, keystore.Authority, key1)
			Expect(ok).To(BeTrue())
			_, success, err = emcli.AuthorityDestroyIssuer(keystore.Authority, key1)
			Expect(success).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())

			// current multisig authority now
			// create-issuer
			execAuthMSigTx(
				newAuthorityAddress, []nt.Key{key1, key2}, "tx", "authority",
				"create-issuer", newAuthorityAddress,
				key1.GetAddress(), "eeur,ejpy", "--generate-only", "--from",
				newAuthorityAddress,
			)

			// destroy-issuer with current multisig authority
			execAuthMSigTx(
				newAuthorityAddress, []nt.Key{key1, key2}, "tx", "authority",
				"destroy-issuer", newAuthorityAddress,
				key1.GetAddress(), "--generate-only", "--from",
				newAuthorityAddress,
			)

			// try with a non-authority account
			ok = nt.AuthCreatesIssuer(emcli, keystore.Key2, key1)
			Expect(ok).To(BeFalse())
		})

		It("set global gas prices", func() {
			// former authority is still in effect
			authorityAddress := sdk.AccAddress(Authority.GetPublicKey().Address()).String()

			newMinGasPrices, _ := sdk.ParseDecCoins("0.0006eeur")

			execAuthMSigTx(authorityAddress, []nt.Key{key1, key2}, "tx", "authority", "set-gas-prices", authorityAddress,
				newMinGasPrices.String(), "--generate-only", "--from", authorityAddress)

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

// execAuthMSigTx signs and broadcasts a multisig trx using the emcli.CustomCommand.
func execAuthMSigTx(authorityAddress string, keys []nt.Key, cmdArgs ...string) {
	var (
		emcli = testnet.NewEmcli()
	)

	jsonPath, err := ioutil.TempDir("", "")
	Expect(err).To(BeNil())
	defer os.RemoveAll(jsonPath)

	cmdTx, err := emcli.CustomCommand(cmdArgs...)
	Expect(err).To(BeNil())

	transactionPath := fmt.Sprintf("%v/transaction.json", jsonPath)
	err = ioutil.WriteFile(transactionPath, []byte(cmdTx), 0777)
	Expect(err).To(BeNil())

	tx, err := emcli.SignTranscation(
		transactionPath, keys[0].GetAddress(), authorityAddress,
	)
	signaturePath1 := fmt.Sprintf("%v/sign%v.json", jsonPath, 0)
	err = ioutil.WriteFile(signaturePath1, []byte(tx), 0777)
	Expect(err).To(BeNil())
	tx, err = emcli.SignTranscation(
		transactionPath, keys[1].GetAddress(), authorityAddress,
	)
	signaturePath2 := fmt.Sprintf("%v/sign%v.json", jsonPath, 1)
	err = ioutil.WriteFile(signaturePath2, []byte(tx), 0777)
	Expect(err).To(BeNil())

	// Combine the two signatures
	tx, err = emcli.CustomCommand(
		"tx", "multisign", transactionPath, "multikey", signaturePath1,
		signaturePath2,
	)
	Expect(err).To(BeNil())
	err = ioutil.WriteFile(transactionPath, []byte(tx), 0777)
	Expect(err).To(BeNil())

	tx, err = emcli.CustomCommand("tx", "broadcast", transactionPath)
	Expect(err).To(BeNil())
	Expect(gjson.Parse(tx).Get("logs.0.success").Exists()).To(Equal(false))
}
