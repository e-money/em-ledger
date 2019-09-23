package cli

import (
	"fmt"
	"github.com/spf13/cobra"

	"emoney/x/liquidityprovider/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	lpCmds := &cobra.Command{
		Use:                "liquidityprovider",
		Aliases:            []string{"lp"},
		DisableFlagParsing: false,
		RunE:               client.ValidateCmd,
	}

	lpCmds.AddCommand(client.PostCommands(
		getCmdMint(cdc),
	)...)

	lpCmds = client.PostCommands(lpCmds)[0]
	return lpCmds
}

func getCmdMint(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:  "mint",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			amount, err := sdk.ParseCoins(args[0])
			if err != nil {
				return err
			}

			fmt.Println(" *** Parsed minting amount : ", amount)

			msg := types.MsgMintTokens{
				Amount:            amount,
				LiquidityProvider: cliCtx.GetFromAddress(),
			}

			result := utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
			return result
		},
	}
}
