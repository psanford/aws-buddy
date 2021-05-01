package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/psanford/aws-buddy/cost"
	"github.com/psanford/aws-buddy/ec2"
	"github.com/psanford/aws-buddy/iam"
	"github.com/psanford/aws-buddy/org"
	"github.com/psanford/aws-buddy/parameterstore"
	"github.com/psanford/aws-buddy/route53"
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

	rootCmd.AddCommand(ec2.Command())
	rootCmd.AddCommand(org.Command())
	rootCmd.AddCommand(route53.Command())
	rootCmd.AddCommand(cost.Command())
	rootCmd.AddCommand(iam.Command())
	rootCmd.AddCommand(parameterstore.Command())
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

func session() *awssession.Session {
	sess, err := awssession.NewSession(&aws.Config{Region: &region})
	if err != nil {
		log.Fatalf("AWS NewSession error: %s", err)
	}

	return sess
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	var result string
	fmt.Scanln(&result)

	return result == "y" || result == "Y" || result == "yes" || result == "Yes"
}
