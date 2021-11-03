package v09

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"
)

const (
	ModuleName = "slashing"
)

// GenesisState - all slashing state that must be provided at genesis
type GenesisState struct {
	Params       Params                          `json:"params" yaml:"params"`
	SigningInfos map[string]ValidatorSigningInfo `json:"signing_infos" yaml:"signing_infos"`
	MissedBlocks map[string][]MissedBlock        `json:"missed_blocks" yaml:"missed_blocks"`
}

// Params - used for initializing default parameter for slashing at genesis
type Params struct {
	MaxEvidenceAge             time.Duration `json:"max_evidence_age" yaml:"max_evidence_age"`
	SignedBlocksWindowDuration time.Duration `json:"signed_blocks_window_duration" yaml:"signed_blocks_window_duration"`
	MinSignedPerWindow         sdk.Dec       `json:"min_signed_per_window" yaml:"min_signed_per_window"`
	DowntimeJailDuration       time.Duration `json:"downtime_jail_duration" yaml:"downtime_jail_duration"`
	SlashFractionDoubleSign    sdk.Dec       `json:"slash_fraction_double_sign" yaml:"slash_fraction_double_sign"`
	SlashFractionDowntime      sdk.Dec       `json:"slash_fraction_downtime" yaml:"slash_fraction_downtime"`
}

// MissedBlock
type MissedBlock struct {
	Index  int64 `json:"index" yaml:"index"`
	Missed bool  `json:"missed" yaml:"missed"`
}

// Signing info for a validator
type ValidatorSigningInfo struct {
	Address     sdk.ConsAddress `json:"address" yaml:"address"`           // validator consensus address
	JailedUntil time.Time       `json:"jailed_until" yaml:"jailed_until"` // timestamp validator cannot be unjailed until
	Tombstoned  bool            `json:"tombstoned" yaml:"tombstoned"`     // whether or not a validator has been tombstoned (killed out of validator set)
}
