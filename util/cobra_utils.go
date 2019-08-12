package util

import (
	"github.com/spf13/cobra"
	"strings"
)

func RemoveCobraCommands(rootCmd *cobra.Command, commandNames ...string) {
	for _, commandName := range commandNames {
		commandPath := strings.Split(commandName, ".")
		parent, child := rootCmd, rootCmd
		for _, segment := range commandPath {
			parent = child
			child = findSubCommand(segment, parent)
			if child == nil {
				return
			}
		}
		parent.RemoveCommand(child)
	}
}

func findSubCommand(name string, cmd *cobra.Command) *cobra.Command {
	for _, childCmd := range cmd.Commands() {
		if childCmd.Use == name {
			return childCmd
		}
	}
	return nil
}
