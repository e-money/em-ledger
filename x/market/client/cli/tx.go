// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"bufio"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"

	"github.com/e-money/em-ledger/x/market/types"
)

const (
	flag_TimeInForce = "time-in-force"

	flag_TimeInForceDescription = "Select the order's time-in-force value (GTC|IOC|FOK)"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Market transaction commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		AddLimitOrderCmd(cdc),
		AddMarketOrderCmd(cdc),
		CancelOrderCmd(cdc),
		CancelReplaceOrder(cdc),
	)
	return txCmd
}

func AddLimitOrderCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-limit [source-amount] [destination-amount] [client-orderid]",
		Short: "Create a limit order and send it to the market",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			src, err := sdk.ParseCoin(args[0])
			if err != nil {
				return
			}

			dst, err := sdk.ParseCoin(args[1])
			if err != nil {
				return
			}

			clientOrderID := args[2]

			timeInForce, err := types.TimeInForceFromString(viper.GetString(flag_TimeInForce))
			if err != nil {
				return err
			}

			msg := types.MsgAddLimitOrder{
				Owner:         cliCtx.GetFromAddress(),
				TimeInForce:   timeInForce.String(),
				Source:        src,
				Destination:   dst,
				ClientOrderId: clientOrderID,
			}

			err = msg.ValidateBasic()
			if err != nil {
				return
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(flag_TimeInForce, types.TimeInForce_GoodTilCancel.String(), flag_TimeInForceDescription)
	cmd = flags.PostCommands(cmd)[0]
	return cmd
}

func AddMarketOrderCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-market [source-denom] [destination-amount] [market-slippage] [client-orderid]",
		Short: "Create a market order",
		Long: `Create an order based on latest pricing information. 

Example:
 emcli tx market add-market eeur 300echf 0.05 order12345
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			srcDenom := args[0]

			dst, err := sdk.ParseCoin(args[1])
			if err != nil {
				return
			}

			slippage, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return err
			}

			clientOrderID := args[3]

			timeInForce, err := types.TimeInForceFromString(viper.GetString(flag_TimeInForce))
			if err != nil {
				return err
			}

			msg := types.MsgAddMarketOrder{
				Owner:         cliCtx.GetFromAddress(),
				TimeInForce:   timeInForce.String(),
				Source:        srcDenom,
				Destination:   dst,
				ClientOrderId: clientOrderID,
				MaxSlippage:   slippage,
			}

			err = msg.ValidateBasic()
			if err != nil {
				return
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(flag_TimeInForce, types.TimeInForce_ImmediateOrCancel.String(), flag_TimeInForceDescription)
	cmd = flags.PostCommands(cmd)[0]
	return cmd

}

func CancelOrderCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel [client-orderid]",
		Short: "Cancel an order in the market",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			clientOrderID := args[0]

			msg := types.MsgCancelOrder{
				Owner:         cliCtx.GetFromAddress(),
				ClientOrderId: clientOrderID,
			}

			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd = flags.PostCommands(cmd)[0]
	return cmd
}

func CancelReplaceOrder(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancelreplace [original-client-order-id] [source-amount] [destination-amount] [client-orderid]",
		Short: "Update an existing order",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			src, err := sdk.ParseCoin(args[1])
			if err != nil {
				return
			}

			dst, err := sdk.ParseCoin(args[2])
			if err != nil {
				return
			}

			origClientOrderID := args[0]
			newClientOrderID := args[3]

			msg := types.MsgCancelReplaceLimitOrder{
				Owner:             cliCtx.GetFromAddress(),
				Source:            src,
				Destination:       dst,
				OrigClientOrderId: origClientOrderID,
				NewClientOrderId:  newClientOrderID,
			}

			err = msg.ValidateBasic()
			if err != nil {
				return
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd = flags.PostCommands(cmd)[0]
	return cmd
}
