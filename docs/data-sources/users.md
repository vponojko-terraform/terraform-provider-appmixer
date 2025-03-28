# Users Data Source

The `appmixer_users` data source allows you to retrieve a list of users from Appmixer. This data source requires admin permissions.

## Example Usage

```hcl
data "appmixer_users" "all" {
  limit   = 50
  offset  = 0
  sort    = "created:-1"
  filter  = "scope:admin"
  pattern = "john"
}

output "admin_users" {
  value = data.appmixer_users.all.users
}
```

## Argument Reference

* `filter` - (Optional) Filter users by specific criteria (e.g., 'scope:admin').
* `pattern` - (Optional) Filter users by pattern in username.
* `sort` - (Optional) Sort users by a specific field and order (e.g., 'created:-1'). Defaults to 'created:-1'.
* `limit` - (Optional) Limit the number of users returned. Defaults to 30.
* `offset` - (Optional) Offset for pagination. Defaults to 0.

## Attribute Reference

* `users` - A list of user objects with the following attributes:
  * `id` - The unique identifier for the user.
  * `username` - The username of the user.
  * `email` - The email address of the user.
  * `is_active` - Whether the user account is active.
  * `plan` - The plan information for the user.
  * `scope` - The list of scope permissions the user has.
  * `created` - The timestamp when the user was created.

## Security Notes

This data source requires the authenticating user to have admin permissions. 