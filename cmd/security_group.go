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

func securityGroupCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:     "security_group",
		Aliases: []string{"sg"},
		Short:   "Security Group Commands",
	}

	cmd.AddCommand(sgListCommand())

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

func sgListAction(cmd *cobra.Command, args []string) {
	sess, err := session.NewSession(&aws.Config{Region: &region})
	if err != nil {
		log.Fatalf("AWS NewSession error: %s", err)
	}
	svc := ec2.New(sess)
	err = svc.DescribeSecurityGroupsPages(nil, func(output *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
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

func str(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
