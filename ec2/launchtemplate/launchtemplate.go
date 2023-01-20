package launchtemplate

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "launch_template <name>",
		Short: "Launch template command",
		Run:   launchTemplateAction,
	}

	return &cmd
}

func launchTemplateAction(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatalf("usage: launch_template <name>")
	}

	name := args[0]
	fname := fmt.Sprintf("%s.yml", name)

	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Fatalf("create file %s err: %s", fname, err)
	}

	defer f.Close()

	svc := ec2.New(config.Session())
	var securityGroups []string
	var defaultSG string
	err = svc.DescribeSecurityGroupsPages(&ec2.DescribeSecurityGroupsInput{}, func(dsgo *ec2.DescribeSecurityGroupsOutput, b bool) bool {
		for _, sg := range dsgo.SecurityGroups {
			var name string
			if sg.GroupName != nil {
				name = *sg.GroupName
			}

			securityGroups = append(securityGroups, fmt.Sprintf("%s (%s)", *sg.GroupId, name))
			if name == "allow-ssh" {
				defaultSG = *sg.GroupId
			}
		}
		return true
	})
	if err != nil {
		log.Fatalf("DescribeSecurityGroups err: %s", err)
	}

	var subnets []string
	err = svc.DescribeSubnetsPages(&ec2.DescribeSubnetsInput{}, func(dso *ec2.DescribeSubnetsOutput, b bool) bool {
		for _, s := range dso.Subnets {
			var name string

			for _, t := range s.Tags {
				if *t.Key == "Name" {
					name = *t.Value
				}
			}

			subnets = append(subnets, fmt.Sprintf("%s %s %s %s", *s.SubnetId, *s.AvailabilityZone, *s.CidrBlock, name))
		}
		return true
	})
	if err != nil {
		log.Fatalf("DescribeSubnets err: %s", err)
	}

	var keyPairs []string
	kps, err := svc.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
	if err != nil {
		log.Fatalf("DescribeKeyPairs err: %s", err)
	}
	for _, kp := range kps.KeyPairs {
		keyPairs = append(keyPairs, *kp.KeyName)
	}

	cfg := tmplConfig{
		Name:                 name,
		SecurityGroups:       securityGroups,
		Subnets:              subnets,
		KeyPairs:             keyPairs,
		DefaultSubnet:        strings.Fields(subnets[rand.Intn(len(subnets))])[0],
		DefaultSecurityGroup: defaultSG,
	}

	if len(keyPairs) > 0 {
		cfg.DefaultKeyPair = keyPairs[0]

	}

	err = tmpl.Execute(f, cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stderr, "wrote %s\n", fname)
}

type tmplConfig struct {
	Name                 string
	SecurityGroups       []string
	DefaultSecurityGroup string
	Subnets              []string
	DefaultSubnet        string
	KeyPairs             []string
	DefaultKeyPair       string
}

var tmpl = template.Must(template.New("tmpl").Parse(tmplText))

var tmplText = `name: {{.Name}}

instance_type: t4g.small
{{range .SecurityGroups}}
# {{.}}
{{- end}}
security_group: {{.DefaultSecurityGroup}}
{{range .Subnets}}
# {{.}}
{{- end}}
subnet: {{.DefaultSubnet}}
{{range .KeyPairs}}
# {{.}}
{{- end}}
key_pair: {{.DefaultKeyPair}}

ubuntu_release: 22.04

# user_data script content
# see CloudInit docs for details
# user_data: |
#   #!/bin/bash
#   set -x
#   mkdir -p /home/ubuntu/.ssh/
#   curl https://api.sanford.io/sshkeys/ > /home/ubuntu/.ssh/authorized_keys
#   chmod 400 /home/ubuntu/.ssh/authorized_keys
#   chown ubuntu:ubuntu /home/ubuntu/.ssh/authorized_keys
