package cmd

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/spf13/cobra"
)

func asgCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "asg",
		Short: "ASG Commands",
	}

	cmd.AddCommand(asgScalingActivitiesCommand())

	return &cmd
}

func asgScalingActivitiesCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "scaling-activites <asg-name>",
		Short: "list scaling activities",
		Run:   asgListScalingActivitiesAction,
	}

	return &cmd
}

func asgListScalingActivitiesAction(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatalf("Missing required asg-name argument")
	}
	svc := autoscaling.New(session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	input := autoscaling.DescribeScalingActivitiesInput{
		AutoScalingGroupName: &args[0],
	}
	err := svc.DescribeScalingActivitiesPages(&input, func(out *autoscaling.DescribeScalingActivitiesOutput, more bool) bool {
		for _, act := range out.Activities {
			jsonOut.Encode(act)
		}
		return true
	})

	if err != nil {
		log.Fatalf("DescribeScalingActivities error: %s", err)
	}
}
