// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"bufio"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/authority/types"

	"github.com/spf13/cobra"
)

func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	authorityCmds := &cobra.Command{
		Use:                "authority",
		Short:              "Manage liquidity providers",
		DisableFlagParsing: false,
	}

	authorityCmds.AddCommand(
		flags.PostCommands(
			getCmdCreateIssuer(cdc),
			getCmdDestroyIssuer(cdc),
			getCmdSetGasPrices(cdc),
		)...,
	)

	return authorityCmds
}

func getCmdSetGasPrices(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-gas-prices [authority_key_or_address] [minimum_gas_prices]",
		Example: "emcli authority set-gas-prices masterkey 0.0005eeur,0.0000001ejpy",
		Short:   "Control the minimum gas prices for the chain",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			gasPrices, err := sdk.ParseDecCoins(args[1])
			if err != nil {
				return err
			}

			msg := types.MsgSetGasPrices{
				GasPrices: gasPrices,
				Authority: cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdCreateIssuer(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "create-issuer [authority_key_or_address] [issuer_address] [denominations]",
		Example: "emcli authority create-issuer masterkey emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu eeur,ejpy",
		Short:   "Create a new issuer",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			issuerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			denoms, err := util.ParseDenominations(args[2])
			if err != nil {
				return err
			}

			msg := types.MsgCreateIssuer{
				Issuer:        issuerAddr,
				Denominations: denoms,
				Authority:     cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdDestroyIssuer(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "destroy-issuer [authority_key_or_address] [issuer_address]",
		Example: "emcli authority destory-issuer masterkey emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu",
		Short:   "Delete an issuer",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			issuerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.MsgDestroyIssuer{
				Issuer:    issuerAddr,
				Authority: cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
