package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var region = "us-east-1"

var rootCmd = &cobra.Command{
	Use:   "aws-buddy",
	Short: "AWS tools",
}

func Execute() error {
	if os.Getenv("AWS_DEFAULT_REGION") != "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}

	rootCmd.AddCommand(ec2Command())
	return rootCmd.Execute()
}
