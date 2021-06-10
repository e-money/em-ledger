package buyback

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	embank "github.com/e-money/em-ledger/hooks/bank"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/e-money/em-ledger/x/buyback/internal/keeper"
	"github.com/e-money/em-ledger/x/market"
	"github.com/e-money/em-ledger/x/market/types"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/tendermint/tendermint/libs/log"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	dbm "github.com/tendermint/tm-db"
)

const stakingDenom = "ungm"

func TestBuyback1(t *testing.T) {
	ctx, k, market, accountKeeper, bankKeeper := createTestComponents(t)

	ctx = ctx.WithBlockHeight(1)

	// Add some ungm sell orders
	acc1 := createAccount(ctx, accountKeeper, bankKeeper, randomAddress(), "50000ungm")
	market.NewOrderSingle(ctx, order(acc1, "5000ungm", "10000eur"))
	market.NewOrderSingle(ctx, order(acc1, "5000ungm", "20000chf"))

	buybackAccount := accountKeeper.GetModuleAccount(ctx, ModuleName).GetAddress()
	err := bankKeeper.AddCoins(ctx, buybackAccount, coins("10000ungm"))
	require.NoError(t, err)

	BeginBlocker(ctx, k, bankKeeper)

	// Verify that staking tokens are burned
	balances := bankKeeper.GetAllBalances(ctx, buybackAccount)
	require.True(t, balances.AmountOf(stakingDenom).IsZero(), balances)

	require.Condition(t, func() bool {
		for _, evt := range ctx.EventManager().ABCIEvents() {
			if evt.Type == EventTypeBuyback {
				return true
			}
		}

		return false
	}, "Burn event not found")

	// Verify that an order was created in the market
	orders := market.GetOrdersByOwner(ctx, buybackAccount)
	require.Len(t, orders, 1)

	// Ensure that the buyback module matches the order and that an order remains in the market with the remaining 40 000 eur
	require.Equal(t, coin("50000eur"), orders[0].Source)
	require.Equal(t, sdk.NewInt(40000).String(), orders[0].SourceRemaining.String())
	require.True(t, strings.HasSuffix(orders[0].ClientOrderID, "1"))

	// Add some echf to to buyback and see it take a bit of the previous sell order
	err = bankKeeper.AddCoins(ctx, buybackAccount, coins("10000chf"))
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(2).WithBlockTime(ctx.BlockTime().Add(2 * time.Hour))
	BeginBlocker(ctx, k, bankKeeper)

	balances = bankKeeper.GetAllBalances(ctx, buybackAccount)
	require.True(t, balances.AmountOf("chf").IsZero(), balances)

	// Account 1 should have one of its orders completely filled and the other partially filled
	orders = market.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, orders, 1)
	require.Equal(t, orders[0].Destination.Denom, "chf")
}

func TestBuyback2(t *testing.T) {
	// Verify that the module does not update its market positions for every block
	ctx, keeper, market, accountKeeper, bankKeeper := createTestComponents(t)

	ctx = ctx.WithBlockHeight(1)

	acc1 := createAccount(ctx, accountKeeper, bankKeeper, randomAddress(), "50000ungm")
	//generateMarketActivity(ctx, market, accountKeeper, bankKeeper)
	market.NewOrderSingle(ctx, order(acc1, "5000ungm", "10000eur"))
	market.NewOrderSingle(ctx, order(acc1, "5000ungm", "20000chf"))

	buybackAccount := accountKeeper.GetModuleAccount(ctx, ModuleName).GetAddress()

	orders := market.GetOrdersByOwner(ctx, buybackAccount)
	require.Empty(t, orders)

	BeginBlocker(ctx, keeper, bankKeeper)

	orders = market.GetOrdersByOwner(ctx, buybackAccount)
	require.Len(t, orders, 1)

	// New balance that should trigger order updates
	bankKeeper.AddCoins(ctx, buybackAccount, coins("10000eur,75000chf"))

	// Time since last update is too short to update the module's market positions.
	ctx = ctx.WithBlockHeight(2).WithBlockTime(ctx.BlockTime().Add(30 * time.Minute))
	BeginBlocker(ctx, keeper, bankKeeper)

	orders = market.GetOrdersByOwner(ctx, buybackAccount)
	require.Len(t, orders, 1)
	require.True(t, strings.HasSuffix(orders[0].ClientOrderID, "1"))

	// Time since last update is sufficient. Market positions must be updated.
	ctx = ctx.WithBlockHeight(3).WithBlockTime(ctx.BlockTime().Add(time.Hour))
	BeginBlocker(ctx, keeper, bankKeeper)

	orders = market.GetOrdersByOwner(ctx, buybackAccount)
	require.Len(t, orders, 1)
	require.True(t, strings.HasSuffix(orders[0].ClientOrderID, "3"))
}

