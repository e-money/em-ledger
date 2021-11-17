// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

//go:build bdd
// +build bdd

package emoney_test

import (
	"encoding/json"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	nt "github.com/e-money/em-ledger/networktest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

const (
	// gjson paths
	QGetInflationEUR = "state.assets.#(denom==\"eeur\").inflation"
)

var _ = Describe("Authority", func() {
	emcli := testnet.NewEmcli()

	var (
		Authority         = testnet.Keystore.Authority
		Issuer            = testnet.Keystore.Key1
		LiquidityProvider = testnet.Keystore.Key2
		OtherIssuer       = testnet.Keystore.Key3
	)

	Describe("Authority manages issuers", func() {
		It("creates a new testnet", createNewTestnet)

		It("Impostor attempts to change the number of validators", func() {
			const validatorsCount = "validators.#"

			var (
				vCntParamValue = 4
				vExpectedCnt   = 4
			)

			// starting with 4 active validators
			validators, err := emcli.QueryActiveValidators()
			Expect(err).ToNot(HaveOccurred())
			validatorCnt := validators.Get(validatorsCount).Num
			Expect(validatorCnt).To(Equal(float64(vExpectedCnt)))

			// Attempt to set 5 validators
			vCntParamValue = 5
			_, success, err := emcli.AuthoritySetParams(Issuer, fmt.Sprintf(`[{"subspace":"staking","key":"MaxValidators","value":%d}]`,
				vCntParamValue,
			))
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())

			// bummer, change max_gas or block size?
			_, success, err = emcli.AuthoritySetParams(Issuer, fmt.Sprintf(`[{"subspace":"baseapp","key":"BlockParams","value":{"max_bytes":"%d","max_gas":"%d"}}]`,
				1 /* block size */, 1, /* max_gas */
			))
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())
		})

		It("Authority changes the number of validators", func() {
			const validatorsCount = "validators.#"

			var (
				vCntParamValue = 4
				vExpectedCnt   = 4
			)

			// starting with 4 active validators
			validators, err := emcli.QueryActiveValidators()
			Expect(err).ToNot(HaveOccurred())
			validatorCnt := validators.Get(validatorsCount).Num
			Expect(validatorCnt).To(Equal(float64(vExpectedCnt)))

			// set 1 validator
			vCntParamValue = 1
			_, success, err := emcli.AuthoritySetParams(Authority, fmt.Sprintf(`[{"subspace":"staking","key":"MaxValidators","value":%d}]`,
				vCntParamValue,
			))
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			nt.IncChain(1)

			// 1 validator is active
			vExpectedCnt = 1
			validators, err = emcli.QueryActiveValidators()
			Expect(err).ToNot(HaveOccurred())
			validatorCnt = validators.Get(validatorsCount).Num
			Expect(validatorCnt).To(Equal(float64(vExpectedCnt)))

			// set validator count to 10 but 4 are available
			vCntParamValue = 10
			_, success, err = emcli.AuthoritySetParams(Authority, fmt.Sprintf(`[{"subspace":"staking","key":"MaxValidators","value":%d}]`,
				vCntParamValue,
			))
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			nt.IncChain(1)

			vExpectedCnt = 4
			validators, err = emcli.QueryActiveValidators()
			Expect(err).ToNot(HaveOccurred())
			validatorCnt = validators.Get(validatorsCount).Num
			Expect(validatorCnt).To(Equal(float64(vExpectedCnt)))
		})

		It("Authority changes the block size", func() {
			type blockParamsType struct {
				MaxBytesStr string `json:"max_bytes"`
				MaxGasStr   string `json:"max_gas"`
			}
			var (
				vExpectedCnt float64
				blockParams  blockParamsType
			)

			// starting with 22020096 bytes
			blockParamsGJson, err := emcli.QueryBlockParams()
			Expect(err).ToNot(HaveOccurred())
			blockBytesStr := blockParamsGJson.Get("value").Str
			// gjson responds with nil for "value.max_bytes"
			// so employing partial json's unmarshalling
			err = json.Unmarshal([]byte(blockBytesStr), &blockParams)
			Expect(err).ToNot(HaveOccurred())
			vExpectedCnt, err = strconv.ParseFloat(blockParams.MaxBytesStr, 64)
			Expect(err).ToNot(HaveOccurred())

			// set 22022120 bytes
			vExpectedCnt += 1024
			_, success, err := emcli.AuthoritySetParams(Authority, fmt.Sprintf(`[{"subspace":"baseapp","key":"BlockParams","value":{"max_bytes":"%d","max_gas":"60000000"}}]`,
				int(vExpectedCnt),
			))
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			nt.IncChain(1)

			blockParamsGJson, err = emcli.QueryBlockParams()
			Expect(err).ToNot(HaveOccurred())
			blockBytesStr = blockParamsGJson.Get("value").Str
			err = json.Unmarshal([]byte(blockBytesStr), &blockParams)
			Expect(err).ToNot(HaveOccurred())
			updMaxBytes, err := strconv.ParseFloat(blockParams.MaxBytesStr, 64)
			Expect(err).ToNot(HaveOccurred())
			Expect(vExpectedCnt).To(Equal(updMaxBytes))
		})

		It("creates an issuer", func() {
			// denomination metadata are not set before a new issuer
			denomList, err := emcli.QueryDenomMetadata()
			Expect(err).NotTo(HaveOccurred())
			Expect(denomList).To(HaveLen(6))

			ok := nt.AuthCreatesIssuer(emcli, Authority, Issuer)
			Expect(ok).To(BeTrue())

			// denomination metadata are set to EEUR, EJPY
			denomList, err = emcli.QueryDenomMetadata()
			Expect(err).NotTo(HaveOccurred())
			Expect(denomList).To(HaveLen(7))
			Expect(denomList[2].Get("base").Str).To(Equal("eeur"))
			Expect(denomList[2].Get("display").Str).To(Equal("EEUR"))
			Expect(denomList[2].Get("description").Str).To(Equal("e-Money EUR stablecoin"))
			Expect(denomList[3].Get("base").Str).To(Equal("ejpy"))
			Expect(denomList[3].Get("display").Str).To(Equal("eJPY"))
			Expect(denomList[3].Get("description").Str).To(Equal("Japanese yen stablecoin"))
		})

		It("imposter attempts to act as authority", func() {
			_, success, err := emcli.AuthorityCreateIssuer(Issuer, LiquidityProvider, "echf", "edkk")
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())
		})

		It("authority assigns a second issuer to same denomination", func() {
			_, success, err := emcli.AuthorityCreateIssuer(Authority, OtherIssuer, "ejpy")
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())
		})

		It("authority creates a second issuer", func() {
			issuers, denoms := nt.CreateIssuer(emcli, Authority, OtherIssuer, `'esek,eSEK,Not a bad stablecoin'`, `"edkk,EDKK,Coolest stablecoin"`, `"echf,eCHF,yet another stablecoin"`)
			Expect(issuers).To(HaveLen(2))
			Expect(denoms).To(Equal("echf,edkk,esek"))

			denomList, err := emcli.QueryDenomMetadata()
			Expect(err).NotTo(HaveOccurred())
			Expect(denomList).To(HaveLen(7))
			Expect(denomList[0].Get("base").Str).To(Equal("echf"))
			Expect(denomList[0].Get("display").Str).To(Equal("eCHF"))
			Expect(denomList[0].Get("description").Str).To(Equal("yet another stablecoin"))
			Expect(denomList[1].Get("base").Str).To(Equal("edkk"))
			Expect(denomList[1].Get("display").Str).To(Equal("EDKK"))
			Expect(denomList[1].Get("description").Str).To(Equal("Coolest stablecoin"))
			Expect(denomList[5].Get("base").Str).To(Equal("esek"))
			Expect(denomList[5].Get("display").Str).To(Equal("eSEK"))
			Expect(denomList[5].Get("description").Str).To(Equal("Not a bad stablecoin"))
		})

		It("creates a liquidity provider", func() {
			// The issuer makes a liquidity provider of EUR
			_, success, err := emcli.IssuerIncreaseMintableAmount(Issuer, LiquidityProvider, "50000eeur")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			bz, err := emcli.QueryMintableJson(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			lpaccount := gjson.ParseBytes(bz)
			mintableAmount := lpaccount.Get("mintable").Array()
			Expect(mintableAmount).To(HaveLen(1))
			Expect(mintableAmount[0].Get("denom").Str).To(Equal("eeur"))
			Expect(mintableAmount[0].Get("amount").Str).To(Equal("50000"))
		})

		It("changes inflation of a denomination", func() {
			bz, err := emcli.QueryInflation()
			Expect(err).ToNot(HaveOccurred())

			s := gjson.ParseBytes(bz).Get(QGetInflationEUR).Str
			inflationBefore, _ := sdk.NewDecFromStr(s)

			_, success, err := emcli.IssuerSetInflation(Issuer, "eeur", "0.1")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			bz, err = emcli.QueryInflation()
			Expect(err).ToNot(HaveOccurred())

			s = gjson.ParseBytes(bz).Get(QGetInflationEUR).Str
			inflationAfter, _ := sdk.NewDecFromStr(s)

			Expect(inflationAfter).ToNot(Equal(inflationBefore))
			Expect(inflationAfter).To(Equal(sdk.MustNewDecFromStr("0.100")))
		})

		It("attempts to change inflation of denomination not under its control", func() {
			_, success, err := emcli.IssuerSetInflation(OtherIssuer, "eeur", "0.5")

			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())
		})

		It("creates an issuer of a completely new denomination", func() {
			_, success, err := emcli.AuthorityCreateIssuer(Authority, OtherIssuer, "caps")
			Expect(err).To(BeNil())
			Expect(success).To(BeTrue())

			bz, err := emcli.QueryInflation()
			Expect(err).To(BeNil())

			fmt.Println(string(bz))

			s := gjson.ParseBytes(bz).Get("state.assets.#(denom==\"caps\").inflation").Str
			inflationCaps, _ := sdk.NewDecFromStr(s)
			Expect(inflationCaps).To(Equal(sdk.ZeroDec()))
		})

		It("liquidity provider draws on its mintable amount", func() {
			balanceBefore, err := emcli.QueryBalance(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())
			mintableBefore, err := emcli.QueryMintable(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			_, success, err := emcli.LiquidityProviderMint(LiquidityProvider, "20000eeur")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue())

			mintableAfter, err := emcli.QueryMintable(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			balanceAfter, err := emcli.QueryBalance(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			Expect(balanceAfter).To(Equal(balanceBefore + 20000))
			Expect(mintableAfter).To(Equal(mintableBefore - 20000))
		})

		It("liquidity provider attempts to overdraw its mintable balance", func() {
			balanceBefore, err := emcli.QueryBalance(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			mintableBefore, err := emcli.QueryMintable(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			_, success, err := emcli.LiquidityProviderMint(LiquidityProvider, "500000eeur")
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())

			mintableAfter, err := emcli.QueryMintable(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			balanceAfter, err := emcli.QueryBalance(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			Expect(balanceAfter).To(Equal(balanceBefore))
			Expect(mintableAfter).To(Equal(mintableBefore))
		})

		It("liquidity provider burns some tokens", func() {
			balanceBefore, err := emcli.QueryBalance(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())
			mintableBefore, err := emcli.QueryMintable(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			_, success, err := emcli.LiquidityProviderBurn(LiquidityProvider, "500000eeur")
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			mintableAfter, err := emcli.QueryMintable(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())
			balanceAfter, err := emcli.QueryBalance(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			Expect(balanceAfter).To(Equal(balanceBefore - 500000))
			Expect(mintableAfter).To(Equal(mintableBefore + 500000))
		})

		It("liquidity provider gets mintable amount reduced", func() {
			mintableBefore, err := emcli.QueryMintable(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			_, success, err := emcli.IssuerDecreaseMintableAmount(Issuer, LiquidityProvider, "10000eeur")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue())

			mintableAfter, err := emcli.QueryMintable(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			Expect(mintableAfter).To(Equal(mintableBefore - 10000))
		})

		It("liquidity provider gets revoked", func() {
			_, success, err := emcli.IssuerRevokeMinting(Issuer, LiquidityProvider)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue())

			bz, err := emcli.QueryMintableJson(LiquidityProvider.GetAddress())
			mintable := gjson.ParseBytes(bz).Get("mintable").Array()
			Expect(mintable).To(HaveLen(0))
		})

		It("former liquidity provider attempts to mint", func() {
			balanceBefore, err := emcli.QueryBalance(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())
			_, err = emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			_, success, err := emcli.LiquidityProviderMint(LiquidityProvider, "10000eeur")
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())

			_, err = emcli.QueryAccount(LiquidityProvider.GetAddress())
			Expect(err).ToNot(HaveOccurred())

			balanceAfter, err := emcli.QueryBalance(LiquidityProvider.GetAddress())
			Expect(err).ShouldNot(HaveOccurred())

			Expect(balanceBefore).To(Equal(balanceAfter))
		})

		It("issuer gets revoked", func() {
			_, success, err := emcli.AuthorityDestroyIssuer(Authority, Issuer)
			Expect(success).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())

			_, success, err = emcli.IssuerSetInflation(Issuer, "eeur", "0.5")
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())
		})

		It("Authority sets new gas prices", func() {
			prices, err := sdk.ParseDecCoins("0.00005eeur")
			Expect(err).ToNot(HaveOccurred())

			_, success, err := emcli.AuthoritySetMinGasPrices(Authority, prices.String())
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())

			bz, err := emcli.QueryMinGasPrices()
			Expect(err).ToNot(HaveOccurred())

			_, success, err = emcli.AuthoritySetMinGasPrices(Authority, prices.String(), "--fees", "50eeur")
			Expect(success).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())

			gasPrices := gjson.ParseBytes(bz).Get("min_gas_prices")

			queriedPrices := sdk.DecCoins{}
			for _, price := range gasPrices.Array() {
				gasPrice := sdk.NewDecCoinFromDec(price.Get("denom").Str, sdk.MustNewDecFromStr(price.Get("amount").Str))
				queriedPrices = append(queriedPrices, gasPrice)
			}

			Expect(queriedPrices).To(Equal(prices))

			// A non-authority attempts to set gas prices
			_, success, err = emcli.AuthoritySetMinGasPrices(Issuer, prices.String(), "--fees", "50eeur")
			Expect(success).To(BeFalse())
			Expect(err).To(HaveOccurred())
		})
	})
})
