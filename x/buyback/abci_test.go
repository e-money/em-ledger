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
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	embank "github.com/e-money/em-ledger/hooks/bank"
	types2 "github.com/e-money/em-ledger/x/authority/types"
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

const (
	stakingDenom = "ungm"
)

func TestBuyback1(t *testing.T) {
	ctx, keeper, market, accountKeeper, bankKeeper := createTestComponents(t)

	ctx = ctx.WithBlockHeight(1)
	generateMarketActivity(ctx, market, accountKeeper, bankKeeper)

	account := accountKeeper.GetModuleAccount(ctx, ModuleName)
	err := bankKeeper.AddCoins(ctx, account.GetAddress(), coins("10000ungm"))
	require.NoError(t, err)

	BeginBlocker(ctx, keeper, bankKeeper)

	// Verify that staking tokens are burned
	account = accountKeeper.GetModuleAccount(ctx, ModuleName)
	balances := bankKeeper.GetAllBalances(ctx, account.GetAddress())
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
	orders := market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Len(t, orders, 1)
	require.Equal(t, sdk.NewCoin("eur", sdk.NewInt(50000)), orders[0].Source)
	require.True(t, strings.HasSuffix(orders[0].ClientOrderID, "1"))

	ctx = ctx.WithBlockHeight(2).WithBlockTime(ctx.BlockTime().Add(2 * time.Hour))
	// Add more stablecoin tokens to the module's buyback account
	bankKeeper.AddCoins(ctx, account.GetAddress(), coins("10000eur,75000chf"))

	// Update account balance information
	account = accountKeeper.GetModuleAccount(ctx, ModuleName)
	BeginBlocker(ctx, keeper, bankKeeper)

	orders = market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Len(t, orders, 2)
	balances = bankKeeper.GetAllBalances(ctx, account.GetAddress())
	require.Equal(t, balances.AmountOf("chf"), orders[0].Source.Amount)
	require.Equal(t, balances.AmountOf("eur"), orders[1].Source.Amount)

	require.True(t, strings.HasSuffix(orders[0].ClientOrderID, "2"))

	// Verify prices.
	require.Equal(t, sdk.NewDecWithPrec(25, 2), orders[0].Price())
	require.Equal(t, sdk.NewDecWithPrec(5, 1), orders[1].Price())
}

func TestBuyback2(t *testing.T) {
	// Verify that the module does not update its market positions for every block
	ctx, keeper, market, accountKeeper, bankKeeper := createTestComponents(t)

	ctx = ctx.WithBlockHeight(1)
	generateMarketActivity(ctx, market, accountKeeper, bankKeeper)

	account := accountKeeper.GetModuleAccount(ctx, ModuleName)

	orders := market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Empty(t, orders)

	BeginBlocker(ctx, keeper, bankKeeper)

	orders = market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Len(t, orders, 1)

	// New balance that should trigger order updates
	bankKeeper.AddCoins(ctx, account.GetAddress(), coins("10000eur,75000chf"))

	// Time since last update is too short to update the module's market positions.
	ctx = ctx.WithBlockHeight(2).WithBlockTime(ctx.BlockTime().Add(30 * time.Minute))
	BeginBlocker(ctx, keeper, bankKeeper)

	orders = market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Len(t, orders, 1)
	require.True(t, strings.HasSuffix(orders[0].ClientOrderID, "1"))

	// Time since last update is sufficient. Market positions must be updated.
	ctx = ctx.WithBlockHeight(3).WithBlockTime(ctx.BlockTime().Add(time.Hour))
	BeginBlocker(ctx, keeper, bankKeeper)

	orders = market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Len(t, orders, 2)
	require.True(t, strings.HasSuffix(orders[0].ClientOrderID, "3"))
}

func TestBuyback3(t *testing.T) {
	// Test very high NGM price with very low balance
	ctx, keeper, market, accountKeeper, bankKeeper := createTestComponents(t)
	account := accountKeeper.GetModuleAccount(ctx, ModuleName)
	bankKeeper.AddCoins(ctx, account.GetAddress(), coins("50pesos"))

	// Generate some prices for the pesos <-> ungm instrument
	var (
		acc1 = createAccount(ctx, accountKeeper, bankKeeper, randomAddress(), "5000000pesos")
		acc2 = createAccount(ctx, accountKeeper, bankKeeper, randomAddress(), "10000ungm")
	)
	_, err := market.NewOrderSingle(ctx, order(acc1, "4000000pesos", "1ungm"))
	require.NoError(t, err)

	_, err = market.NewOrderSingle(ctx, order(acc2, "1ungm", "4000000pesos"))
	require.NoError(t, err)

	// Attempt to create a position using the meager pesos balance of the module
	BeginBlocker(ctx, keeper, bankKeeper)

	account = accountKeeper.GetModuleAccount(ctx, ModuleName)
	balances := bankKeeper.GetAllBalances(ctx, account.GetAddress())
	require.Equal(t, "50", balances.AmountOf("pesos").String())

	orders := market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Empty(t, orders)
}

