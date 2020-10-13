// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/tendermint/tendermint/types"

	app "github.com/e-money/em-ledger"
	apptypes "github.com/e-money/em-ledger/types"

	tmtypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/server"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

var configureConsensus = func() {
	viper.Set("consensus.create_empty_blocks_interval", "60s")
	viper.Set("consensus.create_empty_blocks", false)
	viper.Set("consensus.timeout_commit", "500ms")
	viper.Set("consensus.timeout_propose", "2s")
	viper.Set("consensus.peer_gossip_sleep_duration", "25ms")
}

func main() {
	cobra.EnableCommandSorting = false

	apptypes.ConfigureSDK()
	cdc := app.MakeCodec()

	ctx := server.NewDefaultContext()
	// Add application to logging configuration
	logLevel := ctx.Config.BaseConfig.LogLevel
	ctx.Config.BaseConfig.LogLevel = fmt.Sprintf("emz:info,x/inflation:info,x/liquidityprovider:info,%v", logLevel)

	configureConsensus()
	viper.Set("p2p.flush_throttle_timeout", "25ms")

	rootCmd := &cobra.Command{
		Use:               "emd",
		Short:             "e-money validator node",
		PersistentPreRunE: persistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(genutilcli.InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome))
	rootCmd.AddCommand(addGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome))
	rootCmd.AddCommand(testnetCmd(ctx, cdc, app.ModuleBasics))

	server.AddCommands(ctx, cdc, rootCmd, newAppCreator(ctx), newAppExporter(ctx))

	executor := cli.PrepareBaseCmd(rootCmd, "EMD", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newAppCreator(ctx *server.Context) func(log.Logger, db.DB, io.Writer) tmtypes.Application {
	return func(logger log.Logger, db db.DB, _ io.Writer) tmtypes.Application {
		return app.NewApp(logger, db, ctx,
			baseapp.SetPruning(store.NewPruningOptionsFromString(viper.GetString("pruning"))),
			baseapp.SetHaltHeight(uint64(viper.GetInt(server.FlagHaltHeight))),
			baseapp.SetHaltTime(uint64(viper.GetInt(server.FlagHaltTime))),
		)
	}
}

func newAppExporter(ctx *server.Context) server.AppExporter {
	return func(logger log.Logger, db db.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
	) (json.RawMessage, []types.GenesisValidator, error) {

		if height != -1 {
			a := app.NewApp(logger, db, ctx)
			err := a.LoadHeight(height)
			if err != nil {
				return nil, nil, err
			}

			return a.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
		}

		a := app.NewApp(logger, db, ctx)
		return a.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}
}
