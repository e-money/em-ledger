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
