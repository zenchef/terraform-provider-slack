# ============================================================================
# Data Source Outputs
# ============================================================================

output "user_by_name" {
  description = "User information looked up by name"
  value = {
    id    = data.slack_user.by_name.id
    name  = data.slack_user.by_name.name
    email = data.slack_user.by_name.email
  }
}

output "existing_usergroup" {
  description = "Existing usergroup information"
  value       = data.slack_usergroup.existing_group
}

output "existing_channel" {
  description = "Existing channel information"
  value = {
    id         = data.slack_conversation.existing_channel.id
    name       = data.slack_conversation.existing_channel.name
    topic      = data.slack_conversation.existing_channel.topic
    purpose    = data.slack_conversation.existing_channel.purpose
    is_private = data.slack_conversation.existing_channel.is_private
  }
}

# ============================================================================
# Usergroup Outputs
# ============================================================================

output "engineering_team_id" {
  description = "ID of the engineering usergroup"
  value       = slack_usergroup.engineering.id
}

output "engineering_team_handle" {
  description = "Handle/mention of the engineering usergroup"
  value       = slack_usergroup.engineering.handle
}

output "support_team" {
  description = "Support team usergroup details"
  value = {
    id          = slack_usergroup.support.id
    name        = slack_usergroup.support.name
    handle      = slack_usergroup.support.handle
    description = slack_usergroup.support.description
  }
}

# ============================================================================
# Conversation Outputs
# ============================================================================

output "general_channel_id" {
  description = "ID of the general discussion channel"
  value       = slack_conversation.general_channel.id
}

output "private_channel_info" {
  description = "Private channel information"
  value = {
    id         = slack_conversation.private_channel.id
    name       = slack_conversation.private_channel.name
    topic      = slack_conversation.private_channel.topic
    is_private = slack_conversation.private_channel.is_private
  }
}

output "devops_channel_id" {
  description = "ID of the devops team channel"
  value       = slack_conversation.devops_channel.id
}

# ============================================================================
# Combined Outputs
# ============================================================================

output "devops_team_info" {
  description = "Complete DevOps team information"
  value = {
    usergroup_id   = slack_usergroup.devops.id
    usergroup_name = slack_usergroup.devops.name
    channel_id     = slack_conversation.devops_channel.id
    channel_name   = slack_conversation.devops_channel.name
  }
}
