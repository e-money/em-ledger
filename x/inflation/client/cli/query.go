package cli

import (
	"emoney/x/inflation/internal/types"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"

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

	inflationQueryCmd = client.GetCommands(inflationQueryCmd)[0]

	return inflationQueryCmd
}
