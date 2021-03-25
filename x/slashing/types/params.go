package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"time"
)

const (
	DefaultParamspace                 = ModuleName
	DefaultSignedBlocksWindowDuration = time.Hour
	DefaultDowntimeJailDuration       = DefaultSignedBlocksWindowDuration
)

var (
	DefaultMinSignedPerWindow      = sdk.NewDecWithPrec(1, 1)
	DefaultSlashFractionDoubleSign = sdk.NewDec(1).Quo(sdk.NewDec(20))
	DefaultSlashFractionDowntime   = sdk.NewDec(1).Quo(sdk.NewDec(1000))
)

var (
	//todo (reviewer): this was "SignedBlocksWindowDuration" before
	KeySignedBlocksWindow = slashingtypes.KeySignedBlocksWindow

	KeyMinSignedPerWindow      = slashingtypes.KeyMinSignedPerWindow
	KeyDowntimeJailDuration    = slashingtypes.KeyDowntimeJailDuration
	KeySlashFractionDoubleSign = slashingtypes.KeySlashFractionDoubleSign
	KeySlashFractionDowntime   = slashingtypes.KeySlashFractionDowntime
)

func DefaultParams() slashingtypes.Params {
	return slashingtypes.Params{
		SignedBlocksWindow:      DefaultSignedBlocksWindowDuration.Nanoseconds(),
		MinSignedPerWindow:      DefaultMinSignedPerWindow,
		DowntimeJailDuration:    DefaultDowntimeJailDuration,
		SlashFractionDoubleSign: DefaultSlashFractionDoubleSign,
		SlashFractionDowntime:   DefaultSlashFractionDowntime,
	}
}
