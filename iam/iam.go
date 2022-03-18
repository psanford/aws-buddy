package iam

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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
	cmd.AddCommand(iamAccessCommand())

	return &cmd
}

func iamAccessCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "access",
		Short: "Commands to help review access by iam principals",
	}

	cmd.AddCommand(iamGetAccountAuthorizationDetailsCommand())
	cmd.AddCommand(testAllIamIdentitiesCommand())

	return &cmd
}

func iamUserCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "user",
		Short: "User Commands",
	}

	cmd.AddCommand(iamUserListCommand())
	cmd.AddCommand(iamUserShowCommand())
	cmd.AddCommand(listAccessKeysCommand())

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

func iamUserShowCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "show <username>",
		Short: "Show user",
		Run:   iamShowUser,
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "", false, "Show raw json ouput")

	return &cmd
}

func iamShowUser(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatalf("Missing required <username> argument")
	}
	username := args[0]

	iamSvc := iam.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	userOutput, err := iamSvc.GetUser(&iam.GetUserInput{
		UserName: aws.String(username),
	})

	if err != nil {
		log.Fatalf("GetUser error: %s", err)
	}

	u := userOutput.User

	passwordLastUsed := "never"
	if u.PasswordLastUsed != nil {
		passwordLastUsed = u.PasswordLastUsed.Format("2006-01-02")
	}

	fmt.Printf("========[ %s ]===================\n", *u.UserId)
	fmt.Printf("name         : %s\n", *u.UserName)
	fmt.Printf("arn          : %s\n", *u.Arn)
	fmt.Printf("creation     : %s\n", u.CreateDate.Format(time.RFC3339))
	fmt.Printf("pw last      : %s\n", passwordLastUsed)

	listPolicyInput := iam.ListUserPoliciesInput{
		UserName: aws.String(username),
	}
	err = iamSvc.ListUserPoliciesPages(&listPolicyInput, func(out *iam.ListUserPoliciesOutput, more bool) bool {
		for _, pname := range out.PolicyNames {

			gotPolicy, err := iamSvc.GetUserPolicy(&iam.GetUserPolicyInput{
				UserName:   aws.String(username),
				PolicyName: pname,
			})
			if err != nil {
				log.Printf("Fetch iam user inline policy err: %s", err)
				continue
			}

			fmt.Printf("========[ policy %s ]===================\n", *pname)
			doc, err := url.QueryUnescape(*gotPolicy.PolicyDocument)
			if err != nil {
				log.Printf("Decode policy doc err: %s", err)
				continue
			}

			fmt.Printf("%s\n", doc)
		}

		return true

	})
	if err != nil {
		log.Fatalf("List user policies err: %s", err)
	}

	listAttachedInput := iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(username),
	}

	fmt.Printf("========[ attached policies ]===================\n")
	err = iamSvc.ListAttachedUserPoliciesPages(&listAttachedInput, func(laupo *iam.ListAttachedUserPoliciesOutput, b bool) bool {
		for _, p := range laupo.AttachedPolicies {
			fmt.Printf("%s : %s\n", *p.PolicyArn, *p.PolicyName)
			policyInfo, err := iamSvc.GetPolicy(&iam.GetPolicyInput{
				PolicyArn: p.PolicyArn,
			})
			if err != nil {
				log.Printf("Get policy info err: %s", err)
				continue
			}
			policyDoc, err := iamSvc.GetPolicyVersion(&iam.GetPolicyVersionInput{
				PolicyArn: p.PolicyArn,
				VersionId: policyInfo.Policy.DefaultVersionId,
			})
			if err != nil {
				log.Printf("Get policy doc err: %s", err)
				continue
			}
			doc, err := url.QueryUnescape(*policyDoc.PolicyVersion.Document)
			if err != nil {
				log.Printf("Decode policy doc err: %s", err)
				continue
			}

			fmt.Printf("========[ policy %s ]===================\n", *p.PolicyArn)
			fmt.Printf("%s\n", doc)
		}
		return true
	})
	if err != nil {
		log.Printf("List attched user policies err: %s", err)
	}

	listGroupsInput := iam.ListGroupsForUserInput{
		UserName: aws.String(username),
	}
	err = iamSvc.ListGroupsForUserPages(&listGroupsInput, func(groups *iam.ListGroupsForUserOutput, more bool) bool {
		for _, g := range groups.Groups {
			fmt.Printf("========[ group %s ]===================\n", *g.GroupId)
			fmt.Printf("name         : %s\n", *g.GroupName)
			fmt.Printf("arn          : %s\n", *g.Arn)

			listGroupPoliciesInput := iam.ListGroupPoliciesInput{
				GroupName: g.GroupName,
			}
			err = iamSvc.ListGroupPoliciesPages(&listGroupPoliciesInput, func(p *iam.ListGroupPoliciesOutput, more bool) bool {
				for _, pname := range p.PolicyNames {
					gotPolicy, err := iamSvc.GetGroupPolicy(&iam.GetGroupPolicyInput{
						GroupName:  g.GroupName,
						PolicyName: pname,
					})

					if err != nil {
						log.Printf("Fetch iam group policy err: %s", err)
						continue
					}

					fmt.Printf("========[ group-policy %s ]===================\n", *pname)
					fmt.Printf("%s\n", *gotPolicy.PolicyDocument)
				}

				return true
			})
			if err != nil {
				log.Printf("List group policies err: %s %s", *g.GroupName, err)
			}

			listAttachedGroup := iam.ListAttachedGroupPoliciesInput{
				GroupName: g.GroupName,
			}
			fmt.Printf("========[ attached group policies ]===================\n")
			iamSvc.ListAttachedGroupPoliciesPages(&listAttachedGroup, func(lagpo *iam.ListAttachedGroupPoliciesOutput, b bool) bool {
				for _, p := range lagpo.AttachedPolicies {
					fmt.Printf("%s : %s\n", *p.PolicyArn, *p.PolicyName)
				}

				return true
			})
			if err != nil {
				log.Printf("List attched group policies err: %s", err)
			}

		}

		return true
	})
	if err != nil {
		log.Fatalf("List user groups err: %s", err)
	}
}

func listAccessKeysCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list-access-keys",
		Short: "List all access keys in account",
		Run:   listAccessKeysAction,
	}

	return &cmd
}

func listAccessKeysAction(cmd *cobra.Command, args []string) {
	iamSvc := iam.New(config.Session())

	fmt.Printf("%20s %8s %30s %s\n", "key_id", "status", "creation", "user_arn")

	err := iamSvc.ListUsersPages(&iam.ListUsersInput{}, func(out *iam.ListUsersOutput, b bool) bool {
		for _, user := range out.Users {
			keysResp, err := iamSvc.ListAccessKeys(&iam.ListAccessKeysInput{
				UserName: user.UserName,
			})
			if err != nil {
				log.Printf("ListAccessKeys err for %s: %s", *user.UserName, err)
				continue
			}

			for _, k := range keysResp.AccessKeyMetadata {
				fmt.Printf("%20s %8s %30s %s\n", *k.AccessKeyId, *k.Status, k.CreateDate, *user.Arn)
			}
		}
		return true
	})
	if err != nil {
		log.Fatal(err)
	}
}

var filterPolicyMatch string

func iamGetAccountAuthorizationDetailsCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "account-authorization-datails",
		Short: "Get snapshot of account permissions",
		Run:   iamGetAccountAuthorizationDetailsAction,
	}

	cmd.Flags().StringVarP(&filterPolicyMatch, "filter-by-policy-match", "", "", "Regex string to match on policy documents")

	return &cmd
}

