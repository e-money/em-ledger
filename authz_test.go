// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

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
			Expect(ir.Get("code").String()).To(Equal("4"))
		})

		It("Let Issuee grant auth access for Issuer", func() {
			var _, _, err = emcli.AuthzGrantAuthority(granter, grantee)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Let IssueE sign for IssueR", func() {
			var _, err = emcli.SendOnBehalf(grantee, recipient, granter, "4000ungm")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Let's revoke that grant", func() {
			var _, _, err = emcli.AuthzRevokeAuthority(grantee, granter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Let Issuee sign for Issuer (fail)", func() {
			var msg, err = emcli.SendOnBehalf(grantee, recipient, granter, "4000ungm")
			Expect(err).Should(HaveOccurred())
			ir := gjson.ParseBytes([]byte(msg))
			Expect(ir.Get("code").String()).To(Equal("4"))
		})
	})
})
