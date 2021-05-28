// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney_test

import (
	nt "github.com/e-money/em-ledger/networktest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/sjson"
	"time"
)

var _ = Describe("Staking", func() {
	const (
		QueryNgmRewards            = "total.#(denom==\"ungm\")"
		QueryJailedValidatorsCount = "validators.#(jailed==true)#"
	)

	Describe("Authority manages issuers", func() {
		Context("", func() {
			emcli := testnet.NewEmcli()

			var (
				Validator0Key = testnet.Keystore.Validators[0]
				Validator2Key = testnet.Keystore.Validators[2]
			)

			It("creates a new testnet", func() {
				awaitReady, err := testnet.RestartWithModifications(
					// increase slash fraction so that amount is > 0
					func(bz []byte) []byte {
						bz, _ = sjson.SetBytes(bz, "app_state.slashing.params.slash_fraction_downtime", "0.100000000000000000")
						return bz
					})
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

				Expect(slash()).ToNot(BeNil())

				// wait 2 blocks
				nt.IncChain(2)

				rewardsJson, err := emcli.QueryRewards(Validator0Key.GetAddress())
				Expect(err).ToNot(HaveOccurred())
				Expect(rewardsJson.Get(QueryNgmRewards).Raw).ToNot(BeEmpty())

				validators, err := emcli.QueryValidators()
				Expect(err).ToNot(HaveOccurred())
				validators = validators.Get(QueryJailedValidatorsCount)
				Expect(validators.Array()).To(HaveLen(1))
			})

			It("validator unjails", func() {
				validators, err := emcli.QueryValidators()
				Expect(err).ToNot(HaveOccurred())
				validators = validators.Get(QueryJailedValidatorsCount)
				Expect(validators.Array()).To(HaveLen(1))

				_, success, err := emcli.UnjailValidator(Validator2Key.GetAddress())
				Expect(success).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				validators, err = emcli.QueryValidators()
				Expect(err).ToNot(HaveOccurred())
				validators = validators.Get(QueryJailedValidatorsCount)
				Expect(validators.Array()).To(BeEmpty())
			})
		})
	})
})
