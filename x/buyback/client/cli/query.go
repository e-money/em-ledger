// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/e-money/em-ledger/x/buyback/internal/keeper"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/e-money/em-ledger/x/buyback/internal/types"
)

func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Query commands for the buyback module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetModuleBalanceCmd(cdc),
	)

	return cmd
}

func GetModuleBalanceCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Query for the current buyback balance",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryBalance))
			if err != nil {
				return err
			}

			switch cliCtx.OutputFormat {
			case "text":
				response := keeper.QueryBalanceResponse{}
				json.Unmarshal(bz, &response)

				for _, b := range response.Balance {
					fmt.Println(b.String())
				}

			case "json":
				if cliCtx.Indent {
					buf := new(bytes.Buffer)
					err = json.Indent(buf, bz, "", "  ")
					if err != nil {
						return err
					}

					bz = buf.Bytes()
				}

				fmt.Println(string(bz))
			}

			return nil
		},
	}

	return flags.GetCommands(cmd)[0]
}
