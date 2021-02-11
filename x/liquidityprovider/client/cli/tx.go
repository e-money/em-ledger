// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
	"github.com/spf13/cobra"
)

func GetTxCmd() *cobra.Command {
	lpCmds := &cobra.Command{
		Use:                "liquidityprovider",
		Short:              "Liquidity provider operations",
		Aliases:            []string{"lp"},
		DisableFlagParsing: false,
	}

	lpCmds.AddCommand(
		getCmdMint(),
		getCmdBurn(),
	)

	return lpCmds
}

func getCmdBurn() *cobra.Command {
	return &cobra.Command{
		Use:   "burn [liquidity_provider_key_or_address] [amount]",
		Short: "Destroys the given amount of tokens",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgBurnTokens{
				Amount:            amount,
				LiquidityProvider: clientCtx.GetFromAddress().String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

func getCmdMint() *cobra.Command {
	return &cobra.Command{
		Use:   "mint [liquidity_provider_key_or_address] [amount]",
		Short: "Creates new tokens from the liquidity provider's mintable amount",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgMintTokens{
				Amount:            amount,
				LiquidityProvider: clientCtx.GetFromAddress().String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}
