package org

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

func scpCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "scp",
		Short: "Service Control Policy Commands",
	}

	cmd.AddCommand(scpListCommand())
	cmd.AddCommand(scpShowCommand())

	return &cmd
}

func scpListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List Service Control Policies",
		Run:   scpListAction,
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json output")

	return &cmd
}

func scpListAction(cmd *cobra.Command, args []string) {
	svc := organizations.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	err := svc.ListPoliciesPages(&organizations.ListPoliciesInput{
		Filter: aws.String("SERVICE_CONTROL_POLICY"),
	}, func(output *organizations.ListPoliciesOutput, lastPage bool) bool {
		for _, policy := range output.Policies {
			if jsonOutput {
				jsonOut.Encode(policy)
			} else {
				fmt.Printf("%20s %s\n", *policy.Id, *policy.Name)
			}
		}

		return true
	})

	if err != nil {
		log.Fatalf("ListPolicies error: %s", err)
	}
}

func scpShowCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "show <policy-id>",
		Short: "Show Service Control Policy",
		Run:   scpShowAction,
	}

	return &cmd
}

func scpShowAction(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatalf("Usage: show <policy-id>")
	}

	policyID := args[0]

	svc := organizations.New(config.Session())

	output, err := svc.DescribePolicy(&organizations.DescribePolicyInput{
		PolicyId: aws.String(policyID),
	})

	if err != nil {
		log.Fatalf("DescribePolicy error: %s", err)
	}

	policy := output.Policy

	fmt.Printf("========[ %s ]===================\n", *policy.PolicySummary.Name)
	fmt.Printf("id          : %s\n", *policy.PolicySummary.Id)
	fmt.Printf("name        : %s\n", *policy.PolicySummary.Name)
	fmt.Printf("description : %s\n", *policy.PolicySummary.Description)
	fmt.Printf("type        : %s\n", *policy.PolicySummary.Type)
	fmt.Printf("aws managed : %t\n", *policy.PolicySummary.AwsManaged)
	fmt.Printf("Content     :\n")

	var content any
	err = json.Unmarshal([]byte(*policy.Content), &content)
	if err != nil {
		log.Fatalf("Failed to parse policy content: %s", err)
	}

	contentJSON, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		log.Fatalf("Failed to format policy content: %s", err)
	}
	fmt.Println(string(contentJSON))
}
