package main

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"os"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	app "emoney"
	apptypes "emoney/types"
	"emoney/util"
	issuercli "emoney/x/issuer/client/cli"
	lpcli "emoney/x/liquidityprovider/client/cli"
	lptypes "emoney/x/liquidityprovider/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/version"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/spf13/cobra"
)

func main() {
	cobra.EnableCommandSorting = false

	apptypes.ConfigureSDK()
	cdc := app.MakeCodec()

	rootCmd := &cobra.Command{
		Use:   "emcli",
		Short: "Command line interface for interacting with e-money daemon",
	}
	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")

	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCmds(cdc),
		client.ConfigCmd(app.DefaultCLIHome),
		txCmds(cdc),
		lcd.ServeCommand(cdc, registerLCDRoutes),
		keys.Commands(),
		lpcli.GetTxCmd(cdc),
		issuercli.GetTxCmd(cdc),
		version.Cmd,
	)

	// Remove commands for functionality that is not supported or superfluous to the e-money zone
	util.RemoveCobraCommands(rootCmd,
		"query.distribution.community-pool",
	)

	executor := cli.PrepareMainCmd(rootCmd, "GA", app.DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		fmt.Printf("Failed executing CLI command: %s, exiting...\n", err)
		os.Exit(1)
	}
}

func init() {
	registerTypesInAuthModule()
}

func registerTypesInAuthModule() {
	// The auth module's codec must be updated with the account types introduced by the liquidityprovider module
	// When https://github.com/cosmos/cosmos-sdk/pull/5017 is in the used Cosmos-sdk, consider switching to it.
	// https://github.com/cosmos/cosmos-sdk/blob/1d16d34b1b35cb65405f84b632d228ed8fc329fc/docs/architecture/adr-011-generalize-genesis-accounts.md
	authcdc := codec.New()

	codec.RegisterCrypto(authcdc)
	lptypes.RegisterCodec(authcdc)
	authtypes.RegisterCodec(authcdc)

	authtypes.ModuleCdc = authcdc
	auth.ModuleCdc = authcdc
}

func txCmds(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		authcmd.GetSignCommand(cdc),
		authcmd.GetMultiSignCommand(cdc),
		authcmd.GetBroadcastCommand(cdc),
	)

	app.ModuleBasics.AddTxCommands(txCmd, cdc)

	// remove bank command as it's already mounted under the root tx command
	for _, cmd := range txCmd.Commands() {
		if cmd.Use == bank.ModuleName {
			txCmd.RemoveCommand(cmd)
			break
		}
	}

	return txCmd
}

func queryCmds(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		authcmd.GetAccountCmd(cdc),
	)

	app.ModuleBasics.AddQueryCommands(queryCmd, cdc)
	return queryCmd
}

func registerLCDRoutes(rs *lcd.RestServer) {
	client.RegisterRoutes(rs.CliCtx, rs.Mux)
	authrest.RegisterTxRoutes(rs.CliCtx, rs.Mux)
	app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)
}
