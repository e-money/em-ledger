package cli

import (
	"emoney/x/issuer/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
)

func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	issuanceTxCmd := &cobra.Command{
		Use:                        "issuer",
		Short:                      "control inflation rates and manage liquidity providers",
		Aliases:                    []string{"i"},
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
	}

	issuanceTxCmd.AddCommand(
		client.PostCommands(
			getCmdIncreaseCredit(cdc),
			getCmdDecreaseCredit(cdc),
			getCmdSetInflation(cdc),
			getCmdRevokeLiquidityProvider(cdc),
		)...,
	)

	return issuanceTxCmd
}

func getCmdSetInflation(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-inflation [issuer_key_or_address] [denomination] [inflation]",
		Example: "emcli issuer set-inflation issuerkey x2eur 0.02",
		Short:   "Set the inflation rate for a denomination",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			denom := args[1]

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

func getCmdIncreaseCredit(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "increase-credit [issuer_key_or_address] [liquidity_provider_address] [amount]",
		Short: "Increase the credit of a liquidity provider.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			lpAcc, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			creditIncrease, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}

			msg := types.MsgIncreaseCredit{
				CreditIncrease:    creditIncrease,
				LiquidityProvider: lpAcc,
				Issuer:            cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdDecreaseCredit(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "decrease-credit [issuer_key_or_address] [liquidity_provider_address] [amount]",
		Short: "Decrease the credit of a liquidity provider. Credit cannot be negative",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			lpAcc, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			creditDecrease, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}

			msg := types.MsgDecreaseCredit{
				CreditDecrease:    creditDecrease,
				LiquidityProvider: lpAcc,
				Issuer:            cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdRevokeLiquidityProvider(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "revoke-credit [issuer_key_or_address] [liquidity_provider_address]",
		Short: "Revoke liquidity provider status for account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
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
