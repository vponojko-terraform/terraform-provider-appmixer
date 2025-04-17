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
* `password` - (Required, Sensitive) The password for the user. Changing this attribute will trigger a password update action. See notes below.
* `scope` - (Optional, Computed) List of scope permissions for the user. Requires admin permissions to set. Common values include `["user"]` and `["user", "admin"]`.
* `vendor` - (Optional, Computed) List of vendor associations for the user. Requires admin permissions to set.

-> **Note:** Setting `scope` or `vendor` attributes requires the authenticating user to have admin permissions.

-> **Password Update Note:** 
> - Changing the `password` for a user **other than the authenticated user** triggers a password reset via the admin API endpoint and requires the provider's authenticated user to have **admin permissions**.
> - Changing the `password` for the **currently authenticated user** is **not supported** via Terraform due to API limitations (the old password is required but not available). Use the Appmixer UI or API directly for self-password changes.

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