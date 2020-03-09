package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	cmd.AddCommand(orgEachEC2ListAccountCommand())

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

func orgEachEC2ListAccountCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "each_ec2_list",
		Short: "Run ec2 list against each account",
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
	var (
		origAccessKeyID     = os.Getenv("AWS_ACCESS_KEY_ID")
		origSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
		origSessionToken    = os.Getenv("AWS_SESSION_TOKEN")
	)

	resetEnv := func() {
		os.Setenv("AWS_ACCESS_KEY_ID", origAccessKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", origSecretAccessKey)
		os.Setenv("AWS_SESSION_TOKEN", origSessionToken)
	}

	defer resetEnv()

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
			fmt.Fprintf(os.Stderr, "%s %s\n", *account.Id, *account.Name)
			resetEnv()

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

			os.Setenv("AWS_ACCESS_KEY_ID", *resp.Credentials.AccessKeyId)
			os.Setenv("AWS_SECRET_ACCESS_KEY", *resp.Credentials.SecretAccessKey)
			os.Setenv("AWS_SESSION_TOKEN", *resp.Credentials.SessionToken)

			ec2ListAction(nil, nil)
		}

		resetEnv()
		return true
	})

	if err != nil {
		log.Fatalf("ListAccounts error: %s", err)
	}
}
