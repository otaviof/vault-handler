package main

import (
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Run:   runUploadCmd,
	Short: "",
	Long:  ``,
}

func runUploadCmd(cmd *cobra.Command, args []string) {

}

func init() {
	flags := uploadCmd.PersistentFlags()

	flags.String("input-dir", ".", "Input directory.")

	rootCmd.AddCommand(uploadCmd)
}
