---
subcategory: "Slack"
page_title: "Slack: slack_conversation"
---

# slack_conversation Data Source

Use this data source to get information about a Slack conversation for use in other
resources.

## Required scopes

This resource requires the following scopes:

- [channels:read](https://api.slack.com/scopes/channels:read) (public channels)
- [groups:read](https://api.slack.com/scopes/groups:read) (private channels)

The Slack API methods used by the resource are:

- [conversations.info](https://api.slack.com/methods/conversations.info)
- [conversations.members](https://api.slack.com/methods/conversations.members)

If you get `missing_scope` errors while using this resource check the scopes against
the documentation for the methods above.

## Example Usage

### Lookup by channel ID

```hcl
data "slack_conversation" "by_id" {
  channel_id = "C01234ABCDE"
}

output "channel_name" {
  value = data.slack_conversation.by_id.name
}
```

### Lookup public channel by name

```hcl
data "slack_conversation" "engineering" {
  name = "engineering"
}

output "channel_topic" {
  value = data.slack_conversation.engineering.topic
}
```

### Lookup private channel by name

```hcl
data "slack_conversation" "private_channel" {
  name       = "private-team"
  is_private = true
}

output "channel_purpose" {
  value = data.slack_conversation.private_channel.purpose
}
```

## Argument Reference

The following arguments are supported:

- `channel_id` - (Optional) The ID of the channel
- `name` - (Optional) The name of the public or private channel
- `is_private` - (Optional) The conversation is privileged between two or more members

Either `channel_id` or `name` must be provided. `is_private` only works in conjunction
with `name`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the channel (e.g., `C01234ABCDE`)
- `name` - Name of the public or private channel
- `topic` - Topic for the channel (max 250 characters)
- `purpose` - Purpose of the channel (max 250 characters)
- `creator` - User ID of the member that created this channel (e.g., `U01234ABCDE`)
- `created` - Unix timestamp of when the channel was created
- `is_private` - Whether the conversation is a private channel
- `is_archived` - Whether the conversation is archived (frozen in time, no new messages allowed)
- `is_shared` - Whether the conversation is in some way shared between multiple workspaces
- `is_ext_shared` - Whether this conversation is part of a Shared Channel with a remote organization
- `is_org_shared` - Whether this shared channel is shared between Enterprise Grid workspaces within the same organization
- `is_general` - Whether this is the "general" channel that includes all regular team members in the workspace
