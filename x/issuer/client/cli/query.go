package cli

import (
	"emoney/x/issuer/types"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/spf13/cobra"
)

func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	issuerQueryCmd := &cobra.Command{
		Use:   "issuers",
		Short: "List issuers",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resp, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryIssuers))
			if err != nil {
				return err
			}

			issuers := make(types.Issuers, 0)
			cdc.MustUnmarshalJSON(resp, &issuers)

			return cliCtx.PrintOutput(issuers)
		},
	}

	return flags.GetCommands(issuerQueryCmd)[0]
}
