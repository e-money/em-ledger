// +build bdd

package emoney

import (
	"time"

	nt "emoney/networktest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Staking", func() {
	const QueryNgmRewards = "total.#(denom==\"x3ngm\")"

	Describe("Authority manages issuers", func() {
		Context("", func() {
			emcli := nt.NewEmcli(testnet.Keystore)

			var (
				Validator0Key = testnet.Keystore.Validators[0]
				Validator2Key = testnet.Keystore.Validators[2]
			)

			It("creates a new testnet", createNewTestnet)

			It("kill validator 2 and get jailed", func() {
				listener, err := nt.NewEventListener()
				if err != nil {
					panic(err)
				}

				// Allow for a few blocks
				time.Sleep(5 * time.Second)

				rewardsJson, err := emcli.QueryRewards(Validator0Key.GetAddress())
				Expect(err).ToNot(HaveOccurred())
				Expect(rewardsJson.Get(QueryNgmRewards).Raw).To(BeEmpty())

				slash, err := listener.AwaitSlash()
				Expect(err).ToNot(HaveOccurred())

				payoutEvent, err := listener.AwaitPenaltyPayout()
				Expect(err).ToNot(HaveOccurred())

				_, err = testnet.KillValidator(2)
				Expect(err).ToNot(HaveOccurred())

				Expect(slash()).ToNot(BeNil())
				Expect(payoutEvent()).To(BeTrue())

				time.Sleep(30 * time.Second)

				rewardsJson, err = emcli.QueryRewards(Validator0Key.GetAddress())
				Expect(err).ToNot(HaveOccurred())
				Expect(rewardsJson.Get(QueryNgmRewards).Raw).ToNot(BeEmpty())

				// Ensure that the jailed validator
				rewardsJson, err = emcli.QueryRewards(Validator2Key.GetAddress())
				Expect(err).ToNot(HaveOccurred())
				Expect(rewardsJson.Get(QueryNgmRewards).Raw).To(BeEmpty())
			})
		})
	})
})
