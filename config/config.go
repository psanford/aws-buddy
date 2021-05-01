package config

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	awssession "github.com/aws/aws-sdk-go/aws/session"
)

var DefaultRegion = "us-east-1"

func Session() *awssession.Session {
	region := DefaultRegion
	if os.Getenv("AWS_DEFAULT_REGION") != "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}

	sess, err := awssession.NewSession(&aws.Config{Region: &region})
	if err != nil {
		panic(err)
	}

	return sess
}
