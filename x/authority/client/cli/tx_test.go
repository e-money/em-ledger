package cli_test

import (
	"fmt"
	"github.com/e-money/em-ledger/x/authority/client/cli"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdkcli "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtest "github.com/cosmos/cosmos-sdk/x/auth/client/testutil"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 2

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	kb := s.network.Validators[0].ClientCtx.Keyring
	_, _, err := kb.NewMnemonic("newAccount", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	account1, _, err := kb.NewMnemonic("newAccount1", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	account2, _, err := kb.NewMnemonic("newAccount2", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	multi := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{account1.GetPubKey(), account2.GetPubKey()})
	_, err = kb.SaveMultisig("multi", multi)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestCLIMultisign() {
	val1 := *s.network.Validators[0]

	// Generate 2 accounts and a multisig.
	account1, err := val1.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	account2, err := val1.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	multisigInfo, err := val1.ClientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)
	_, err = bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		multisigInfo.GetAddress(),
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)

	s.Require().NoError(s.network.WaitForNextBlock())

	resp, err := bankcli.QueryBalancesExec(val1.ClientCtx, multisigInfo.GetAddress())
	s.Require().NoError(err)

	var balRes banktypes.QueryAllBalancesResponse
	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(resp.Bytes(), &balRes)
	s.Require().NoError(err)
	s.Require().Equal(sendTokens.Amount, balRes.Balances.AmountOf(s.cfg.BondDenom))

	// Generate multisig transaction.
	multiGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		multisigInfo.GetAddress(),
		val1.Address,
		sdk.NewCoins(
			sdk.NewInt64Coin(s.cfg.BondDenom, 5),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())

	// Sign with account1
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtest.TxSignExec(val1.ClientCtx, account1.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())

	// Sign with account1
	account2Signature, err := authtest.TxSignExec(val1.ClientCtx, account2.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign2File := testutil.WriteToNewTempFile(s.T(), account2Signature.String())

	// Does not work in offline mode.
	_, err = authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), "--offline", sign1File.Name(), sign2File.Name())
	s.Require().EqualError(err, "couldn't verify signature: unable to verify single signer signature")

	val1.ClientCtx.Offline = false
	multiSigWith2Signatures, err := authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())

	_, err = authtest.TxValidateSignaturesExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	val1.ClientCtx.BroadcastMode = flags.BroadcastBlock
	_, err = authtest.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *IntegrationTestSuite) TestMultisign() {
	val1 := *s.network.Validators[0]

	// Generate 2 accounts and a multisig.
	account1, err := val1.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	account2, err := val1.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	multisigInfo, err := val1.ClientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	args := []string{
		multisigInfo.GetAddress().String(),
		fmt.Sprintf(
			"%s,%s", account1.GetAddress().String(),
			account2.GetAddress().String(),
		),
		"2",
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	}

	_, err = sdkcli.ExecTestCLICmd(val1.ClientCtx, cli.GetCmdReplaceAuthority(),args)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())

	//resp, err := bankcli.QueryBalancesExec(val1.ClientCtx, multisigInfo.GetAddress())
	//s.Require().NoError(err)

	//var balRes banktypes.QueryAllBalancesResponse
	//err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(resp.Bytes(), &balRes)
	//s.Require().NoError(err)
	// s.Require().Equal(sendTokens.Amount, balRes.Balances.AmountOf(s.cfg.BondDenom))

	/************************************************************
	// Generate multisig transaction.
	multiGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		multisigInfo.GetAddress(),
		val1.Address,
		sdk.NewCoins(
			sdk.NewInt64Coin(s.cfg.BondDenom, 5),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())

	// Sign with account1
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtest.TxSignExec(val1.ClientCtx, account1.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())

	// Sign with account1
	account2Signature, err := authtest.TxSignExec(val1.ClientCtx, account2.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign2File := testutil.WriteToNewTempFile(s.T(), account2Signature.String())

	// Does not work in offline mode.
	_, err = authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), "--offline", sign1File.Name(), sign2File.Name())
	s.Require().EqualError(err, "couldn't verify signature: unable to verify single signer signature")

	val1.ClientCtx.Offline = false
	multiSigWith2Signatures, err := authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())

	_, err = authtest.TxValidateSignaturesExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	val1.ClientCtx.BroadcastMode = flags.BroadcastBlock
	_, err = authtest.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())

	********************************************/
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}