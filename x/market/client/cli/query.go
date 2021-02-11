// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/market/keeper"
	"github.com/e-money/em-ledger/x/market/types"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"strings"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the market module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetInstrumentsCmd(),
		GetInstrumentCmd(),
		GetByAccountCmd(),
	)

	return cmd
}

func GetByAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [key_or_address]",
		Short: "Query orders placed by a specific account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				// Named key specified
				addr = clientCtx.FromAddress
			}

			bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, types.QueryByAccount, addr))
			if err != nil {
				return err
			}

			switch clientCtx.OutputFormat {
			case "text":
				return clientCtx.PrintString(stringifyOrders(bz))

			case "json":
				return clientCtx.PrintBytes(bz)
			}

			return nil
		},
	}

	return cmd
}

func stringifyOrders(bz []byte) string {
	sb := new(strings.Builder)

	allOrders := gjson.ParseBytes(bz).Get("orders")

	for _, order := range allOrders.Array() {
		srcDenom, dstDenom := order.Get("source.denom").Str, order.Get("destination.denom").Str

		sb.WriteString(
			fmt.Sprintf("%v : %v -> %v @ %v (%v)\n - (%v%v remaining) (%v%v filled) (%v%v filled)\n",
				order.Get("order_id").Raw,
				srcDenom,
				dstDenom,
				order.Get("price").Str,
				order.Get("owner").Str,
				order.Get("source_remaining").Str, srcDenom,
				order.Get("source_filled").Str, srcDenom,
				order.Get("destination_filled").Str, dstDenom,
			),
		)
	}

	return sb.String()
}

func GetInstrumentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instrument [source-denomination] [destination-denomination]",
		Short: "Query the order book of a specific instrument",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			source, destination := args[0], args[1]

			bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, types.QueryInstrument, source, destination))
			if err != nil {
				return err
			}

			resp := new(keeper.QueryInstrumentResponse)
			err = json.Unmarshal(bz, resp)
			if err != nil {
				return err
			}

			return clientCtx.PrintBytes(bz)
		},
	}

	return cmd
}

func GetInstrumentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instruments",
		Short: "Query the current instruments",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryInstruments))
			if err != nil {
				return err
			}

			resp := new(keeper.QueryInstrumentsWrapperResponse)
			err = json.Unmarshal(bz, resp)
			if err != nil {
				return err
			}

			return clientCtx.PrintBytes(bz)
		},
	}

	return cmd
}
