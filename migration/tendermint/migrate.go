package tendermint

import (
	tmlegacy "github.com/e-money/em-ledger/migration/tendermint/v0.32.12"
	tmtypes "github.com/tendermint/tendermint/types"
	"time"
)

func ToV033(importGenesis string) (tmtypes.GenesisDoc, error) {
	// Previous mainnet used a legacy version of Tendermint.
	gendoc, err := tmlegacy.GenesisDocFromFile(importGenesis)
	if err != nil {
		return tmtypes.GenesisDoc{}, err
	}

	newConsParams := tmtypes.ConsensusParams{
		Block: tmtypes.BlockParams(gendoc.ConsensusParams.Block),
		Evidence: tmtypes.EvidenceParams{
			MaxAgeDuration:  2 * time.Hour,
			MaxAgeNumBlocks: gendoc.ConsensusParams.Evidence.MaxAge,
		},
		Validator: tmtypes.ValidatorParams(gendoc.ConsensusParams.Validator),
	}

	return tmtypes.GenesisDoc{
		GenesisTime:     gendoc.GenesisTime,
		ChainID:         gendoc.ChainID,
		ConsensusParams: &newConsParams,
		Validators:      gendoc.Validators,
		AppHash:         gendoc.AppHash,
		AppState:        gendoc.AppState,
	}, nil
}
