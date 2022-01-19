package awsconfig

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "config",
		Short: "AWS Config Commands",
	}

	cmd.AddCommand(queryPublicIPCommand())
	cmd.AddCommand(queryResourceIDCommand())
	return &cmd
}

func queryPublicIPCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "query_eni_by_public_ip",
		Short: "Query for ENIs matching a public ip",
		Run:   queryPublicIPAction,
	}

	cmd.Flags().StringVarP(&aggregatorName, "aggregator-name", "", "AllAccounts", "AWS Config Aggretator Name")

	return &cmd
}

var (
	aggregatorName string
)

func queryPublicIPAction(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatalf("Usage: query_eni_by_public_ip <PUBLIC_IP>")
	}

	publicIP := args[0]

	svc := configservice.New(config.Session())

	queryTmpl := `SELECT
  resourceId,
  resourceName,
  resourceType,
  tags,
  configuration.privateIpAddress,
  configuration.association.publicIp,
  configuration.attachment.instanceId,
  availabilityZone,
  configuration
WHERE
  resourceType = 'AWS::EC2::NetworkInterface'
  AND configuration.association.publicIp = '%s'
`

	query := fmt.Sprintf(queryTmpl, publicIP)

	input := &configservice.SelectAggregateResourceConfigInput{
		ConfigurationAggregatorName: &aggregatorName,
		Expression:                  &query,
	}
	err := svc.SelectAggregateResourceConfigPages(input, func(resp *configservice.SelectAggregateResourceConfigOutput, b bool) bool {
		for _, result := range resp.Results {
			fmt.Println(*result)
		}

		return true
	})
	if err != nil {
		log.Fatal(err)
	}
}

func queryResourceIDCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "query_by_id",
		Short: "Query by resource id for resoucre",
		Run:   queryResourceIDAction,
	}

	cmd.Flags().StringVarP(&aggregatorName, "aggregator-name", "", "AllAccounts", "AWS Config Aggretator Name")

	return &cmd
}

func queryResourceIDAction(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatalf("Usage: query_by_id <RESOURCE_ID>")
	}

	resourceID := args[0]

	svc := configservice.New(config.Session())

	queryTmpl := `
SELECT
  resourceId,
  resourceName,
  resourceType,
  accountId,
  arn,
  configuration.instanceType,
  tags,
  availabilityZone,
  configuration.state.name
WHERE
  resourceId = '%s'
`

	query := fmt.Sprintf(queryTmpl, resourceID)

	input := &configservice.SelectAggregateResourceConfigInput{
		ConfigurationAggregatorName: &aggregatorName,
		Expression:                  &query,
	}
	err := svc.SelectAggregateResourceConfigPages(input, func(resp *configservice.SelectAggregateResourceConfigOutput, b bool) bool {
		for _, result := range resp.Results {
			fmt.Println(*result)
		}

		return true
	})
	if err != nil {
		log.Fatal(err)
	}
}