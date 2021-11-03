package v040

import banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

func GetDenomMetaData() []banktypes.Metadata {
	return []banktypes.Metadata{
		{
			Description: "e-Money NGM staking token",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "ungm",
					Exponent: 0,
					Aliases:  nil,
				},
				{
					Denom:    "ngm",
					Exponent: 6,
					Aliases:  nil,
				},
			},
			Base:    "ungm",
			Display: "NGM",
		},
		{
			Base:        "echf",
			Description: "e-Money CHF stablecoin",
			Display:     "ECHF",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "echf",
					Exponent: 6,
					Aliases:  nil,
				},
			},
		},
		{
			Base:        "edkk",
			Description: "e-Money DKK stablecoin",
			Display:     "EDKK",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "edkk",
					Exponent: 6,
					Aliases:  nil,
				},
			},
		},
		{
			Base:        "eeur",
			Description: "e-Money EUR stablecoin",
			Display:     "EEUR",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "eeur",
					Exponent: 6,
					Aliases:  nil,
				},
			},
		},
		{
			Base:        "enok",
			Description: "e-Money NOK stablecoin",
			Display:     "ENOK",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "enok",
					Exponent: 6,
					Aliases:  nil,
				},
			},
		},
		{
			Base:        "esek",
			Description: "e-Money SEK stablecoin",
			Display:     "ESEK",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "esek",
					Exponent: 6,
					Aliases:  nil,
				},
			},
		},
	}
}
