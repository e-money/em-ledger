// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/issuer/types"
	"github.com/spf13/cobra"
)

func GetTxCmd() *cobra.Command {
	issuanceTxCmd := &cobra.Command{
		Use:                        "issuer",
		Short:                      "Control inflation rates and manage liquidity providers",
		Aliases:                    []string{"i"},
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
	}

	issuanceTxCmd.AddCommand(
		getCmdIncreaseMintableAmount(),
		getCmdDecreaseMintableAmount(),
		getCmdSetInflation(),
		getCmdRevokeLiquidityProvider(),
	)

	return issuanceTxCmd
}

func getCmdSetInflation() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-inflation [issuer_key_or_address] [denomination] [inflation]",
		Example: "emd tx issuer set-inflation issuerkey eeur 0.02",
		Short:   "Set the inflation rate for a denomination",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denom := args[1]
			if !util.ValidateDenom(denom) {
				return fmt.Errorf("invalid denomination: %v", denom)
			}

			inflation, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgSetInflation{
				Denom:         denom,
				InflationRate: inflation,
				Issuer:        clientCtx.GetFromAddress().String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func getCmdIncreaseMintableAmount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "increase-mintable [issuer_key_or_address] [liquidity_provider_address] [amount]",
		Short: "Increase the amount mintable for a liquidity provider.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			lpAcc, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			mintableIncrease, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgIncreaseMintable{
				MintableIncrease:  mintableIncrease,
				LiquidityProvider: lpAcc.String(),
				Issuer:            clientCtx.GetFromAddress().String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func getCmdDecreaseMintableAmount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decrease-mintable [issuer_key_or_address] [liquidity_provider_address] [amount]",
		Short: "Decrease the amount mintable for a liquidity provider. Result cannot be negative",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			lpAcc, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			mintableDecrease, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgDecreaseMintable{
				MintableDecrease:  mintableDecrease,
				LiquidityProvider: lpAcc.String(),
				Issuer:            clientCtx.GetFromAddress().String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func getCmdRevokeLiquidityProvider() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-mint [issuer_key_or_address] [liquidity_provider_address]",
		Short: "Revoke liquidity provider status for an account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			lpAcc, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgRevokeLiquidityProvider{
				LiquidityProvider: lpAcc.String(),
				Issuer:            clientCtx.GetFromAddress().String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