func TestGroupMarketData(t *testing.T) {
	price := func(i int64) *sdk.Dec {
		r := sdk.NewDec(i)
		return &r
	}

	md := []types.MarketData{
		{
			Source:      "eur",
			Destination: "ungm",
		},
		{
			Source:      "chf",
			Destination: "ungm",
		},
		{
			Source:      "eur",
			Destination: "chf",
			LastPrice:   price(5),
		},
		{
			Source:      "ungm",
			Destination: "chf",
			LastPrice:   price(8),
		},
	}

	groupedOrders := groupMarketDataBySource(md, "ungm")

	require.Len(t, groupedOrders, 2)
	require.Contains(t, groupedOrders, "eur")
	require.Contains(t, groupedOrders, "chf")

}

// Create some basic pricing information that market orders can be made from
func generateMarketActivity(ctx sdk.Context, marketKeeper *market.Keeper, ak banktypes.AccountKeeper, bk bankkeeper.SendKeeper) {
	var (
		acc1 = createAccount(ctx, ak, bk, randomAddress(), "50000ungm")
		acc2 = createAccount(ctx, ak, bk, randomAddress(), "150000eur,290000chf")
	)

	marketKeeper.NewOrderSingle(ctx, order(acc1, "5000ungm", "10000eur"))
	marketKeeper.NewOrderSingle(ctx, order(acc1, "5000ungm", "20000chf"))
	marketKeeper.NewOrderSingle(ctx, order(acc2, "10000eur", "5000ungm"))
	marketKeeper.NewOrderSingle(ctx, order(acc2, "20000chf", "5000ungm"))
}

func order(account authtypes.AccountI, src, dst string) types.Order {
	s, _ := sdk.ParseCoinNormalized(src)
	d, _ := sdk.ParseCoinNormalized(dst)
	o, err := types.NewOrder(
		time.Now(), types.TimeInForce_GoodTillCancel, s, d,
		account.GetAddress(), tmrand.Str(10), time.Time{},
	)
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
		keyMarket  = sdk.NewKVStoreKey(types.StoreKey)
		keyIndices = sdk.NewKVStoreKey(types.StoreKeyIdx)
		authCapKey = sdk.NewKVStoreKey("authCapKey")
		keyParams  = sdk.NewKVStoreKey("params")
		stakingKey = sdk.NewKVStoreKey("staking")
		buybackKey = sdk.NewKVStoreKey("buyback")
		bankKey    = sdk.NewKVStoreKey(banktypes.ModuleName)

		tkeyParams = sdk.NewTransientStoreKey(paramstypes.TStoreKey)

		blockedAddr = make(map[string]bool)
		maccPerms   = map[string][]string{
			AccountName: {authtypes.Burner},
		}
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(stakingKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyMarket, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyIndices, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(buybackKey, sdk.StoreTypeIAVL, db)
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
		allowAllDenoms = embank.RestrictedKeeperFunc(func(context sdk.Context) types2.RestrictedDenoms {
			return types2.RestrictedDenoms{}
		})
		bk = embank.Wrap(bankkeeper.NewBaseKeeper(
			encConfig.Marshaler, bankKey, ak, pk.Subspace(banktypes.ModuleName), blockedAddr,
		), allowAllDenoms)

		marketKeeper = market.NewKeeper(
			encConfig.Marshaler, keyMarket, keyIndices, ak, bk, mockAuthority{},
			pk.Subspace(market.ModuleName),
		)
	)

	initialSupply := coins(fmt.Sprintf("1000000eur,1000000usd,1000000chf,1000000jpy,1000000gbp,1000000%v,500000000pesos", stakingDenom))
	bk.SetSupply(ctx, banktypes.NewSupply(initialSupply))

	marketKeeper.InitParamsStore(ctx)

	keeper := NewKeeper(encConfig.Marshaler, buybackKey, marketKeeper, ak, mockStakingKeeper{}, bk)
	keeper.SetUpdateInterval(ctx, time.Hour)

	// Deposit a working balance on the buyback module account.
	buybackAccount := ak.GetModuleAddress(ModuleName)
	bk.SetBalances(ctx, buybackAccount, coins("50000eur"))

	return ctx, keeper, marketKeeper, ak, bk
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

var (
	_ types.RestrictedKeeper = (*mockAuthority)(nil)
	_ keeper.StakingKeeper   = (*mockStakingKeeper)(nil)
)

type (
	mockAuthority     struct{}
	mockStakingKeeper struct{}
)

func (mockStakingKeeper) BondDenom(sdk.Context) string {
	return stakingDenom
}

func (mockAuthority) GetRestrictedDenoms(sdk.Context) types2.RestrictedDenoms {
	return types2.RestrictedDenoms{}
}
