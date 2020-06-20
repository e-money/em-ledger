// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"bufio"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/issuer/types"
	"github.com/spf13/cobra"
)

func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	issuanceTxCmd := &cobra.Command{
		Use:                        "issuer",
		Short:                      "Control inflation rates and manage liquidity providers",
		Aliases:                    []string{"i"},
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
	}

	issuanceTxCmd.AddCommand(
		flags.PostCommands(
			getCmdIncreaseMintableAmount(cdc),
			getCmdDecreaseMintableAmount(cdc),
			getCmdSetInflation(cdc),
			getCmdRevokeLiquidityProvider(cdc),
		)...,
	)

	return issuanceTxCmd
}

func getCmdSetInflation(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-inflation [issuer_key_or_address] [denomination] [inflation]",
		Example: "emcli issuer set-inflation issuerkey eeur 0.02",
		Short:   "Set the inflation rate for a denomination",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			denom := args[1]
			if !util.ValidateDenom(denom) {
				return fmt.Errorf("invalid denomination: %v", denom)
			}

			inflation, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return err
			}

			msg := types.MsgSetInflation{
				Denom:         denom,
				InflationRate: inflation,
				Issuer:        cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdIncreaseMintableAmount(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "increase-mintable [issuer_key_or_address] [liquidity_provider_address] [amount]",
		Short: "Increase the amount mintable for a liquidity provider.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			lpAcc, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			mintableIncrease, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}

			msg := types.MsgIncreaseMintable{
				MintableIncrease:  mintableIncrease,
				LiquidityProvider: lpAcc,
				Issuer:            cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdDecreaseMintableAmount(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "decrease-mintable [issuer_key_or_address] [liquidity_provider_address] [amount]",
		Short: "Decrease the amount mintable for a liquidity provider. Result cannot be negative",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			lpAcc, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			mintableDecrease, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}

			msg := types.MsgDecreaseMintable{
				MintableDecrease:  mintableDecrease,
				LiquidityProvider: lpAcc,
				Issuer:            cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdRevokeLiquidityProvider(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "revoke-mint [issuer_key_or_address] [liquidity_provider_address]",
		Short: "Revoke liquidity provider status for an account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			lpAcc, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.MsgRevokeLiquidityProvider{
				LiquidityProvider: lpAcc,
				Issuer:            cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
