package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

func ec2Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "ec2",
		Short: "EC2 Commands",
	}

	cmd.AddCommand(ec2ListCommand())
	cmd.AddCommand(ec2ShowCommand())
	cmd.AddCommand(securityGroupCommand())
	cmd.AddCommand(asgCommand())
	cmd.AddCommand(tagCommands())
	return &cmd
}

func ec2ListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List instances",
		Run:   ec2ListAction,
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json ouput")
	cmd.Flags().BoolVarP(&truncateFields, "truncate", "", true, "Trucate fields")
	cmd.Flags().BoolVarP(&verboseOutput, "verbose", "v", false, "Show verbose (multi-line) output")
	cmd.Flags().StringVarP(&filterFlag, "filter", "f", "", "Filter results by name or id")
	cmd.Flags().StringVarP(&filterNameFlag, "filter-name", "", "", "API Filter by Tag:Name")

	return &cmd
}

func ec2ListAction(cmd *cobra.Command, args []string) {
	showInstances(nil)
}

func ec2ShowCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "show <instance-id>",
		Short: "Show Instance",
		Run:   ec2ShowAction,
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json ouput")
	cmd.Flags().BoolVarP(&verboseOutput, "verbose", "v", false, "Show verbose (multi-line) output")
	cmd.Flags().StringVarP(&filterNameFlag, "filter-name", "", "", "API Filter by Tag:Name")

	return &cmd
}

func showInstances(input *ec2.DescribeInstancesInput) {
	svc := ec2.New(session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	if input == nil {
		input = &ec2.DescribeInstancesInput{}
	}
	if len(input.InstanceIds) == 0 {
		input.MaxResults = aws.Int64(1000)
	}

	if filterNameFlag != "" {
		if input.Filters == nil {
			input.Filters = []*ec2.Filter{}
		}
		input.Filters = append(input.Filters, &ec2.Filter{
			Name:   aws.String("tag:Name"),
			Values: []*string{&filterNameFlag},
		})
	}

	err := svc.DescribeInstancesPages(input, func(output *ec2.DescribeInstancesOutput, lastPage bool) bool {
		for _, inst := range instancesFromDesc(output) {
			tags := make(map[string]string)
			for _, t := range inst.Tags {
				tags[*t.Key] = *t.Value
			}
			name := tags["Name"]
			if filterFlag != "" {
				filterFlag = strings.ToLower(filterFlag)
				if strings.Index(strings.ToLower(name), filterFlag) < 0 && strings.Index(*inst.InstanceId, filterFlag) < 0 {
					continue
				}
			}

			if jsonOutput {
				jsonOut.Encode(inst)
			} else {
				instType := shortType(*inst.InstanceType)
				state := shortState(*inst.State.Name)

				az := *inst.Placement.AvailabilityZone

				var (
					privateIPs           []string
					publicIPs            []string
					securityGroupNames   []string
					securityGroupNameIDs []string
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
					securityGroupNames = append(securityGroupNames, *sg.GroupName)
					securityGroupNameIDs = append(securityGroupNameIDs, fmt.Sprintf("%s(%s)", *sg.GroupName, *sg.GroupId))
				}

				var ifaces []string
				for _, iface := range inst.NetworkInterfaces {
					ifaces = append(ifaces, *iface.NetworkInterfaceId)
				}

				if verboseOutput {
					fmt.Printf("========[ %s ]===================\n", *inst.InstanceId)
					fmt.Printf("name     : %s\n", name)
					fmt.Printf("id       : %s\n", *inst.InstanceId)
					fmt.Printf("type     : %s\n", instType)
					fmt.Printf("az       : %s\n", az)
					fmt.Printf("state    : %s\n", state)
					fmt.Printf("priv IPs : %s\n", strings.Join(privateIPs, ","))
					fmt.Printf("pub  IPs : %s\n", strings.Join(publicIPs, ","))
					fmt.Printf("SGs      : %s\n", strings.Join(securityGroupNameIDs, ","))
					fmt.Printf("Launch   : %s\n", inst.LaunchTime.Format(time.RFC3339))
					fmt.Printf("IFaces   : %s\n", strings.Join(ifaces, ";"))
					fmt.Printf("Tags     : %v\n", tags)
				} else {
					formatStr := "%s %-35.35s %6.6s %4.4s %3.3s %15s %15s %s\n"
					if !truncateFields {
						formatStr = "%s %s %s %s %s %s %s %s\n"
					}
					fmt.Printf(formatStr, *inst.InstanceId, name, instType, shortAZ(az), state, strings.Join(privateIPs, ","), strings.Join(publicIPs, ","), strings.Join(securityGroupNames, ","))
				}
			}
		}
		return true
	})
	if err != nil {
		log.Fatalf("DescribeInstance error: %s", err)
	}
}

var instanceIDRegex = regexp.MustCompile(`\Ai-[0-9a-f]+\z`)

func ec2ShowAction(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatalf("Usage: show <instance-id> [...<instance-id>]")
	}

	instanceIDs := args

	for _, instanceID := range instanceIDs {
		if !instanceIDRegex.MatchString(instanceID) {
			log.Fatalf("<instance-id> must be of the form i-[0-9a-f]+")
		}
	}

	input := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice(instanceIDs),
	}
	showInstances(input)
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

func instancesFromDesc(desc *ec2.DescribeInstancesOutput) []ec2.Instance {
	out := make([]ec2.Instance, 0)
	for _, res := range desc.Reservations {
		for _, instPtr := range res.Instances {
			inst := *instPtr
			out = append(out, inst)
		}
	}

	return out
}

func getInstance(instanceID string) (*ec2.Instance, error) {
	svc := ec2.New(session())
	input := ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("instance-id"),
				Values: []*string{&instanceID},
			},
		},
	}
	output, err := svc.DescribeInstances(&input)
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances err: %w", err)
	}

	instances := instancesFromDesc(output)
	if len(instances) < 1 {
		return nil, fmt.Errorf("No instance found")
	}
	if len(instances) > 1 {
		ids := make([]string, 0, len(instances))
		for _, inst := range instances {
			ids = append(ids, *inst.InstanceId)
		}
		return nil, fmt.Errorf("Multiple instances found found: %s", strings.Join(ids, ","))
	}

	return &instances[0], nil
}
