// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/e-money/em-ledger/x/inflation/internal/types"

	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Commands for querying the inflation state",
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryInflation)
			res, _, err := clientCtx.QueryWithData(route, nil)

			if err != nil {
				return err
			}

			// TODO Consider introducing a more presentation-friendly struct
			var is types.InflationState
			if err := clientCtx.LegacyAmino.UnmarshalJSON(res, &is); err != nil {
				return err
			}

			return clientCtx.PrintBytes(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
