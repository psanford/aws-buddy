package terminate

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
	"github.com/psanford/aws-buddy/config"
	"github.com/psanford/aws-buddy/console"
	"github.com/psanford/aws-buddy/ec2/instance"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "terminate <i-instanceid>",
		Short: "Terminate instance",
		Run:   terminateAction,
	}

	return &cmd
}

func terminateAction(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatalf("usage: terminate <i-instanceid>")
	}

	instanceID := args[0]
	inst, err := instance.Get(instanceID)
	if err != nil {
		log.Fatalf("fetch instance err: %s", err)
	}

	tags := make(map[string]string)
	for _, t := range inst.Tags {
		tags[*t.Key] = *t.Value
	}
	name := tags["Name"]

	fmt.Printf("name     : %s\n", name)
	fmt.Printf("id       : %s\n", *inst.InstanceId)
	fmt.Printf("type     : %s\n", *inst.InstanceType)
	fmt.Printf("az       : %s\n", *inst.Placement.AvailabilityZone)
	fmt.Printf("state    : %s\n", *inst.State.Name)

	ok := console.Confirm(fmt.Sprintf("Are you sure you want to terminate %s %s? [yN]?", color.New(color.FgRed).Sprint(instanceID), name))
	if !ok {
		log.Fatalln("Aborting")
	}

	// give a few seconds to change your mind
	time.Sleep(3 * time.Second)

	svc := ec2.New(config.Session())
	_, err = svc.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice([]string{instanceID}),
	})

	if err != nil {
		log.Fatalf("Terminate instance err: %s", err)
	}
}
