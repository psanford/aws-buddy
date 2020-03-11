package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/spf13/cobra"
)

func orgCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "org",
		Short: "Organization Commands",
	}

	cmd.AddCommand(orgListAccountsCommand())
	cmd.AddCommand(orgEachAccountCommand())

	return &cmd
}

func orgListAccountsCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List accounts",
		Run:   orgListAccountsAction,
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json ouput")

	return &cmd
}

func orgEachAccountCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "each",
		Short: "Run command against each account",
		Run:   orgEachAccountAction,
	}

	cmd.Flags().StringVarP(&assumeRoleName, "role", "", "", "Role name to assume in each account")

	return &cmd
}

func orgListAccountsAction(cmd *cobra.Command, args []string) {
	svc := organizations.New(session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	err := svc.ListAccountsPages(nil, func(output *organizations.ListAccountsOutput, lastPage bool) bool {
		for _, account := range output.Accounts {
			if jsonOutput {
				jsonOut.Encode(account)
			} else {
				fmt.Printf("%s %s\n", *account.Id, *account.Name)
			}
		}

		return true
	})

	if err != nil {
		log.Fatalf("ListAccounts error: %s", err)
	}
}

func orgEachAccountAction(cmd *cobra.Command, args []string) {
	buddyPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to find my own executable: %s", err)
	}

	svc := organizations.New(session())

	stsClient := sts.New(session())

	if assumeRoleName == "" {
		log.Fatal("-role is a required flag")
	}

	ident, err := stsClient.GetCallerIdentity(nil)
	if err != nil {
		log.Fatalf("DescribeAccount (root) error: %s", err)
	}
	rootAccountID := *ident.Account

	err = svc.ListAccountsPages(nil, func(output *organizations.ListAccountsOutput, lastPage bool) bool {
		for _, account := range output.Accounts {
			if *account.Id == rootAccountID {
				continue
			}
			fmt.Fprintf(os.Stderr, "# Account %s %s\n", *account.Id, *account.Name)

			roleARN := fmt.Sprintf("arn:aws:iam::%s:role/%s", *account.Id, assumeRoleName)
			roleSessionName := fmt.Sprintf("%d", time.Now().UTC().UnixNano())

			assumeRoleInput := &sts.AssumeRoleInput{
				RoleArn:         aws.String(roleARN),
				RoleSessionName: aws.String(roleSessionName),
				DurationSeconds: aws.Int64(900),
			}

			resp, err := stsClient.AssumeRole(assumeRoleInput)
			if err != nil {
				log.Fatalf("Assume role error: %s", err)
			}

			cmd := exec.Command(buddyPath, args...)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", *resp.Credentials.AccessKeyId),
				fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", *resp.Credentials.SecretAccessKey),
				fmt.Sprintf("AWS_SESSION_TOKEN=%s", *resp.Credentials.SessionToken),
			)

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			fmt.Fprintf(os.Stderr, "# Running %s %s\n", buddyPath, strings.Join(args, " "))
			err = cmd.Run()
			if err != nil {
				log.Fatalf("Failed to execute cmd on account %s/%s: %s, cmd: %s %s", *account.Id, *account.Name, buddyPath, err, strings.Join(args, " "))
			}
		}

		return true
	})

	if err != nil {
		log.Fatalf("ListAccounts error: %s", err)
	}
}
