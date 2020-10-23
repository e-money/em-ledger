// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package main

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/spf13/viper"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	app "github.com/e-money/em-ledger"
	apptypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/authority"
	issuercli "github.com/e-money/em-ledger/x/issuer/client/cli"
	lpcli "github.com/e-money/em-ledger/x/liquidityprovider/client/cli"
	lptypes "github.com/e-money/em-ledger/x/liquidityprovider/types"
	marketcli "github.com/e-money/em-ledger/x/market/client/cli"
	queries "github.com/e-money/em-ledger/x/queries/client/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	ckeys "github.com/cosmos/cosmos-sdk/crypto/keys"
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
	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "Chain ID of tendermint node")

	keysRootCmd := keys.Commands()
	// Configured erroneously in v0.39.1. Seems to be fixed in future releases. (https://github.com/cosmos/cosmos-sdk/blob/v0.39.1/client/keys/root.go#L36)
	viper.BindPFlag(flags.FlagKeyringBackend, keysRootCmd.PersistentFlags().Lookup(flags.FlagKeyringBackend))

	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCmds(cdc),
		client.ConfigCmd(app.DefaultCLIHome),
		txCmds(cdc),
		lcd.ServeCommand(cdc, registerLCDRoutes),
		keysRootCmd,

		version.Cmd,
	)

	// Remove commands for functionality that is not supported or superfluous to the e-money zone
	util.RemoveCobraCommands(rootCmd,
		"query.distribution.community-pool",
	)

	viper.SetDefault(flags.FlagBroadcastMode, "block")
	viper.SetDefault(flags.FlagKeyringBackend, ckeys.BackendFile)
	// TODO Appears to be necessary after the upgrade from cosmos-sdk v0.37.3 -> v0.37.8
	// TODO This also upgraded viper v1.5.0 -> v1.6.1 which may be the cause of change in behaviour
	// TODO The createVerifier() funcion in cosmos-sdk@v0.37.8/client/context/context.go seems to be the issue
	viper.SetDefault(flags.FlagTrustNode, "false")

	overrideDefaults(rootCmd)

	executor := cli.PrepareMainCmd(rootCmd, "EM", app.DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		fmt.Printf("Failed executing CLI command: %s, exiting...\n", err)
		os.Exit(1)
	}
}

func init() {
	registerTypesInAuthModule()
}

// Change some of the default values for emcli usage flags:
//  - Switch the default value of --broadcast-mode to "block"
//  - Switch the default value of --keyring-backend to "file"
func overrideDefaults(cmd *cobra.Command) {
	if flag := cmd.Flag(flags.FlagBroadcastMode); flag != nil {
		flag.DefValue = flags.BroadcastBlock
	}

	if flag := cmd.Flag(flags.FlagKeyringBackend); flag != nil {
		flag.DefValue = ckeys.BackendFile
	}

	for _, child := range cmd.Commands() {
		overrideDefaults(child)
	}
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
		marketcli.GetTxCmd(cdc),
		lpcli.GetTxCmd(cdc),
		issuercli.GetTxCmd(cdc),
		authority.GetTxCmd(cdc),
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
		queries.GetQuerySpendableBalance(cdc),
		authcmd.QueryTxCmd(cdc),
		authcmd.QueryTxsByEventsCmd(cdc),
	)

	// Make sure account querying supports vesting accounts.
	authtypes.ModuleCdc = cdc

	app.ModuleBasics.AddQueryCommands(queryCmd, cdc)

	// Extend some of the standard SDK queries
	cmd, _, _ := queryCmd.Find([]string{"supply"})
	cmd.AddCommand(queries.GetQueryCirculatingSupplyCmd(cdc))

	return queryCmd
}

func registerLCDRoutes(rs *lcd.RestServer) {
	client.RegisterRoutes(rs.CliCtx, rs.Mux)
	authrest.RegisterTxRoutes(rs.CliCtx, rs.Mux)
	app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)
}
