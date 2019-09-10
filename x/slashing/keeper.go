package slashing

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"emoney/x/slashing/types"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper of the slashing store
type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	sk            types.StakingKeeper
	supplyKeeper  types.SupplyKeeper
	feeModuleName string
	paramspace    params.Subspace

	// codespace
	codespace sdk.CodespaceType

	// Alternative to IAVL KV storage. For data that should not be part of consensus.
	database db.DB
}

const (
	dbKeyMissedByVal      = "%v.missedBlocks"
	dbKeyPendingPenalties = "activePenalties"
)

// NewKeeper creates a slashing keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, sk types.StakingKeeper, supplyKeeper types.SupplyKeeper, feeModuleName string, paramspace params.Subspace, codespace sdk.CodespaceType, database db.DB) Keeper {
	keeper := Keeper{
		storeKey:      key,
		cdc:           cdc,
		sk:            sk,
		supplyKeeper:  supplyKeeper,
		feeModuleName: feeModuleName,
		paramspace:    paramspace.WithKeyTable(ParamKeyTable()),
		codespace:     codespace,
		database:      database,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// handle a validator signing two blocks at the same height
// power: power of the double-signing validator at the height of infraction
func (k Keeper) HandleDoubleSign(ctx sdk.Context, addr crypto.Address, infractionHeight int64, timestamp time.Time, power int64) {
	logger := k.Logger(ctx)

	// calculate the age of the evidence
	time := ctx.BlockHeader().Time
	age := time.Sub(timestamp)

	// fetch the validator public key
	consAddr := sdk.ConsAddress(addr)
	pubkey, err := k.getPubkey(ctx, addr)
	if err != nil {
		// Ignore evidence that cannot be handled.
		// NOTE:
		// We used to panic with:
		// `panic(fmt.Sprintf("Validator consensus-address %v not found", consAddr))`,
		// but this couples the expectations of the app to both Tendermint and
		// the simulator.  Both are expected to provide the full range of
		// allowable but none of the disallowed evidence types.  Instead of
		// getting this coordination right, it is easier to relax the
		// constraints and ignore evidence that cannot be handled.
		return
	}

	// Reject evidence if the double-sign is too old
	if age > k.MaxEvidenceAge(ctx) {
		logger.Info(fmt.Sprintf("Ignored double sign from %s at height %d, age of %d past max age of %d",
			sdk.ConsAddress(pubkey.Address()), infractionHeight, age, k.MaxEvidenceAge(ctx)))
		return
	}

	// Get validator and signing info
	validator := k.sk.ValidatorByConsAddr(ctx, consAddr)
	if validator == nil || validator.IsUnbonded() {
		// Defensive.
		// Simulation doesn't take unbonding periods into account, and
		// Tendermint might break this assumption at some point.
		return
	}

	// fetch the validator signing info
	signInfo, found := k.getValidatorSigningInfo(consAddr)
	if !found {
		panic(fmt.Sprintf("Expected signing info for validator %s but not found", consAddr))
	}

	// validator is already tombstoned
	if signInfo.Tombstoned {
		logger.Info(fmt.Sprintf("Ignored double sign from %s at height %d, validator already tombstoned", sdk.ConsAddress(pubkey.Address()), infractionHeight))
		return
	}

	// double sign confirmed
	logger.Info(fmt.Sprintf("Confirmed double sign from %s at height %d, age of %d", sdk.ConsAddress(pubkey.Address()), infractionHeight, age))

	// We need to retrieve the stake distribution which signed the block, so we subtract ValidatorUpdateDelay from the evidence height.
	// Note that this *can* result in a negative "distributionHeight", up to -ValidatorUpdateDelay,
	// i.e. at the end of the pre-genesis block (none) = at the beginning of the genesis block.
	// That's fine since this is just used to filter unbonding delegations & redelegations.
	distributionHeight := infractionHeight - sdk.ValidatorUpdateDelay

	// get the percentage slash penalty fraction
	fraction := k.SlashFractionDoubleSign(ctx)

	// Slash validator
	// `power` is the int64 power of the validator as provided to/by
	// Tendermint. This value is validator.Tokens as sent to Tendermint via
	// ABCI, and now received as evidence.
	// The fraction is passed in to separately to slash unbonding and rebonding delegations.
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlash,
			sdk.NewAttribute(types.AttributeKeyAddress, consAddr.String()),
			sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", power)),
			sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueDoubleSign),
		),
	)
	k.slashValidator(ctx, consAddr, distributionHeight, power, fraction)

	// Jail validator if not already jailed
	// begin unbonding validator if not already unbonding (tombstone)
	if !validator.IsJailed() {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSlash,
				sdk.NewAttribute(types.AttributeKeyJailed, consAddr.String()),
			),
		)
		k.sk.Jail(ctx, consAddr)
	}

	// Set tombstoned to be true
	signInfo.Tombstoned = true

	// Set jailed until to be forever (max time)
	signInfo.JailedUntil = types.DoubleSignJailEndTime

	// Set validator signing info
	k.SetValidatorSigningInfo(consAddr, signInfo)
}

