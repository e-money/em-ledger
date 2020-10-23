// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/market/keeper"
	"github.com/tidwall/gjson"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/e-money/em-ledger/x/market/types"
)

func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the market module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetInstrumentsCmd(cdc),
		GetInstrumentCmd(cdc),
		GetByAccountCmd(cdc),
	)

	return cmd
}

func GetByAccountCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [key_or_address]",
		Short: "Query orders placed by a specific account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				// Named key specified
				addr, _, err = context.GetFromFields(os.Stdin, args[0], viper.GetBool(flags.FlagGenerateOnly))
				if err != nil {
					return err
				}
			}

			bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, types.QueryByAccount, addr))
			if err != nil {
				return err
			}

			switch cliCtx.OutputFormat {
			case "text":
				fmt.Println(stringifyOrders(bz))

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

func stringifyOrders(bz []byte) string {
	sb := new(strings.Builder)

	allOrders := gjson.ParseBytes(bz).Get("orders")

	for _, order := range allOrders.Array() {
		srcDenom, dstDenom := order.Get("source.denom").Str, order.Get("destination.denom").Str

		sb.WriteString(
			fmt.Sprintf("%v : %v -> %v @ %v (%v)\n - (%v%v remaining) (%v%v filled) (%v%v filled)\n",
				order.Get("id").Raw,
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

func GetInstrumentCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instrument [source-denomination] [destination-denomination]",
		Short: "Query the order book of a specific instrument",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			source, destination := args[0], args[1]

			bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, types.QueryInstrument, source, destination))
			if err != nil {
				return err
			}

			resp := new(keeper.QueryInstrumentResponse)
			err = json.Unmarshal(bz, resp)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}

	return flags.GetCommands(cmd)[0]
}

func GetInstrumentsCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instruments",
		Short: "Query the current instruments",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryInstruments))
			if err != nil {
				return err
			}

			resp := new(keeper.QueryInstrumentsWrapperResponse)
			err = json.Unmarshal(bz, resp)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}

	return flags.GetCommands(cmd)[0]
}
