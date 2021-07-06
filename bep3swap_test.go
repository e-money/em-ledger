// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/e-money/em-ledger/networktest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

const (
	trxAmount = 5
	denom     = "ungm"
)

var _ = Describe("BEP3 Swap", func() {
	var (
		emcli = testnet.NewEmcli()
		key1  = testnet.Keystore.Key1
		key2  = testnet.Keystore.Key2

		deputy = testnet.Keystore.DeputyKey
	)

	var swapSecret, swapId string

	JustAfterEach(
		func() {
			if CurrentGinkgoTestDescription().Failed {
				fmt.Printf(
					"Failed test in %s\n",
					CurrentGinkgoTestDescription().TestText,
				)
				fmt.Printf(
					"Failed test: %s\n",
					CurrentGinkgoTestDescription().ComponentTexts,
				)
				fmt.Printf(
					"Failed line: %d\n",
					CurrentGinkgoTestDescription().LineNumber,
				)
			}
		},
	)

	Describe("Bep3 Swapping, Claiming, Refunding...", func() {
		It("creates a new testnet", createNewTestnet)

		It("Creates a new incoming swap", func() {
			var (
				height int64
				err    error
			)
			height, err = networktest.GetHeight()
			Expect(err).ToNot(HaveOccurred())

			// deputy has sent 5ungm to key1 on another chain.
			// The deputy creates the swap on the e-money chain.
			secretNumber, randomNumberHash, _, err := emcli.BEP3Create(
				deputy, key1.GetAddress(),
				"0xotherchainrecipient",
				"0xotherchainsender",
				fmt.Sprintf("%d%s", trxAmount, denom),
				600,
			)
			swapSecret = secretNumber
			Expect(err).ToNot(HaveOccurred())
			height = incChainHeight(height)

			list, err := emcli.BEP3ListSwaps()
			fmt.Println(" --- List swaps output\n", list)
			Expect(err).ToNot(HaveOccurred())

			swapList := gjson.Parse(list).Get("swaps.augmented_atomic_swaps")
			Expect(swapList.IsArray()).To(BeTrue())
			Expect(swapList.Array()).NotTo(HaveLen(0))

			swap := swapList.Array()[0]

			swapId = swap.Get("id").Str
			Expect(strings.ToUpper(swap.Get("random_number_hash").Str)).To(Equal(strings.ToUpper(randomNumberHash)))
		},
		)

		It("Uses the wrong secret", func() {
			randomNumber := make([]byte, 32)
			_, err := rand.Read(randomNumber)

			Expect(err).ToNot(HaveOccurred())
			wrongSecret := hex.EncodeToString(randomNumber)

			networktest.IncChain(1)

			output, err := emcli.BEP3Claim(
				key1, swapId, wrongSecret,
			)
			Expect(err).To(HaveOccurred())
			networktest.IncChain(1)

			jsonOutput := gjson.Parse(output)
			Expect(jsonOutput.Get("codespace").Str).To(Equal("bep3"))
			Expect(jsonOutput.Get("code").Int()).To(Equal(int64(13)))
		})

		It("Intended recipient claims the swap", func() {
			var (
				height int64
				err    error
			)
			height, err = networktest.GetHeight()

			// Check state before claiming swap
			totalSupply, err := emcli.QueryTotalSupply()
			Expect(err).ToNot(HaveOccurred())
			ungmSupplyBefore := gjson.ParseBytes(totalSupply).Get(`supply.#(denom=="ungm").amount`).Int()

			ungmBalanceBefore, err := emcli.QueryBalanceDenom(
				key1.GetAddress(), denom,
			)
			Expect(err).ToNot(HaveOccurred())

			// Claim swap
			_, err = emcli.BEP3Claim(key2, swapId, swapSecret)
			Expect(err).ToNot(HaveOccurred())

			height = incChainHeight(height)

			// Check updated state
			totalSupplyAfter, err := emcli.QueryTotalSupply()
			Expect(err).ToNot(HaveOccurred())

			fmt.Println("supply before:", ungmSupplyBefore)
			ungmSupplyAfter := gjson.ParseBytes(totalSupplyAfter).Get(`supply.#(denom=="ungm").amount`).Int()
			fmt.Println("supply after:", ungmSupplyAfter)
			Expect(ungmSupplyAfter).To(Equal(ungmSupplyBefore + trxAmount))

			ungmBalanceAfter, err := emcli.QueryBalanceDenom(
				key1.GetAddress(), denom,
			)
			Expect(err).ToNot(HaveOccurred())

			Expect(ungmBalanceAfter).To(Equal(ungmBalanceBefore + trxAmount))
		})

		It("Allows a swap to expire", func() {
			var (
				height int64
				err    error
			)
			height, err = networktest.GetHeight()

			const swapStatusQuery = "swaps.augmented_atomic_swaps.#(sender_other_chain==\"0x001\").status"
			const swapIdQuery = "swaps.augmented_atomic_swaps.#(sender_other_chain==\"0x001\").id"

			secretNumber, _, _, err := emcli.BEP3Create(
				deputy,
				key1.GetAddress(), "0x002", "0x001",
				fmt.Sprintf("%d%s", trxAmount, denom), 1,
			)
			Expect(err).ToNot(HaveOccurred())
			height = incChainHeight(height)

			list, _ := emcli.BEP3ListSwaps()

			id := gjson.Parse(list).Get(swapIdQuery).Str
			Expect(gjson.Parse(list).Get(swapStatusQuery).Str).To(Equal("Open"))

			fmt.Println("Waiting up to 61 secs for the expiration swap event...")
			listener, err := networktest.NewEventListener()
			Expect(err).ToNot(HaveOccurred())
			err = listener.SubscribeExpirations(id, 61*time.Second)
			Expect(err).ToNot(HaveOccurred())

			// Verify state
			list, _ = emcli.BEP3ListSwaps()
			Expect(gjson.Parse(list).Get(swapStatusQuery).Str).To(Equal("Expired"))

			output, _ := emcli.BEP3Claim(key1, id, secretNumber)
			height = incChainHeight(height)
			jsonOutput := gjson.Parse(output)
			Expect(jsonOutput.Get("codespace").Str).To(Equal("bep3"))
			Expect(jsonOutput.Get("code").Int()).To(Equal(int64(17)))
		})

		It("Creates a new outgoing swap", func() {
			var (
				height int64
				err    error
			)
			height, err = networktest.GetHeight()

			const (
				//swapStatusQuery = "#(sender_other_chain==\"0x0075\").status"
				swapIdQuery = "#(sender_other_chain==\"0x0075\").id"
			)

			// key1 has sent 5000ungm to key2 on another chain. The deputy creates the swap on the e-money chain.
			secretNumber, randomNumberHash, _, err := emcli.BEP3Create(
				key1, deputy.GetAddress(), "0x0050", "0x0075",
				"27000ungm", 120,
			)
			Expect(err).ToNot(HaveOccurred())
			fmt.Println(secretNumber, randomNumberHash)

			height = incChainHeight(height)

			list, _ := emcli.BEP3ListSwaps()
			swapId := gjson.Parse(list).Get(swapIdQuery).Str

			supply, _ := emcli.QueryTotalSupply()
			fmt.Println("Supply before\n", string(supply))

			claim, err := emcli.BEP3Claim(
				deputy, swapId, secretNumber,
			)
			Expect(err).To(HaveOccurred())
			fmt.Println(claim)

			height = incChainHeight(height)

			supply, _ = emcli.QueryTotalSupply()
			fmt.Println("Supply after\n", string(supply))
			// TODO https://github.com/e-money/bep3/issues/1

			swapList := gjson.Parse(list).Get("swaps.augmented_atomic_swaps")
			Expect(swapList.IsArray()).To(BeTrue())
			Expect(swapList.Array()).NotTo(HaveLen(0))
		})
	})
})

func incChainHeight(height int64) int64 {
	// Will not exhaust it in most cases
	var timeOutDur = 8 * time.Second

	newHeight, err := networktest.WaitForHeightWithTimeout(
		height+1, timeOutDur,
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(newHeight > height).To(Equal(true))

	return newHeight
}
