package keeper

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	sdkslashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	sdkslashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/e-money/em-ledger/x/slashing/types"
	db "github.com/tendermint/tm-db"
	"time"
)

var _ evidencetypes.SlashingKeeper = Keeper{}

const (
	dbKeyMissedByVal = "%v.missedBlocks"
	dbKeyBlockTimes  = "blocktimes"
)

type Keeper struct {
	sdkslashingkeeper.Keeper
	StoreKey sdk.StoreKey

	paramspace sdkslashingtypes.ParamSubspace

	cdc        codec.BinaryCodec
	sk         sdkslashingtypes.StakingKeeper
	bankKeeper sdkslashingtypes.BankKeeper
	// Alternative to IAVL KV storage. For data that should not be part of consensus.
	database      types.ReadOnlyDB
	feeModuleName string
}

func NewKeeper(
	cdc codec.Codec,
	key sdk.StoreKey,
	sk sdkslashingtypes.StakingKeeper,
	paramspace sdkslashingtypes.ParamSubspace,
	bankKeeper sdkslashingtypes.BankKeeper,
	database types.ReadOnlyDB,
	feeModuleName string,
) Keeper {
	return Keeper{
		Keeper:        sdkslashingkeeper.NewKeeper(cdc, key, sk, paramspace),
		StoreKey:      key,
		paramspace:    paramspace,
		cdc:           cdc,
		sk:            sk,
		bankKeeper:    bankKeeper,
		database:      database,
		feeModuleName: feeModuleName,
	}
}

func (k Keeper) getMissingBlocksForValidator(address sdk.ConsAddress) []time.Time {
	key := []byte(fmt.Sprintf(dbKeyMissedByVal, address.String()))
	bz, err := k.database.Get(key)
	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return nil
	}

	b := bytes.NewBuffer(bz)
	dec := gob.NewDecoder(b)

	res := make([]time.Time, 0)
	err = dec.Decode(&res)
	if err != nil {
		panic(err)
	}

	return res
}

func (k Keeper) setMissingBlocksForValidator(batch db.Batch, address sdk.ConsAddress, missingBlocks []time.Time) {
	bz := new(bytes.Buffer)
	enc := gob.NewEncoder(bz)
	err := enc.Encode(missingBlocks)
	if err != nil {
		panic(err)
	}

	key := []byte(fmt.Sprintf(dbKeyMissedByVal, address.String()))
	batch.Set(key, bz.Bytes())
}

func (k Keeper) deleteMissingBlocksForValidator(batch db.Batch, address sdk.ConsAddress) {
	key := []byte(fmt.Sprintf(dbKeyMissedByVal, address.String()))
	batch.Delete(key)
}

func (k Keeper) getBlockTimes() []time.Time {
	bz, err := k.database.Get([]byte(dbKeyBlockTimes))
	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return make([]time.Time, 0)
	}

	b := bytes.NewBuffer(bz)
	blockTimes := make([]time.Time, 0)
	dec := gob.NewDecoder(b)
	_ = dec.Decode(&blockTimes)
	return blockTimes
}

func (k Keeper) setBlockTimes(batch db.Batch, blockTimes []time.Time) {
	bz := new(bytes.Buffer)
	enc := gob.NewEncoder(bz)
	_ = enc.Encode(blockTimes)
	batch.Set([]byte(dbKeyBlockTimes), bz.Bytes())
}
