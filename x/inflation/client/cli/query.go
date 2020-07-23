// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/e-money/em-ledger/x/inflation/internal/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	inflationQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Commands for querying the inflation state",
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryInflation)
			res, _, err := cliCtx.QueryWithData(route, nil)

			if err != nil {
				return err
			}

			// TODO Consider introducing a more presentation-friendly struct
			var is types.InflationState
			if err := cdc.UnmarshalJSON(res, &is); err != nil {
				return err
			}

			return cliCtx.PrintOutput(is)
		},
	}

	inflationQueryCmd = flags.GetCommands(inflationQueryCmd)[0]

	return inflationQueryCmd
}
