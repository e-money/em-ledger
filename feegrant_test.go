//go:build bdd

package emoney_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strconv"
)

var _ = Describe("FeeGrant", func() {

	var (
		emcli          = testnet.NewEmcli()
		authority      = testnet.Keystore.Authority
		granter        = testnet.Keystore.Key1
		grantee        = testnet.Keystore.Key2
		reciever       = testnet.Keystore.Key3
		initialBalance = 0 // toBeOverridden
		sendValue      = 2500000
		spendLimit     = 1000000
		feeAmount      = 5000
		totalGasSpent  = 0
		denom          = "ungm"
	)

	Describe("Let's Test This FeeGrant Module", func() {

		It("creates a new testnet", createNewTestnet)

		It("Let's get the initial funds value", func() {
			var err error
			initialBalance, err = emcli.QueryBalanceDenom(granter.GetAddress(), denom)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Let's set gasprice and confirm", func() {
			var prices, err = sdk.ParseDecCoins("0.00005" + denom)
			Expect(err).ToNot(HaveOccurred())

			_, success, err := emcli.AuthoritySetMinGasPrices(authority, prices.String())
			Expect(success).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())

			_, err = emcli.QueryMinGasPrices()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Let's check some initial balances", func() {
			var granterBalance, err = emcli.QueryBalanceDenom(granter.GetAddress(), denom)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(granterBalance).To(Equal(initialBalance))

			var granteeBalance, err2 = emcli.QueryBalanceDenom(grantee.GetAddress(), denom)
			Expect(err2).ShouldNot(HaveOccurred())
			Expect(granteeBalance).To(Equal(initialBalance))
		})

		It("Let's make a grant", func() {
			_, _, err := emcli.FeegrantGrant(granter, grantee, strconv.Itoa(spendLimit)+denom, strconv.Itoa(feeAmount)+denom)
			totalGasSpent += feeAmount
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Let's check that the grant is there", func() {
			var msg, err = emcli.FeegrantQuery(grantee)
			Expect(err).ShouldNot(HaveOccurred())
			ir := gjson.ParseBytes(msg)
			Expect(len(ir.Get("allowances").Array())).To(Equal(1))
		})

		It("Let's check some balances", func() {
			var granterBalance, err = emcli.QueryBalanceDenom(granter.GetAddress(), denom)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(granterBalance).To(Equal(initialBalance - feeAmount))

			granteeBalance, err := emcli.QueryBalanceDenom(grantee.GetAddress(), denom)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(granteeBalance).To(Equal(initialBalance))
		})

		It("Let grantee send message and let granter pay fee", func() {
			var msg, err = emcli.SendGrantfee(grantee, reciever, granter, strconv.Itoa(sendValue)+denom, strconv.Itoa(feeAmount)+denom)
			totalGasSpent += feeAmount
			ir := gjson.ParseBytes([]byte(msg))
			code := ir.Get("code").Int()
			Expect(int(code)).To(Equal(0))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Let's check some balances after send", func() {
			var granteeBalance, err = emcli.QueryBalanceDenom(grantee.GetAddress(), denom)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(granteeBalance).To(Equal(initialBalance - sendValue))

			granterBalance, err := emcli.QueryBalanceDenom(granter.GetAddress(), denom)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(granterBalance).To(Equal(initialBalance - totalGasSpent))
		})

		It("Let's revoke that grant", func() {
			var _, _, err = emcli.FeegrantRevoke(granter, grantee, strconv.Itoa(feeAmount)+denom)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Let's check that the grant is gone", func() {
			var msg, err = emcli.FeegrantQuery(grantee)
			Expect(err).ShouldNot(HaveOccurred())
			ir := gjson.ParseBytes(msg)
			Expect(len(ir.Get("allowances").Array())).To(Equal(0))
		})
	})
})
