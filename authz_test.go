//go:build bdd

package emoney_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Authz", func() {

	var (
		emcli     = testnet.NewEmcli()
		granter   = testnet.Keystore.Key1
		grantee   = testnet.Keystore.Key2
		recipient = testnet.Keystore.Key3
	)

	Describe("Let's Test This Authz Module", func() {

		It("creates a new testnet", createNewTestnet)

		It("Let Issuee sign for Issuer (fail)", func() {
			var msg, err = emcli.SendOnBehalf(grantee, recipient, granter, "4000ungm")
			Expect(err).Should(HaveOccurred())
			ir := gjson.ParseBytes([]byte(msg))
			// error 4 = unauthorized
			Expect(ir.Get("code").String()).To(Equal("4"))
		})

		It("Let Issuee grant auth access for Issuer", func() {
			var _, _, err = emcli.AuthzGrantAuthority(granter, grantee, "1000000ungm,1000eeur")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Let IssueE sign for IssueR (NGM)", func() {
			var _, err = emcli.SendOnBehalf(grantee, recipient, granter, "1000000ungm")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Let IssueE sign for IssueR (EEUR)", func() {
			var _, err = emcli.SendOnBehalf(grantee, recipient, granter, "999eeur")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Let IssueE sign for IssueR (spendLimit exceeded - should fail)", func() {
			var msg, err = emcli.SendOnBehalf(grantee, recipient, granter, "1000001ungm")
			Expect(err).Should(HaveOccurred())
			ir := gjson.ParseBytes([]byte(msg))
			// error 5 = insufficient funds
			Expect(ir.Get("code").String()).To(Equal("5"))
		})

		It("Let's revoke that grant", func() {
			var _, _, err = emcli.AuthzRevokeAuthority(grantee, granter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Let Issuee sign for Issuer (NGM - should fail)", func() {
			var msg, err = emcli.SendOnBehalf(grantee, recipient, granter, "1000ungm")
			Expect(err).Should(HaveOccurred())
			ir := gjson.ParseBytes([]byte(msg))
			// error 4 = unauthorized
			Expect(ir.Get("code").String()).To(Equal("4"))
		})

		It("Let Issuee sign for Issuer (EEUR - should fail)", func() {
			var msg, err = emcli.SendOnBehalf(grantee, recipient, granter, "1000eeur")
			Expect(err).Should(HaveOccurred())
			ir := gjson.ParseBytes([]byte(msg))
			// error 4 = unauthorized
			Expect(ir.Get("code").String()).To(Equal("4"))
		})
	})
})
