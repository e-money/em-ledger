package cli_test

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	"github.com/e-money/em-ledger/x/authority/client/cli"
	"github.com/e-money/em-ledger/x/authority/types"

	"testing"

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
	"github.com/stretchr/testify/suite"
)

type ReplacementTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *ReplacementTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.LegacyAmino.RegisterConcrete(&types.MsgReplaceAuthority{}, "MsgReplaceAuthority", nil)
	cfg.InterfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &types.MsgReplaceAuthority{})
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

func (s *ReplacementTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *ReplacementTestSuite) TestReplaceAuth() {
	val1 := *s.network.Validators[0]

	// Generate 2 accounts and a multisig.
	const (
		acc1UID    = "newAccount1"
		acc2UID    = "newAccount2"
		acc3UID    = "newAccount3"
		msigUID    = "multi"
		newMsigUID = "newMulti"
	)

	account1, err := val1.ClientCtx.Keyring.Key(acc1UID)
	s.Require().NoError(err)
	fmt.Println("acc 1:", account1.GetAddress().String())

	account2, err := val1.ClientCtx.Keyring.Key(acc2UID)
	s.Require().NoError(err)
	fmt.Println("acc 2:", account2.GetAddress().String())

	authMultiSigAcc, err := val1.ClientCtx.Keyring.Key(msigUID)
	s.Require().NoError(err)
	s.Require().Equal(keyring.TypeMulti, authMultiSigAcc.GetType())
	fmt.Println("multi:", authMultiSigAcc.GetAddress().String())

	var (
		pks            []cryptotypes.PubKey
		authorityAddrs []string
	)

	// create the new multisig authority key adding a 3rd key
	kb := val1.ClientCtx.Keyring
	account3, _, err := kb.NewMnemonic(acc3UID, keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)
	s.Require().NotNil(account3)

	// test getting it from the keystore
	account3, err = val1.ClientCtx.Keyring.Key(acc3UID)
	s.Require().NoError(err)
	s.Require().NotNil(account3)
	fmt.Println("acc 3:", account3.GetAddress().String())

	authkeys := []string{acc1UID, acc2UID, acc3UID}
	multisigThreshold := len(authkeys) - 1
	// generate new multisig authority key
	for _, keyname := range authkeys {
		k, err := kb.Key(keyname)
		s.Require().NoError(err)

		pks = append(pks, k.GetPubKey())
		authorityAddrs = append(authorityAddrs, k.GetAddress().String())
	}

	sort.Slice(
		pks, func(i, j int) bool {
			return bytes.Compare(pks[i].Address(), pks[j].Address()) < 0
		},
	)

	pk := kmultisig.NewLegacyAminoPubKey(multisigThreshold, pks)
	newAuthMultiSigAcc, err := kb.SaveMultisig(newMsigUID, pk)
	s.Require().NoError(err)
	s.Require().Equal(keyring.TypeMulti, newAuthMultiSigAcc.GetType())

	fmt.Println("New authority multisig key:", newAuthMultiSigAcc.GetAddress().String())

	// set the multisig account in the state
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)
	_, err = bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		authMultiSigAcc.GetAddress(),
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)

	// Generate the unsigned multisig transaction json with the existing authority key.
	args := []string{
		authMultiSigAcc.GetAddress().String(),
		newAuthMultiSigAcc.GetAddress().String(),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	}

	multiGeneratedTx, err := sdkcli.ExecTestCLICmd(val1.ClientCtx, cli.GetCmdReplaceAuthority(), args)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())

	// Sign with account1
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtest.TxSignExec(
		val1.ClientCtx, account1.GetAddress(), multiGeneratedTxFile.Name(),
		"--multisig", authMultiSigAcc.GetAddress().String(),
	)
	s.Require().NoError(err)

	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())

	// Sign with account2
	account2Signature, err := authtest.TxSignExec(
		val1.ClientCtx, account2.GetAddress(), multiGeneratedTxFile.Name(),
		"--multisig", authMultiSigAcc.GetAddress().String(),
	)
	s.Require().NoError(err)

	sign2File := testutil.WriteToNewTempFile(s.T(), account2Signature.String())

	// Does not work in offline mode.
	_, err = authtest.TxMultiSignExec(
		val1.ClientCtx, authMultiSigAcc.GetName(), multiGeneratedTxFile.Name(),
		"--offline", sign1File.Name(), sign2File.Name(),
	)
	s.Require().EqualError(
		err,
		"couldn't verify signature: unable to verify single signer signature",
	)

	val1.ClientCtx.Offline = false
	multiSigWith2Signatures, err := authtest.TxMultiSignExec(
		val1.ClientCtx, authMultiSigAcc.GetName(), multiGeneratedTxFile.Name(),
		sign1File.Name(), sign2File.Name(),
	)
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

func TestReplacementTestSuite(t *testing.T) {
	suite.Run(t, new(ReplacementTestSuite))
}
