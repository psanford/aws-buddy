package org

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
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

var (
	jsonOutput       bool
	assumeRoleName   string
	orgListFileName  string
	externalCommand  string
	includeAccts     bool
	includeSuspended bool
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "org",
		Short: "Organization Commands",
	}

	cmd.AddCommand(orgListAccountsCommand())
	cmd.AddCommand(orgEachAccountCommand())
	cmd.AddCommand(orgListOrgUnitsCommand())
	cmd.AddCommand(scpCommand())

	return &cmd
}

func orgListAccountsCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List accounts",
		Run:   orgListAccountsAction,
	}

	cmd.Flags().BoolVarP(&includeSuspended, "include-suspended", "", false, "Include suspended accounts")
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
	svc := organizations.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	err := svc.ListAccountsPages(nil, func(output *organizations.ListAccountsOutput, lastPage bool) bool {
		for _, account := range output.Accounts {
			if jsonOutput {
				jsonOut.Encode(account)
			} else {
				if *account.Status == "SUSPENDED" && !includeSuspended {
					continue
				}

				status := *account.Status
				if status == "ACTIVE" {
					status = ""
				} else {
					status = " " + status
				}

				fmt.Printf("%s %s%s\n", *account.Id, *account.Name, status)
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

	svc := organizations.New(config.Session())

	stsClient := sts.New(config.Session())

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
		data, err := os.ReadFile(orgListFileName)
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

func orgListOrgUnitsCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list-ou-tree",
		Short: "List organizational units",
		Run:   orgListOrgUnitsAction,
	}

	cmd.Flags().BoolVarP(&includeAccts, "include-accts", "", false, "Include Child accounts")

	return &cmd
}

func orgListOrgUnitsAction(cmd *cobra.Command, args []string) {
	svc := organizations.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	lri := organizations.ListRootsInput{}
	depth := -1
	err := svc.ListRootsPages(&lri, func(lro *organizations.ListRootsOutput, b bool) bool {
		depth += 1
		defer func() { depth -= 1 }()

		var handleOU func(loufpo *organizations.ListOrganizationalUnitsForParentOutput, b bool) bool
		handleOU = func(loufpo *organizations.ListOrganizationalUnitsForParentOutput, b bool) bool {
			depth += 1
			defer func() { depth -= 1 }()

			for _, ou := range loufpo.OrganizationalUnits {
				fmt.Printf("%s%s %s\n", strings.Repeat(" ", depth), *ou.Id, *ou.Name)

				if includeAccts {
					lafp := organizations.ListAccountsForParentInput{
						ParentId: ou.Id,
					}
					err := svc.ListAccountsForParentPages(&lafp, func(lafpo *organizations.ListAccountsForParentOutput, b bool) bool {
						for _, acct := range lafpo.Accounts {
							fmt.Printf("%s%s %s\n", strings.Repeat(" ", depth+1), *acct.Id, *acct.Name)
						}
						return true
					})
					if err != nil {
						log.Fatalf("list accounts for parent err: %s", err)
					}
				}

				loufpi := organizations.ListOrganizationalUnitsForParentInput{
					ParentId: ou.Id,
				}
				err := svc.ListOrganizationalUnitsForParentPages(&loufpi, handleOU)
				if err != nil {
					log.Fatalf("list ou for parent err: %s", err)
				}
			}

			return true
		}

		loufpo := organizations.ListOrganizationalUnitsForParentOutput{
			OrganizationalUnits: make([]*organizations.OrganizationalUnit, len(lro.Roots)),
		}

		for i, ou := range lro.Roots {
			loufpo.OrganizationalUnits[i] = &organizations.OrganizationalUnit{
				Arn:  ou.Arn,
				Id:   ou.Id,
				Name: ou.Name,
			}
		}

		handleOU(&loufpo, true)
		return true
	})
	if err != nil {
		log.Fatalf("list roots err: %s", err)
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
