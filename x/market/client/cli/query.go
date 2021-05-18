// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	apptypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/market/types"
	"github.com/spf13/cobra"
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
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ByAccount(cmd.Context(), &types.QueryByAccountRequest{Address: addr.String()})
			if err != nil {
				return err
			}

			return clientCtx.WithJSONMarshaler(apptypes.NewMarshaller(clientCtx)).PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
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

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Instrument(cmd.Context(), &types.QueryInstrumentRequest{
				Source:      args[0],
				Destination: args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.WithJSONMarshaler(apptypes.NewMarshaller(clientCtx)).PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
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

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Instruments(cmd.Context(), &types.QueryInstrumentsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.WithJSONMarshaler(apptypes.NewMarshaller(clientCtx)).PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
