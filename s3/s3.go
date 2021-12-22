package s3

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mitchellh/mapstructure"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "s3",
		Short: "S3 Commands",
	}

	cmd.AddCommand(catCommand())
	cmd.AddCommand(headCommand())

	return &cmd
}

func catCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "cat <[s3://]bucket/path/to/object>",
		Short: "Cat object",
		Run:   catAction,
	}
	return &cmd
}

func bucketPath(raw string) (string, string) {
	bucketPath := strings.TrimPrefix(raw, "s3://")

	// add a prefix / to make clean remove /.. segments
	bucketPath = path.Clean("/" + bucketPath)

	parts := strings.Split(bucketPath, "/")
	// remove extra segment
	parts = parts[1:]

	bucket := parts[0]
	path := strings.Join(parts[1:], "/")
	return bucket, path
}

func catAction(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatalf("Usage: cat <[s3://]bucket/path/to/obj")
	}

	bucket, path := bucketPath(args[0])

	svc := s3.New(config.Session())

	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &path,
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(os.Stdout, obj.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func headCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "head <[s3://]bucket/path/to/object>",
		Short: "Head object",
		Run:   headAction,
	}
	return &cmd
}

func headAction(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatalf("Usage: head <[s3://]bucket/path/to/obj")
	}

	bucket, path := bucketPath(args[0])

	svc := s3.New(config.Session())

	obj, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &path,
	})
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[string]interface{})
	err = mapstructure.Decode(obj, &m)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range m {
		if v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil()) {
			delete(m, k)
		}
	}

	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", out)
}
