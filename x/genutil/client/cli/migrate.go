package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	v09 "github.com/e-money/em-ledger/x/genutil/legacy/v090"
)

const (
	flagGenesisTime        = "genesis-time"
	flagNoConsensusUpgrade = "no-consensus-upgrade"
)

// Allow applications to extend and modify the migration process.
//
// Ref: https://github.com/cosmos/cosmos-sdk/issues/5041
var migrationMap = types.MigrationMap{
	"v0.9": v09.Migrate,
}

// GetMigrationCallback returns a MigrationCallback for a given version.
func GetMigrationCallback(version string) types.MigrationCallback {
	return migrationMap[version]
}

// GetMigrationVersions get all migration version in a sorted slice.
func GetMigrationVersions() []string {
	versions := make([]string, len(migrationMap))

	var i int

	for version := range migrationMap {
		versions[i] = version
		i++
	}

	sort.Strings(versions)

	return versions
}

// MigrateGenesisCmd returns a command to execute genesis state migration.
func MigrateGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [target-version] [genesis-file]",
		Short: "Migrate genesis to a specified target version",
		Long: fmt.Sprintf(`Migrate the source genesis into the target version and print to STDOUT.

Example:
$ %s migrate v0.9 /path/to/genesis.json --chain-id=cosmoshub-3 --genesis-time=2019-04-22T17:00:00Z
`, version.AppName),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			var err error

			target := args[0]
			importGenesis := args[1]

			genDoc, err := validateGenDoc(importGenesis)
			if err != nil {
				return err
			}

			var initialState types.AppMap
			if err := json.Unmarshal(genDoc.AppState, &initialState); err != nil {
				return errors.Wrap(err, "failed to JSON unmarshal initial genesis state")
			}

			migrationFunc := GetMigrationCallback(target)
			if migrationFunc == nil {
				migrations := make([]string, len(migrationMap))
				for k, _ := range migrationMap {
					migrations = append(migrations, k)
				}

				avail := fmt.Sprintf("Available migrations:%v", strings.Join(migrations, " "))
				return fmt.Errorf("unknown migration function for version: %s\n%v", target, avail)
			}

			// TODO: handler error from migrationFunc call
			newGenState := migrationFunc(initialState, clientCtx)

			noUpgrade, _ := cmd.Flags().GetBool(flagNoConsensusUpgrade)
			if !noUpgrade {
				genDoc.ConsensusParams.Evidence.MaxBytes = 4194304

				upgradeModuleParams(clientCtx.JSONMarshaler, newGenState)
			}

			genDoc.AppState, err = json.Marshal(newGenState)
			if err != nil {
				return errors.Wrap(err, "failed to JSON marshal migrated genesis state")
			}

			genesisTime, _ := cmd.Flags().GetString(flagGenesisTime)
			if genesisTime != "" {
				var t time.Time

				err := t.UnmarshalText([]byte(genesisTime))
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal genesis time")
				}

				genDoc.GenesisTime = t
			}

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID != "" {
				genDoc.ChainID = chainID
			}

			bz, err := tmjson.Marshal(genDoc)
			if err != nil {
				return errors.Wrap(err, "failed to marshal genesis doc")
			}

			sortedBz, err := sdk.SortJSON(bz)
			if err != nil {
				return errors.Wrap(err, "failed to sort JSON genesis doc")
			}

			fmt.Println(string(sortedBz))
			return nil
		},
	}

	cmd.Flags().String(flagGenesisTime, "", "override genesis_time with this flag")
	cmd.Flags().String(flags.FlagChainID, "", "override chain_id with this flag")
	cmd.Flags().Bool(flagNoConsensusUpgrade, false, "do not apply planned changes to consensus parameters")

	return cmd
}
