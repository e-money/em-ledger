package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/market/types"
	"github.com/spf13/cobra"
)

const (
	flag_TimeInForce = "time-in-force"

	flag_TimeInForceDescription = "Select the order's time-in-force value (GTC|IOC|FOK)"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Market transaction commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		AddLimitOrderCmd(),
		AddMarketOrderCmd(),
		CancelOrderCmd(),
		CancelReplaceOrder(),
	)
	return txCmd
}

func AddLimitOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-limit [source-amount] [destination-amount] [client-orderid]",
		Short: "Create a limit order and send it to the market",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			src, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return
			}

			dst, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return
			}

			clientOrderID := args[2]

			tif, err := cmd.Flags().GetString(flag_TimeInForce)
			if err != nil {
				return err
			}
			timeInForce, err := types.TimeInForceFromString(tif)
			if err != nil {
				return err
			}

			msg := &types.MsgAddLimitOrder{
				Owner:         clientCtx.GetFromAddress().String(),
				TimeInForce:   timeInForce,
				Source:        src,
				Destination:   dst,
				ClientOrderId: clientOrderID,
			}

			err = msg.ValidateBasic()
			if err != nil {
				return
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(flag_TimeInForce, "GTC", flag_TimeInForceDescription)
	return cmd
}

func AddMarketOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-market [source-denom] [destination-amount] [market-slippage] [client-orderid]",
		Short: "Create a market order",
		Long: `Create an order based on latest pricing information. 

Example:
 emd tx market add-market eeur 300echf 0.05 order12345
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			srcDenom := args[0]

			dst, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return
			}

			slippage, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return err
			}

			clientOrderID := args[3]

			tif, err := cmd.Flags().GetString(flag_TimeInForce)
			if err != nil {
				return err
			}
			timeInForce, err := types.TimeInForceFromString(tif)
			if err != nil {
				return err
			}

			msg := &types.MsgAddMarketOrder{
				Owner:         clientCtx.GetFromAddress().String(),
				TimeInForce:   timeInForce,
				Source:        srcDenom,
				Destination:   dst,
				ClientOrderId: clientOrderID,
				MaxSlippage:   slippage,
			}

			err = msg.ValidateBasic()
			if err != nil {
				return
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(flag_TimeInForce, "GTC", flag_TimeInForceDescription)
	return cmd

}

func CancelOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel [client-orderid]",
		Short: "Cancel an order in the market",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			clientOrderID := args[0]

			msg := &types.MsgCancelOrder{
				Owner:         clientCtx.GetFromAddress().String(),
				ClientOrderId: clientOrderID,
			}

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CancelReplaceOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancelreplace [original-client-order-id] [source-amount] [destination-amount] [client-orderid]",
		Short: "Update an existing order",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			src, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return
			}

			dst, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return
			}

			origClientOrderID := args[0]
			newClientOrderID := args[3]

			tif, err := cmd.Flags().GetString(flag_TimeInForce)
			if err != nil {
				return err
			}

			timeInForce, err := types.TimeInForceFromString(tif)
			if err != nil {
				return err
			}

			msg := &types.MsgCancelReplaceLimitOrder{
				Owner:             clientCtx.GetFromAddress().String(),
				TimeInForce:       timeInForce,
				Source:            src,
				Destination:       dst,
				OrigClientOrderId: origClientOrderID,
				NewClientOrderId:  newClientOrderID,
			}

			err = msg.ValidateBasic()
			if err != nil {
				return
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(flag_TimeInForce, "GTC", flag_TimeInForceDescription)

	return cmd
}
