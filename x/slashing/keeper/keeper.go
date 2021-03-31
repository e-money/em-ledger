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
	dbKeyMissedByVal      = "%v.missedBlocks"
	dbKeyPendingPenalties = "activePenalties"
	dbKeyBlockTimes       = "blocktimes"
)

type Keeper struct {
	sdkslashingkeeper.Keeper
	paramspace types.ParamSubspace

	cdc        codec.BinaryMarshaler
	sk         types.StakingKeeper
	bankKeeper types.BankKeeper
	// Alternative to IAVL KV storage. For data that should not be part of consensus.
	database      types.ReadOnlyDB
	feeModuleName string
}

func NewKeeper(
	cdc codec.BinaryMarshaler,
	key sdk.StoreKey,
	sk types.StakingKeeper,
	paramspace types.ParamSubspace,
	bankKeeper types.BankKeeper,
	database types.ReadOnlyDB,
	feeModuleName string,
) Keeper {
	return Keeper{
		Keeper:        sdkslashingkeeper.NewKeeper(cdc, key, sk, paramspace),
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
		panic(err) // TODO Better handling
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

func (k Keeper) handlePendingPenalties(ctx sdk.Context, batch db.Batch, vfn func() map[string]bool) {
	activePenalties := k.getPendingPenalties()
	if activePenalties.Empty() {
		return
	}

	validatorSet := vfn()
	nextActivePenalties := types.Penalties{Elements: make([]types.Penalty, 0)}
	for _, e := range activePenalties.Elements {
		if _, present := validatorSet[e.Validator]; present {
			// Penalized validator is still in the validator set. Do not pay out slashing fine.
			nextActivePenalties.Add(e.Validator, e.Amounts)
			continue
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePenaltyPayout,
				sdk.NewAttribute(types.AttributeKeyAmount, e.Amounts.String()),
				sdk.NewAttribute(sdkslashingtypes.AttributeKeyAddress, e.Validator),
			),
		)

		err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.PenaltyAccount, k.feeModuleName, e.Amounts)
		if err != nil {
			panic(err)
		}
	}

	k.setPendingPenalties(batch, nextActivePenalties)
}

func (k Keeper) setPendingPenalties(batch db.Batch, penalties types.Penalties) {
	if penalties.Empty() {
		batch.Delete([]byte(dbKeyPendingPenalties))
		return
	}

	bz := k.cdc.MustMarshalBinaryBare(&penalties)
	batch.Set([]byte(dbKeyPendingPenalties), bz)
}

func (k Keeper) getPendingPenalties() types.Penalties {
	bz, err := k.database.Get([]byte(dbKeyPendingPenalties))
	if err != nil {
		panic(err) // TODO Better handling
	}

	if len(bz) == 0 {
		return types.Penalties{}
	}

	var activePenalties types.Penalties
	k.cdc.MustUnmarshalBinaryBare(bz, &activePenalties)

	return activePenalties
}

func (k Keeper) getBlockTimes() []time.Time {
	bz, err := k.database.Get([]byte(dbKeyBlockTimes))
	if err != nil {
		panic(err) // TODO Better handling
	}

	if len(bz) == 0 {
		return make([]time.Time, 0)
	}

	b := bytes.NewBuffer(bz)
	blockTimes := make([]time.Time, 0)
	dec := gob.NewDecoder(b) // todo (reviewer): you may want to use protobuf instead for consistency
	_ = dec.Decode(&blockTimes)
	return blockTimes
}

func (k Keeper) setBlockTimes(batch db.Batch, blockTimes []time.Time) {
	bz := new(bytes.Buffer)
	enc := gob.NewEncoder(bz)
	_ = enc.Encode(blockTimes)
	batch.Set([]byte(dbKeyBlockTimes), bz.Bytes())
}

func (k Keeper) slashValidator(ctx sdk.Context, batch db.Batch, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor sdk.Dec) {
	k.sk.Slash(ctx, consAddr, infractionHeight, power, slashFactor)

	// Mint the slashed coins and assign them to the distribution pool.
	slashAmount := calculateSlashingAmount(power, slashFactor)
	k.Logger(ctx).Info("calculating slash amount",
		"power", power, "factor", slashFactor, "amount", slashAmount.String(),
		"tokensFPower", sdk.TokensFromConsensusPower(power).String(),
		"calc", sdk.TokensFromConsensusPower(power).ToDec().Mul(slashFactor).String())
	stakingDenom := k.sk.BondDenom(ctx)
	coins := sdk.NewCoins(sdk.NewCoin(stakingDenom, slashAmount))
	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins)
	if err != nil {
		panic(err)
	}

	k.Logger(ctx).Info("transfer to penalty account", "amounts", coins)
	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.PenaltyAccount, coins)
	if err != nil {
		panic(err)
	}

	activePenalties := k.getPendingPenalties()
	activePenalties.Add(consAddr.String(), coins)
	k.setPendingPenalties(batch, activePenalties)
	k.Logger(ctx).Info("set penalties", "active", activePenalties)
}

// Adopted from cosmos-sdk/x/staking/keeper/slash.go
func calculateSlashingAmount(power int64, slashFactor sdk.Dec) sdk.Int {
	amount := sdk.TokensFromConsensusPower(power)
	slashAmountDec := amount.ToDec().Mul(slashFactor)
	slashAmount := slashAmountDec.TruncateInt()
	return slashAmount
}
