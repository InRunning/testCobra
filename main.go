/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/spf13/cobra"
	"github.com/testCobra/pcap"
)

func main() {
	rootCmd := cobra.Command{Use: "axyomctl"}
	rootCmd.AddCommand(pcap.NewCmdPcap())
	rootCmd.Execute()
}
