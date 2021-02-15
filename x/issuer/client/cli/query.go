// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/e-money/em-ledger/x/issuer/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issuers",
		Short: "List issuers",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			resp, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryIssuers))
			if err != nil {
				return err
			}

			issuers := make(types.Issuers, 0)
			clientCtx.LegacyAmino.MustUnmarshalJSON(resp, &issuers)

			return clientCtx.PrintBytes(resp)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
