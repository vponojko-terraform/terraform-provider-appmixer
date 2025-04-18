# Appmixer Terraform Provider

⚠️ AI generated, partially verified

Based on 
[Appmixer API documentation](https://docs.appmixer.com/)
version 6.0.  
Tested against [self-hosted version of an Appmixer](https://docs.appmixer.com/appmixer-self-managed/appmixer-deployment-models)

## Requirements
- Terraform >= 0.13.x
- Go >= 1.18

## Setup

```sh
make build
make install
```

## Configuration

```hcl
terraform {
  required_providers {
    appmixer = {
      source  = "vponojko-terraform/appmixer"
      version = "0.1.0"
    }
  }
}

provider "appmixer" {
  api_url  = "https://api.your-tenant.appmixer.cloud"
  email    = "user@example.com"
  password = "password"
}
```

## Environment Variables

```sh
export APPMIXER_API_URL="https://api.your-tenant.appmixer.cloud"
export APPMIXER_EMAIL="user@example.com"
export APPMIXER_PASSWORD="password"
```

## Security Notes
- Admin scope required for: user listing, user count, permission modification, password resets
- Cannot modify own permissions or update own password

## Data Sources

```hcl
# Current User
data "appmixer_user" "current" {}

# List Users (Admin)
data "appmixer_users" "all" {
  limit   = 50
  offset  = 0
  sort    = "created:-1"
  filter  = "scope:admin"
  pattern = "john"
}

# User Count (Admin)
data "appmixer_users_count" "count" {}
```

## Resources

```hcl
# User Resource
resource "appmixer_user" "example" {
  username = "new-user@example.com"
  email    = "new-user@example.com"
  password = "secure-password"
  
  # Admin only
  scope  = ["user", "admin"]
  vendor = ["vendor1"]
}
```

## Quick Test

```bash
# Setup
make build
make install
mkdir -p ~/terraform-test && cd ~/terraform-test

# Create main.tf
cat > main.tf << 'EOF'
terraform {
  required_providers {
    appmixer = {
      source  = "vponojko-terraform/appmixer"
      version = "0.1.0"
    }
  }
}

provider "appmixer" {
  api_url  = "https://api.your-tenant.appmixer.cloud"
  email    = "your-email@vponojko-terraform.com"
  password = "your-password"
}

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
EOF

# Provider installation config
cat > ~/.terraformrc << 'EOF'
provider_installation {
  dev_overrides {
    "vponojko-terraform/appmixer" = "${pwd}/appmixer-terraform-provider"
  }
  direct {}
}
EOF

# Run
terraform init
terraform apply
terraform output current_user
```