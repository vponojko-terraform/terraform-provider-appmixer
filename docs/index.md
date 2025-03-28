# Appmixer Provider

The Appmixer provider allows you to interact with the Appmixer API to manage users and their properties. This provider is suitable for Appmixer self-hosted instances and cloud environments.

## Example Usage

```hcl
terraform {
  required_providers {
    appmixer = {
      source = "vponojko-terraform/appmixer"
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

## Authentication

The Appmixer provider requires an API URL, email, and password to authenticate with the Appmixer API.

### Environment Variables

You can provide your credentials via the `APPMIXER_API_URL`, `APPMIXER_EMAIL`, and `APPMIXER_PASSWORD` environment variables.

```hcl
provider "appmixer" {}
```

```sh
export APPMIXER_API_URL="https://api.your-tenant.appmixer.cloud"
export APPMIXER_EMAIL="admin@example.com"
export APPMIXER_PASSWORD="your-secure-password"
```

## Argument Reference

* `api_url` - (Required) The URL of the Appmixer API. This can also be provided via the `APPMIXER_API_URL` environment variable.
* `email` - (Required) The email used for authentication. This can also be provided via the `APPMIXER_EMAIL` environment variable.
* `password` - (Required) The password used for authentication. This can also be provided via the `APPMIXER_PASSWORD` environment variable.

## Security Notes

* Admin scope is required for certain operations:
  * Listing all users
  * Getting user count
  * Modifying user permissions
  * Creating users with specific permissions or vendor settings
* Users cannot modify their own permissions or update their own passwords through this provider 