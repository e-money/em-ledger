// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("BEP3 Swap", func() {
	var (
		emcli = testnet.NewEmcli()
		key1  = testnet.Keystore.Key1
		key2  = testnet.Keystore.Key2

		deputy = testnet.Keystore.DeputyKey
	)

	var swapSecret, swapId string

	It("creates a new testnet", createNewTestnet)

	It("Creates a new incoming swap", func() {
		time.Sleep(5 * time.Second)

		// key1 has sent 5000ungm to key2 on another chain. The deputy creates the swap on the e-money chain.
		secretNumber, randomNumberHash, _, err := emcli.BEP3Create(deputy, key1.GetAddress(), "0xotherchainrecipient", "0xotherchainsender", "5000ungm", 5)
		swapSecret = secretNumber

		Expect(err).ToNot(HaveOccurred())

		list, err := emcli.BEP3ListSwaps()
		fmt.Println(" --- List swaps output\n", list)
		Expect(err).ToNot(HaveOccurred())

		swapList := gjson.Parse(list)
		Expect(swapList.IsArray()).To(BeTrue())
		Expect(swapList.Array()).To(HaveLen(1))

		swap := swapList.Array()[0]

		swapId = swap.Get("id").Str
		Expect(strings.ToUpper(swap.Get("random_number_hash").Str)).To(Equal(strings.ToUpper(randomNumberHash)))
	})

	It("Uses the wrong secret", func() {
		randomNumber := make([]byte, 32)
		_, err := rand.Read(randomNumber)
		Expect(err).ToNot(HaveOccurred())
		wrongSecret := hex.EncodeToString(randomNumber)

		output, err := emcli.BEP3Claim(key1, swapId, wrongSecret)
		Expect(err).ToNot(HaveOccurred())

		jsonOutput := gjson.Parse(output)
		Expect(jsonOutput.Get("codespace").Str).To(Equal("bep3"))
		Expect(jsonOutput.Get("code").Int()).To(Equal(int64(13)))

	})

	It("Intended recpient claims the swap", func() {
		// Check state before claiming swap
		totalSupply, err := emcli.QueryTotalSupply()
		Expect(err).ToNot(HaveOccurred())
		ungmSupplyBefore := gjson.ParseBytes(totalSupply).Get("#(denom==\"ungm\").amount").Int()

		accountBalance, err := emcli.QueryAccountJson(key1.GetAddress())
		Expect(err).ToNot(HaveOccurred())
		ungmBalanceBefore := gjson.ParseBytes(accountBalance).Get("value.coins.#(denom==\"ungm\").amount").Int()

		// Claim swap
		_, err = emcli.BEP3Claim(key2, swapId, swapSecret)
		Expect(err).ToNot(HaveOccurred())

		// Check updated state

		totalSupply, err = emcli.QueryTotalSupply()
		Expect(err).ToNot(HaveOccurred())
		ungmSupplyAfter := gjson.ParseBytes(totalSupply).Get("#(denom==\"ungm\").amount").Int()
		Expect(ungmSupplyAfter).To(Equal(ungmSupplyBefore + 5000))

		accountBalance, err = emcli.QueryAccountJson(key1.GetAddress())
		Expect(err).ToNot(HaveOccurred())
		ungmBalanceAfter := gjson.ParseBytes(accountBalance).Get("value.coins.#(denom==\"ungm\").amount").Int()

		Expect(ungmBalanceAfter).To(Equal(ungmBalanceBefore + 5000))
	})

	It("Allows a swap to expire", func() {
		const (
			swapStatusQuery = "#(sender_other_chain==\"0x001\").status"
			swapIdQuery     = "#(sender_other_chain==\"0x001\").id"
		)

		secretNumber, _, _, err := emcli.BEP3Create(deputy, key1.GetAddress(), "0x002", "0x001", "1000ungm", 5)
		Expect(err).ToNot(HaveOccurred())

		list, _ := emcli.BEP3ListSwaps()

		id := gjson.Parse(list).Get(swapIdQuery).Str
		Expect(gjson.Parse(list).Get(swapStatusQuery).Str).To(Equal("Open"))

		time.Sleep(7 * time.Second) // Swap expires after 5 seconds

		// Verify state
		list, _ = emcli.BEP3ListSwaps()
		Expect(gjson.Parse(list).Get(swapStatusQuery).Str).To(Equal("Expired"))

		output, _ := emcli.BEP3Claim(key1, id, secretNumber)
		jsonOutput := gjson.Parse(output)
		Expect(jsonOutput.Get("codespace").Str).To(Equal("bep3"))
		Expect(jsonOutput.Get("code").Int()).To(Equal(int64(17)))
	})

	It("Creates a new outgoing swap", func() {
		const (
			//swapStatusQuery = "#(sender_other_chain==\"0x0075\").status"
			swapIdQuery = "#(sender_other_chain==\"0x0075\").id"
		)

		// key1 has sent 5000ungm to key2 on another chain. The deputy creates the swap on the e-money chain.
		secretNumber, randomNumberHash, _, err := emcli.BEP3Create(key1, deputy.GetAddress(), "0x0050", "0x0075", "27000ungm", 120)
		fmt.Println(err)
		fmt.Println(secretNumber, randomNumberHash)

		list, _ := emcli.BEP3ListSwaps()
		swapId := gjson.Parse(list).Get(swapIdQuery).Str

		supply, _ := emcli.QueryTotalSupply()
		fmt.Println("Supply before\n", string(supply))

		claim, err := emcli.BEP3Claim(deputy, swapId, secretNumber)
		fmt.Println(err)
		fmt.Println(claim)

		supply, _ = emcli.QueryTotalSupply()
		fmt.Println("Supply after\n", string(supply))
		// TODO https://github.com/e-money/bep3/issues/1
	})
})
