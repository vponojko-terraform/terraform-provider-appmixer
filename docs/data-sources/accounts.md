`appmixer_accounts` Data Source
===============================

Provides a list of Appmixer accounts, optionally filtered by service or other criteria.

Use this data source to retrieve multiple accounts based on the provided filter.

Example Usage
-------------

### Get all accounts

```hcl
data "appmixer_accounts" "all" {}

output "all_account_ids" {
  value = [for acc in data.appmixer_accounts.all.accounts : acc.account_id]
}
```

### Get all AWS accounts

```hcl
data "appmixer_accounts" "aws_accounts" {
  filter = "service:appmixer:aws"
}

output "aws_account_names" {
  value = [for acc in data.appmixer_accounts.aws_accounts.accounts : acc.name]
}
```

### Get all accounts *except* AWS and Acme accounts

```hcl
data "appmixer_accounts" "non_aws_acme" {
  # Note: The API uses '!' for negation and requires separate filter parameters for multiple conditions.
  # Terraform provider currently supports a single filter string.
  # For complex filtering, you might need multiple data sources or post-processing.
  # This example shows filtering out AWS, assuming the API supports multiple filter params
  # or a more complex syntax that the provider could eventually pass through.
  # The current implementation passes the filter string as-is.
  filter = "service:!appmixer:aws&service:!appmixer:acme" # Check Appmixer API docs for exact filter syntax
}

output "other_account_count" {
  value = length(data.appmixer_accounts.non_aws_acme.accounts)
}
```

Argument Reference
------------------

*   `filter` - (Optional, String) Filter accounts based on Appmixer API filter syntax. Refer to the [Appmixer API documentation](https://docs.appmixer.com/6.0/6.1/api/accounts#get-all-accounts) for supported filter options (e.g., `service:appmixer:slack`, `service:!appmixer:aws`).

Attribute Reference
-------------------

*   `id` - A unique ID for the data source itself, derived from the filter and result count.
*   `accounts` - (List of Objects) A list of accounts matching the filter criteria. Each object in the list has the following attributes:
    *   `account_id` - (String) The unique ID assigned to the account by Appmixer.
    *   `service` - (String) The service identifier for the account.
    *   `display_name` - (String) The user-friendly name assigned to the account.
    *   `name` - (String) The primary identifier/name returned by the API.
    *   `profile_info` - (Map of String) Additional profile details returned by the API.
    *   `icon` - (String) Base64 encoded icon for the service.
    *   `label` - (String) The user-friendly label for the service type.
    *   `user_id` - (String) The Appmixer user ID associated with this account. 