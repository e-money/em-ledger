package cli

import (
	"emoney/x/issuance/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

func GetTxCmd(_ string, cdc *codec.Codec) *cobra.Command {
	issuanceTxCmd := &cobra.Command{
		Use:                        "issuance",
		Short:                      " --- ",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	issuanceTxCmd.AddCommand(client.PostCommands(
		getCmdMintTokens(cdc),
	)...)

	return issuanceTxCmd
}

func getCmdMintTokens(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "mint [amount]",
		Short: "mint new tokens",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			coins, err := sdk.ParseCoins(args[0])
			if err != nil {
				return err
			}

			msg := types.MsgMintTokens{
				Coins:  coins,
				Issuer: cliCtx.GetFromAddress(),
			}

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
