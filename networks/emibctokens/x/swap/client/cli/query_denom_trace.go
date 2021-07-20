package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/e-money/stargate/networks/emibctokens/x/swap/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func CmdListDenomTrace() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-denom-trace",
		Short: "list all denomTrace",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllDenomTraceRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.DenomTraceAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowDenomTrace() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-denom-trace [index]",
		Short: "shows a denomTrace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argsIndex, err := cast.ToStringE(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryGetDenomTraceRequest{
				Index: argsIndex,
			}

			res, err := queryClient.DenomTrace(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
