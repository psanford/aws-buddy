package ami

import (
	"fmt"
	"log"
	"sort"

	"github.com/psanford/aws-buddy/config"
	"github.com/psanford/ubuntuami"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "ami",
		Short: "AMI subcommands",
	}

	cmd.AddCommand(listUbuntuCommand())

	return &cmd
}

func listUbuntuCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list_ubuntu",
		Short: "list ubuntu AMIs",
		Run:   listUbuntuAction,
	}

	return &cmd
}

func listUbuntuAction(cmd *cobra.Command, args []string) {
	amis, err := ubuntuami.Fetch()
	if err != nil {
		log.Fatalf("fetch ubuntu ami err: %s", err)
	}

	sort.Slice(amis, func(i, j int) bool {
		if amis[i].ReleaseVersion == amis[j].ReleaseVersion {
			return amis[i].ReleaseTime.Before(amis[j].ReleaseTime)
		}

		return amis[i].ReleaseVersion < amis[j].ReleaseVersion
	})

	for _, ami := range amis {
		if ami.Region != config.DefaultRegion {
			continue
		}

		fmt.Printf("%20s %10.10s %10.10s %10.10s %7.7s %20.20s %s\n", ami.ID, ami.Region, ami.ReleaseName, ami.ReleaseVersion, ami.Arch, ami.InstanceType, ami.ReleaseTime)
	}
}
