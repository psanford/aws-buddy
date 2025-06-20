package iam

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

func identityCenterCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "identity-center",
		Short: "Commands to help review access by iam principals",
	}

	cmd.AddCommand(identityCenterListPermissionSets())

	return &cmd
}

func identityCenterListPermissionSets() *cobra.Command {
	return &cobra.Command{
		Use:   "list-permission-sets",
		Short: "List Identity Center Permission Sets",
		Run:   listPermissionSets,
	}
}

func listPermissionSets(cmd *cobra.Command, args []string) {
	ssoAdminSvc := ssoadmin.New(config.Session())

	instances, err := ssoAdminSvc.ListInstances(&ssoadmin.ListInstancesInput{})
	if err != nil {
		log.Fatalf("Failed to list SSO instances: %v", err)
	}

	if len(instances.Instances) == 0 {
		log.Fatalf("No SSO instances found")
	}

	instanceArn := *instances.Instances[0].InstanceArn

	err = ssoAdminSvc.ListPermissionSetsPages(&ssoadmin.ListPermissionSetsInput{
		InstanceArn: &instanceArn,
	}, func(output *ssoadmin.ListPermissionSetsOutput, lastPage bool) bool {
		for _, permissionSet := range output.PermissionSets {
			describeOutput, err := ssoAdminSvc.DescribePermissionSet(&ssoadmin.DescribePermissionSetInput{
				InstanceArn:      &instanceArn,
				PermissionSetArn: permissionSet,
			})
			if err != nil {
				log.Printf("Failed to describe permission set %s: %v", *permissionSet, err)
				continue
			}

			fmt.Printf("%s %s\n", *permissionSet, *describeOutput.PermissionSet.Name)
		}
		return true
	})

	if err != nil {
		log.Fatalf("Failed to list permission sets: %v", err)
	}
}
