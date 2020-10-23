// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"

	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	emtypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/authority"
	"github.com/e-money/em-ledger/x/inflation"

	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/p2p"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	flagNumValidators      = "validators"
	flagOutputDir          = "output-dir"
	flagStartingIPAddress  = "starting-ip-address"
	flagAddKeybaseAccounts = "keyaccounts"

	nodeMonikerTemplate = "node%v"

	defaultKeyPass = "pwd123456"
)

func testnetCmd(ctx *server.Context, cdc *codec.Codec, mbm module.BasicManager) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "testnet [chain-id] [authority_key_or_address] ]",
		Short: "Initialize files for an e-money testnet",
		Long: `testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Example:
	emd testnet -v 4 --output-dir ./output --starting-ip-address 192.168.10.2
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ctx.Config
			chainID := args[0]

			outputDir := viper.GetString(flagOutputDir)

			startingIPAddress := viper.GetString(flagStartingIPAddress)
			numValidators := viper.GetInt(flagNumValidators)
			addKeybaseAccounts := viper.GetString(flagAddKeybaseAccounts)

			authorityKey := getAuthorityKey(args[1], addKeybaseAccounts)

			//return InitTestnet(cmd, config, cdc, mbm, genAccIterator, outputDir, chainID,
			//	minGasPrices, nodeDirPrefix, nodeDaemonHome, nodeCLIHome, startingIPAddress, numValidators)
			return initializeTestnet(cdc, mbm, cfg, outputDir, numValidators, startingIPAddress, addKeybaseAccounts, chainID, authorityKey)
		},
		Args: cobra.ExactArgs(2),
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
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	//cmd.Flags().String(
	//	client.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	//cmd.Flags().String(
	//	server.FlagMinGasPrices, fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
	//	"Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 0.01photino,0.001stake)")
	return cmd
}

func initializeTestnet(
	cdc *codec.Codec, mbm module.BasicManager, config *cfg.Config,
	outputDir string, validatorCount int, baseIPAddress,
	addRandomAccounts, chainID string, authorityKey sdk.AccAddress) error {

	config.Genesis = "genesis.json"

	gen := mbm.DefaultGenesis()
	gen["authority"] = createAuthorityGenesis(authorityKey)
	gen["inflation"] = createInflationGenesis()

	appState, err := codec.MarshalJSONIndent(cdc, gen)
	if err != nil {
		return err
	}

	if chainID == "" {
		chainID = fmt.Sprintf("emoney-%v", tmrand.Str(6))
	}

	genDoc := &tmtypes.GenesisDoc{
		Validators: []tmtypes.GenesisValidator{},
		ChainID:    chainID,
		AppState:   appState,
	}

	nodeIDs := make([]string, validatorCount)
	createValidatorTXs := make([]types.StdTx, validatorCount)
	validatorAccounts := make([]exported.GenesisAccount, validatorCount)

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

	var genaccounts exported.GenesisAccounts
	if addRandomAccounts != "" {
		genaccounts = addRandomTestAccounts(addRandomAccounts)
	}

	// Update genesis file with the created validators
	allAccounts := append(validatorAccounts, genaccounts...)
	addGenesisValidators(cdc, genDoc, createValidatorTXs, allAccounts)

	// Update consensus-parameters
	genDoc.ConsensusParams = tmtypes.DefaultConsensusParams()
	genDoc.ConsensusParams.Block.MaxBytes = 1024 * 1024 * 16
	genDoc.ConsensusParams.Block.TimeIotaMs = 1

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

func createInflationGenesis() json.RawMessage {
	state := inflation.NewInflationState("ejpy", "0.05", "echf", "0.10", "eeur", "0.01", "ungm", "0.1")

	gen := inflation.GenesisState{
		InflationState: state,
	}

	bz, err := json.Marshal(gen)
	if err != nil {
		panic(err)
	}

	return json.RawMessage(bz)
}

func createAuthorityGenesis(akey sdk.AccAddress) json.RawMessage {
	gen := authority.NewGenesisState(akey, emtypes.RestrictedDenoms{}, sdk.NewDecCoins())

	bz, err := json.Marshal(gen)
	if err != nil {
		panic(err)
	}

	return json.RawMessage(bz)
}

func addRandomTestAccounts(keystorepath string) exported.GenesisAccounts {
	kb, err := keys.NewKeyring(sdk.KeyringServiceName(), keys.BackendTest, keystorepath, nil)
	if err != nil {
		panic(err)
	}

	allKeys, err := kb.List()
	if err != nil {
		panic(err)
	}

	result := make(exported.GenesisAccounts, len(allKeys))
	for i, k := range allKeys {
		fmt.Printf("Creating genesis account for key %v.\n", k.GetName())
		coins := sdk.NewCoins(
			sdk.NewCoin("ungm", sdk.NewInt(99000000000)),
			sdk.NewCoin("eeur", sdk.NewInt(10000000000)),
			sdk.NewCoin("ejpy", sdk.NewInt(3500000000000)),
			sdk.NewCoin("echf", sdk.NewInt(10000000000)),
		)

		//genAcc := auth.NewBaseAccount(k.GetAddress(), coins, k.GetPubKey(), 0, 0)
		genAcc := auth.NewBaseAccount(k.GetAddress(), coins, nil, 0, 0)
		//genAcc := exported.NewGenesisAccountRaw(k.GetAddress(), coins, sdk.NewCoins(), 0, 0, "")
		result[i] = genAcc
	}

	return result
}

func createValidatorAccounts(address crypto.Address) exported.GenesisAccount {
	accStakingTokens := sdk.TokensFromConsensusPower(100000)
	account := &auth.BaseAccount{
		Address: sdk.AccAddress(address),
		Coins: sdk.Coins{
			sdk.NewCoin("ungm", accStakingTokens),
		},
	}

	return account
}

func addGenesisValidators(cdc *codec.Codec, genDoc *tmtypes.GenesisDoc, txs []types.StdTx, accounts []exported.GenesisAccount) {
	var appState map[string]json.RawMessage
	if err := cdc.UnmarshalJSON(genDoc.AppState, &appState); err != nil {
		panic(err)
	}

	genutil.SetGenTxsInAppGenesisState(cdc, appState, txs)

	authGenesis := auth.NewGenesisState(auth.DefaultParams(), accounts)
	genesisStateBz := cdc.MustMarshalJSON(authGenesis)
	appState[auth.ModuleName] = genesisStateBz

	genDoc.AppState = cdc.MustMarshalJSON(appState)
}

func createValidatorTransaction(i int, validatorpk crypto.PubKey, chainID string) (types.StdTx, crypto.Address) {
	kb := keys.NewInMemory()
	info, secret, err := kb.CreateMnemonic("nodename", keys.English, defaultKeyPass, keys.Secp256k1)
	if err != nil {
		panic(err)
	}

	moniker := fmt.Sprintf("Validator-%v", i)
	valTokens := sdk.TokensFromConsensusPower(60000)
	msg := staking.NewMsgCreateValidator(
		sdk.ValAddress(info.GetPubKey().Address()),
		validatorpk,
		sdk.NewCoin("ungm", valTokens),
		staking.NewDescription(moniker, "", "", "", ""),
		staking.NewCommissionRates(sdk.NewDecWithPrec(15, 2), sdk.NewDecWithPrec(100, 2), sdk.NewDecWithPrec(100, 2)),
		sdk.OneInt())

	fmt.Printf("Key mnemonic for %v : %v\n", moniker, secret)

	tx := auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, []auth.StdSignature{}, " - ")
	txBldr := auth.NewTxBuilderFromCLI(strings.NewReader("")).WithChainID(chainID).WithMemo(" - ").WithKeybase(kb)
	signedTx, err := txBldr.SignStdTx("nodename", defaultKeyPass, tx, false)

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
	conf.TxIndex.IndexAllKeys = true
	conf.Consensus.TimeoutCommit = 2 * time.Second
	conf.RPC.ListenAddress = "tcp://0.0.0.0:26657"

	cfg.WriteConfigFile(configFilePath, conf)

	appConfigFilePath := filepath.Join(rootDir, "config/app.toml")
	appConf, _ := config.ParseConfig()
	config.WriteConfigFile(appConfigFilePath, appConf)
}

func getAuthorityKey(param string, keystorePath string) sdk.AccAddress {
	key, err := sdk.AccAddressFromBech32(param)
	if err == nil {
		return key
	}

	kb, err := keys.NewKeyring(sdk.KeyringServiceName(), keys.BackendTest, keystorePath, nil)
	if err != nil {
		panic(err)
	}

	keys, err := kb.List()
	if err != nil {
		panic(err)
	}

	for _, key := range keys {
		if key.GetName() == param {
			return key.GetAddress()
		}
	}

	panic(fmt.Errorf("unable to find key %s", param))
}
