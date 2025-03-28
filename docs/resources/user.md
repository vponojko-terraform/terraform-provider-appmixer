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
* `password` - (Required, Sensitive) The password for the user.
* `scope` - (Optional, Computed) List of scope permissions for the user. Requires admin permissions to set. Common values include `["user"]` and `["user", "admin"]`.
* `vendor` - (Optional, Computed) List of vendor associations for the user. Requires admin permissions to set.

-> **Note:** Setting `scope` or `vendor` attributes requires the authenticating user to have admin permissions.

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