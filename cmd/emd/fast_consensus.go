// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build fast_consensus

package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func init() {
	previousConfig := configureConsensus
	configureConsensus = func() {
		previousConfig()
		fmt.Fprintln(os.Stderr, " --- Overriding consensus parameters for tests!")

		configChanges := map[string]string{
			"consensus.create_empty_blocks_interval": "4s",
			"consensus.timeout_commit":               "1500ms",
		}

		for k, v := range configChanges {
			fmt.Fprintln(os.Stderr, " --- Overriding:", k, v)
			viper.Set(k, v)
		}
	}
}
