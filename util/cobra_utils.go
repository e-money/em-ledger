// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package util

import (
	"strings"

	"github.com/spf13/cobra"
)

func RemoveCobraCommands(rootCmd *cobra.Command, commandNames ...string) {
	for _, commandName := range commandNames {
		commandPath := strings.Split(commandName, ".")
		cmd, _, err := rootCmd.Find(commandPath)
		if err != nil {
			panic(err)
		}

		cmd.Parent().RemoveCommand(cmd)
	}
}
