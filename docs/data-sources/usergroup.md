---
subcategory: "Slack"
page_title: "Slack: slack_usergroup"
---

# slack_usergroup Data Source

Use this data source to get information about a usergroups for use in other
resources. The data source returns enabled groups only.

## Required scopes

This resource requires the following scopes:

- [usergroups:read](https://api.slack.com/scopes/usergroups:read)

The Slack API methods used by the resource are:

- [usergroups.list](https://api.slack.com/methods/usergroups.list)

If you get `missing_scope` errors while using this resource check the scopes against
the documentation for the methods above.

## Example Usage

### Lookup by name

```hcl
data "slack_usergroup" "engineering" {
  name = "engineering"
}

output "engineering_usergroup_id" {
  value = data.slack_usergroup.engineering.id
}
```

### Lookup by ID

```hcl
data "slack_usergroup" "by_id" {
  usergroup_id = "S01234ABCDE"
}

output "usergroup_members" {
  value = data.slack_usergroup.by_id.users
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Optional) The name of the usergroup
- `usergroup_id` - (Optional) The id of the usergroup

The data source expects exactly one of these fields, you can't set both.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the usergroup (e.g., `S01234ABCDE`)
- `name` - The name of the usergroup
- `description` - The short description of the usergroup
- `handle` - The mention handle for the usergroup (e.g., `engineers` for `@engineers`)
- `users` - Set of user IDs that are members of the usergroup
- `channels` - Set of channel IDs that are set as default channels for the usergroup
