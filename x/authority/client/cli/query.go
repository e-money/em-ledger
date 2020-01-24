// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/e-money/em-ledger/x/authority/types"
)

func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the authority module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetGasPricesCmd(cdc),
	)

	return cmd
}

func GetGasPricesCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gas-prices",
		Short: "Query the current minimum gas prices",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryGasPrices))
			if err != nil {
				return err
			}

			var out string
			if cliCtx.Indent {
				var buf bytes.Buffer
				err = json.Indent(&buf, bz, "", "  ")
				out = buf.String()
			} else {
				out = string(bz)
			}

			if err != nil {
				return err
			}

			_, err = fmt.Println(out)
			return err
		},
	}

	return flags.GetCommands(cmd)[0]
}
