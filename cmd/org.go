package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/spf13/cobra"
)

func orgCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "org",
		Short: "Organization Commands",
	}

	cmd.AddCommand(orgListAccountsCommand())

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
