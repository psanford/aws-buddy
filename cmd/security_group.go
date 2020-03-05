package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

func securityGroupCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:     "security_group",
		Aliases: []string{"sg"},
		Short:   "Security Group Commands",
	}

	cmd.AddCommand(sgListCommand())
	cmd.AddCommand(sgShowCommand())

	return &cmd
}

func sgListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "list security groups",
		Run:   sgListAction,
	}
	return &cmd
}

func sgShowCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "show <sg-id>",
		Short: "Show a security group",
		Run:   sgShowAction,
	}
	return &cmd
}

func sgListAction(cmd *cobra.Command, args []string) {
	svc := ec2.New(session())
	err := svc.DescribeSecurityGroupsPages(nil, func(output *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		for _, sg := range output.SecurityGroups {
			id := str(sg.GroupId)
			name := str(sg.GroupName)
			desc := str(sg.Description)

			tags := make([]string, 0, len(sg.Tags))
			for _, t := range sg.Tags {
				tags = append(tags, fmt.Sprintf("%s:%q", *t.Key, *t.Value))
			}

			fmt.Printf("%20s %-40s %q %s\n", id, name, desc, strings.Join(tags, ","))
		}
		return true
	})
	if err != nil {
		log.Fatalf("DescribeSecurityGroups error: %s", err)
	}
}

func sgShowAction(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatalf("Missing required sg-id argument")
	}

	svc := ec2.New(session())

	groups := findSGs(svc, "group-id", args[0])
	if len(groups) == 0 {
		groups = findSGs(svc, "group-name", args[0])
	}

	if len(groups) == 0 {
		log.Fatal("No matching security group found")
	}

	var matchGroup ec2.SecurityGroup
	if len(groups) > 1 {
		var matchByID bool

		names := make([]string, 0)

		for _, sg := range groups {
			if *sg.GroupId == args[0] {
				matchByID = true
				matchGroup = *sg
				break
			}

			names = append(names, fmt.Sprintf("%s(%s)", *sg.GroupId, *sg.GroupName))
		}

		if !matchByID {
			log.Fatalf("Multiple matching groups found: %s\n", strings.Join(names, ","))
		}
	} else {
		matchGroup = *groups[0]
	}

	sg := matchGroup
	id := str(sg.GroupId)
	name := str(sg.GroupName)
	desc := str(sg.Description)

	tags := make([]string, 0, len(sg.Tags))
	for _, t := range sg.Tags {
		tags = append(tags, fmt.Sprintf("%s:%q", *t.Key, *t.Value))
	}

	fmt.Printf("%20s %-40s %q %s\n", id, name, desc, strings.Join(tags, ","))

	fmt.Printf("ingress rules:\n")
	for _, perm := range sg.IpPermissions {
		fmt.Printf("%s\n", perm)
	}

	fmt.Printf("egress rules:\n")
	for _, perm := range sg.IpPermissionsEgress {
		fmt.Printf("%s\n", perm)
	}
}

func findSGs(svc *ec2.EC2, attr, val string) []*ec2.SecurityGroup {
	input := ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String(attr),
				Values: []*string{aws.String(val)},
			},
		},
	}
	output, err := svc.DescribeSecurityGroups(&input)
	if err != nil {
		log.Fatalf("DescribeSecurityGroups error: %s", err)
	}

	return output.SecurityGroups
}

func str(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