func (k Keeper) HandleValidatorSignature(ctx sdk.Context, addr crypto.Address, power int64, signed bool, blockCount int64) {
	logger := k.Logger(ctx)
	height := ctx.BlockHeight()
	consAddr := sdk.ConsAddress(addr)

	missedBlocks := k.getMissingBlocksForValidator(consAddr)
	missedBlocks = truncateByWindow(ctx.BlockTime(), missedBlocks, k.SignedBlocksWindowDuration(ctx))

	if !signed {
		missedBlocks = append(missedBlocks, ctx.BlockTime())
	}

	k.setMissingBlocksForValidator(consAddr, missedBlocks)
	missedBlockCount := sdk.NewInt(int64(len(missedBlocks))).ToDec()

	missedRatio := missedBlockCount.QuoInt64(blockCount)

	// TODO Only do this if missing is true?
	minSignedPerWindow := k.MinSignedPerWindow(ctx)

	// if we are past the minimum height and the validator has missed too many blocks, punish them
	if minSignedPerWindow.LT(missedRatio) {
		validator := k.sk.ValidatorByConsAddr(ctx, consAddr)
		if validator != nil && !validator.IsJailed() {

			// Downtime confirmed: slash and jail the validator
			logger.Info(fmt.Sprintf("Validator %s is below signed blocks threshold of %d during the last %d",
				consAddr, k.MinSignedPerWindow(ctx), k.SignedBlocksWindowDuration(ctx)))

			// We need to retrieve the stake distribution which signed the block, so we subtract ValidatorUpdateDelay from the evidence height,
			// and subtract an additional 1 since this is the LastCommit.
			// Note that this *can* result in a negative "distributionHeight" up to -ValidatorUpdateDelay-1,
			// i.e. at the end of the pre-genesis block (none) = at the beginning of the genesis block.
			// That's fine since this is just used to filter unbonding delegations & redelegations.
			distributionHeight := height - sdk.ValidatorUpdateDelay - 1

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeSlash,
					sdk.NewAttribute(types.AttributeKeyAddress, consAddr.String()),
					sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", power)),
					sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueMissingSignature),
					sdk.NewAttribute(types.AttributeKeyJailed, consAddr.String()),
				),
			)

			k.slashValidator(ctx, consAddr, distributionHeight, power, k.SlashFractionDowntime(ctx))
			k.sk.Jail(ctx, consAddr)

			// fetch signing info
			signInfo, found := k.getValidatorSigningInfo(consAddr)
			if !found {
				panic(fmt.Sprintf("Expected signing info for validator %s but not found", consAddr))
			}

			signInfo.JailedUntil = ctx.BlockHeader().Time.Add(k.DowntimeJailDuration(ctx))

			// Reset number of blocks missed.
			k.deleteMissingBlocksForValidator(consAddr)
			k.SetValidatorSigningInfo(consAddr, signInfo)
		} else {
			// Validator was (a) not found or (b) already jailed, don't slash
			logger.Info(
				fmt.Sprintf("Validator %s would have been slashed for downtime, but was either not found in store or already jailed", consAddr),
			)
		}
	}
}

func (k Keeper) getMissingBlocksForValidator(address sdk.ConsAddress) []time.Time {
	key := []byte(fmt.Sprintf(dbKeyMissedByVal, address.String()))
	bz := k.database.Get(key)
	if len(bz) == 0 {
		return nil
	}

	b := bytes.NewBuffer(bz)
	dec := gob.NewDecoder(b)

	res := make([]time.Time, 0)
	err := dec.Decode(&res)
	if err != nil {
		panic(err)
	}

	return res
}

