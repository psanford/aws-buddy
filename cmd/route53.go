package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/spf13/cobra"
)

func route53Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "route53",
		Short: "Route53 Commands",
	}

	cmd.AddCommand(route53ListCommand())

	return &cmd
}

func route53ListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List Records",
		Run:   route53ListRecords,
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json ouput")

	return &cmd
}

func route53ListRecords(cmd *cobra.Command, args []string) {
	svc := route53.New(session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	err := svc.ListHostedZonesPages(nil, func(zoneOut *route53.ListHostedZonesOutput, more bool) bool {
		for _, zone := range zoneOut.HostedZones {

			listRRS := route53.ListResourceRecordSetsInput{
				HostedZoneId: zone.Id,
			}
			svc.ListResourceRecordSetsPages(&listRRS, func(recOut *route53.ListResourceRecordSetsOutput, more bool) bool {
				for _, rrs := range recOut.ResourceRecordSets {
					if jsonOutput {
						jsonOut.Encode(rrs)
					} else {
						for _, val := range rrs.ResourceRecords {
							fmt.Printf("%-60.60s %5.5s %6d %s\n", *rrs.Name, *rrs.Type, *rrs.TTL, *val.Value)
						}
					}
				}
				return true
			})

		}
		return true
	})

	if err != nil {
		log.Fatalf("ListHostedZones error: %s", err)
	}
}
