package testapp

import "encoding/json"

// Migrated from github/kava/app/genesis.go
// GenesisState represents the genesis state of the blockchain. It is a map from module names to module genesis states.
type GenesisState map[string]json.RawMessage
