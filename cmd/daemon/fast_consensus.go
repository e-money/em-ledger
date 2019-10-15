// +build fast_consensus

package main

import (
	"fmt"
	"github.com/spf13/viper"
)

func init() {
	previousConfig := configureConsensus
	configureConsensus = func() {
		fmt.Println("Overriding consensus parameters to achieve 5 second block time")
		previousConfig()

		viper.Set("consensus.create_empty_blocks_interval", "5s")
	}
}
