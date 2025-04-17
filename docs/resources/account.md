`appmixer_account` Resource
============================

Manages an Appmixer account.

Accounts represent authenticated connections to third-party services (like AWS, Slack, etc.) or custom services within Appmixer.

This resource allows you to create, update, and delete accounts using API keys or username/password combinations.

**Note:** Currently, only API Key (`appmixer:aws` style) and PWD (`appmixer:acme` style) account types can be created directly via this resource due to the nature of their non-interactive authentication. OAuth2 based accounts typically require user interaction and cannot be fully managed through Terraform creation, but existing OAuth2 accounts can be read and potentially have their display name updated.

Example Usage
-------------

### Creating an AWS Account (API Key type)

```hcl
resource "appmixer_account" "aws_main" {
  service = "appmixer:aws"
  token = {
    accessKeyId = "YOUR_AWS_ACCESS_KEY_ID"
    secretKey   = "YOUR_AWS_SECRET_KEY"
  }
  display_name = "Main AWS Account"
}
```

### Creating a Custom Service Account (PWD type)

```hcl
resource "appmixer_account" "custom_crm" {
  service = "appmixer:mycrm"
  token = {
    username = "crm_api_user"
    password = "supersecretpassword"
  }
  display_name = "My Custom CRM Connection"
}
```

Argument Reference
------------------

*   `service` - (Required, String, Forces new resource) The service identifier for the account (e.g., `appmixer:aws`, `appmixer:acme`, `appmixer:slack`). This determines the structure expected in the `token` map.
*   `token` - (Required, Map of String, Forces new resource, Sensitive) A map containing the authentication credentials. Keys depend on the `service` type (e.g., `accessKeyId`, `secretKey` for AWS; `username`, `password` for PWD). Values must be strings.
*   `display_name` - (Optional, String) An optional user-friendly name for the account. This is the only attribute that can be updated after creation.

Attribute Reference
-------------------

In addition to the arguments above, the following computed attributes are exported:

*   `id` - The unique ID assigned to the account by Appmixer.
*   `name` - The primary identifier/name returned by the API (e.g., username, email, AWS account ID, etc., depending on the service).
*   `profile_info` - (Map of String) Additional profile details returned by the API for the account.
*   `icon` - (String) Base64 encoded icon for the service associated with the account.
*   `label` - (String) The user-friendly label for the service type (e.g., "Slack", "Pipedrive", "AWS").
*   `user_id` - (String) The Appmixer user ID associated with this account.

Import
------

Appmixer accounts can be imported using their ID, e.g.

```bash
terraform import appmixer_account.aws_main 5a6e21f3b266224186ac7d03
``` 