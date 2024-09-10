package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/psanford/aws-buddy/awsconfig"
	"github.com/psanford/aws-buddy/cost"
	"github.com/psanford/aws-buddy/ec2"
	"github.com/psanford/aws-buddy/iam"
	"github.com/psanford/aws-buddy/org"
	"github.com/psanford/aws-buddy/parameterstore"
	"github.com/psanford/aws-buddy/route53"
	"github.com/psanford/aws-buddy/s3"
	"github.com/psanford/aws-buddy/sqs"
	"github.com/psanford/aws-buddy/textract"
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
	if os.Getenv("AWS_REGION") != "" {
		region = os.Getenv("AWS_REGION")
	}

	os.Setenv("AWS_DEFAULT_REGION", region)

	rootCmd.AddCommand(ec2.Command())
	rootCmd.AddCommand(s3.Command())
	rootCmd.AddCommand(org.Command())
	rootCmd.AddCommand(route53.Command())
	rootCmd.AddCommand(cost.Command())
	rootCmd.AddCommand(iam.Command())
	rootCmd.AddCommand(parameterstore.Command())
	rootCmd.AddCommand(awsconfig.Command())
	rootCmd.AddCommand(sqs.Command())
	rootCmd.AddCommand(helpTreeCommand())
	rootCmd.AddCommand(textract.Command())

	return rootCmd.Execute()
}

func helpTreeCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "help-tree",
		Short: "Print Help for all commands",
		Run: func(cmd *cobra.Command, args []string) {
			var printHelp func(cmd *cobra.Command)

			printHelp = func(cmd *cobra.Command) {
				fmt.Println("\n========================================")
				cmd.Help()
				for _, childCmd := range cmd.Commands() {
					printHelp(childCmd)
				}
			}

			printHelp(rootCmd)
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
