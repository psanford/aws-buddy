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

	cmd.AddCommand(queryIPCommand())
	cmd.AddCommand(queryResourceIDCommand())
	return &cmd
}

func queryIPCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "query_eni_by_ip",
		Short: "Query for ENIs matching an ip",
		Run:   queryIPAction,
	}

	cmd.Flags().StringVarP(&aggregatorName, "aggregator-name", "", "AllAccounts", "AWS Config Aggretator Name")

	return &cmd
}

var (
	aggregatorName string
)

func queryIPAction(cmd *cobra.Command, args []string) {
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
  availabilityZone,
  relationships,
  configuration
WHERE
  resourceType = 'AWS::EC2::NetworkInterface'
  and (
    configuration.association.publicIp = '%s'
    or configuration.privateIpAddresses.privateIpAddress = '%s'
  )
`

	query := fmt.Sprintf(queryTmpl, publicIP, publicIP)

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
