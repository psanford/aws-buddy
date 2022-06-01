package awsconfig

import (
	"fmt"
	"log"
	"sort"

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
	cmd.AddCommand(resourceInventoryByTypeCommand())
	cmd.AddCommand(listResourceTypesCommand())
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
	resourceType   string
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

func listResourceTypesCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "resource_types",
		Short: "List the resource types for aws config",
		Run:   listResourceTypes,
	}

	return &cmd
}

func listResourceTypes(cmd *cobra.Command, args []string) {
	types := configservice.ResourceType_Values()
	sort.Strings(types)
	for _, t := range types {
		fmt.Println(t)
	}
}

func resourceInventoryByTypeCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "inventory_by_type",
		Short: "List all config resources for a given type",
		Run:   resourceInventoryByType,
	}

	cmd.Flags().StringVarP(&aggregatorName, "aggregator-name", "", "AllAccounts", "AWS Config Aggretator Name")
	cmd.Flags().StringVarP(&resourceType, "type", "", "AWS::S3::Bucket", "Resource type (aws-buddy config resource_types)")

	return &cmd
}

func resourceInventoryByType(cmd *cobra.Command, args []string) {
	// resource types https://docs.aws.amazon.com/config/latest/developerguide/resource-config-reference.html

	svc := configservice.New(config.Session())

	queryTmpl := `
SELECT
  resourceId,
  resourceName,
  resourceType,
  accountId,
  arn,
  tags,
  availabilityZone
WHERE
  resourceType = '%s'
`

	query := fmt.Sprintf(queryTmpl, resourceType)

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
