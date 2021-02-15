// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/e-money/em-ledger/x/authority/keeper"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the authority module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetGasPricesCmd(),
	)

	return cmd
}

func GetGasPricesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gas-prices",
		Short: "Query the current minimum gas prices",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			//queryClient := types.NewQueryClient(clientCtx)

			bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryGasPrices))
			if err != nil {
				return err
			}

			resp := new(keeper.QueryGasPricesResponse)
			err = json.Unmarshal(bz, resp)
			if err != nil {
				return err
			}
			clientCtx.PrintBytes(bz)
			// todo (Alex): refactor to GRPC
			//return clientCtx.PrintProto(res.Balance)
			return nil
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
