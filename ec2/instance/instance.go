package instance

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/psanford/aws-buddy/config"
)

func InstancesFromDesc(desc *ec2.DescribeInstancesOutput) []ec2.Instance {
	out := make([]ec2.Instance, 0)
	for _, res := range desc.Reservations {
		for _, instPtr := range res.Instances {
			inst := *instPtr
			out = append(out, inst)
		}
	}

	return out
}

func Get(instanceID string) (*ec2.Instance, error) {
	svc := ec2.New(config.Session())
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

	instances := InstancesFromDesc(output)
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
