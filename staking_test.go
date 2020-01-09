// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney

import (
	"time"

	nt "github.com/e-money/em-ledger/networktest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Staking", func() {
	const (
		QueryNgmRewards            = "total.#(denom==\"ungm\")"
		QueryJailedValidatorsCount = "#(jailed==true)#"
	)

	Describe("Authority manages issuers", func() {
		Context("", func() {
			emcli := testnet.NewEmcli()

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

				rewardsJson, err = emcli.QueryRewards(Validator0Key.GetAddress())
				Expect(err).ToNot(HaveOccurred())
				Expect(rewardsJson.Get(QueryNgmRewards).Raw).ToNot(BeEmpty())

				// Ensure that the jailed validator does not get any of the fine.
				rewardsJson, err = emcli.QueryRewards(Validator2Key.GetAddress())
				Expect(err).ToNot(HaveOccurred())
				Expect(rewardsJson.Get(QueryNgmRewards).Raw).To(BeEmpty())

				validators, err := emcli.QueryValidators()
				Expect(err).ToNot(HaveOccurred())
				validators = validators.Get(QueryJailedValidatorsCount)
				Expect(validators.Array()).To(HaveLen(1))
			})

			It("validator unjails", func() {
				_, success, err := emcli.UnjailValidator(Validator2Key)
				Expect(success).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				validators, err := emcli.QueryValidators()
				Expect(err).ToNot(HaveOccurred())
				validators = validators.Get(QueryJailedValidatorsCount)
				Expect(validators.Array()).To(BeEmpty())
			})
		})
	})
})
