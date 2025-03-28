# User Data Source

The `appmixer_user` data source allows you to retrieve information about the currently authenticated user in Appmixer.

## Example Usage

```hcl
data "appmixer_user" "current" {}

output "user_details" {
  value = {
    id       = data.appmixer_user.current.id
    email    = data.appmixer_user.current.email
    username = data.appmixer_user.current.username
    scope    = data.appmixer_user.current.scope
  }
}
```

## Argument Reference

The data source doesn't require any configuration.

## Attribute Reference

* `id` - The unique identifier for the user.
* `username` - The username of the user.
* `email` - The email address of the user.
* `is_active` - Whether the user account is active.
* `plan` - The plan information for the user.
* `scope` - The list of scope permissions the user has.
* `created` - The timestamp when the user was created. 