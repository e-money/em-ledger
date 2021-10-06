// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	sdkslashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	embank "github.com/e-money/em-ledger/hooks/bank"
	apptypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/slashing/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"math"
	"testing"
	"time"
)

var (
	pks = []cryptotypes.PubKey{
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
	}
	addrs = []sdk.ValAddress{
		sdk.ValAddress(pks[0].Address()),
		sdk.ValAddress(pks[1].Address()),
		sdk.ValAddress(pks[2].Address()),
	}


	initTokens = sdk.TokensFromConsensusPower(200, sdk.OneInt())
	initCoins  = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
)

func TestBeginBlocker(t *testing.T) {
	ctx, keeper, _, bankKeeper, stakingKeeper, database := createTestComponents(t)

	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power, sdk.OneInt())
	addr1, pk1 := addrs[2], pks[2]
	addr2, pk2 := addrs[1], pks[1]

	// bond the validators
	_, err := staking.NewHandler(stakingKeeper)(ctx, NewTestMsgCreateValidator(addr1, pk1, amt))
	require.NoError(t, err)
	_, err = staking.NewHandler(stakingKeeper)(ctx, NewTestMsgCreateValidator(addr2, pk2, amt))
	require.NoError(t, err)

	staking.EndBlocker(ctx, stakingKeeper)
	require.Equal(
		t, bankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr1)),
		sdk.NewCoins(sdk.NewCoin(stakingKeeper.GetParams(ctx).BondDenom, initTokens.Sub(amt))),
	)
	require.Equal(t, amt.String(), stakingKeeper.Validator(ctx, addr1).GetBondedTokens().String())

	totalSupplyBefore := getTotalSupply(t, ctx, bankKeeper).AmountOf("stake")

	val := abci.Validator{
		Address: pk1.Address(),
		Power:   power,
	}

	val2 := abci.Validator{
		Address: pk2.Address(),
		Power:   power,
	}

	// mark the validator as having signed
	req := abci.RequestBeginBlock{
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{
				{
					Validator:       val,
					SignedLastBlock: true,
				},
				{
					Validator:       val2,
					SignedLastBlock: true,
				},
			},
		},
	}

	batch := database.NewBatch()
	ctx = apptypes.WithCurrentBatch(ctx, batch)
	BeginBlocker(ctx, req, keeper)
	batch.Write()

	info, found := keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(pk1.Address()))
	require.True(t, found)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	require.Equal(t, sdk.ConsAddress(pk1.Address()).String(), info.Address)

	height := int64(0)

	// for 1000 blocks, mark the validators as having signed
	now := time.Now()
	for ; height < 1000; height++ {
		now = now.Add(5 * time.Minute)
		ctx = ctx.WithBlockHeight(height).WithBlockTime(now)
		req = abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val,
						SignedLastBlock: true,
					},
					{
						Validator:       val2,
						SignedLastBlock: true,
					},
				},
			},
		}
		batch = database.NewBatch()
		ctx = apptypes.WithCurrentBatch(ctx, batch)
		BeginBlocker(ctx, req, keeper)
		batch.Write()
	}

	// for 500 blocks, mark the validator as having not signed. Other validator keeps signing.
	for ; height < 1500; height++ {
		now = now.Add(time.Minute)
		ctx = ctx.WithBlockHeight(height).WithBlockTime(now)
		req = abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val,
						SignedLastBlock: false,
					},
					{
						Validator:       val2,
						SignedLastBlock: true,
					},
				},
			},
		}
		batch = database.NewBatch()
		ctx = apptypes.WithCurrentBatch(ctx, batch)
		BeginBlocker(ctx, req, keeper)
		batch.Write()
	}

	// end block
	staking.EndBlocker(ctx, stakingKeeper)

	// validator should be jailed
	validator, found := stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk1))
	require.True(t, found)
	require.Equal(t, stakingtypes.Unbonding, validator.GetStatus())

	now = now.Add(5 * time.Minute)
	ctx = ctx.WithBlockHeight(height).WithBlockTime(now)
	req = abci.RequestBeginBlock{
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{
				{
					Validator:       val2,
					SignedLastBlock: true,
				},
			},
		},
	}
	batch = database.NewBatch()
	ctx = apptypes.WithCurrentBatch(ctx, batch)
	BeginBlocker(ctx, req, keeper)
	batch.Write()

	// Verify that slashed tokens are burned
	slashingPenalty := amt.ToDec().Mul(keeper.SlashFractionDowntime(ctx)).TruncateInt()
	totalSupplyAfter := getTotalSupply(t, ctx, bankKeeper).AmountOf("stake")
	require.Equal(t, totalSupplyBefore.Sub(slashingPenalty), totalSupplyAfter)

}

