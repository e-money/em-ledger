package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/offer/types"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
)

func TestName(t *testing.T) {
	ctx := createTestComponents(t)

	k := NewKeeper()

	order := types.NewOrder("EUR", "USD", 100, 120)
	k.ProcessOrder(ctx, order)

	fmt.Println(k.instruments.String())

	order = types.NewOrder("USD", "EUR", 120, 100)
	k.ProcessOrder(ctx, order)

	fmt.Println(k.instruments.String())

}

func createTestComponents(t *testing.T) sdk.Context {
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, true, log.NewNopLogger())

	return ctx
}
