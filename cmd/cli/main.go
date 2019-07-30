package main

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"os"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	app "emoney"
	"emoney/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"

	"github.com/spf13/cobra"
)

var (
	DefaultCLIHome = os.ExpandEnv(".")
)

func main() {
	cobra.EnableCommandSorting = false
	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(types.Bech32PrefixAccAddr, types.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(types.Bech32PrefixValAddr, types.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(types.Bech32PrefixConsAddr, types.Bech32PrefixConsPub)
	config.Seal()

	rootCmd := &cobra.Command{
		Use:   "emcli",
		Short: "Command line interface for interacting with e-money daemon",
	}
	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")

	rootCmd.AddCommand(queryCmds(cdc), txCmds(cdc), keys.Commands())

	executor := cli.PrepareMainCmd(rootCmd, "GA", DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		fmt.Printf("Failed executing CLI command: %s, exiting...\n", err)
		os.Exit(1)
	}
}

func txCmds(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(bankcmd.SendTxCmd(cdc))

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

	return queryCmd
}
