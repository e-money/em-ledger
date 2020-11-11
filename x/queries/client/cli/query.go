// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/queries/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

func GetQuerySpendableBalance(cdc *codec.Codec) *cobra.Command {
	spendableBalanceCmd := &cobra.Command{
		Use:   "spendable",
		Short: "Display the vested balance of an account",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			key, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			resp, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, types.QuerySpendable, key))

			var bal sdk.Coins
			err = cdc.UnmarshalJSON(resp, &bal)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(bal)
		},
	}

	return flags.GetCommands(spendableBalanceCmd)[0]
}

// Meant as an extension to the "emcli query supply" queries.
func GetQueryCirculatingSupplyCmd(cdc *codec.Codec) *cobra.Command {
	circulatingSupplyCmd := &cobra.Command{
		Use:   "circulating",
		Short: "Display circulating (ie non-vesting) token supply",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resp, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryCirculating))
			if err != nil {
				return err
			}

			var totalSupply sdk.Coins
			err = cdc.UnmarshalJSON(resp, &totalSupply)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(totalSupply)
		},
	}

	return flags.GetCommands(circulatingSupplyCmd)[0]
}

func GetQueryStatementCmd(cdc *codec.Codec) *cobra.Command {
	statementCmd := &cobra.Command{
		Use:   "statement",
		Short: "Display a statement for the given account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			// TODO
			// Parse and verify argument address
			// Use events query api to find transactions where the account is either "sender" or "recipient"

			return nil
		},
	}

	return flags.GetCommands(statementCmd)[0]
}
