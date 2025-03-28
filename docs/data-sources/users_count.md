# Users Count Data Source

The `appmixer_users_count` data source allows you to retrieve the total count of users in Appmixer. This data source requires admin permissions.

## Example Usage

```hcl
data "appmixer_users_count" "count" {}

output "total_users" {
  value = data.appmixer_users_count.count.total
}
```

## Argument Reference

This data source doesn't require any configuration.

## Attribute Reference

* `total` - The total number of users in the Appmixer instance.

## Security Notes

This data source requires the authenticating user to have admin permissions. 