---
subcategory: "Resource Manager"
page_title: "Yandex: yandex_resourcemanager_cloud_iam_binding"
description: |-
  Allows management of a single IAM binding for a Yandex Resource Manager cloud.
---

# yandex_resourcemanager_cloud_iam_binding (Resource)

Allows creation and management of a single binding within IAM policy for an existing Yandex Resource Manager cloud.

## Example usage

```terraform
//
// Create a new Cloud and new IAM Binding for it.
//
data "yandex_resourcemanager_cloud" "project1" {
  name = "Project 1"
}

resource "yandex_resourcemanager_cloud_iam_binding" "admin" {
  cloud_id = data.yandex_resourcemanager_cloud.project1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `cloud_id` - (Required) ID of the cloud to attach the policy to.

* `role` - (Required) The role that should be assigned. Only one `yandex_resourcemanager_cloud_iam_binding` can be used per role.

* `members` - (Required) An array of identities that will be granted the privilege in the `role`. Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **federatedUser:{federated_user_id}**: A unique federated user ID.
  * **federatedUser:{federated_user_id}:**: A unique saml federation user account ID.
  * **group:{group_id}**: A unique group ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://yandex.cloud/docs/iam/concepts/access-control/system-group).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_resourcemanager_cloud_iam_binding.<resource Name> "<resource Id> <resource Role>"
terraform import yandex_resourcemanager_cloud_iam_binding.admin "b1g3a**********qc5df admin"
```
