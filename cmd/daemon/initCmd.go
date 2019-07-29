package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"
)

func initCmd(ctx *server.Context, cdc *codec.Codec, mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize genesis configuration",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			config.Moniker = "Node1"
			config.SetRoot(viper.GetString(cli.HomeFlag))
			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

			_, pk, err := genutil.InitializeNodeValidatorFiles(config)
			if err != nil {
				return err
			}

			_, _, validator, err := simpleAppGenTx(cdc, pk)
			if err != nil {
				return err
			}

			appState, err := codec.MarshalJSONIndent(cdc, mbm.DefaultGenesis())
			if err != nil {
				return err
			}

			chainID := fmt.Sprintf("emoney-%v", common.RandStr(6))
			genDoc := &tmtypes.GenesisDoc{
				Validators: []tmtypes.GenesisValidator{validator},
				ChainID:    chainID,
				AppState:   appState,
			}

			genFile := config.GenesisFile()
			if _, err := os.Stat(genFile); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				genDoc, err = tmtypes.GenesisDocFromFile(genFile)
				if err != nil {
					return err
				}
			}

			if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String(cli.HomeFlag, DefaultNodeHome, "node's home directory")

	return cmd
}

func simpleAppGenTx(cdc *codec.Codec, pk crypto.PubKey) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	addr, secret, err := server.GenerateCoinKey()
	if err != nil {
		return
	}

	bz, err := cdc.MarshalJSON(struct {
		Addr sdk.AccAddress `json:"addr"`
	}{addr})
	if err != nil {
		return
	}

	appGenTx = json.RawMessage(bz)

	bz, err = cdc.MarshalJSON(map[string]string{"secret": secret})
	if err != nil {
		return
	}

	cliPrint = json.RawMessage(bz)

	validator = tmtypes.GenesisValidator{
		Address: pk.Address(),
		PubKey:  pk,
		Power:   10,
	}

	return
}
