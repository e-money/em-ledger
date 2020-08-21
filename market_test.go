// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney

import (
	"fmt"
	market "github.com/e-money/em-ledger/x/market/types"
	"io/ioutil"
	"os"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"

	emoney "github.com/e-money/em-ledger" // To get around this issue: https://stackoverflow.com/q/14723229
	"github.com/e-money/em-ledger/networktest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"

	tmrand "github.com/tendermint/tendermint/libs/rand"
)

var _ = Describe("Market", func() {
	emcli := testnet.NewEmcli()

	var (
		acc1 = testnet.Keystore.Key1
		acc2 = testnet.Keystore.Key2
		acc3 = testnet.Keystore.Key3

		Authority = testnet.Keystore.Authority
	)

	Describe("Authority manages issuers", func() {
		It("creates a new testnet", createNewTestnet)

		It("Basic creation of simple orders", func() {
			time.Sleep(5 * time.Second)
			//bz, err := emcli.QueryAccountJson(acc1.GetAddress())
			//fmt.Println(string(bz))
			//Expect(err).ShouldNot(HaveOccurred())

			for i := 0; i < 10; i++ {
				output, success, err := emcli.MarketAddOrder(acc1, "120000eeur", fmt.Sprintf("%dechf", 90000-i*100), tmrand.Str(10))
				Expect(err).ToNot(HaveOccurred(), "Error output: %v", output)
				Expect(success).To(BeTrue())
			}

			bz, err := emcli.QueryMarketInstrument("eeur", "echf")
			Expect(err).ToNot(HaveOccurred())
			ir := gjson.ParseBytes(bz)
			Expect(ir.Get("orders").Array()).To(HaveLen(10))
		})

		It("Crashing validator can catch up", func() {
			_, success, err := emcli.MarketAddOrder(acc2, "5000eeur", "100000ejpy", "acc2cid1")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			// Kill a validator
			_, err = testnet.KillValidator(2)
			Expect(err).ToNot(HaveOccurred())

			// Create and execute a couple of orders
			for i := 0; i < 7; i++ {
				_, success, err := emcli.MarketAddOrder(acc2, "90500echf", fmt.Sprintf("%deeur", 11000-i*400), tmrand.Str(10))
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			}

			for i := 0; i < 15; i++ {
				_, success, err := emcli.MarketAddOrder(acc3, "440000ejpy", fmt.Sprintf("%deeur", 100000-i*100), tmrand.Str(10))
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			}

			_, success, err = emcli.MarketCancelOrder(acc2, "acc2cid1")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			//bz, err := emcli.QueryMarketInstruments()
			//Expect(err).ToNot(HaveOccurred())
			//fmt.Println("Order book:\n", string(bz))

			time.Sleep(4 * time.Second)

			_, err = testnet.ResurrectValidator(2)
			Expect(err).ToNot(HaveOccurred())

			for i := 0; i < 10; i++ {
				_, success, err := emcli.MarketAddOrder(acc3, "90000eeur", fmt.Sprintf("%dechf", 50000-i*100), tmrand.Str(10))
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			}

			// Wait and while and attempt to discover consensus failure in the logs of the resurrected validator
			time.Sleep(8 * time.Second)

			log, err := testnet.GetValidatorLogs(2)
			Expect(err).ToNot(HaveOccurred())
			if strings.Contains(log, "Wrong Block.Header.AppHash") ||
				strings.Contains(log, "panic") {
				Fail(fmt.Sprintf("Validator 2 does not appear to have re-established consensus:\n%v", log))
			}
		})

		// Create a vanilla testnet to reset market state
		It("creates a new testnet", createNewTestnet)

		It("Runs out of gas while using the market", func() {
			time.Sleep(5 * time.Second)

			prices, err := sdk.ParseDecCoins("0.00005eeur")
			Expect(err).ToNot(HaveOccurred())

			_, success, err := emcli.AuthoritySetMinGasPrices(Authority, prices.String())
			Expect(success).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())

			s, success, err := emcli.MarketAddOrder(acc2, "5000eeur", "100000ejpy", "acc2cid1", "--fees", "50eeur")

			// Create one transaction that includes a market order and a lot transfers, which will make the tx run out of gas.
			jsonPath, err := ioutil.TempDir("", "")
			Expect(err).To(BeNil())
			defer os.RemoveAll(jsonPath)

			msgs := make([]sdk.Msg, 0)

			addr3, err := sdk.AccAddressFromBech32(acc3.GetAddress())
			if err != nil {
				panic(err)
			}

			clientOrderId := "ShouldNotBePresent"
			addOrder := market.MsgAddLimitOrder{
				Owner:         addr3,
				Source:        sdk.NewCoin("echf", sdk.NewInt(50000)),
				Destination:   sdk.NewCoin("eeur", sdk.NewInt(60000)),
				ClientOrderId: clientOrderId,
			}

			msgs = append(msgs, addOrder)

			// Add a few transfers to make sure that gas is exhausted
			addr2, err := sdk.AccAddressFromBech32(acc2.GetAddress())
			if err != nil {
				panic(err)
			}

			coins := sdk.NewCoins(sdk.NewCoin("eeur", sdk.NewInt(5000)))
			for i := 0; i < 5; i++ {
				msgs = append(msgs, bank.NewMsgSend(addr3, addr2, coins))
			}

			accountJson, err := emcli.QueryAccountJson(acc3.GetAddress())
			Expect(err).To(BeNil())
			accNum := gjson.ParseBytes(accountJson).Get("value.account_number").Uint()
			accSeq := gjson.ParseBytes(accountJson).Get("value.sequence").Uint()

			tx := networktest.CreateMultiMsgTx(acc3, testnet.ChainID(), "500eeur", accNum, accSeq, msgs...)

			cdc := emoney.MakeCodec()
			json := cdc.MustMarshalJSON(tx)

			transactionPath := fmt.Sprintf("%v/tx.json", jsonPath)
			ioutil.WriteFile(transactionPath, json, 0777)

			s, err = emcli.CustomCommand("tx", "broadcast", transactionPath)
			Expect(err).To(BeNil())
			// Transaction must have failed due to insufficient gas
			Expect(gjson.Parse(s).Get("logs.0.success").Exists()).To(Equal(false))

			// Order must not be available for querying
			bz, err := emcli.QueryMarketByAccount(addr3.String())
			Expect(err).To(BeNil())

			query := fmt.Sprintf("orders.#(client_order_id==\"%v\")#", clientOrderId)
			Expect(gjson.ParseBytes(bz).Get(query).Array()).To(BeEmpty())
		})
	})
})
