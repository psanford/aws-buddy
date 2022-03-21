# aws-buddy

`aws-buddy` is my personal aws cli tool. Its main goal is to make common tasks I do easy and fast.

I mostly use this tool to avoid having to load the aws web interface or use the clunky aws cli tool.

Maybe you'll find it useful too.

## Commands:

```
AWS tools

Usage:
  aws-buddy [command]

Available Commands:
  completion  Generates bash completion scripts
  config      AWS Config Commands
  cost        COST Commands
  ec2         EC2 Commands
  help        Help about any command
  help-tree   Print Help for all commands
  iam         IAM Commands
  org         Organization Commands
  param       SSM Parameter Store Commands
  route53     Route53 Commands
  s3          S3 Commands

Use "aws-buddy [command] --help" for more information about a command.

========================================
To load completion run

. <(aws-buddy completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(aws-buddy completion)

Usage:
  aws-buddy completion

========================================
AWS Config Commands

Usage:
  aws-buddy config [command]

Available Commands:
  query_by_id            Query by resource id for resoucre
  query_eni_by_public_ip Query for ENIs matching a public ip

Use "aws-buddy config [command] --help" for more information about a command.

========================================
Query by resource id for resoucre

Usage:
  aws-buddy config query_by_id [flags]

Flags:
      --aggregator-name string   AWS Config Aggretator Name (default "AllAccounts")

========================================
Query for ENIs matching a public ip

Usage:
  aws-buddy config query_eni_by_public_ip [flags]

Flags:
      --aggregator-name string   AWS Config Aggretator Name (default "AllAccounts")

========================================
COST Commands

Usage:
  aws-buddy cost [command]

Available Commands:
  daily       Show daily costs

Use "aws-buddy cost [command] --help" for more information about a command.

========================================
Show daily costs

Usage:
  aws-buddy cost daily [flags]

Flags:
      --days int   Number of days to fetch (default 14)

========================================
EC2 Commands

Usage:
  aws-buddy ec2 [command]

Available Commands:
  ami             AMI subcommands
  asg             ASG Commands
  ip              IP Commands
  launch          Launch instance
  launch_template Launch template command
  list            List instances
  security_group  Security Group Commands
  show            Show Instance
  tag             Tag Commands
  volume          Volume Commands

Use "aws-buddy ec2 [command] --help" for more information about a command.

========================================
AMI subcommands

Usage:
  aws-buddy ec2 ami [command]

Available Commands:
  list_ubuntu list ubuntu AMIs

Use "aws-buddy ec2 ami [command] --help" for more information about a command.

========================================
list ubuntu AMIs

Usage:
  aws-buddy ec2 ami list_ubuntu

========================================
ASG Commands

Usage:
  aws-buddy ec2 asg [command]

Available Commands:
  scaling-activites list scaling activities

Use "aws-buddy ec2 asg [command] --help" for more information about a command.

========================================
list scaling activities

Usage:
  aws-buddy ec2 asg scaling-activites <asg-name>

========================================
IP Commands

Usage:
  aws-buddy ec2 ip [command]

Available Commands:
  list        list ips by eni association

Use "aws-buddy ec2 ip [command] --help" for more information about a command.

========================================
list ips by eni association

Usage:
  aws-buddy ec2 ip list [flags]

Flags:
      --json   Show raw json ouput

========================================
Launch instance

Usage:
  aws-buddy ec2 launch <launch_tmpl.yml>

========================================
Launch template command

Usage:
  aws-buddy ec2 launch_template <name>

========================================
List instances

Usage:
  aws-buddy ec2 list [flags]

Flags:
  -f, --filter string        Filter results by name or id
      --filter-name string   API Filter by Tag:Name
      --json                 Show raw json ouput
      --truncate             Trucate fields (default true)
  -v, --verbose              Show verbose (multi-line) output

========================================
Security Group Commands

Usage:
  aws-buddy ec2 security_group [command]

Aliases:
  security_group, sg

Available Commands:
  list        list security groups
  show        Show a security group

Use "aws-buddy ec2 security_group [command] --help" for more information about a command.

========================================
list security groups

Usage:
  aws-buddy ec2 security_group list [flags]

Flags:
      --json   Show raw json ouput

========================================
Show a security group

Usage:
  aws-buddy ec2 security_group show <sg-id>

========================================
Show Instance

Usage:
  aws-buddy ec2 show <instance-id> [flags]

Flags:
      --filter-name string   API Filter by Tag:Name
      --json                 Show raw json ouput
  -v, --verbose              Show verbose (multi-line) output

========================================
Tag Commands

Usage:
  aws-buddy ec2 tag [command]

Available Commands:
  list        list tags on instance
  rm          remove tag on instance
  set         set tag on instance

Use "aws-buddy ec2 tag [command] --help" for more information about a command.

========================================
list tags on instance

Usage:
  aws-buddy ec2 tag list <instance-id>

Aliases:
  list, ls

========================================
remove tag on instance

Usage:
  aws-buddy ec2 tag rm <instance-id> <tag-name>

========================================
set tag on instance

Usage:
  aws-buddy ec2 tag set <instance-id> <tag-name> <tag-value>

========================================
Volume Commands

Usage:
  aws-buddy ec2 volume [command]

Available Commands:
  list        list volumes

Use "aws-buddy ec2 volume [command] --help" for more information about a command.

========================================
list volumes

Usage:
  aws-buddy ec2 volume list [flags]

Aliases:
  list, ls

Flags:
      --csv    Output as csv
      --json   Show raw json ouput

========================================
Help provides help for any command in the application.
Simply type aws-buddy help [path to command] for full details.

Usage:
  aws-buddy help [command]

========================================
Print Help for all commands

Usage:
  aws-buddy help-tree [flags]

Flags:
  -h, --help   help for help-tree

========================================
IAM Commands

Usage:
  aws-buddy iam [command]

Available Commands:
  access      Commands to help review access by iam principals
  user        User Commands

Use "aws-buddy iam [command] --help" for more information about a command.

========================================
Commands to help review access by iam principals

Usage:
  aws-buddy iam access [command]

Available Commands:
  account-authorization-details Get snapshot of account permissions
  test-all-iam-identites        Test access permission to a action+resource for all iam identites in account

Use "aws-buddy iam access [command] --help" for more information about a command.

========================================
Get snapshot of account permissions

Usage:
  aws-buddy iam access account-authorization-details [flags]

Flags:
      --filter-by-policy-match string   Regex string to match on policy documents

========================================
Test access permission to a action+resource for all iam identites in account

Usage:
  aws-buddy iam access test-all-iam-identites [flags]

Flags:
      --actions stringArray     List of api operations (e.g kms:Decrypt)
      --resources stringArray   Resources to test access against (e.g. arn:aws:kms:us-east-1:123456789012:key/e50f9eee-b521-47c8-8d67-3058d3409969

========================================
User Commands

Usage:
  aws-buddy iam user [command]

Available Commands:
  list             List users
  list-access-keys List all access keys in account
  show             Show user

Use "aws-buddy iam user [command] --help" for more information about a command.

========================================
List users

Usage:
  aws-buddy iam user list [flags]

Flags:
      --csv        Show csv ouput
      --full-arn   Show full arn for username
      --json       Show raw json ouput

========================================
List all access keys in account

Usage:
  aws-buddy iam user list-access-keys

========================================
Show user

Usage:
  aws-buddy iam user show <username> [flags]

Flags:
      --json   Show raw json ouput

========================================
Organization Commands

Usage:
  aws-buddy org [command]

Available Commands:
  each         Run command against each account
  list         List accounts
  list-ou-tree List organizational units

Use "aws-buddy org [command] --help" for more information about a command.

========================================
Run command against each account

Usage:
  aws-buddy org each [flags]

Flags:
      --external-cmd string   External command to run instead of a buddy command
      --org-list string       File with list of org ids (empty means use the current accounts org list)
      --role string           Role name to assume in each account

========================================
List accounts

Usage:
  aws-buddy org list [flags]

Flags:
      --json   Show raw json ouput

========================================
List organizational units

Usage:
  aws-buddy org list-ou-tree [flags]

Flags:
      --include-accts   Include Child accounts

========================================
SSM Parameter Store Commands

Usage:
  aws-buddy param [command]

Available Commands:
  cp          Copy param from old to new path
  get         Get parameter value
  list        List parameter
  put         Set or create parameter value
  rm          Delete param at path

Use "aws-buddy param [command] --help" for more information about a command.

========================================
Copy param from old to new path

Usage:
  aws-buddy param cp

========================================
Get parameter value

Usage:
  aws-buddy param get

========================================
List parameter

Usage:
  aws-buddy param list [flags]

Flags:
      --json   Show raw json ouput

========================================
Set or create parameter value

Usage:
  aws-buddy param put [flags]

Flags:
      --description string   Param description
      --type string          Param type (String, StringList, SecureString) (default "SecureString")

========================================
Delete param at path

Usage:
  aws-buddy param rm

========================================
Route53 Commands

Usage:
  aws-buddy route53 [command]

Available Commands:
  list        List Records
  zones       List Zones

Use "aws-buddy route53 [command] --help" for more information about a command.

========================================
List Records

Usage:
  aws-buddy route53 list [flags]

Flags:
      --json          Show raw json ouput
      --zone string   Filter by zone name

========================================
List Zones

Usage:
  aws-buddy route53 zones [flags]

Flags:
      --json   Show raw json ouput

========================================
S3 Commands

Usage:
  aws-buddy s3 [command]

Available Commands:
  cat         Cat object
  head        Head object

Use "aws-buddy s3 [command] --help" for more information about a command.

========================================
Cat object

Usage:
  aws-buddy s3 cat <[s3://]bucket/path/to/object>

========================================
Head object

Usage:
  aws-buddy s3 head <[s3://]bucket/path/to/object>
```
