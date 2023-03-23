package pcap

import (
	"github.com/spf13/cobra"
)

var cmdPcap = &cobra.Command{
	Use:   "pcap",
	Short: "pcap ls/delete pcap files",
}

func NewCmdPcap() *cobra.Command {

	cmd := cmdPcap
	cmd.AddCommand(NewCmdLs())
	// cmdPcap.AddCommand(cmdDelete)

	return cmd
}
