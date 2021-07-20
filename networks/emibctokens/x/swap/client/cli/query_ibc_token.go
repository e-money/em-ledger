package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/e-money/stargate/networks/emibctokens/x/swap/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func CmdListIbcToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-ibc-token",
		Short: "list all ibcToken",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllIbcTokenRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.IbcTokenAll(context.Background(), params)
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

func CmdShowIbcToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-ibc-token [index]",
		Short: "shows a ibcToken",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argsIndex, err := cast.ToStringE(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryGetIbcTokenRequest{
				Index: argsIndex,
			}

			res, err := queryClient.IbcToken(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
