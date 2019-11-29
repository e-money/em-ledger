// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build fast_consensus

package main

import (
	"fmt"
	"github.com/spf13/viper"
)

func init() {
	previousConfig := configureConsensus
	configureConsensus = func() {
		fmt.Println("Overriding consensus parameters to achieve 4 second block time")
		previousConfig()

		viper.Set("consensus.create_empty_blocks_interval", "4s")
	}
}
