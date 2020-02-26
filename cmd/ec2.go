package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

func ec2Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "ec2",
		Short: "EC2 Commands",
	}

	cmd.AddCommand(ec2ListCommand())
	cmd.AddCommand(securityGroupCommand())
	return &cmd
}

func ec2ListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List instances",
		Run:   ec2ListAction,
	}

	return &cmd
}

func ec2ListAction(cmd *cobra.Command, args []string) {
	sess, err := session.NewSession(&aws.Config{Region: &region})
	if err != nil {
		log.Fatalf("AWS NewSession error: %s", err)
	}
	svc := ec2.New(sess)

	err = svc.DescribeInstancesPages(nil, func(output *ec2.DescribeInstancesOutput, lastPage bool) bool {
		for _, res := range output.Reservations {
			for _, inst := range res.Instances {
				instType := shortType(*inst.InstanceType)
				state := shortState(*inst.State.Name)

				tags := make(map[string]string)
				for _, t := range inst.Tags {
					tags[*t.Key] = *t.Value
				}
				name := tags["Name"]
				az := *inst.Placement.AvailabilityZone

				var (
					privateIPs     []string
					publicIPs      []string
					securityGroups []string
				)

				for _, iface := range inst.NetworkInterfaces {
					for _, privIP := range iface.PrivateIpAddresses {
						privateIPs = append(privateIPs, *privIP.PrivateIpAddress)
						if privIP.Association != nil {
							publicIPs = append(publicIPs, *privIP.Association.PublicIp)
						}
					}
				}

				for _, sg := range inst.SecurityGroups {
					securityGroups = append(securityGroups, *sg.GroupName)
				}

				fmt.Printf("%-35.35s %6.6s %4.4s %3.3s %15s %15s %s\n", name, instType, shortAZ(az), state, strings.Join(privateIPs, ","), strings.Join(publicIPs, ","), strings.Join(securityGroups, ","))
			}
		}
		return true
	})
	if err != nil {
		log.Fatalf("DescribeInstance error: %s", err)
	}
}

func shortAZ(fullAZ string) string {
	parts := strings.SplitN(fullAZ, "-", 3)

	for i, part := range parts {
		if i == len(parts)-1 {
			break
		}
		parts[i] = part[:1]
	}
	return strings.Join(parts, "")
}

func shortState(state string) string {
	if state == "stopping" {
		state = "sin"
	} else if state == "stopped" {
		state = "sed"
	} else {
		state = state[:3]
	}
	return state
}

var typeReplacer = strings.NewReplacer(
	"large", "l",
	"medium", "m",
	"metal", "⛁",
	"micro", "μ",
	"nano", "n",
	"small", "s",
)

func shortType(fullType string) string {
	return typeReplacer.Replace(fullType)
}
