# Getting Started with Appmixer Provider

This guide covers how to get started with the Appmixer Terraform Provider.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) 0.13+
- Access to an Appmixer instance (self-hosted or cloud)
- Valid admin credentials for the Appmixer instance

## Provider Configuration

To use the Appmixer provider, you first need to configure it in your Terraform configuration:

```hcl
terraform {
  required_providers {
    appmixer = {
      source  = "vponojko-terraform/appmixer"
      version = ">= 0.1.0"
    }
  }
}

provider "appmixer" {
  api_url  = "https://api.your-tenant.appmixer.cloud"
  email    = "admin@example.com"
  password = "your-secure-password"
}
```

Alternatively, you can use environment variables to configure the provider:

```bash
export APPMIXER_API_URL="https://api.your-tenant.appmixer.cloud"
export APPMIXER_EMAIL="admin@example.com"
export APPMIXER_PASSWORD="your-secure-password"
```

And then configure the provider without inline credentials:

```hcl
provider "appmixer" {}
```

## Basic Usage Examples

### Retrieving the Current User

```hcl
data "appmixer_user" "current" {}

output "current_user" {
  value = {
    id       = data.appmixer_user.current.id
    email    = data.appmixer_user.current.email
    username = data.appmixer_user.current.username
    scope    = data.appmixer_user.current.scope
  }
  sensitive = true
}
```

### Creating a New User

```hcl
resource "appmixer_user" "new_user" {
  username = "new-user@example.com"
  email    = "new-user@example.com"
  password = "secure-password"
}

output "new_user_id" {
  value = appmixer_user.new_user.id
}
```

### Advanced: Creating an Admin User

```hcl
resource "appmixer_user" "admin_user" {
  username = "admin-user@example.com"
  email    = "admin-user@example.com"
  password = "secure-admin-password"
  scope    = ["user", "admin"]
}
```

### Retrieving All Users (Admin Required)

```hcl
data "appmixer_users" "all_users" {
  limit  = 100
  sort   = "created:-1"
}

output "user_list" {
  value = [
    for user in data.appmixer_users.all_users.users : {
      id       = user.id
      username = user.username
      email    = user.email
    }
  ]
}
```

### Getting User Count (Admin Required)

```hcl
data "appmixer_users_count" "count" {}

output "total_users" {
  value = data.appmixer_users_count.count.total
}
```

## Security Considerations

1. Always use environment variables or a secure secret management solution for storing credentials
2. Ensure that admin credentials are properly secured
3. Remember that certain operations require admin permissions (listing users, counting users, setting scopes/vendors, resetting other users' passwords, deleting users)
4. Consider using dedicated service accounts with appropriate permissions for Terraform operations
5. Be aware that updating the password for the currently authenticated user via Terraform is not supported. 