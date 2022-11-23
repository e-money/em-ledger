package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/queries/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for circulation and vested balance",
		SuggestionsMinimumDistance: 2,
	}
	cmd.AddCommand(
		GetQuerySpendableBalance(),
		GetQueryCirculatingSupplyCmd(),
		GetQueryMissedBlocksCmd(),
	)

	return cmd
}

func GetQuerySpendableBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spendable [account_address]",
		Short: "Display the vested balance of an account",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Spendable(cmd.Context(), &types.QuerySpendableRequest{Address: addr.String()})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// Meant as an extension to the "emcli query supply" queries.
func GetQueryCirculatingSupplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "circulating",
		Short: "Display circulating (ie non-vesting) token supply",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Circulating(cmd.Context(), &types.QueryCirculatingRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryMissedBlocksCmd Replacing the SDK slashing signing info
func GetQueryMissedBlocksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "missedblocks [validator-conspub]",
		Short: "Query the validator's missed blocks and blocks counter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			consAddr, err := sdk.ConsAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.MissedBlocks(cmd.Context(), &types.QueryMissedBlocksRequest{
				ConsAddress: consAddr.String(),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
