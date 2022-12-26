package launch

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/psanford/aws-buddy/config"
	"github.com/psanford/ubuntuami"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "launch <launch_tmpl.yml>",
		Short: "Launch instance",
		Run:   launchAction,
	}

	return &cmd
}

func launchAction(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatalf("usage: launch <launch_tmpl.yml>")
	}

	f, err := os.Open(args[0])
	if err != nil {
		log.Fatalf("open %s err: %s", args[0], err)
	}

	defer f.Close()

	dec := yaml.NewDecoder(f)
	var cfg launchCfg
	err = dec.Decode(&cfg)
	if err != nil {
		log.Fatalf("decode err: %s", err)
	}

	if cfg.Name == "" {
		log.Fatal("name: is required")
	}

	if cfg.SecurityGroup == "" {
		log.Fatal("security_group: is required")
	}

	arch := instanceTypeArch(cfg.InstanceType)

	sgID := strings.Fields(cfg.SecurityGroup)[0]

	amis, err := ubuntuami.Fetch()
	if err != nil {
		log.Fatalf("ubuntu ami fetch err: %s", err)
	}
	var (
		matchAMI ubuntuami.AMI
	)
	for _, ami := range amis {
		if ami.Region != config.DefaultRegion {
			continue
		}
		version := strings.TrimSuffix(ami.ReleaseVersion, " LTS")
		if version != cfg.UbuntuRelease {
			continue
		}

		if ami.Arch != arch {
			continue
		}

		if ami.ReleaseTime.After(matchAMI.ReleaseTime) {
			matchAMI = ami
		}
	}

	if matchAMI.ID == "" {
		log.Fatalf("No matching AMI found")
	}

	svc := ec2.New(config.Session())

	runCfg := &ec2.RunInstancesInput{
		InstanceType:                      &cfg.InstanceType,
		ImageId:                           &matchAMI.ID,
		SecurityGroupIds:                  aws.StringSlice([]string{sgID}),
		SubnetId:                          &cfg.Subnet,
		MinCount:                          aws.Int64(1),
		MaxCount:                          aws.Int64(1),
		InstanceInitiatedShutdownBehavior: aws.String(ec2.ShutdownBehaviorTerminate),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: &cfg.Name,
					},
				},
			},
		},
	}

	if cfg.KeyPair != "" {
		runCfg.KeyName = &cfg.KeyPair
	}

	if cfg.UserData != "" {
		var buf bytes.Buffer
		enc := base64.NewEncoder(base64.StdEncoding, &buf)
		enc.Write([]byte(cfg.UserData))
		enc.Close()
		ud := buf.String()
		runCfg.UserData = &ud
	}

	r, err := svc.RunInstances(runCfg)
	if err != nil {
		log.Fatalf("RunInstances err: %s", err)
	}

	fmt.Printf("instance: %s\n", *r.Instances[0].InstanceId)
}

type launchCfg struct {
	Name          string `yaml:"name"`
	InstanceType  string `yaml:"instance_type"`
	SecurityGroup string `yaml:"security_group"`
	Subnet        string `yaml:"subnet"`
	UbuntuRelease string `yaml:"ubuntu_release"`
	KeyPair       string `yaml:"key_pair"`
	UserData      string `yaml:"user_data"`
}

var gravitonRe = regexp.MustCompile(`\dg`)

func instanceTypeArch(t string) string {
	if gravitonRe.MatchString(t) {
		return "arm64"
	}
	return "amd64"
}