func iamGetAccountAuthorizationDetailsAction(cmd *cobra.Command, args []string) {
	iamSvc := iam.New(config.Session())

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")

	type IamObject struct {
		Type   string                   `json:"type"`
		Group  *iam.GroupDetail         `json:"group,omitempty"`
		Policy *iam.ManagedPolicyDetail `json:"policy,omitempty"`
		Role   *iam.RoleDetail          `json:"role,omitempty"`
		User   *iam.UserDetail          `json:"user,omitempty"`
	}

	var filterMatch *regexp.Regexp
	if filterPolicyMatch != "" {
		re, err := regexp.Compile("(?i)" + filterPolicyMatch)
		if err != nil {
			log.Fatalf("-filter-by-policy-match doesn't compile as a Go regex: %s", err)
		}
		filterMatch = re
	}

	input := &iam.GetAccountAuthorizationDetailsInput{}
	err := iamSvc.GetAccountAuthorizationDetailsPages(input, func(details *iam.GetAccountAuthorizationDetailsOutput, b bool) bool {
		for _, g := range details.GroupDetailList {
			include := filterMatch == nil
			for _, pol := range g.GroupPolicyList {
				if pol.PolicyDocument != nil {
					doc, err := url.QueryUnescape(*pol.PolicyDocument)
					if err != nil {
						log.Fatalf("Unescape group policy doc err doc=%q err=%s", *pol.PolicyDocument, err)
					}
					pol.PolicyDocument = &doc
					if include == false && filterMatch != nil {
						if filterMatch.MatchString(doc) {
							include = true
						}
					}
				}
			}
			if include {
				jsonOut.Encode(IamObject{
					Type:  "group",
					Group: g,
				})
			}
		}
		for _, m := range details.Policies {
			include := filterMatch == nil
			for _, pol := range m.PolicyVersionList {
				if pol.Document != nil {
					doc, err := url.QueryUnescape(*pol.Document)
					if err != nil {
						log.Fatalf("Unescape policy doc err doc=%q err=%s", *pol.Document, err)
					}
					pol.Document = &doc

					if include == false && filterMatch != nil {
						if filterMatch.MatchString(doc) {
							include = true
						}
					}
				}
			}
			if include {
				jsonOut.Encode(IamObject{
					Type:   "policy",
					Policy: m,
				})
			}
		}
		for _, r := range details.RoleDetailList {
			include := filterMatch == nil
			if r.AssumeRolePolicyDocument != nil {
				doc, err := url.QueryUnescape(*r.AssumeRolePolicyDocument)
				if err != nil {
					log.Fatalf("Unescape assume role policy doc err doc=%q err=%s", *r.AssumeRolePolicyDocument, err)
				}
				r.AssumeRolePolicyDocument = &doc
				if include == false && filterMatch != nil {
					if filterMatch.MatchString(doc) {
						include = true
					}
				}
			}
			for _, ip := range r.InstanceProfileList {
				for _, ipr := range ip.Roles {
					if ipr.AssumeRolePolicyDocument != nil {
						doc, err := url.QueryUnescape(*ipr.AssumeRolePolicyDocument)
						if err != nil {
							log.Fatalf("Unescape assume role policy doc err doc=%q err=%s", *ipr.AssumeRolePolicyDocument, err)
						}
						ipr.AssumeRolePolicyDocument = &doc
						if include == false && filterMatch != nil {
							if filterMatch.MatchString(doc) {
								include = true
							}
						}
					}
				}
			}
			for _, pol := range r.RolePolicyList {
				if pol.PolicyDocument != nil {
					doc, err := url.QueryUnescape(*pol.PolicyDocument)
					if err != nil {
						log.Fatalf("Unescape policy doc err doc=%q err=%s", *pol.PolicyDocument, err)
					}
					pol.PolicyDocument = &doc
					if include == false && filterMatch != nil {
						if filterMatch.MatchString(doc) {
							include = true
						}
					}
				}
			}
			if include {
				jsonOut.Encode(IamObject{
					Type: "role",
					Role: r,
				})
			}
		}
		for _, u := range details.UserDetailList {
			include := filterMatch == nil
			for _, pol := range u.UserPolicyList {
				doc, err := url.QueryUnescape(*pol.PolicyDocument)
				if err != nil {
					log.Fatalf("Unescape policy doc err doc=%q err=%s", *pol.PolicyDocument, err)
				}
				pol.PolicyDocument = &doc
				if include == false && filterMatch != nil {
					if filterMatch.MatchString(doc) {
						include = true
					}
				}
			}
			if include {
				jsonOut.Encode(IamObject{
					Type: "user",
					User: u,
				})
			}
		}
		return true
	})

	if err != nil {
		log.Fatalf("GetAccountAuthorizationDetails err: %s", err)
	}
}

