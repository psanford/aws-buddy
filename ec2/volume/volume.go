package volume

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	csvOutput  bool
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "volume",
		Short: "Volume Commands",
	}

	cmd.AddCommand(volumeListCommand())

	return &cmd
}

func volumeListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list volumes",
		Run:     volumeListAction,
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json ouput")
	cmd.Flags().BoolVarP(&csvOutput, "csv", "", false, "Output as csv")

	return &cmd
}

func volumeListAction(cmd *cobra.Command, args []string) {
	ec2Svc := ec2.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	csvOut := csv.NewWriter(os.Stdout)

	err := ec2Svc.DescribeVolumesPages(&ec2.DescribeVolumesInput{}, func(dvo *ec2.DescribeVolumesOutput, b bool) bool {
		for _, vol := range dvo.Volumes {
			if jsonOutput {
				jsonOut.Encode(vol)
				continue
			}

			var instances []string
			for _, attach := range vol.Attachments {
				instances = append(instances, *attach.InstanceId)
			}
			enc := "unencrypted"
			if vol.Encrypted != nil && *vol.Encrypted {
				enc = "encrypted"
			}
			if csvOutput {
				csvOut.Write([]string{*vol.VolumeId, enc, strings.Join(instances, ",")})
			} else {
				fmt.Printf("%20.20s %3s %s\n", *vol.VolumeId, enc, strings.Join(instances, ","))
			}
		}
		return true
	})
	if err != nil {
		log.Fatal(err)
	}

	if csvOutput {
		csvOut.Flush()
	}
}
