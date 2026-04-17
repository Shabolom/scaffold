package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "CLI для генерации структуры Go-проекта",
	Long:  "CLI для генерации структуры Go-проекта по заданному шаблону",
}

func Execute() error {
	return rootCmd.Execute()
}
