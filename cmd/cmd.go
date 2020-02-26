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
	rootCmd.AddCommand(completionCommand())

	return rootCmd.Execute()
}

func completionCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "completion",
		Short: "Generates bash completion scripts",
		Long: `To load completion run

. <(aws-buddy completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(aws-buddy completion)
`,
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
	}

	return cmd
}
