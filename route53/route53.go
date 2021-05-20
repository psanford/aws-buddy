package route53

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	filterZone string
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "route53",
		Short: "Route53 Commands",
	}

	cmd.AddCommand(route53ListCommand())
	cmd.AddCommand(route53ListZonesCommand())

	return &cmd
}

func route53ListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List Records",
		Run:   route53ListRecords,
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json ouput")
	cmd.Flags().StringVarP(&filterZone, "zone", "", "", "Filter by zone name")

	return &cmd
}

func route53ListRecords(cmd *cobra.Command, args []string) {
	svc := route53.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	if filterZone != "" && !strings.HasSuffix(filterZone, ".") {
		filterZone += "."
	}

	err := svc.ListHostedZonesPages(nil, func(zoneOut *route53.ListHostedZonesOutput, more bool) bool {
		for _, zone := range zoneOut.HostedZones {

			if filterZone != "" && *zone.Name != filterZone {
				continue
			}

			listRRS := route53.ListResourceRecordSetsInput{
				HostedZoneId: zone.Id,
			}
			svc.ListResourceRecordSetsPages(&listRRS, func(recOut *route53.ListResourceRecordSetsOutput, more bool) bool {
				for _, rrs := range recOut.ResourceRecordSets {
					if jsonOutput {
						jsonOut.Encode(rrs)
					} else {
						var (
							recordName string
							recordType string
							ttl        int64
						)

						if rrs.Name != nil {
							recordName = *rrs.Name
						}

						if rrs.Type != nil {
							recordType = *rrs.Type
						}

						if rrs.TTL != nil {
							ttl = *rrs.TTL
						}

						for _, val := range rrs.ResourceRecords {
							if val.Value != nil {
								fmt.Printf("%-60.60s %5.5s %6d %s\n", recordName, recordType, ttl, *val.Value)
							}

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

func route53ListZonesCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "zones",
		Short: "List Zones",
		Run:   route53ListZones,
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json ouput")

	return &cmd
}

func route53ListZones(cmd *cobra.Command, args []string) {
	svc := route53.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	err := svc.ListHostedZonesPages(nil, func(zoneOut *route53.ListHostedZonesOutput, more bool) bool {
		for _, zone := range zoneOut.HostedZones {

			if jsonOutput {
				jsonOut.Encode(zone)
			} else {
				fmt.Printf("%-40.40s %6.6d %s\n", *zone.Id, *zone.ResourceRecordSetCount, *zone.Name)
			}
		}

		return true
	})

	if err != nil {
		log.Fatalf("ListHostedZones error: %s", err)
	}
}
