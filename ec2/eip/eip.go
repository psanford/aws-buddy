package eip

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "ip",
		Short: "IP Commands",
	}

	cmd.AddCommand(ipListCommand())

	return &cmd
}

func ipListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "list ips by eni association",
		Run:   ipListAction,
	}
	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json ouput")

	return &cmd
}

func ipListAction(cmd *cobra.Command, args []string) {
	svc := ec2.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	err := svc.DescribeNetworkInterfacesPages(nil, func(output *ec2.DescribeNetworkInterfacesOutput, b bool) bool {
		for _, nic := range output.NetworkInterfaces {
			if jsonOutput {
				jsonOut.Encode(nic)
			} else {

				var (
					instanceID = "-"
					groups     []string
					publicIPs  []string
					privateIPs []string
					status     = "-"
					subnetID   = "-"
					vpcID      = "-"
				)

				if att := nic.Attachment; att != nil {
					if att.InstanceId != nil {
						instanceID = *att.InstanceId
					}
				}
				for _, g := range nic.Groups {
					name := fmt.Sprintf("%s (%s)", *g.GroupId, *g.GroupName)
					groups = append(groups, name)
				}
				for _, pip := range nic.PrivateIpAddresses {
					if pip.Association != nil && pip.Association.PublicIp != nil {
						publicIPs = append(publicIPs, *pip.Association.PublicIp)
					}
					if pip.PrivateIpAddress != nil {
						privateIPs = append(privateIPs, *pip.PrivateIpAddress)
					}
				}
				if nic.Status != nil {
					status = *nic.Status
				}
				if nic.SubnetId != nil {
					subnetID = *nic.SubnetId
				}
				if nic.VpcId != nil {
					vpcID = *nic.VpcId
				}

				fmt.Printf("%s %s %s %s %s %s %s\n", strings.Join(publicIPs, ","), strings.Join(privateIPs, ","), instanceID, strings.Join(groups, "-"), status, subnetID, vpcID)
			}

		}
		return true
	})
	if err != nil {
		log.Fatalf("DescribeSecurityGroups error: %s", err)
	}
}
