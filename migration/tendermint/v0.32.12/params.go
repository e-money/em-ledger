package types

import (
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/types"
)

const (
	// MaxBlockSizeBytes is the maximum permitted size of the blocks.
	MaxBlockSizeBytes = 104857600 // 100MB

	// BlockPartSizeBytes is the size of one block part.
	BlockPartSizeBytes = 65536 // 64kB

	// MaxBlockPartsCount is the maximum number of block parts.
	MaxBlockPartsCount = (MaxBlockSizeBytes / BlockPartSizeBytes) + 1
)

// ConsensusParams contains consensus critical parameters that determine the
// validity of blocks.
type ConsensusParams struct {
	Block     BlockParams     `json:"block"`
	Evidence  EvidenceParams  `json:"evidence"`
	Validator ValidatorParams `json:"validator"`
}

// BlockParams define limits on the block size and gas plus minimum time
// between blocks.
type BlockParams struct {
	MaxBytes int64 `json:"max_bytes"`
	MaxGas   int64 `json:"max_gas"`
	// Minimum time increment between consecutive blocks (in milliseconds)
	// Not exposed to the application.
	TimeIotaMs int64 `json:"time_iota_ms"`
}

// EvidenceParams determine how we handle evidence of malfeasance.
type EvidenceParams struct {
	MaxAge int64 `json:"max_age"` // only accept new evidence more recent than this
}

// ValidatorParams restrict the public key types validators can use.
// NOTE: uses ABCI pubkey naming, not Amino names.
type ValidatorParams struct {
	PubKeyTypes []string `json:"pub_key_types"`
}

// DefaultConsensusParams returns a default ConsensusParams.
func DefaultConsensusParams() *ConsensusParams {
	return &ConsensusParams{
		DefaultBlockParams(),
		DefaultEvidenceParams(),
		DefaultValidatorParams(),
	}
}

// DefaultBlockParams returns a default BlockParams.
func DefaultBlockParams() BlockParams {
	return BlockParams{
		MaxBytes:   22020096, // 21MB
		MaxGas:     -1,
		TimeIotaMs: 1000, // 1s
	}
}

// DefaultEvidenceParams Params returns a default EvidenceParams.
func DefaultEvidenceParams() EvidenceParams {
	return EvidenceParams{
		MaxAge: 100000, // 27.8 hrs at 1block/s
	}
}

// DefaultValidatorParams returns a default ValidatorParams, which allows
// only ed25519 pubkeys.
func DefaultValidatorParams() ValidatorParams {
	return ValidatorParams{[]string{types.ABCIPubKeyTypeEd25519}}
}

// Validate validates the ConsensusParams to ensure all values are within their
// allowed limits, and returns an error if they are not.
func (params *ConsensusParams) Validate() error {
	if params.Block.MaxBytes <= 0 {
		return errors.Errorf("Block.MaxBytes must be greater than 0. Got %d",
			params.Block.MaxBytes)
	}
	if params.Block.MaxBytes > MaxBlockSizeBytes {
		return errors.Errorf("Block.MaxBytes is too big. %d > %d",
			params.Block.MaxBytes, MaxBlockSizeBytes)
	}

	if params.Block.MaxGas < -1 {
		return errors.Errorf("Block.MaxGas must be greater or equal to -1. Got %d",
			params.Block.MaxGas)
	}

	if params.Block.TimeIotaMs <= 0 {
		return errors.Errorf("Block.TimeIotaMs must be greater than 0. Got %v",
			params.Block.TimeIotaMs)
	}

	if params.Evidence.MaxAge <= 0 {
		return errors.Errorf("EvidenceParams.MaxAge must be greater than 0. Got %d",
			params.Evidence.MaxAge)
	}

	if len(params.Validator.PubKeyTypes) == 0 {
		return errors.New("len(Validator.PubKeyTypes) must be greater than 0")
	}

	// Check if keyType is a known ABCIPubKeyType
	for i := 0; i < len(params.Validator.PubKeyTypes); i++ {
		keyType := params.Validator.PubKeyTypes[i]
		if _, ok := types.ABCIPubKeyTypesToAminoNames[keyType]; !ok {
			return errors.Errorf("params.Validator.PubKeyTypes[%d], %s, is an unknown pubkey type",
				i, keyType)
		}
	}

	return nil
}
