// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package slashing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	db "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetSetValidatorSigningInfo(t *testing.T) {
	ctx, _, _, _, keeper, _ := createTestInput(t, DefaultParams(), db.NewMemDB())
	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(addrs[0]))
	require.False(t, found)
	newInfo := NewValidatorSigningInfo(
		sdk.ConsAddress(addrs[0]),
		time.Unix(2, 0),
		false,
	)
	keeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrs[0]), newInfo)
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(addrs[0]))
	require.True(t, found)
	//require.Equal(t, info.StartHeight, int64(4))
	//require.Equal(t, info.IndexOffset, int64(3))
	require.Equal(t, info.JailedUntil, time.Unix(2, 0).UTC())
	//require.Equal(t, info.MissedBlocksCounter, int64(10))
}

func TestGetSetValidatorMissedBlockBitArray(t *testing.T) {
	database := db.NewMemDB()
	_, _, _, _, keeper, _ := createTestInput(t, DefaultParams(), database)

	missed := keeper.getValidatorMissedBlockBitArray(sdk.ConsAddress(addrs[0]), 0)
	require.False(t, missed) // treat empty key as not missed
	batch := database.NewBatch()
	keeper.setValidatorMissedBlockBitArray(batch, sdk.ConsAddress(addrs[0]), 0, true)
	batch.Write()
	missed = keeper.getValidatorMissedBlockBitArray(sdk.ConsAddress(addrs[0]), 0)
	require.True(t, missed) // now should be missed
}
