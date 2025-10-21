# ============================================================================
# Data Sources - Lookup existing Slack resources
# ============================================================================

# Lookup user by name
data "slack_user" "by_name" {
  name = "john.doe"
}

# Lookup user by email
data "slack_user" "by_email" {
  email = "john.doe@example.com"
}

# Lookup existing usergroup by name
data "slack_usergroup" "existing_group" {
  name = "existing-group"
}

# Lookup existing conversation by ID
data "slack_conversation" "existing_channel" {
  id = "C1234567890"
}

# ============================================================================
# Usergroups - Manage Slack user groups
# ============================================================================

# Create a basic usergroup
resource "slack_usergroup" "my_team" {
  name        = "my-team"
  handle      = "myteam"
  description = "My awesome team"
}

# Create a usergroup with users
resource "slack_usergroup" "engineering" {
  name        = "engineering"
  handle      = "engineers"
  description = "Engineering team"
  users       = [data.slack_user.by_name.id, data.slack_user.by_email.id]
}

# Create a usergroup with default channels
resource "slack_usergroup" "support" {
  name        = "support"
  handle      = "support"
  description = "Support team"
  users       = [data.slack_user.by_name.id]
  channels    = [slack_conversation.support_channel.id]
}

# ============================================================================
# Conversations - Manage Slack channels
# ============================================================================

# Create a basic public channel
resource "slack_conversation" "general_channel" {
  name       = "general-discussion"
  topic      = "General team discussions"
  purpose    = "A place for general team discussions"
  is_private = false
}

# Create a private channel with members
resource "slack_conversation" "private_channel" {
  name              = "private-team-channel"
  topic             = "Private team discussions"
  purpose           = "A private channel for the team"
  is_private        = true
  permanent_members = [data.slack_user.by_name.id, data.slack_user.by_email.id]
}

# Create a channel with archive settings
resource "slack_conversation" "project_channel" {
  name              = "project-alpha"
  topic             = "Project Alpha discussions"
  purpose           = "Channel for Project Alpha"
  is_private        = false
  permanent_members = [data.slack_user.by_name.id]

  # Archive the channel when destroyed (default behavior)
  action_on_destroy = "archive"

  # Kick users when removed from permanent_members (default behavior)
  action_on_update_permanent_members = "kick"
}

# Create an archived channel
resource "slack_conversation" "archived_channel" {
  name        = "old-project"
  topic       = "Old project discussions"
  purpose     = "Channel for an old project"
  is_private  = false
  is_archived = true
}

# Create a support channel
resource "slack_conversation" "support_channel" {
  name       = "support"
  topic      = "Customer support discussions"
  purpose    = "Channel for customer support team"
  is_private = true

  # Don't archive the channel when destroyed
  action_on_destroy = "none"
}

# ============================================================================
# Example: Complete workflow
# ============================================================================

# 1. Create a usergroup with users and default channels
resource "slack_usergroup" "devops" {
  name        = "devops"
  handle      = "devops"
  description = "DevOps team"
  users       = [data.slack_user.by_name.id, data.slack_user.by_email.id]
  channels    = [slack_conversation.devops_channel.id]

  depends_on = [slack_conversation.devops_channel]
}

# 2. Create a channel for the team
resource "slack_conversation" "devops_channel" {
  name              = "devops-team"
  topic             = "DevOps team channel"
  purpose           = "Channel for DevOps team communications"
  is_private        = true
  permanent_members = [data.slack_user.by_name.id, data.slack_user.by_email.id]
}
