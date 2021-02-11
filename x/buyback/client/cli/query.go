// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/e-money/em-ledger/x/buyback/internal/keeper"
	"github.com/e-money/em-ledger/x/buyback/internal/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Query commands for the buyback module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetModuleBalanceCmd(),
	)

	return cmd
}

func GetModuleBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Query for the current buyback balance",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryBalance))
			if err != nil {
				return err
			}

			switch clientCtx.OutputFormat {
			case "text":
				response := keeper.QueryBalanceResponse{}
				if err := json.Unmarshal(bz, &response); err != nil {
					return err
				}

				for _, b := range response.Balance {
					clientCtx.PrintString(b.String())
				}
			case "json":
				clientCtx.PrintBytes(bz)
			}
			return nil
		},
	}

	return cmd
}
