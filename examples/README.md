# Terraform Slack Provider Examples

This directory contains example Terraform configurations for using the Slack provider.

## Prerequisites

1. A Slack workspace with admin access
2. A Slack Bot Token with the following scopes:
   - `channels:read`, `channels:write`, `channels:manage`
   - `groups:read`, `groups:write`
   - `usergroups:read`, `usergroups:write`
   - `users:read`, `users:read.email`
3. Terraform >= 1.0

## Getting Started

### 1. Set up your Slack token

Export your Slack bot token as an environment variable:

```bash
export SLACK_TOKEN="xoxb-your-slack-bot-token"
```

Alternatively, you can set it in the provider configuration in `provider.tf`.

### 2. Update the examples

Before running the examples, update the following in `main.tf`:

- Replace `john.doe` with an actual username in your workspace
- Replace `john.doe@example.com` with an actual email in your workspace
- Replace `existing-group` with an actual usergroup name (or remove that data source)

### 3. Initialize Terraform

```bash
terraform init
```

### 4. Review the plan

```bash
terraform plan
```

### 5. Apply the configuration

```bash
terraform apply
```

## Examples Overview

### Data Sources

- **slack_user**: Lookup users by name or email
- **slack_usergroup**: Lookup existing usergroups by name or ID

### Resources

#### slack_usergroup

Create and manage Slack usergroups:

```hcl
resource "slack_usergroup" "my_team" {
  name        = "my-team"
  handle      = "myteam"
  description = "My awesome team"
  users       = [data.slack_user.by_name.id]
  channels    = [slack_conversation.my_channel.id]
}
```

**Attributes:**
- `name` (required): Name of the usergroup
- `handle` (optional): Handle/mention name (e.g., @myteam)
- `description` (optional): Description of the usergroup
- `users` (optional): List of user IDs who are members
- `channels` (optional): List of channel IDs where the group is mentioned

#### slack_conversation

Create and manage Slack channels:

```hcl
resource "slack_conversation" "my_channel" {
  name              = "my-channel"
  topic             = "Channel topic"
  purpose           = "Channel purpose"
  is_private        = false
  is_archived       = false
  permanent_members = [data.slack_user.by_name.id]

  action_on_destroy                  = "archive"
  action_on_update_permanent_members = "kick"
}
```

**Attributes:**
- `name` (required): Name of the channel
- `topic` (optional): Channel topic
- `purpose` (optional): Channel purpose
- `is_private` (required): Whether the channel is private
- `is_archived` (optional): Whether the channel is archived (default: false)
- `permanent_members` (optional): List of user IDs who are permanent members
- `action_on_destroy` (optional): Action when destroying (`archive` or `none`, default: `archive`)
- `action_on_update_permanent_members` (optional): Action when removing members (`kick` or `none`, default: `kick`)

## Common Use Cases

### Create a team with a dedicated channel

```hcl
resource "slack_usergroup" "engineering" {
  name        = "engineering"
  handle      = "engineers"
  description = "Engineering team"
  users       = [data.slack_user.engineer1.id, data.slack_user.engineer2.id]
}

resource "slack_conversation" "engineering_channel" {
  name              = "engineering"
  topic             = "Engineering discussions"
  is_private        = true
  permanent_members = slack_usergroup.engineering.users
}
```

### Manage channel membership

```hcl
resource "slack_conversation" "project_channel" {
  name              = "project-alpha"
  is_private        = false
  permanent_members = [
    data.slack_user.pm.id,
    data.slack_user.dev1.id,
    data.slack_user.dev2.id
  ]

  # Automatically remove users when they're removed from permanent_members
  action_on_update_permanent_members = "kick"
}
```

### Create an archived channel

```hcl
resource "slack_conversation" "old_project" {
  name        = "old-project"
  is_private  = false
  is_archived = true
}
```

## Cleanup

To destroy all resources created by these examples:

```bash
terraform destroy
```

**Note**: By default, channels are archived (not deleted) when destroyed. To prevent archiving, set `action_on_destroy = "none"` on the conversation resource.