func TestOldGenesisTime(t *testing.T) {
	// The launch of emoney-3 revealed that the first block contains the genesis time from genesis.json. In the case of emoney-3, this was inherited directly from
	// emoney-2 and was thus some time in the past, leading to immediate jailing and slashing of validators that were not ready.
	// Block 1 should be treated differently to avoid this in future upgrades.
	genesisTime, _ := time.Parse(time.RFC3339, "2020-11-04T13:00:00Z") // Genesis time from emoney-2 and emoney-3

	ctx, keeper, _, _, stakingKeeper, database := createTestComponents(t)

	var power1, power2 int64 = 70, 30
	addr1, pk1 := addrs[2], pks[2]
	addr2, pk2 := addrs[1], pks[1]

	// bond the validators
	_, err := staking.NewHandler(stakingKeeper)(ctx, NewTestMsgCreateValidator(addr1, pk1, sdk.TokensFromConsensusPower(power1, sdk.OneInt())))
	require.NoError(t, err)
	_, err = staking.NewHandler(stakingKeeper)(ctx, NewTestMsgCreateValidator(addr2, pk2, sdk.TokensFromConsensusPower(power2, sdk.OneInt())))
	require.NoError(t, err)

	val1 := abci.Validator{Address: pk1.Address(), Power: power1}
	val2 := abci.Validator{Address: pk2.Address(), Power: power2}

	// Val2 fails to sign
	req := abci.RequestBeginBlock{
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{
				{
					Validator:       val1,
					SignedLastBlock: true,
				},
				{
					Validator:       val2,
					SignedLastBlock: false,
				},
			},
		},
	}

	// Start processing blocks
	now := time.Now()
	ctx = ctx.WithBlockHeight(1).WithBlockTime(genesisTime)

	// block 1
	{
		batch := database.NewBatch()
		ctx = apptypes.WithCurrentBatch(ctx, batch)
		BeginBlocker(ctx, req, keeper)
		staking.EndBlocker(ctx, stakingKeeper)
		batch.Write()
	}

	ctx = ctx.WithBlockHeight(2).WithBlockTime(now)
	// block 2
	{
		batch := database.NewBatch()
		ctx = apptypes.WithCurrentBatch(ctx, batch)
		BeginBlocker(ctx, req, keeper)
		staking.EndBlocker(ctx, stakingKeeper)
		batch.Write()
	}

	// Validator 2 must remain bonded.
	validator, found := stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk2))
	require.True(t, found)
	require.Equal(t, stakingtypes.Bonded, validator.GetStatus())

	// create a third block 25 hours later and ensure jailing
	ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(61 * time.Minute))
	// block 3
	{
		batch := database.NewBatch()
		ctx = apptypes.WithCurrentBatch(ctx, batch)
		BeginBlocker(ctx, req, keeper)
		staking.EndBlocker(ctx, stakingKeeper)
		batch.Write()
	}

	// Validator 2 must be unbonding.
	validator, found = stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk2))
	require.True(t, found)
	require.Equal(t, stakingtypes.Unbonding.String(), validator.GetStatus().String())
}

