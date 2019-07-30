package main

import (
	app "emoney"
	"emoney/types"
	"io"
	"os"

	tmtypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	DefaultNodeHome = os.ExpandEnv(".")
)

func main() {
	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(types.Bech32PrefixAccAddr, types.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(types.Bech32PrefixValAddr, types.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(types.Bech32PrefixConsAddr, types.Bech32PrefixConsPub)
	config.Seal()

	ctx := server.NewDefaultContext()

	viper.Set("consensus.create_empty_blocks_interval", "60s")
	viper.Set("consensus.create_empty_blocks", false)
	viper.Set("consensus.timeout_commit", "0s")

	rootCmd := &cobra.Command{
		Use:               "daemon",
		Short:             "e-money validator node",
		PersistentPreRunE: persistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(initCmd(ctx, cdc, app.ModuleBasics))
	rootCmd.AddCommand(addGenesisAccountCmd(ctx, cdc, DefaultNodeHome, DefaultNodeHome))
	rootCmd.AddCommand(testnetCmd(ctx, cdc, app.ModuleBasics, nil))

	server.AddCommands(ctx, cdc, rootCmd, newApp, nil)

	executor := cli.PrepareBaseCmd(rootCmd, "TMSND", ".")
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db db.DB, _ io.Writer) tmtypes.Application {
	return app.NewApp(logger, db)
}
