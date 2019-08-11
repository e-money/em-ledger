package main

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/tendermint/tendermint/crypto"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	ckeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/common"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/p2p"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	flagNumValidators      = "validators"
	flagOutputDir          = "output-dir"
	flagStartingIPAddress  = "starting-ip-address"
	flagAddKeybaseAccounts = "keyaccounts"

	nodeMonikerTemplate = "node%v"
)

func testnetCmd(ctx *server.Context, cdc *codec.Codec, mbm module.BasicManager) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Initialize files for an e-money testnet",
		Long: `testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Example:
	emd testnet -v 4 --output-dir ./output --starting-ip-address 192.168.10.2
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config := ctx.Config

			outputDir := viper.GetString(flagOutputDir)

			//chainID := viper.GetString(client.FlagChainID)
			//minGasPrices := viper.GetString(server.FlagMinGasPrices)
			//nodeDaemonHome := viper.GetString(flagNodeDaemonHome)
			//nodeCLIHome := viper.GetString(flagNodeCLIHome)
			startingIPAddress := viper.GetString(flagStartingIPAddress)
			numValidators := viper.GetInt(flagNumValidators)
			addKeybaseAccounts := viper.GetString(flagAddKeybaseAccounts)

			//return InitTestnet(cmd, config, cdc, mbm, genAccIterator, outputDir, chainID,
			//	minGasPrices, nodeDirPrefix, nodeDaemonHome, nodeCLIHome, startingIPAddress, numValidators)
			return initializeTestnet(cdc, mbm, config, outputDir, numValidators, startingIPAddress, addKeybaseAccounts)
		},
	}

	cmd.Flags().IntP(flagNumValidators, "v", 4,
		"Number of validators to initialize the testnet with")
	cmd.Flags().StringP(flagOutputDir, "o", "./testnet",
		"Directory to store initialization data for the testnet")
	cmd.Flags().String(flagAddKeybaseAccounts, "", "Generate accounts for each key in the keystore at the specified path.")
	cmd.Flags().Lookup(flagAddKeybaseAccounts).NoOptDefVal = ""

	//cmd.Flags().String(flagNodeDaemonHome, "gaiad",
	//	"Home directory of the node's daemon configuration")
	//cmd.Flags().String(flagNodeCLIHome, "gaiacli",
	//	"Home directory of the node's cli configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.10.2",
		"Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	//cmd.Flags().String(
	//	client.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	//cmd.Flags().String(
	//	server.FlagMinGasPrices, fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
	//	"Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 0.01photino,0.001stake)")
	return cmd
}

func initializeTestnet(cdc *codec.Codec, mbm module.BasicManager, config *cfg.Config, outputDir string, validatorCount int, baseIPAddress string, addRandomAccounts string) error {
	config.Genesis = "genesis.json"

	appState, err := codec.MarshalJSONIndent(cdc, mbm.DefaultGenesis())
	if err != nil {
		return err
	}

	chainID := fmt.Sprintf("emoney-%v", common.RandStr(6))
	genDoc := &tmtypes.GenesisDoc{
		Validators: []tmtypes.GenesisValidator{},
		ChainID:    chainID,
		AppState:   appState,
	}

	nodeIDs := make([]string, validatorCount)
	createValidatorTXs := make([]types.StdTx, validatorCount)
	validatorAccounts := make([]genaccounts.GenesisAccount, validatorCount)

	for i := 0; i < validatorCount; i++ {
		nodeMoniker := fmt.Sprintf(nodeMonikerTemplate, i)
		nodeDir := filepath.Join(outputDir, nodeMoniker)
		config.SetRoot(nodeDir)
		config.Moniker = nodeMoniker

		createConfigurationFiles(nodeDir)
		_, pk, err := genutil.InitializeNodeValidatorFiles(config)
		if err != nil {
			panic(err)
		}

		nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
		if err != nil {
			panic(err)
		}
		nodeIDs[i] = string(nodeKey.ID())

		tx, validatorAccountAddress := createValidatorTransaction(i, pk, chainID)
		createValidatorTXs[i] = tx
		validatorAccounts[i] = createValidatorAccounts(validatorAccountAddress)
	}

	var genaccounts genaccounts.GenesisAccounts
	if addRandomAccounts != "" {
		genaccounts = addRandomTestAccounts(addRandomAccounts)
	}

	// Update genesis file with the created validators
	allAccounts := append(validatorAccounts, genaccounts...)
	addGenesisValidators(cdc, genDoc, createValidatorTXs, allAccounts)

	for i := 0; i < validatorCount; i++ {
		// Add genesis file to each node directory
		nodeMoniker := fmt.Sprintf(nodeMonikerTemplate, i)
		nodeDir := filepath.Join(outputDir, nodeMoniker)
		genFile := filepath.Join(nodeDir, "config", "genesis.json")

		if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
			return err
		}

		// Update config.toml with peer lists
		updateConfigWithPeers(nodeDir, i, nodeIDs, baseIPAddress)
		if i != 0 {
			updateLoggingConfig(nodeDir)
		}
	}

	return nil
}

func addRandomTestAccounts(keystorepath string) genaccounts.GenesisAccounts {
	kb, err := keys.NewKeyBaseFromDir(keystorepath)
	if err != nil {
		panic(err)
	}

	keys, err := kb.List()
	if err != nil {
		panic(err)
	}

	result := make(genaccounts.GenesisAccounts, len(keys))
	for i, k := range keys {
		fmt.Printf("Creating genesis account for key %v.\n", k.GetName())
		coins := sdk.NewCoins(
			sdk.NewCoin("ungm", sdk.TokensFromConsensusPower(100)),
			sdk.NewCoin("caps", sdk.TokensFromConsensusPower(5000)),
			sdk.NewCoin("kredits", sdk.TokensFromConsensusPower(2700)),
		)

		genAcc := genaccounts.NewGenesisAccountRaw(k.GetAddress(), coins, sdk.NewCoins(), 0, 0, "")
		result[i] = genAcc
	}

	return result
}

func createValidatorAccounts(address crypto.Address) genaccounts.GenesisAccount {
	accStakingTokens := sdk.TokensFromConsensusPower(500)
	account := genaccounts.GenesisAccount{
		Address: sdk.AccAddress(address),
		Coins: sdk.Coins{
			sdk.NewCoin("ungm", accStakingTokens),
		},
	}

	return account
}

func addGenesisValidators(cdc *codec.Codec, genDoc *tmtypes.GenesisDoc, txs []types.StdTx, accounts []genaccounts.GenesisAccount) {
	var appState map[string]json.RawMessage
	if err := cdc.UnmarshalJSON(genDoc.AppState, &appState); err != nil {
		panic(err)
	}
	genutil.SetGenesisStateInAppState(cdc, appState, genutil.NewGenesisStateFromStdTx(txs))
	genaccounts.SetGenesisStateInAppState(cdc, appState, accounts)

	genDoc.AppState = cdc.MustMarshalJSON(appState)
}

func createValidatorTransaction(i int, validatorpk crypto.PubKey, chainID string) (types.StdTx, crypto.Address) {
	kb := keys.NewInMemoryKeyBase()
	info, secret, err := kb.CreateMnemonic("nodename", ckeys.English, "12345678", ckeys.Secp256k1)
	if err != nil {
		panic(err)
	}

	moniker := fmt.Sprintf("Validator-%v", i)
	valTokens := sdk.TokensFromConsensusPower(1)
	msg := staking.NewMsgCreateValidator(
		sdk.ValAddress(info.GetPubKey().Address()),
		validatorpk,
		sdk.NewCoin("ungm", valTokens),
		staking.NewDescription(moniker, "", "", ""),
		staking.NewCommissionRates(sdk.NewDecWithPrec(15, 2), sdk.NewDecWithPrec(100, 2), sdk.NewDecWithPrec(100, 2)),
		sdk.OneInt())

	// TODO Write mnemonic to file in the validator directory.
	fmt.Printf("Key mnemonic for %v : %v\n", moniker, secret)

	tx := auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, []auth.StdSignature{}, " - ")
	txBldr := auth.NewTxBuilderFromCLI().WithChainID(chainID).WithMemo(" - ").WithKeybase(kb)
	signedTx, err := txBldr.SignStdTx("nodename", client.DefaultKeyPass, tx, false)

	if err != nil {
		panic(err)
	}

	return signedTx, info.GetPubKey().Address()
}

// Remove emz-module logging from all but the first node.
func updateLoggingConfig(nodeDir string) {
	configFilePath := filepath.Join(nodeDir, "config/config.toml")

	configFile := viper.New()
	configFile.SetConfigFile(configFilePath)
	err := configFile.ReadInConfig()
	if err != nil {
		panic(err)
	}

	logLevel := configFile.Get("log_level").(string)
	configFile.Set("log_level", strings.Replace(logLevel, "emz:info,", "", 1))

	err = configFile.WriteConfig()
	if err != nil {
		panic(err)
	}
}

func updateConfigWithPeers(nodeDir string, i int, nodeIDs []string, baseIPAddress string) {
	configFilePath := filepath.Join(nodeDir, "config/config.toml")

	configFile := viper.New()
	configFile.SetConfigFile(configFilePath)
	err := configFile.ReadInConfig()
	if err != nil {
		panic(err)
	}

	peers := make([]string, 0)
	for j := 0; j < len(nodeIDs); j++ {
		if j == i {
			continue
		}

		peer := fmt.Sprintf("%v@%v:%v", nodeIDs[j], nodeIPAddress(j, baseIPAddress), 26656)
		peers = append(peers, peer)
	}

	peerList := strings.Join(peers, ",")

	configFile.Set("p2p.persistent_peers", peerList)
	configFile.Set("p2p.laddr", fmt.Sprintf("tcp://%v:%v", nodeIPAddress(i, baseIPAddress), 26656))
	err = configFile.WriteConfig()
	if err != nil {
		panic(err)
	}
}

func nodeIPAddress(i int, baseIPAddress string) string {
	ip := net.ParseIP(baseIPAddress).To4() // Only IPv4 for now.
	ip[3] += byte(i)

	return ip.String()
}

func createConfigurationFiles(rootDir string) {
	cfg.EnsureRoot(rootDir)
	// Create config.toml to control Tendermint options
	configFilePath := filepath.Join(rootDir, "config/config.toml")

	conf, _ := tcmd.ParseConfig() // NOTE: ParseConfig() creates dir/files as necessary.
	conf.ProfListenAddress = "localhost:6060"
	conf.P2P.RecvRate = 5120000
	conf.P2P.SendRate = 5120000
	conf.TxIndex.IndexAllTags = true
	conf.Consensus.TimeoutCommit = 2 * time.Second
	conf.RPC.ListenAddress = "tcp://0.0.0.0:26657"

	cfg.WriteConfigFile(configFilePath, conf)

	appConfigFilePath := filepath.Join(rootDir, "config/app.toml")
	appConf, _ := config.ParseConfig()
	config.WriteConfigFile(appConfigFilePath, appConf)
}
