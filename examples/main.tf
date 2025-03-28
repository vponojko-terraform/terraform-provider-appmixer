terraform {
  required_providers {
    appmixer = {
      source  = "vponojko-terraform/appmixer"
      version = "0.1.0"
    }
  }
}

provider "appmixer" {
  api_url  = "https://api.your-tenant.appmixer.cloud"  # Replace with your API URL
  email    = "admin@example.com"                      # Replace with your email or use environment variable APPMIXER_EMAIL
  password = "password"                               # Replace with your password or use environment variable APPMIXER_PASSWORD
}

# Get information about the current authenticated user
data "appmixer_user" "current" {
}

output "current_user" {
  value     = data.appmixer_user.current
  sensitive = true
}

# Create a new user
resource "appmixer_user" "test_user" {
  username = "test-user@example.com"
  email    = "test-user@example.com"
  password = "secure-password123"
}

output "created_user" {
  value = {
    id        = appmixer_user.test_user.id
    username  = appmixer_user.test_user.username
    email     = appmixer_user.test_user.email
    is_active = appmixer_user.test_user.is_active
    plan      = appmixer_user.test_user.plan
    created   = appmixer_user.test_user.created
  }
}

# Admin-only: List all users with pagination and filtering
data "appmixer_users" "all_users" {
  limit   = 10
  offset  = 0
  sort    = "created:-1"
  # Uncomment if needed:
  # filter  = "scope:user"
  # pattern = "example.com"
}

output "all_users" {
  value = data.appmixer_users.all_users.users
}

# Admin-only: Get total user count
data "appmixer_users_count" "count" {
}

output "user_count" {
  value = data.appmixer_users_count.count.count
} 