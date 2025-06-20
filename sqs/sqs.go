package sqs

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "sqs",
		Short: "SQS Commands",
	}

	cmd.AddCommand(listCommand())
	cmd.AddCommand(peekCommand())
	cmd.AddCommand(consumeCommand())

	return &cmd
}

func listCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List SQS queues",
		Run:   listAction,
	}
	return &cmd
}

func listAction(cmd *cobra.Command, args []string) {
	svc := sqs.New(config.Session())

	result, err := svc.ListQueues(&sqs.ListQueuesInput{})
	if err != nil {
		log.Fatal(err)
	}

	for _, url := range result.QueueUrls {
		attrs, err := svc.GetQueueAttributes(&sqs.GetQueueAttributesInput{
			AttributeNames: aws.StringSlice([]string{"All"}),
			QueueUrl:       url,
		})
		if err != nil {
			log.Fatal(err)
		}
		arn := attrs.Attributes["QueueArn"]
		arnParts := strings.Split(*arn, ":")
		name := arnParts[len(arnParts)-1]
		messages := attrs.Attributes["ApproximateNumberOfMessages"]
		messagesPending := attrs.Attributes["ApproximateNumberOfMessagesNotVisible"]
		messagesDelayed := attrs.Attributes["ApproximateNumberOfMessagesDelayed"]

		fmt.Printf("========[ %s ]===================\n", name)
		fmt.Printf("arn              : %s\n", *arn)
		fmt.Printf("url              : %s\n", *url)
		fmt.Printf("messages         : %s\n", *messages)
		fmt.Printf("pending messages : %s\n", *messagesPending)
		fmt.Printf("delayed messages : %s\n", *messagesDelayed)
	}
}

func peekCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "peek <queue-url>",
		Short: "Peek at messages in an SQS queue",
		Run:   peekAction,
	}
	return &cmd
}

func peekAction(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatalf("Usage: peek <queue-url>")
	}

	queueURL := args[0]

	svc := sqs.New(config.Session())

	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: aws.Int64(1),
		VisibilityTimeout:   aws.Int64(0),
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(result.Messages) == 0 {
		log.Println("No messages in the queue")
		return
	}

	for _, message := range result.Messages {
		out, err := json.MarshalIndent(message, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s\n", out)
	}
}

var (
	countFlag int
)

func consumeCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "consume <queue-url>",
		Short: "Consume (receive and delete) messages from an SQS queue",
		Run:   consumeAction,
	}
	cmd.Flags().IntVarP(&countFlag, "count", "n", 1, "Number of messages to consume")
	return &cmd
}

func consumeAction(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatalf("Usage: consume <queue-url>")
	}

	queueURL := args[0]

	svc := sqs.New(config.Session())

	consumed := 0
	for consumed < countFlag {
		remaining := countFlag - consumed
		maxMessages := int64(10)
		if remaining < 10 {
			maxMessages = int64(remaining)
		}

		result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:              aws.String(queueURL),
			MaxNumberOfMessages:   aws.Int64(maxMessages),
			WaitTimeSeconds:       aws.Int64(20),
			MessageAttributeNames: aws.StringSlice([]string{"All"}),
			AttributeNames:        aws.StringSlice([]string{"All"}),
		})
		if err != nil {
			log.Fatalf("Error receiving messages: %v", err)
		}

		if len(result.Messages) == 0 {
			fmt.Printf("No more messages available. Consumed %d message(s).\n", consumed)
			break
		}

		for _, message := range result.Messages {
			if consumed >= countFlag {
				break
			}

			fmt.Printf("%s\n", *message.Body)

			_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			if err != nil {
				log.Printf("Error deleting message %s: %v", *message.MessageId, err)
			} else {
				consumed++
			}
		}
	}

	if consumed == countFlag {
		fmt.Printf("Successfully consumed %d message(s).\n", consumed)
	}
}
