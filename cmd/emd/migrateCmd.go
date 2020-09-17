package main

import (
	"fmt"
	v039 "github.com/cosmos/cosmos-sdk/x/genutil/legacy/v0_39"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	extypes "github.com/cosmos/cosmos-sdk/x/genutil"
	v038 "github.com/cosmos/cosmos-sdk/x/genutil/legacy/v0_38"

	tmmigrate "github.com/e-money/em-ledger/migration/tendermint"
)

const (
	flagGenesisTime = "genesis-time"
	flagChainId     = "chain-id"
)

func MigrateGenesisCmd(cdc *codec.Codec, writer io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [genesis-file]",
		Short: "Migrate export file from emoney-1 to genesis file for emoney-2",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			importGenesis := args[0]

			// Migrate the Tendermint consensus configuration from v0.32.x to v0.33.x
			genDoc, err := tmmigrate.ToV033(importGenesis)
			if err != nil {
				return err
			}

			var initialState extypes.AppMap
			cdc.MustUnmarshalJSON(genDoc.AppState, &initialState)

			// Let the standard libraries do their thing
			newGenState := v038.Migrate(initialState)
			newGenState = v039.Migrate(newGenState)

			genDoc.AppState = cdc.MustMarshalJSON(newGenState)

			genesisTime := cmd.Flag(flagGenesisTime).Value.String()
			if genesisTime != "" {
				var t time.Time

				err := t.UnmarshalText([]byte(genesisTime))
				if err != nil {
					return err
				}

				genDoc.GenesisTime = t
			}

			chainId := cmd.Flag(flagChainId).Value.String()
			if chainId != "" {
				genDoc.ChainID = chainId
			}

			out, err := cdc.MarshalJSONIndent(genDoc, "", "  ")
			if err != nil {
				return err
			}

			_, err = fmt.Fprint(writer, string(sdk.MustSortJSON(out)))
			return err
		},
	}

	cmd.Flags().String(flagGenesisTime, "", "Override genesis_time with this flag")
	cmd.Flags().String(flagChainId, "", "Override chain_id with this flag")

	return cmd
}
