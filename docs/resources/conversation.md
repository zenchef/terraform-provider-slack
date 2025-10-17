---
subcategory: "Slack"
page_title: "Slack: slack_conversation"
---

# slack_conversation Resource

Manages a Slack channel

## Required scopes

This resource requires the following scopes:

If using `bot` tokens:

- [channels:read](https://api.slack.com/scopes/channels:read)
(public channels)
- [channels:manage](https://api.slack.com/scopes/channels:manage)
(public channels)
- [channels:join](https://api.slack.com/scopes/channels:join)
(adopting existing public channels)
- [groups:read](https://api.slack.com/scopes/groups:read)
(private channels)
- [groups:write](https://api.slack.com/scopes/groups:write)
(private channels)

If using `user` tokens:

- [channels:read](https://api.slack.com/scopes/channels:read) (public channels)
- [channels:write](https://api.slack.com/scopes/channels:manage) (public channels)
- [groups:read](https://api.slack.com/scopes/groups:read) (private channels)
- [groups:write](https://api.slack.com/scopes/groups:write) (private channels)

The Slack API methods used by the resource are:

- [conversations.create](https://api.slack.com/methods/conversations.create)
- [conversations.join](https://api.slack.com/methods/conversations.join)
- [conversations.setTopic](https://api.slack.com/methods/conversations.setTopic)
- [conversations.setPurpose](https://api.slack.com/methods/conversations.setPurpose)
- [conversations.info](https://api.slack.com/methods/conversations.info)
- [conversations.members](https://api.slack.com/methods/conversations.members)
- [conversations.kick](https://api.slack.com/methods/conversations.kick)
- [conversations.invite](https://api.slack.com/methods/conversations.invite)
- [conversations.rename](https://api.slack.com/methods/conversations.rename)
- [conversations.archive](https://api.slack.com/methods/conversations.archive)
- [conversations.unarchive](https://api.slack.com/methods/conversations.unarchive)

If you get `missing_scope` errors while using this resource check the scopes against
the documentation for the methods above.

## Example Usage

### Basic public channel

```hcl
resource "slack_conversation" "general" {
  name       = "general-discussion"
  topic      = "General team discussions"
  purpose    = "A place for general team discussions"
  is_private = false
}
```

### Private channel with permanent members

```hcl
data "slack_user" "team_lead" {
  email = "lead@example.com"
}

data "slack_user" "developer" {
  email = "dev@example.com"
}

resource "slack_conversation" "private_project" {
  name              = "project-alpha"
  topic             = "Project Alpha discussions"
  purpose           = "Private channel for Project Alpha team"
  is_private        = true
  permanent_members = [
    data.slack_user.team_lead.id,
    data.slack_user.developer.id
  ]

  # Kick users when removed from permanent_members
  action_on_update_permanent_members = "kick"
}
```

### Channel integrated with usergroup

```hcl
resource "slack_usergroup" "devops" {
  name        = "devops"
  handle      = "devops"
  description = "DevOps team"
}

resource "slack_conversation" "devops_channel" {
  name              = "devops-team"
  topic             = "DevOps team channel"
  purpose           = "Channel for DevOps team communications"
  is_private        = true
  permanent_members = slack_usergroup.devops.users
}
```

### Archived channel

```hcl
resource "slack_conversation" "old_project" {
  name        = "old-project"
  topic       = "Old project archive"
  is_private  = false
  is_archived = true
}
```

### Channel that won't be archived on destroy

```hcl
resource "slack_conversation" "persistent" {
  name              = "important-channel"
  topic             = "This channel won't be archived on destroy"
  is_private        = false
  action_on_destroy = "none"
}
```

### Adopting an existing channel

```hcl
resource "slack_conversation" "adopted" {
  name                               = "existing-channel"
  topic                              = "Adopt existing channel, don't kick members"
  adopt_existing_channel             = true
  action_on_update_permanent_members = "none"
  is_private                         = false
}
```

## Argument Reference

The following arguments are supported:

### Required Arguments

- `name` - (Required) Name of the public or private channel. Channel names can only contain lowercase letters, numbers, hyphens, and underscores, and must be 80 characters or less.
- `is_private` - (Required) Create a private channel instead of a public one. This cannot be changed after creation.

### Optional Arguments

- `topic` - (Optional) Topic for the channel (max 250 characters).
- `purpose` - (Optional) Purpose of the channel (max 250 characters).
- `permanent_members` - (Optional) Set of user IDs to manage as permanent members of the channel. The channel creator is automatically a member and does not need to be included in this list. When users are removed from this list, the behavior is controlled by `action_on_update_permanent_members`.
- `is_archived` - (Optional, Default: `false`) Whether the conversation is archived. Archived channels are frozen in time - no messages can be posted and membership cannot be changed.
- `action_on_destroy` - (Optional, Default: `archive`) Action to take when the resource is destroyed. Valid values:
  - `archive` - Archive the channel on destroy (default behavior)
  - `none` - Leave the channel as-is. **Warning**: Subsequent applies with the same name will fail
- `action_on_update_permanent_members` - (Optional, Default: `kick`) Action to take when users are removed from `permanent_members`. Valid values:
  - `kick` - Remove users from the channel when removed from `permanent_members` (default behavior)
  - `none` - Do not remove users. Useful for public channels where users can self-join
- `adopt_existing_channel` - (Optional, Default: `false`) Adopt an existing channel with the same name and bring it under Terraform management. If the existing channel is archived, it will be unarchived. **Note**: For unarchiving existing channels, you must use a user token, not a bot token, due to Slack API limitations.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The channel ID (e.g. C015QDUB7ME).
- `creator` - is the user ID of the member that created this channel.
- `created` - is a unix timestamp.
- `is_shared` - means the conversation is in some way shared between multiple workspaces.
- `is_ext_shared` - represents this conversation as being part of a Shared Channel
with a remote organization.
- `is_org_shared` - explains whether this shared channel is shared between Enterprise
Grid workspaces within the same organization.
- `is_general` - will be true if this channel is the "general" channel that includes
all regular team members.

## Import

`slack_conversation` can be imported using the ID of the conversation/channel, e.g.

```shell
terraform import slack_conversation.my_conversation C023X7QTFHQ
```
