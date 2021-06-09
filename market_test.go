// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	emoney "github.com/e-money/em-ledger"
	"github.com/e-money/em-ledger/networktest"
	market "github.com/e-money/em-ledger/x/market/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"time"
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
			//bz, err := emcli.QueryAccountJson(acc1.GetAddress())
			//fmt.Println(string(bz))
			//Expect(err).ShouldNot(HaveOccurred())

			for i := 0; i < 10; i++ {
				output, success, err := emcli.MarketAddLimitOrder(acc1, "120000eeur", fmt.Sprintf("%dechf", 90000-i*100), tmrand.Str(10))
				Expect(err).ToNot(HaveOccurred(), "Error output: %v", output)
				Expect(success).To(BeTrue())
			}

			bz, err := emcli.QueryMarketInstrument("eeur", "echf")
			Expect(err).ToNot(HaveOccurred())
			ir := gjson.ParseBytes(bz)
			Expect(ir.Get("orders").Array()).To(HaveLen(10))
		})

		It("Crashing validator can catch up", func() {
			var (
				err    error
			)
			_, success, err := emcli.MarketAddLimitOrder(acc2, "5000eeur", "100000ejpy", "acc2cid1")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			// Kill a validator
			_, err = testnet.KillValidator(2)
			Expect(err).ToNot(HaveOccurred())

			// Create and execute a couple of orders
			for i := 0; i < 7; i++ {
				_, success, err := emcli.MarketAddLimitOrder(acc2, "90500echf", fmt.Sprintf("%deeur", 11000-i*400), tmrand.Str(10))
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			}

			for i := 0; i < 15; i++ {
				_, success, err := emcli.MarketAddLimitOrder(acc3, "440000ejpy", fmt.Sprintf("%deeur", 100000-i*100), tmrand.Str(10))
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			}

			_, success, err = emcli.MarketCancelOrder(acc2, "acc2cid1")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			//bz, err := emcli.QueryMarketInstruments()
			//Expect(err).ToNot(HaveOccurred())
			//fmt.Println("Order book:\n", string(bz))

			_, err = networktest.IncChain(2)
			Expect(err).ToNot(HaveOccurred())

			_, err = testnet.ResurrectValidator(2)
			Expect(err).ToNot(HaveOccurred())

			for i := 0; i < 10; i++ {
				_, success, err := emcli.MarketAddLimitOrder(acc3, "90000eeur", fmt.Sprintf("%dechf", 50000-i*100), tmrand.Str(10))
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			}

			// Among the market transactions, send a transaction that we will
			// search its committed hash from the revitalized node 2.
			txHashes       := make(map[string]bool)
			coin, _ := sdk.ParseCoinsNormalized("15000eeur")
			txHash, err := testnet.SendTx(
				testnet.Keystore.Key1, testnet.Keystore.Key2, coin,
				testnet.ChainID(),
			)
			Expect(err).ToNot(HaveOccurred())
			txHashes[txHash] = true

			node2Lic, err := networktest.NewEventListenerNode(2)
			Expect(err).ToNot(HaveOccurred())

			var (
				foundCnt int32
				expiration = 20*time.Second
			)
			Eventually(func() int {
				// find the trx hash among the emitted events
				foundCnt, err = node2Lic.SubTx(
					nil, txHashes, 1, expiration,
				)
				return int(foundCnt)
			}, expiration, time.Second).Should(Equal(1))
			Expect(err).ToNot(HaveOccurred())
		})

		// Create a vanilla testnet to reset market state
		It("creates a new testnet", createNewTestnet)

		It("Runs out of gas while using the market", func() {
			prices, err := sdk.ParseDecCoins("0.00005eeur")
			Expect(err).ToNot(HaveOccurred())

			_, success, err := emcli.AuthoritySetMinGasPrices(Authority, prices.String())
			Expect(success).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())

			s, success, err := emcli.MarketAddLimitOrder(acc2, "5000eeur", "100000ejpy", "acc2cid1", "--fees", "50eeur")
			Expect(err).To(BeNil())
			Expect(success).To(BeTrue())

			// Create one transaction that includes a market order and a lot transfers, which will make the tx run out of gas.
			jsonPath, err := ioutil.TempDir("", "")
			Expect(err).To(BeNil())
			defer os.RemoveAll(jsonPath)

			msgs := make([]sdk.Msg, 0)

			addr3, err := sdk.AccAddressFromBech32(acc3.GetAddress())
			Expect(err).To(BeNil())

			const clientOrderId = "ShouldNotBePresent"

			// Order must not be available for querying or fail fast
			bz, err := emcli.QueryMarketByAccount(addr3.String())
			Expect(err).To(BeNil())
			query := fmt.Sprintf("orders.#(client_order_id==\"%v\")#", clientOrderId)
			Expect(gjson.ParseBytes(bz).Get(query).Array()).To(BeEmpty())

			addOrder := &market.MsgAddLimitOrder{
				TimeInForce:   market.TimeInForce_GoodTillCancel,
				Owner:         acc3.GetAddress(),
				Source:        sdk.NewCoin("echf", sdk.NewInt(50000)),
				Destination:   sdk.NewCoin("eeur", sdk.NewInt(60000)),
				ClientOrderId: clientOrderId,
			}

			msgs = append(msgs, addOrder)

			// Add a few transfers to make sure that gas is exhausted
			addr2, err := sdk.AccAddressFromBech32(acc2.GetAddress())
			Expect(err).To(BeNil())

			coins := sdk.NewCoins(sdk.NewCoin("eeur", sdk.NewInt(5000)))
			for i := 0; i < 5; i++ {
				msgs = append(msgs, banktypes.NewMsgSend(addr3, addr2, coins))
			}

			accountJson, err := emcli.QueryAccountJson(acc3.GetAddress())
			Expect(err).To(BeNil())
			accNum := gjson.ParseBytes(accountJson).Get("account_number").Uint()
			accSeq := gjson.ParseBytes(accountJson).Get("sequence").Uint()

			// todo (reviewer): reduced the fee to ensure the TX fails.
			tx := networktest.CreateMultiMsgTx(acc3, testnet.ChainID(), "0.1eeur", accNum, accSeq, msgs...)

			cfg := emoney.MakeEncodingConfig()
			txBz, err := cfg.TxConfig.TxJSONEncoder()(tx)
			Expect(err).To(BeNil())

			transactionPath := fmt.Sprintf("%v/tx.json", jsonPath)
			ioutil.WriteFile(transactionPath, txBz, 0777)

			s, err = emcli.CustomCommand("tx", "broadcast", transactionPath)
			Expect(err).To(BeNil())
			// Transaction must have failed due to insufficient gas
			Expect(gjson.Parse(s).Get("logs.0.success").Exists()).To(Equal(false))

			// Order must not be available for querying
			bz, err = emcli.QueryMarketByAccount(addr3.String())
			Expect(err).To(BeNil())

			query = fmt.Sprintf("orders.#(client_order_id==\"%v\")#", clientOrderId)
			Expect(gjson.ParseBytes(bz).Get(query).Array()).To(BeEmpty())
		})
	})
})
