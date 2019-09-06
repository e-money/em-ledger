package cli

import (
	"emoney/x/inflation/internal/types"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
)

func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	inflationTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Commands for managing inflation",
	}

	inflationTxCmd.AddCommand(client.PostCommands(
		GetCmdSetInflation(cdc),
	)...)

	return inflationTxCmd
}

func GetCmdSetInflation(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:  "set [denomination] [inflation-rate]",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			inflation, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return err
			}

			denom := args[0]
			// TODO Validate denomination

			msg := types.MsgSetInflation{
				Denom:     denom,
				Inflation: inflation,
				Principal: cliCtx.GetFromAddress(),
			}

			fmt.Println(" *** Attempting to broadcast", msg)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
