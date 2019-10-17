// +build bdd

package emoney

import (
	"fmt"
	"github.com/tidwall/gjson"
	"time"

	nt "emoney/networktest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Staking", func() {
	emcli := nt.NewEmcli(testnet.Keystore)

	Describe("Authority manages issuers", func() {
		Context("", func() {
			It("starts a new testnet", func() {
				awaitReady, err := testnet.Restart()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(awaitReady()).To(BeTrue())
			})

			It("kill validator 2 and get jailed", func() {
				_, err := testnet.KillValidator(2)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(12 * time.Second)

				bz, err := emcli.QueryValidators()
				Expect(err).ToNot(HaveOccurred())

				json := gjson.ParseBytes(bz).Get("#(description.moniker==\"Validator-2\")")
				Expect(json.Get("jailed").Bool()).To(BeTrue())
			})
		})
	})
})
