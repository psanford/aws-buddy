package parameterstore

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/psanford/aws-buddy/config"
	"github.com/psanford/aws-buddy/console"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	paramType  string
	paramDescr string
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "param",
		Short: "SSM Parameter Store Commands",
	}

	cmd.AddCommand(paramListCommand())
	cmd.AddCommand(paramGetCommand())
	cmd.AddCommand(paramPutCommand())
	cmd.AddCommand(paramCpCommand())
	cmd.AddCommand(paramRmCommand())

	return &cmd
}

func paramListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List parameter",
		Run:   paramList,
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json ouput")

	return &cmd
}

func paramList(cmd *cobra.Command, args []string) {
	ssmClient := ssm.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	err := ssmClient.DescribeParametersPages(&ssm.DescribeParametersInput{}, func(dpo *ssm.DescribeParametersOutput, b bool) bool {
		for _, pm := range dpo.Parameters {
			if jsonOutput {
				jsonOut.Encode(pm)
			} else {
				fmt.Printf("%-12.12s %s\n", *pm.Type, *pm.Name)
			}
		}
		return true
	})

	if err != nil {
		log.Fatalf("DescribeParameters err: %s", err)
	}
}

func paramGetCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "get",
		Short: "Get parameter value",
		Run:   paramGet,
	}

	return &cmd
}

func paramGet(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatalf("Usage: get <path/to/parameter>")
	}

	ssmClient := ssm.New(config.Session())

	resp, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           &args[0],
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("GetParameter err: %s", err)
	}

	fmt.Printf("%s\n", *resp.Parameter.Value)
}

func paramPutCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "put",
		Short: "Set or create parameter value",
		Run:   paramPut,
	}

	cmd.Flags().StringVarP(&paramType, "type", "", "SecureString", "Param type (String, StringList, SecureString)")
	cmd.Flags().StringVarP(&paramDescr, "description", "", "", "Param description")

	return &cmd
}

func paramPut(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		log.Fatalf("Usage: get <path/to/parameter> <value>")
	}

	ssmClient := ssm.New(config.Session())

	name := args[0]
	value := args[1]

	resp, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: aws.Bool(true),
	})

	var (
		create bool
		oldVal string
	)
	if err != nil {
		create = true
		oldVal = "(create)"
	} else {
		oldVal = *resp.Parameter.Value
	}

	fmt.Printf("param %s: %s => %s\n\n", name, oldVal, value)

	ok := console.Confirm("Are you sure you want to make this change [yN]? ")
	if !ok {
		log.Fatalln("Aborting")
	}

	overwrite := !create
	input := ssm.PutParameterInput{
		Name:      &name,
		Value:     &value,
		Overwrite: &overwrite,
		Type:      &paramType,
	}

	if paramDescr != "" {
		input.Description = &paramDescr
	}

	_, err = ssmClient.PutParameter(&input)
	if err != nil {
		log.Fatalf("PutParameter err: %s", err)
	}
}

func paramCpCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "cp",
		Short: "Copy param from old to new path",
		Run:   paramCp,
	}

	return &cmd
}

func paramCp(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		log.Fatalf("Usage: cp <old/path> <new/path>")
	}

	ssmClient := ssm.New(config.Session())

	oldPath := args[0]
	newPath := args[1]

	resp, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           &oldPath,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("GetParameter err: %s", err)
	}

	fmt.Printf("param %s => %s (%s)\n\n", oldPath, newPath, *resp.Parameter.Value)

	ok := console.Confirm("Are you sure you want to make this change [yN]? ")
	if !ok {
		log.Fatalln("Aborting")
	}

	input := ssm.PutParameterInput{
		Name:      &newPath,
		Value:     resp.Parameter.Value,
		Overwrite: aws.Bool(true),
		Type:      resp.Parameter.Type,
	}

	_, err = ssmClient.PutParameter(&input)
	if err != nil {
		log.Fatalf("PutParameter err: %s", err)
	}
}

func paramRmCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "rm",
		Short: "Delete param at path",
		Run:   paramRm,
	}

	return &cmd
}

func paramRm(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatalf("Usage: rm <some/path/to/delete>")
	}

	ssmClient := ssm.New(config.Session())

	path := args[0]

	resp, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           &path,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("GetParameter err: %s", err)
	}

	fmt.Printf("param %s (%s) => *deleted\n\n", path, *resp.Parameter.Value)

	ok := console.Confirm("Are you sure you want to make this change [yN]? ")
	if !ok {
		log.Fatalln("Aborting")
	}

	input := ssm.DeleteParameterInput{
		Name: &path,
	}

	_, err = ssmClient.DeleteParameter(&input)
	if err != nil {
		log.Fatalf("DeleteParameter err: %s", err)
	}
}
