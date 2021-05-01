package tag

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/psanford/aws-buddy/config"
	"github.com/psanford/aws-buddy/console"
	"github.com/psanford/aws-buddy/ec2/instance"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "tag",
		Short: "Tag Commands",
	}

	cmd.AddCommand(tagListCommand())
	cmd.AddCommand(tagSetCommand())
	cmd.AddCommand(tagRemoveCommand())

	return &cmd
}

func tagListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list <instance-id>",
		Aliases: []string{"ls"},
		Short:   "list tags on instance",
		Run:     tagListAction,
	}

	return &cmd
}

func tagListAction(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatal("Missing required <instance-id>")
	}

	instanceID := args[0]

	inst, err := instance.Get(instanceID)
	if err != nil {
		log.Fatal(err)
	}

	var name string
	tbl := make([][]string, 0, len(inst.Tags))
	for _, t := range inst.Tags {
		tbl = append(tbl, []string{*t.Key, *t.Value})
		if *t.Key == "Name" {
			name = *t.Value
		}
	}

	fmt.Printf("%s (%s) tags:\n", name, instanceID)
	fmt.Print(console.FormatTable(tbl))
}

func tagSetCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "set <instance-id> <tag-name> <tag-value>",
		Short: "set tag on instance",
		Run:   setTagAction,
	}

	return &cmd
}

func setTagAction(cmd *cobra.Command, args []string) {
	if len(args) < 3 {
		log.Fatal("Missing required <instance-id> <tag-name> <tag-value>")
	}

	instanceID := args[0]
	inst, err := instance.Get(instanceID)
	if err != nil {
		log.Fatal(err)
	}

	var (
		instName string
		oldVal   = "<unset>"
		tagName  = args[1]
		newVal   = args[2]
	)
	for _, t := range inst.Tags {
		if *t.Key == tagName {
			oldVal = *t.Value
		}
		if *t.Key == "Name" {
			instName = *t.Value
		}
	}

	fmt.Printf("%s (%s)\n\n", instanceID, instName)
	fmt.Printf("tag %s: %s => %s\n\n", tagName, oldVal, newVal)

	ok := console.Confirm("Are you sure you want to make this change [yN]? ")
	if !ok {
		log.Fatalln("Aborting")
	}

	// give you a chance to reconsider and ctrl-c
	time.Sleep(3 * time.Second)

	svc := ec2.New(config.Session())
	_, err = svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{&instanceID},
		Tags: []*ec2.Tag{
			{
				Key:   &tagName,
				Value: &newVal,
			},
		},
	})
	if err != nil {
		log.Fatalf("CreateTag err: %s\n", err)
	}
}

func tagRemoveCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "rm <instance-id> <tag-name>",
		Short: "remove tag on instance",
		Run:   removeTagAction,
	}

	return &cmd
}

func removeTagAction(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		log.Fatal("Missing required <instance-id> <tag-name>")
	}

	instanceID := args[0]
	inst, err := instance.Get(instanceID)
	if err != nil {
		log.Fatal(err)
	}

	var (
		instName string
		oldVal   *string
		tagName  = args[1]
	)
	for _, t := range inst.Tags {
		if *t.Key == tagName {
			v := *t.Value
			oldVal = &v
		}
		if *t.Key == "Name" {
			instName = *t.Value
		}
	}

	fmt.Printf("%s (%s)\n\n", instanceID, instName)
	if oldVal == nil {
		log.Fatalf("Tag not set on instance")
	}
	fmt.Printf("tag %s: %s => (deleted)\n\n", tagName, *oldVal)

	ok := console.Confirm("Are you sure you want to make this change [yN]? ")
	if !ok {
		log.Fatalln("Aborting")
	}

	// give you a chance to reconsider and ctrl-c
	time.Sleep(3 * time.Second)

	svc := ec2.New(config.Session())
	svc.DeleteTags(&ec2.DeleteTagsInput{
		Resources: []*string{&instanceID},
		Tags: []*ec2.Tag{
			{
				Key:   &tagName,
				Value: oldVal,
			},
		},
	})
}
