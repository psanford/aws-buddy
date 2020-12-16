package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	cmd.Flags().StringVarP(&orgListFileName, "org-list", "", "", "File with list of org ids (empty means use the current accounts org list)")
	cmd.Flags().StringVarP(&externalCommand, "external-cmd", "", "", "External command to run instead of a buddy command")

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
	var cmdPath string
	if externalCommand != "" {
		p, err := exec.LookPath(externalCommand)
		if err != nil {
			log.Fatalf("Failed to find full path for cmd: %s %s", externalCommand, err)
		}
		cmdPath = p
	} else {
		buddyPath, err := os.Executable()
		if err != nil {
			log.Fatalf("Failed to find my own executable: %s", err)
		}
		cmdPath = buddyPath
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

	var orgIDs []orgInfo

	if orgListFileName != "" {
		data, err := ioutil.ReadFile(orgListFileName)
		if err != nil {
			log.Fatalf("Read %s err: %s", orgListFileName, err)
		}

		lines := strings.Split(string(data), "\n")
		for _, l := range lines {
			fields := strings.Fields(l)

			if len(fields) > 1 {
				orgIDs = append(orgIDs, orgInfo{
					id:   fields[0],
					name: fields[1],
					arn:  fmt.Sprintf("arn:aws:organizations::%s:", fields[0]),
				})
			} else if len(fields) == 1 {
				orgIDs = append(orgIDs, orgInfo{
					id:  fields[0],
					arn: fmt.Sprintf("arn:aws:organizations::%s:", fields[0]),
				})
			}
		}
	} else {
		err = svc.ListAccountsPages(nil, func(output *organizations.ListAccountsOutput, lastPage bool) bool {
			for _, account := range output.Accounts {
				if *account.Id == rootAccountID {
					continue
				}
				orgIDs = append(orgIDs, orgInfo{
					id:   *account.Id,
					name: *account.Name,
					arn:  *account.Arn,
				})
			}
			return true
		})
		if err != nil {
			log.Fatalf("ListAccount err: %s", err)
		}
	}

	var errors []error

	for _, orgInfo := range orgIDs {
		fmt.Fprintf(os.Stderr, "# Account %s %s\n", orgInfo.arn, orgInfo)

		// get the correct arn prefix (for other aws partitions)
		arnParts := strings.SplitN(orgInfo.arn, ":", 3)

		roleARN := fmt.Sprintf("%s:%s:iam::%s:role/%s", arnParts[0], arnParts[1], orgInfo.id, assumeRoleName)
		roleSessionName := fmt.Sprintf("%d", time.Now().UTC().UnixNano())

		assumeRoleInput := &sts.AssumeRoleInput{
			RoleArn:         aws.String(roleARN),
			RoleSessionName: aws.String(roleSessionName),
			DurationSeconds: aws.Int64(900),
		}

		resp, err := stsClient.AssumeRole(assumeRoleInput)
		if err != nil {
			errors = append(errors, fmt.Errorf("Assume role error: %s", err))
			log.Printf("Assume role error: %s", err)
			continue
		}

		cmd := exec.Command(cmdPath, args...)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", *resp.Credentials.AccessKeyId),
			fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", *resp.Credentials.SecretAccessKey),
			fmt.Sprintf("AWS_SESSION_TOKEN=%s", *resp.Credentials.SessionToken),
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Fprintf(os.Stderr, "# Running %s %s\n", cmdPath, strings.Join(args, " "))
		err = cmd.Run()
		if err != nil {
			fullErr := fmt.Errorf("Failed to execute cmd on account %s: %s cmd: %s %s", orgInfo, cmdPath, err, strings.Join(args, " "))
			errors = append(errors, fullErr)
			log.Print(fullErr)
		}
	}

	if len(errors) > 0 {
		log.Printf("Errors:")
		for _, err := range errors {
			log.Print(err)
		}
		os.Exit(1)
	}
}

type orgInfo struct {
	id   string
	name string
	arn  string
}

func (i orgInfo) String() string {
	return fmt.Sprintf("%s/%s", i.id, i.name)
}
