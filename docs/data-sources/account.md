`appmixer_account` Data Source
==============================

Provides information about a specific Appmixer account.

Use this data source to retrieve details of an existing account using its ID.

Example Usage
-------------

```hcl
data "appmixer_account" "specific_account" {
  account_id = "5a6e21f3b266224186ac7d03"
}

output "specific_account_service" {
  value = data.appmixer_account.specific_account.service
}

output "specific_account_name" {
  value = data.appmixer_account.specific_account.name
}
```

Argument Reference
------------------

*   `account_id` - (Required, String) The unique ID of the account to retrieve.

Attribute Reference
-------------------

In addition to the arguments above, the following attributes are exported:

*   `id` - The unique ID assigned to the account by Appmixer (same as `account_id`).
*   `service` - (String) The service identifier for the account (e.g., `appmixer:aws`, `appmixer:slack`).
*   `display_name` - (String) The user-friendly name assigned to the account.
*   `name` - (String) The primary identifier/name returned by the API (e.g., username, email, AWS account ID).
*   `profile_info` - (Map of String) Additional profile details returned by the API for the account.
*   `icon` - (String) Base64 encoded icon for the service associated with the account.
*   `label` - (String) The user-friendly label for the service type (e.g., "Slack", "Pipedrive", "AWS").
*   `user_id` - (String) The Appmixer user ID associated with this account. 