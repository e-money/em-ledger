// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/queries/types"
	"github.com/spf13/cobra"
)

func GetQuerySpendableBalance() *cobra.Command {
	spendableBalanceCmd := &cobra.Command{
		Use:   "spendable",
		Short: "Display the vested balance of an account",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			key, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			resp, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, types.QuerySpendable, key))

			var bal sdk.Coins
			err = clientCtx.LegacyAmino.UnmarshalJSON(resp, &bal)
			if err != nil {
				return err
			}

			return clientCtx.PrintBytes(resp)
		},
	}

	return spendableBalanceCmd
}

// Meant as an extension to the "emcli query supply" queries.
func GetQueryCirculatingSupplyCmd() *cobra.Command {
	circulatingSupplyCmd := &cobra.Command{
		Use:   "circulating",
		Short: "Display circulating (ie non-vesting) token supply",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			resp, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryCirculating))
			if err != nil {
				return err
			}

			var totalSupply sdk.Coins
			err = clientCtx.LegacyAmino.UnmarshalJSON(resp, &totalSupply)
			if err != nil {
				return err
			}

			return clientCtx.PrintBytes(resp)
		},
	}

	return circulatingSupplyCmd
}