func TestBuyback3(t *testing.T) {
	// Test very high NGM price with very low balance
	ctx, keeper, market, accountKeeper, bankKeeper := createTestComponents(t)
	buybackAccount := accountKeeper.GetModuleAccount(ctx, ModuleName).GetAddress()
	bankKeeper.AddCoins(ctx, buybackAccount, coins("50pesos"))

	// Generate some prices for the pesos <-> ungm instrument
	acc2 := createAccount(ctx, accountKeeper, bankKeeper, randomAddress(), "10000ungm")

	_, err := market.NewOrderSingle(ctx, order(acc2, "1ungm", "4000000pesos"))
	require.NoError(t, err)

	// Attempt to create a position using the meager pesos balance of the module
	BeginBlocker(ctx, keeper, bankKeeper)

	balances := bankKeeper.GetAllBalances(ctx, buybackAccount)
	require.Equal(t, "50", balances.AmountOf("pesos").String())

	orders := market.GetOrdersByOwner(ctx, buybackAccount)
	require.Empty(t, orders)
}

func order(account authtypes.AccountI, src, dst string) types.Order {
	s, _ := sdk.ParseCoinNormalized(src)
	d, _ := sdk.ParseCoinNormalized(dst)
	o, err := types.NewOrder(time.Now(), types.TimeInForce_GoodTillCancel, s, d, account.GetAddress(), tmrand.Str(10))
	if err != nil {
		panic(err)
	}

	return o
}

func createAccount(ctx sdk.Context, ak banktypes.AccountKeeper, bk bankkeeper.SendKeeper, address sdk.AccAddress, balance string) authtypes.AccountI {
	acc := ak.NewAccountWithAddress(ctx, address)
	if err := bk.SetBalances(ctx, address, coins(balance)); err != nil {
		panic(err)
	}
	ak.SetAccount(ctx, acc)
	return acc
}

func createTestComponents(t *testing.T) (sdk.Context, keeper.Keeper, *market.Keeper, banktypes.AccountKeeper, bankkeeper.Keeper) {
	t.Helper()
	encConfig := MakeTestEncodingConfig()

	var (
		keyMarket  = sdk.NewKVStoreKey(types.ModuleName)
		keyIndices = sdk.NewKVStoreKey(types.StoreKeyIdx)
		authCapKey = sdk.NewKVStoreKey("authCapKey")
		keyParams  = sdk.NewKVStoreKey("params")
		stakingKey = sdk.NewKVStoreKey("staking")
		buybackKey = sdk.NewKVStoreKey("buyback")
		bankKey    = sdk.NewKVStoreKey(banktypes.ModuleName)

		tkeyParams = sdk.NewTransientStoreKey("transient_params")

		blockedAddr = make(map[string]bool)
		maccPerms   = map[string][]string{
			AccountName: {authtypes.Burner},
		}
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(stakingKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyMarket, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyIndices, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(buybackKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(bankKey, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "test-chain", Time: time.Now()}, true, log.NewNopLogger())
	var (
		pk = paramskeeper.NewKeeper(encConfig.Marshaler, encConfig.Amino, keyParams, tkeyParams)
		ak = authkeeper.NewAccountKeeper(
			encConfig.Marshaler, authCapKey, pk.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
		)

		bk = embank.Wrap(bankkeeper.NewBaseKeeper(encConfig.Marshaler, bankKey, ak, pk.Subspace(banktypes.ModuleName), blockedAddr))
	)

	initialSupply := coins(fmt.Sprintf("1000000eur,1000000usd,1000000chf,1000000jpy,1000000gbp,1000000%v,500000000pesos", stakingDenom))
	bk.SetSupply(ctx, banktypes.NewSupply(initialSupply))

	marketKeeper := market.NewKeeper(encConfig.Marshaler, keyMarket, keyIndices, ak, bk)

	k := NewKeeper(encConfig.Marshaler, buybackKey, marketKeeper, ak, mockStakingKeeper{}, bk)
	k.SetUpdateInterval(ctx, time.Hour)

	// Deposit a working balance on the buyback module account.
	buybackAccount := ak.GetModuleAddress(ModuleName)
	bk.SetBalances(ctx, buybackAccount, coins("50000eur"))

	return ctx, k, marketKeeper, ak, bk
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

func coin(s string) sdk.Coin {
	c, err := sdk.ParseCoinNormalized(s)
	if err != nil {
		panic(err)
	}

	return c
}

func coins(c string) sdk.Coins {
	coins, err := sdk.ParseCoinsNormalized(c)
	if err != nil {
		panic(err)
	}

	return coins
}

func randomAddress() sdk.AccAddress {
	return tmrand.Bytes(sdk.AddrLen)
}

var _ keeper.StakingKeeper = (*mockStakingKeeper)(nil)

type mockStakingKeeper struct{}

func (mockStakingKeeper) BondDenom(sdk.Context) string {
	return stakingDenom
}