func createTestComponents(t *testing.T) (sdk.Context, Keeper, banktypes.AccountKeeper, bankkeeper.Keeper, stakingkeeper.Keeper, *dbm.MemDB) {
	t.Helper()
	encConfig := MakeTestEncodingConfig()

	sdk.DefaultPowerReduction = sdk.OneInt()

	var (
		authCapKey  = sdk.NewKVStoreKey("authCapKey")
		keyParams   = sdk.NewKVStoreKey("params")
		stakingKey  = sdk.NewKVStoreKey("staking")
		slashingKey = sdk.NewKVStoreKey(sdkslashingtypes.StoreKey)
		bankKey     = sdk.NewKVStoreKey(banktypes.ModuleName)

		tkeyParams = sdk.NewTransientStoreKey("transient_params")

		blockedAddr = make(map[string]bool)
		maccPerms   = map[string][]string{
			sdkslashingtypes.ModuleName:    {authtypes.Minter},
			stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
			stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
			authtypes.FeeCollectorName:     nil,
		}
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(stakingKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(slashingKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(bankKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "test-chain", Time: time.Now()}, true, log.NewNopLogger())
	var (
		pk = paramskeeper.NewKeeper(encConfig.Marshaler, encConfig.Amino, keyParams, tkeyParams)
		ak = authkeeper.NewAccountKeeper(
			encConfig.Marshaler, authCapKey, pk.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
		)

		bk = embank.Wrap(bankkeeper.NewBaseKeeper(encConfig.Marshaler, bankKey, ak, pk.Subspace(banktypes.ModuleName), blockedAddr))

		sk = stakingkeeper.NewKeeper(encConfig.Marshaler, stakingKey, ak, bk, pk.Subspace(stakingtypes.ModuleName))
	)
	// set staking params
	_ = staking.InitGenesis(ctx, sk, ak, bk, stakingtypes.DefaultGenesisState())

	// fund test accounts
	for _, addr := range addrs {
		address := sdk.AccAddress(addr)
		fundAccount(t, ctx, address, bk, initCoins)
		require.NoError(t, err)
		ak.SetAccount(ctx, authtypes.NewBaseAccountWithAddress(address))
	}
	// set module address
	require.Equal(t, getAllBalances(ctx, addrs, bk).String(), getTotalSupply(t, ctx, bk).String())

	keeper := NewKeeper(encConfig.Marshaler, slashingKey, sk, pk.Subspace(sdkslashingtypes.ModuleName), bk, db, authtypes.FeeCollectorName)
	keeper.SetParams(ctx, types.DefaultParams())
	sk.SetHooks(keeper.Hooks())
	return ctx, keeper, ak, bk, sk, db
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
	)

	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

func newPubKey(pk string) (res cryptotypes.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	return &ed25519.PubKey{Key: pkBytes}
}

func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey cryptotypes.PubKey, amt sdk.Int) *stakingtypes.MsgCreateValidator {
	commission := stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
	m, err := stakingtypes.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(sdk.DefaultBondDenom, amt),
		stakingtypes.Description{}, commission, sdk.OneInt(),
	)
	if err != nil {
		panic(err)
	}
	return m
}

func getTotalSupply(t *testing.T, ctx sdk.Context, bk bankkeeper.Keeper) sdk.Coins {
	totalSupply, _, err := bk.GetPaginatedTotalSupply(
		ctx, &query.PageRequest{Limit: math.MaxUint64},
	)
	require.NoError(t, err)

	return totalSupply
}

func mintBalance(t *testing.T, ctx sdk.Context, bk bankkeeper.Keeper, supply sdk.Coins) {
	err := bk.MintCoins(ctx, sdkslashingtypes.ModuleName, supply)
	require.NoError(t, err)
}

func setAccBalance(
	t *testing.T, ctx sdk.Context, acc sdk.AccAddress, bk bankkeeper.Keeper,
	balance sdk.Coins,
) {
	err := bk.SendCoinsFromModuleToAccount(
		ctx, sdkslashingtypes.ModuleName, acc, balance.Sub(bk.GetAllBalances(ctx, acc)),
	)
	require.NoError(t, err)
}

func fundAccount(t *testing.T, ctx sdk.Context, acc sdk.AccAddress, bk bankkeeper.Keeper,
	balance sdk.Coins) {
	mintBalance(t, ctx, bk, balance)
	setAccBalance(t, ctx, acc, bk, balance)
}

func getAllBalances(ctx sdk.Context, addresses []sdk.ValAddress, bk bankkeeper.Keeper) sdk.Coins {
	totBalances := sdk.NewCoins()
	for _, addr := range addresses {
		act := sdk.AccAddress(addr)
		coins := bk.GetAllBalances(ctx, act)
		totBalances = totBalances.Add(coins...)
	}

	return totBalances
}