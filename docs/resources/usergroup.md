---
subcategory: "Slack"
page_title: "Slack: slack_usergroup"
---

# slack_usergroup Resource

Manages a Slack User Group.

## Required scopes

This resource requires the following scopes:

- [usergroups:write](https://api.slack.com/scopes/usergroups:write)
- [usergroups:read](https://api.slack.com/scopes/usergroups:read)

The Slack API methods used by the resource are:

- [usergroups.create](https://api.slack.com/methods/usergroups.create)
- [usergroups.enable](https://api.slack.com/methods/usergroups.enable)
- [usergroups.disable](https://api.slack.com/methods/usergroups.disable)
- [usergroups.update](https://api.slack.com/methods/usergroups.update)
- [usergroups.list](https://api.slack.com/methods/usergroups.list)
- [usergroups.users.update](https://api.slack.com/methods/usergroups.users.update)

If you get `missing_scope` errors while using this resource check the scopes against
the documentation for the methods above.

## Example Usage

### Basic usergroup

```hcl
resource "slack_usergroup" "engineering" {
  name        = "engineering"
  handle      = "engineers"
  description = "Engineering team"
}
```

### Usergroup with users

```hcl
data "slack_user" "engineer1" {
  email = "engineer1@example.com"
}

data "slack_user" "engineer2" {
  email = "engineer2@example.com"
}

resource "slack_usergroup" "engineering" {
  name        = "engineering"
  handle      = "engineers"
  description = "Engineering team"
  users       = [
    data.slack_user.engineer1.id,
    data.slack_user.engineer2.id
  ]
}
```

### Usergroup with default channels

```hcl
resource "slack_conversation" "engineering_channel" {
  name       = "engineering"
  topic      = "Engineering discussions"
  is_private = true
}

resource "slack_usergroup" "engineering" {
  name        = "engineering"
  handle      = "engineers"
  description = "Engineering team"
  channels    = [slack_conversation.engineering_channel.id]
}
```

### Integrated usergroup and channel with synchronized membership

**Important**: When a channel is removed from the `channels` list, users are **not** automatically removed from the channel. To keep usergroup membership and channel membership synchronized, use `permanent_members` in the channel resource:

```hcl
resource "slack_usergroup" "devops" {
  name        = "devops"
  handle      = "devops"
  description = "DevOps team"
  users       = [
    data.slack_user.devops1.id,
    data.slack_user.devops2.id
  ]
}

resource "slack_conversation" "devops_channel" {
  name              = "devops-team"
  topic             = "DevOps team channel"
  is_private        = true

  # Keep channel membership synchronized with usergroup
  permanent_members = slack_usergroup.devops.users

  # Automatically kick users when removed from the usergroup
  action_on_update_permanent_members = "kick"
}
```

## Argument Reference

The following arguments are supported:

### Required Arguments

- `name` - (Required) Name for the usergroup. Must be unique among all usergroups in the workspace.

### Optional Arguments

- `handle` - (Optional, Computed) Mention handle for the usergroup (e.g., `engineers` for `@engineers`). Must be unique among channels, users, and usergroups. If not specified, Slack will generate one based on the name.
- `description` - (Optional, Computed) Short description of the usergroup. If not specified, defaults to empty string.
- `users` - (Optional) Set of user IDs that represent the complete membership of the usergroup. When updated, this replaces the entire membership list.
- `channels` - (Optional) Set of channel IDs where this usergroup should be set as a default. Members of the usergroup will see these channels as suggestions when they join Slack or when mentioned.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The usergroup ID

## Import

`slack_usergroup` can be imported using the ID of the group, e.g.

```shell
terraform import slack_usergroup.my_group S022GE79E9G
```
