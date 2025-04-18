`appmixer_account` Resource
============================

Manages an Appmixer account.

Accounts represent authenticated connections to third-party services (like AWS, Slack, etc.) or custom services within Appmixer.

This resource allows you to create, update, and delete accounts using API keys or username/password combinations.

**Note:** Currently, only API Key (`appmixer:aws` style) and PWD (`appmixer:acme` style) account types can be created directly via this resource due to the nature of their non-interactive authentication. OAuth2 based accounts typically require user interaction and cannot be fully managed through Terraform creation, but existing OAuth2 accounts can be read and potentially have their display name updated.  

[accounts API documentation](https://docs.appmixer.com/api/accounts)

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

Troubleshooting
---------------

### Error: "failed to create account... API request failed with status 500: The AWS Access Key Id you provided does not exist in our records."

This error indicates that when Appmixer attempted to validate the AWS credentials provided in the `token` attribute during the creation request, AWS reported that the Access Key ID was invalid or unknown.

1.  **Verify Credentials:** Double-check that the `accessKeyId` and `secretKey` values provided in the `token` map are correct, active, and belong to the intended IAM user in AWS.
2.  **IAM Propagation Delay (Race Condition):** If you are creating the AWS IAM Access Key (e.g., using the `aws_iam_access_key` resource) in the **same Terraform configuration** as this `appmixer_account` resource, there might be a short delay (seconds to potentially a minute) before the new key is fully active and recognized across all AWS APIs. Appmixer's validation check might occur during this brief window.

    **Workaround:** To mitigate this race condition, you can introduce an explicit delay between the creation of the AWS credentials and the creation of the `appmixer_account` resource using the `time_sleep` resource from the `hashicorp/time` provider.

    ```hcl
    # Example: Creating AWS Key and Appmixer Account with delay

    resource "aws_iam_user" "example" {
      name = "appmixer-example-user"
    }

    resource "aws_iam_access_key" "example" {
      user = aws_iam_user.example.name
    }

    # Introduce a delay after the access key is created
    resource "time_sleep" "wait_for_aws_key" {
      create_duration = "30s" # Adjust duration as needed (e.g., 30s, 60s)

      # Ensure this sleep happens only after the key is created
      depends_on = [
        aws_iam_access_key.example
      ]
    }

    # Create the Appmixer account, depending on the sleep
    resource "appmixer_account" "aws_example" {
      service = "appmixer:aws"
      token = {
        accessKeyId = aws_iam_access_key.example.id
        secretKey   = aws_iam_access_key.example.secret
      }
      display_name = "Example AWS Account (Delayed)"

      # Ensure this resource waits for the sleep to complete
      depends_on = [
        time_sleep.wait_for_aws_key
      ]
    }
    ```

    *Requires configuration for the `hashicorp/time` provider.* 