---
subcategory: "Lockbox (Secret Management)"
page_title: "Yandex: yandex_lockbox_secret_iam_member"
description: |-
  Allows management of a single member for a single IAM binding for a Lockbox Secret.
---

# yandex_lockbox_secret_iam_member (Resource)

Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex Lockbox Secret.

~> Roles controlled by `yandex_lockbox_secret_iam_binding` should not be assigned using `yandex_lockbox_secret_iam_member`.

## Example usage

```terraform
//
// Create a new Lockbox Secret and new IAM Member for it.
//
resource "yandex_lockbox_secret" "your-secret" {
  name = "secret-name"
}

resource "yandex_lockbox_secret_iam_member" "viewer" {
  secret_id = yandex_lockbox_secret.your-secret.id
  role      = "viewer"

  member = "userAccount:foo_user_id"
}
```

## Argument Reference

The following arguments are supported:

* `secret_id` - (Required) The [Yandex Lockbox Secret](https://yandex.cloud/docs/lockbox/) Secret ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://yandex.cloud/docs/lockbox/security/).

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field. This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://yandex.cloud/docs/iam/concepts/access-control/system-group).


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_lockbox_secret_iam_member.<resource Name> "<resource Id> <resource Role> <subject Id>"
terraform import yandex_lockbox_secret_iam_member.viewer "abjjf**********p3gp8 viewer foo@example.com"
```