func (k Keeper) setMissingBlocksForValidator(address sdk.ConsAddress, missingBlocks []time.Time) {
	bz := new(bytes.Buffer)
	enc := gob.NewEncoder(bz)
	err := enc.Encode(missingBlocks)
	if err != nil {
		panic(err)
	}

	key := []byte(fmt.Sprintf(dbKeyMissedByVal, address.String()))
	k.database.Set(key, bz.Bytes())
}

func (k Keeper) deleteMissingBlocksForValidator(address sdk.ConsAddress) {
	key := []byte(fmt.Sprintf(dbKeyMissedByVal, address.String()))
	k.database.Delete(key)
}

func (k Keeper) handlePendingPenalties(ctx sdk.Context, vfn func() map[string]bool) {
	activePenalties := k.getPendingPenalties()
	if len(activePenalties) == 0 {
		return
	}

	validatorSet := vfn()
	for val, coins := range activePenalties {
		if _, present := validatorSet[val]; present {
			// Penalized validator is still in the validator set. Do not pay out slashing fine.
			continue
		}

		delete(activePenalties, val)

		err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.PenaltyAccount, k.feeModuleName, coins)
		if err != nil {
			panic(err)
		}
	}

	k.setPendingPenalties(activePenalties)
}

func (k Keeper) setPendingPenalties(penalties map[string]sdk.Coins) {
	if len(penalties) == 0 {
		k.database.Delete([]byte(dbKeyPendingPenalties))
		return
	}

	bz := k.cdc.MustMarshalJSON(penalties)
	k.database.Set([]byte(dbKeyPendingPenalties), bz)
}

func (k Keeper) getPendingPenalties() map[string]sdk.Coins {
	bz := k.database.Get([]byte(dbKeyPendingPenalties))
	if len(bz) == 0 {
		return make(map[string]sdk.Coins)
	}

	activePenalties := make(map[string]sdk.Coins)
	k.cdc.MustUnmarshalJSON(bz, &activePenalties)

	return activePenalties
}

func (k Keeper) slashValidator(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor sdk.Dec) {
	k.sk.Slash(ctx, consAddr, infractionHeight, power, slashFactor)

	// Mint the slashed coins and assign them to the distribution pool.
	slashAmount := calculateSlashingAmount(power, slashFactor)
	stakingDenom := k.sk.BondDenom(ctx)
	coins := sdk.NewCoins(sdk.NewCoin(stakingDenom, slashAmount))
	err := k.supplyKeeper.MintCoins(ctx, types.ModuleName, coins)
	if err != nil {
		panic(err)
	}

	err = k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.PenaltyAccount, coins)
	if err != nil {
		panic(err)
	}

	activePenalties := k.getPendingPenalties()
	activePenalties[consAddr.String()] = coins

	k.setPendingPenalties(activePenalties)
}

// Adopted from cosmos-sdk/x/staking/keeper/slash.go
func calculateSlashingAmount(power int64, slashFactor sdk.Dec) sdk.Int {
	amount := sdk.TokensFromConsensusPower(power)
	slashAmountDec := amount.ToDec().Mul(slashFactor)
	slashAmount := slashAmountDec.TruncateInt()
	return slashAmount
}

func (k Keeper) addPubkey(ctx sdk.Context, pubkey crypto.PubKey) {
	addr := pubkey.Address()
	k.setAddrPubkeyRelation(ctx, addr, pubkey)
}

func (k Keeper) getPubkey(ctx sdk.Context, address crypto.Address) (crypto.PubKey, error) {
	store := ctx.KVStore(k.storeKey)
	var pubkey crypto.PubKey
	err := k.cdc.UnmarshalBinaryLengthPrefixed(store.Get(types.GetAddrPubkeyRelationKey(address)), &pubkey)
	if err != nil {
		return nil, fmt.Errorf("address %s not found", sdk.ConsAddress(address))
	}
	return pubkey, nil
}

func (k Keeper) setAddrPubkeyRelation(ctx sdk.Context, addr crypto.Address, pubkey crypto.PubKey) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(pubkey)
	store.Set(types.GetAddrPubkeyRelationKey(addr), bz)
}

func (k Keeper) deleteAddrPubkeyRelation(ctx sdk.Context, addr crypto.Address) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetAddrPubkeyRelationKey(addr))
}
