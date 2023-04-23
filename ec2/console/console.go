package console

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

var waitForOutput bool

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "console <i-instanceid>",
		Short: "Get console output from instance",
		Run:   consoleAction,
	}

	cmd.Flags().BoolVarP(&waitForOutput, "wait", "", false, "Wait for output")

	return &cmd
}

func consoleAction(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatalf("usage: console <i-instanceid>")
	}

	svc := ec2.New(config.Session())

	maxCount := 1
	maxTimeout := time.Now().Add(10 * time.Minute)
	var gotOutput bool

	if waitForOutput {
		maxCount = 20
	}

	for i := 0; i < maxCount; i++ {
		out, err := svc.GetConsoleOutput(&ec2.GetConsoleOutputInput{
			InstanceId: aws.String(args[0]),
		})
		if err != nil {
			log.Fatal(err)
		}

		if out.Timestamp != nil {
			fmt.Fprintf(os.Stderr, "# %s\n", out.Timestamp)
		}

		if out.Output == nil {
			fmt.Fprintf(os.Stderr, "no output\n")
		} else {
			b, err := base64.StdEncoding.DecodeString(*out.Output)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(b))
			gotOutput = true
			break
		}

		if time.Now().After(maxTimeout) {
			break
		}
	}

	if !gotOutput && maxCount > 1 {
		log.Printf("Giving up")
	}
}
