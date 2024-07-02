package eni

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "eni",
		Short: "ENI Commands",
	}

	cmd.AddCommand(showENICommand())
	cmd.AddCommand(listENICommand())

	return &cmd
}

var (
	jsonOutput bool
)

func showENICommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "show <eni-id>",
		Short: "Show eni details",
		Run:   showENIAction,
	}
	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json ouput")

	return &cmd
}

func showENIAction(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatalf("Usage: show <eni-id> [...<eni-id>]")
	}

	eniIDs := args

	svc := ec2.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: aws.StringSlice(eniIDs),
	}
	err := svc.DescribeNetworkInterfacesPages(input, func(dnio *ec2.DescribeNetworkInterfacesOutput, b bool) bool {
		for _, eni := range dnio.NetworkInterfaces {
			jsonOut.Encode(eni)
		}
		return true
	})
	if err != nil {
		log.Fatalf("DescribeNetworkInterfaces err: %s", err)
	}
}

func listENICommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List eni devices",
		Run:   listENIAction,
	}
	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "List raw json ouput")

	return &cmd
}

func listENIAction(cmd *cobra.Command, args []string) {
	svc := ec2.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	input := &ec2.DescribeNetworkInterfacesInput{
		MaxResults: aws.Int64(500),
	}
	err := svc.DescribeNetworkInterfacesPages(input, func(dnio *ec2.DescribeNetworkInterfacesOutput, b bool) bool {
		for _, eni := range dnio.NetworkInterfaces {
			jsonOut.Encode(eni)
		}
		return true
	})
	if err != nil {
		log.Fatalf("DescribeNetworkInterfaces err: %s", err)
	}
}
