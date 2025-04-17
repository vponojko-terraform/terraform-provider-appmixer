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
* `password_force_update` - (Optional, Boolean) Defaults to `false`. Set to `true` to explicitly trigger a password reset API call on the next `terraform apply` for this user (requires admin permissions for other users). This flag will be automatically reset to `false` in the Terraform state after a successful forced update.
* `scope` - (Optional, Computed) List of scope permissions for the user. Requires admin permissions to set. Common values include `["user"]` and `["user", "admin"]`.
* `vendor` - (Optional, Computed) List of vendor associations for the user. Requires admin permissions to set.

-> **Note:** Setting `scope` or `vendor` attributes requires the authenticating user to have admin permissions.

## Password Update Logic

A password reset API call for a user **other than the currently authenticated user** is triggered only under the following conditions:

1.  The `password_force_update` attribute is set to `true` in the configuration.
2.  **OR** the `password` attribute in the configuration is changed **AND** a password value was previously stored in the Terraform state for this resource (i.e., it wasn't the initial creation or the state value wasn't empty).

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