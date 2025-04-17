# User Resource

The `appmixer_user` resource allows you to create and manage users in Appmixer.

## Example Usage

```hcl
resource "appmixer_user" "example" {
  username = "new-user@example.com"
  email    = "new-user@example.com"
  password = "secure-password"
  
  # Admin only fields
  scope  = ["user", "admin"]
  vendor = ["vendor1"]
}
```

## Argument Reference

* `username` - (Required, ForceNew) The username for the new user. This cannot be changed after creation.
* `email` - (Required) The email address for the user.
* `password` - (Required, Sensitive) The password for the user. See the **Password Update Logic** section below for details on when changes trigger an update.
* `scope` - (Optional, Computed) List of scope permissions for the user. Requires admin permissions to set. Common values include `["user"]` and `["user", "admin"]`.
* `vendor` - (Optional, Computed) List of vendor associations for the user. Requires admin permissions to set.

-> **Note:** Setting `scope` or `vendor` attributes requires the authenticating user to have admin permissions.

## Password Update Logic

A password reset API call for a user **other than the currently authenticated user** is triggered only if:

1.  The `lifecycle { ignore_changes = [password] }` block is **NOT** present in the resource configuration.
2.  **AND** Terraform detects a difference between the `password` value in the configuration and the value stored in the Terraform state.

**Important Considerations:**

*   Triggering a password update (reset) for another user always requires the provider's authenticated user to have **admin permissions**.
*   Changing the `password` for the **currently authenticated user** via Terraform is **not supported** due to API limitations (the old password is required but not available). Use the Appmixer UI or API directly for self-password changes.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The unique identifier for the user.
* `is_active` - Whether the user account is active.
* `plan` - The plan information for the user.
* `created` - The timestamp when the user was created.

## Import

User resources can be imported by their ID:

```shell
terraform import appmixer_user.example <user_id>
``` 
User id can be found inside of MongoDB - db.getCollection('users').find({})