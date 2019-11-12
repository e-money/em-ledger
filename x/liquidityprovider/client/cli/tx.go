package cli

import (
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
		Short:              "Liquidity provider operations",
		Aliases:            []string{"lp"},
		DisableFlagParsing: false,
	}

	lpCmds.AddCommand(client.PostCommands(
		getCmdMint(cdc),
		getCmdBurn(cdc),
	)...)

	lpCmds = client.PostCommands(lpCmds)[0]
	return lpCmds
}

func getCmdBurn(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "burn [liquidity_provider_key_or_address] [amount]",
		Short: "Destroys the given amount of tokens",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			amount, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			msg := types.MsgBurnTokens{
				Amount:            amount,
				LiquidityProvider: cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdMint(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "mint [liquidity_provider_key_or_address] [amount]",
		Short: "Creates new tokens from the liquidity provider's mintable amount",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			amount, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			msg := types.MsgMintTokens{
				Amount:            amount,
				LiquidityProvider: cliCtx.GetFromAddress(),
			}

			result := utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
			return result
		},
	}
}
