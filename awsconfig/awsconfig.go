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

	cmd.AddCommand(queryCommand())
	return &cmd
}

func queryCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "query_eni_by_public_ip",
		Short: "Query for ENIs matching a public ip",
		Run:   queryAction,
	}

	cmd.Flags().StringVarP(&aggregatorName, "aggregator-name", "", "AllAccounts", "AWS Config Aggretator Name")

	return &cmd
}

var (
	aggregatorName string
)

func queryAction(cmd *cobra.Command, args []string) {
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
	fmt.Printf("query: <%s>\n", query)

	input := &configservice.SelectAggregateResourceConfigInput{
		ConfigurationAggregatorName: &aggregatorName,
		Expression:                  &query,
	}
	err := svc.SelectAggregateResourceConfigPages(input, func(resp *configservice.SelectAggregateResourceConfigOutput, b bool) bool {
		for _, result := range resp.Results {
			fmt.Println(result)
		}

		return true
	})
	if err != nil {
		log.Fatal(err)
	}
}