var (
	principalActions []string
	resoucesFlag     []string
)

func testAllIamIdentitiesCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "test-all-iam-identites",
		Short: "Test access permission to a action+resource for all iam identites in account",
		Run:   testAllIamIdentitiesAction,
	}

	cmd.Flags().StringArrayVarP(&principalActions, "actions", "", nil, "List of api operations (e.g kms:Decrypt)")
	cmd.Flags().StringArrayVarP(&resoucesFlag, "resources", "", nil, "Resources to test access against (e.g. arn:aws:kms:us-east-1:123456789012:key/e50f9eee-b521-47c8-8d67-3058d3409969")

	return &cmd
}

func testAllIamIdentitiesAction(cmd *cobra.Command, args []string) {
	iamSvc := iam.New(config.Session())

	if len(principalActions) < 1 {
		log.Fatalf("--actions is required")
	}

	if len(resoucesFlag) < 1 {
		log.Fatalf("--resources is required")
	}

	actionNames := aws.StringSlice(principalActions)
	resourceArns := aws.StringSlice(resoucesFlag)

	csvOut := csv.NewWriter(os.Stdout)
	defer csvOut.Flush()

	input := &iam.GetAccountAuthorizationDetailsInput{}
	err := iamSvc.GetAccountAuthorizationDetailsPages(input, func(details *iam.GetAccountAuthorizationDetailsOutput, b bool) bool {
		for _, m := range details.Policies {
			scpi := &iam.SimulateCustomPolicyInput{
				ActionNames:     actionNames,
				ResourceArns:    resourceArns,
				PolicyInputList: make([]*string, 0, len(m.PolicyVersionList)),
			}

			for _, pol := range m.PolicyVersionList {
				if pol.Document != nil {
					doc, err := url.QueryUnescape(*pol.Document)
					if err != nil {
						log.Fatalf("Unescape policy doc err doc=%q err=%s", *pol.Document, err)
					}

					scpi.PolicyInputList = append(scpi.PolicyInputList, &doc)
				}
			}
			simResult, err := iamSvc.SimulateCustomPolicy(scpi)
			if err != nil {
				log.Fatalf("Failed to simulate policy for %s: %s", *m.Arn, err)
			}

			for _, res := range simResult.EvaluationResults {
				csvOut.Write([]string{*res.EvalDecision, *res.EvalActionName, *res.EvalResourceName, *m.Arn})
				csvOut.Flush()
			}
		}

		simulate := func(arn string) {
			sppi := &iam.SimulatePrincipalPolicyInput{
				ActionNames:     actionNames,
				ResourceArns:    resourceArns,
				PolicySourceArn: &arn,
			}
			simResult, err := iamSvc.SimulatePrincipalPolicy(sppi)
			if err != nil {
				log.Fatalf("Failed to simulate permission for %s: %s", arn, err)
			}
			for _, res := range simResult.EvaluationResults {
				csvOut.Write([]string{*res.EvalDecision, *res.EvalActionName, *res.EvalResourceName, arn})
				csvOut.Flush()
			}
		}

		for _, g := range details.GroupDetailList {
			simulate(*g.Arn)
		}
		for _, r := range details.RoleDetailList {
			simulate(*r.Arn)
		}
		for _, u := range details.UserDetailList {
			simulate(*u.Arn)
		}
		return true
	})

	if err != nil {
		log.Fatalf("GetAccountAuthorizationDetails err: %s", err)
	}
}
