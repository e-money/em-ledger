package ante_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/e-money/em-ledger/x/auth/ante"
	"github.com/e-money/em-ledger/x/buyback"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

func (suite *AnteTestSuite) TestInsufficientFunds() {
	suite.setup()
	ctx := suite.ctx

	payerAccount := suite.createAccount(ctx, coins("250ungm"))

	tx := mockFeeTX{fee: coins("500ungm"), feePayer: payerAccount.GetAddress()}
	_, err := suite.anteHandler(ctx, tx, false)

	require.True(suite.T(), sdkerrors.ErrInsufficientFunds.Is(err))
}

func (suite *AnteTestSuite) TestDeductFees() {
	suite.setup()
	ctx := suite.ctx

	payerAccount := suite.createAccount(ctx, coins("2500ungm,19000eeur"))

	{ // Pay fee with staking token
		tx := mockFeeTX{fee: coins("500ungm"), feePayer: payerAccount.GetAddress()}

		_, err := suite.anteHandler(ctx, tx, false)
		require.NoError(suite.T(), err)

		rewardsBalance := suite.getModuleBalance(ctx, authtypes.FeeCollectorName)
		require.Equal(suite.T(), rewardsBalance, coins("500ungm"))

		buybackBalance := suite.getModuleBalance(ctx, buyback.ModuleName)
		require.True(suite.T(), buybackBalance.IsZero())
	}

	{ // Pay fee with stablecoin token
		tx := mockFeeTX{fee: coins("400eeur"), feePayer: payerAccount.GetAddress()}
		_, err := suite.anteHandler(ctx, tx, false)
		require.NoError(suite.T(), err)

		rewardsBalance := suite.getModuleBalance(ctx, authtypes.FeeCollectorName)
		require.Equal(suite.T(), coins("500ungm"), rewardsBalance)

		buybackBalance := suite.getModuleBalance(ctx, buyback.ModuleName)
		require.Equal(suite.T(), coins("400eeur").String(), buybackBalance.String())
	}
}

func (suite *AnteTestSuite) TestMultipleFeeDenoms() {
	suite.setup()
	ctx := suite.ctx

	payerAccount := suite.createAccount(ctx, coins("500ungm,8000eeur,1500chf"))
	tx := mockFeeTX{fee: coins("500ungm,5000eeur,450chf"), feePayer: payerAccount.GetAddress()}

	_, err := suite.anteHandler(ctx, tx, false)
	require.NoError(suite.T(), err)

	rewardsBalance := suite.getModuleBalance(ctx, authtypes.FeeCollectorName)
	require.Equal(suite.T(), coins("500ungm"), rewardsBalance)

	buybackBalance := suite.getModuleBalance(ctx, buyback.ModuleName)
	require.Equal(suite.T(), coins("450chf,5000eeur").String(), buybackBalance.String())
}

type AnteTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	ak          authkeeper.AccountKeeper
	bk          bankkeeper.BaseKeeper
	anteHandler sdk.AnteHandler
}

func (suite *AnteTestSuite) getModuleBalance(ctx sdk.Context, module string) sdk.Coins {
	rewardsAccount := suite.ak.GetModuleAccount(ctx, module)
	return suite.bk.GetAllBalances(ctx, rewardsAccount.GetAddress())

}

func (suite *AnteTestSuite) createAccount(ctx sdk.Context, balance sdk.Coins) authtypes.AccountI {
	_, _, addr1 := testdata.KeyTestPubAddr()
	account := suite.ak.NewAccountWithAddress(ctx, addr1)
	suite.ak.SetAccount(ctx, account)
	fundAccount(suite, ctx, addr1, suite.bk, balance)

	return account
}

func coins(s string) sdk.Coins {
	c, err := sdk.ParseCoinsNormalized(s)
	if err != nil {
		panic(err)
	}
	return c
}

func (suite *AnteTestSuite) setup() {
	encConfig := MakeTestEncodingConfig()

	var (
		keyAuthCap = sdk.NewKVStoreKey("authCapKey")
		keyParams  = sdk.NewKVStoreKey("params")
		keyBank    = sdk.NewKVStoreKey(banktypes.ModuleName)
		tkeyParams = sdk.NewTransientStoreKey("transient_params")

		blockedAddr = make(map[string]bool)
		maccPerms   = map[string][]string{
			authtypes.ModuleName: {authtypes.Minter},
			authtypes.FeeCollectorName: nil,
			buyback.AccountName:        nil,
		}
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAuthCap, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyBank, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(suite.T(), err)

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	ctx = ctx.WithBlockTime(time.Now())

	pk := paramskeeper.NewKeeper(encConfig.Marshaler, encConfig.Amino, keyParams, tkeyParams)
	ak := authkeeper.NewAccountKeeper(
		encConfig.Marshaler, keyAuthCap, pk.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
	)
	bk := bankkeeper.NewBaseKeeper(
		encConfig.Marshaler, keyBank, ak, pk.Subspace(banktypes.ModuleName), blockedAddr,
	)

	dfd := ante.NewDeductFeeDecorator(ak, bk, mockStakingKeeper{"ungm"})

	suite.anteHandler = sdk.ChainAnteDecorators(dfd)
	suite.ak = ak
	suite.bk = bk
	suite.ctx = ctx
}

func MakeTestEncodingConfig() simappparams.EncodingConfig {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	encodingConfig := simappparams.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             cdc,
	}

	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics := module.NewBasicManager(
		bank.AppModuleBasic{},
		auth.AppModuleBasic{},
		vesting.AppModuleBasic{},
	)

	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

var _ sdk.FeeTx = mockFeeTX{}

// Mocked types
type (
	mockFeeTX struct {
		fee      sdk.Coins
		feePayer sdk.AccAddress
	}
	mockStakingKeeper struct {
		bondDenom string
	}
)

func (m mockFeeTX) GetMsgs() []sdk.Msg {
	return []sdk.Msg{}
}

func (m mockFeeTX) ValidateBasic() error {
	return nil
}

func (m mockFeeTX) GetGas() uint64 {
	return 0
}

func (m mockFeeTX) GetFee() sdk.Coins {
	return m.fee
}

func (m mockFeeTX) FeePayer() sdk.AccAddress {
	return m.feePayer
}

func (m mockFeeTX) FeeGranter() sdk.AccAddress {
	return nil
}

func (msk mockStakingKeeper) BondDenom(sdk.Context) string {
	return msk.bondDenom
}

func setAccBalance(suite *AnteTestSuite, ctx sdk.Context, acc sdk.AccAddress, bk bankkeeper.Keeper,	balance sdk.Coins) {
	err := bk.SendCoinsFromModuleToAccount(
		ctx, authtypes.ModuleName, acc, balance.Sub(bk.GetAllBalances(ctx, acc)),
	)
	suite.NoError(err)
}

func mintBalance(suite *AnteTestSuite, ctx sdk.Context, bk bankkeeper.Keeper, supply sdk.Coins) {
	err := bk.MintCoins(ctx, authtypes.ModuleName, supply)
	suite.NoError(err)
}

func fundAccount(suite *AnteTestSuite, ctx sdk.Context, acc sdk.AccAddress, bk bankkeeper.Keeper,
	balance sdk.Coins) {
	mintBalance(suite, ctx, bk, balance)
	setAccBalance(suite, ctx, acc, bk, balance)
}