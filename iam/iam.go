package iam

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

var (
	jsonOutput     bool
	csvOutput      bool
	iamUserFullArn bool
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "iam",
		Short: "IAM Commands",
	}

	cmd.AddCommand(iamUserCommand())

	return &cmd
}

func iamUserCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "user",
		Short: "User Commands",
	}

	cmd.AddCommand(iamUserListCommand())

	return &cmd
}

func iamUserListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List users",
		Run:   iamListUsers,
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json ouput")
	cmd.Flags().BoolVarP(&csvOutput, "csv", "", false, "Show csv ouput")
	cmd.Flags().BoolVarP(&iamUserFullArn, "full-arn", "", false, "Show full arn for username")

	return &cmd
}

func iamListUsers(cmd *cobra.Command, args []string) {
	iamSvc := iam.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	csvOut := csv.NewWriter(os.Stdout)

	if !jsonOutput && !csvOutput {
		fmt.Printf("%-20.20s %10s %10s %10s %10s %10s %10s\n", "user", "has_pass", "pw_creation", "pw_last_used", "keys", "active_keys", "mfa")
	} else if csvOutput {
		csvOut.Write([]string{"user", "has_pass", "pw_creation", "pw_last_used", "keys", "active_keys", "mfa"})
	}

	format := "%-20.20s %10t %10s %10s %10d %10d %10d\n"
	if iamUserFullArn {
		format = "%-50s %10t %10s %10s %10d %10d %10d\n"
	}

	err := iamSvc.ListUsersPages(&iam.ListUsersInput{}, func(out *iam.ListUsersOutput, b bool) bool {
		for _, user := range out.Users {
			passwordLastUsed := "never"
			passwordCreation := "no-pass"
			hasPassword := false

			if user.PasswordLastUsed != nil {
				passwordLastUsed = user.PasswordLastUsed.Format("2006-01-02")
			}

			lp, err := iamSvc.GetLoginProfile(&iam.GetLoginProfileInput{
				UserName: user.UserName,
			})

			if awserr, ok := err.(awserr.Error); ok && awserr.Code() == iam.ErrCodeNoSuchEntityException {
				// user does not have a password set
			} else if err != nil {
				log.Fatalf("GetLoginProfile err for %s: %s", *user.UserName, err)
			} else {
				if lp.LoginProfile.CreateDate != nil {
					passwordCreation = lp.LoginProfile.CreateDate.Format("2006-01-02")
					hasPassword = true
				}
			}

			mfaResp, err := iamSvc.ListMFADevices(&iam.ListMFADevicesInput{
				UserName: user.UserName,
			})
			if err != nil {
				log.Printf("ListMFADevices err for %s: %s", *user.UserName, err)
				continue
			}

			keysResp, err := iamSvc.ListAccessKeys(&iam.ListAccessKeysInput{
				UserName: user.UserName,
			})
			if err != nil {
				log.Printf("ListAccessKeys err for %s: %s", *user.UserName, err)
				continue
			}

			if jsonOutput {
				jsonOut.Encode(struct {
					User         *iam.User
					LoginProfile *iam.LoginProfile
					ApiKeys      []*iam.AccessKeyMetadata
					MFADevices   []*iam.MFADevice
				}{
					User:         user,
					LoginProfile: lp.LoginProfile,
					ApiKeys:      keysResp.AccessKeyMetadata,
					MFADevices:   mfaResp.MFADevices,
				})
			} else {
				liveKeys := 0
				for _, k := range keysResp.AccessKeyMetadata {
					if k.Status != nil && *k.Status == "Active" {
						liveKeys++
					}
				}

				username := *user.UserName
				if iamUserFullArn {
					username = *user.Arn
				}
				if csvOutput {
					csvOut.Write([]string{username, fmt.Sprintf("%t", hasPassword), passwordCreation, passwordLastUsed, fmt.Sprintf("%d", len(keysResp.AccessKeyMetadata)), fmt.Sprintf("%d", liveKeys), fmt.Sprintf("%d", len(mfaResp.MFADevices))})
					csvOut.Flush()
				} else {
					fmt.Printf(format, username, hasPassword, passwordCreation, passwordLastUsed, len(keysResp.AccessKeyMetadata), liveKeys, len(mfaResp.MFADevices))
				}
			}
		}

		return true
	})

	if err != nil {
		log.Fatalf("ListUsers err: %s", err)
	}
}
