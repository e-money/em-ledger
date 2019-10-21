// +build bdd

package emoney

import (
	"time"

	nt "emoney/networktest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Staking", func() {
	Describe("Authority manages issuers", func() {
		Context("", func() {
			It("starts a new testnet", func() {
				awaitReady, err := testnet.Restart()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(awaitReady()).To(BeTrue())
			})

			It("kill validator 2 and get jailed", func() {
				listener, err := nt.NewEventListener()
				if err != nil {
					panic(err)
				}

				// Allow for a few blocks
				time.Sleep(5 * time.Second)

				slash, err := listener.AwaitSlash()
				Expect(err).ToNot(HaveOccurred())

				_, err = testnet.KillValidator(2)
				Expect(err).ToNot(HaveOccurred())

				slashevent := slash()
				Expect(slashevent).ToNot(BeNil())

				time.Sleep(5 * time.Second)
			})
		})
	})
})
