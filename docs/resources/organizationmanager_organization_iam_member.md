---
subcategory: "Cloud Organization"
page_title: "Yandex: yandex_organizationmanager_organization_iam_member"
description: |-
  Allows management of a single member for a single IAM binding on a Yandex Cloud Organization Manager organization.
---

# yandex_organizationmanager_organization_iam_member (Resource)

Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex Organization Manager organization.

~> Roles controlled by `yandex_organizationmanager_organization_iam_binding` should not be assigned using `yandex_organizationmanager_organization_iam_member`.

~> When you delete `yandex_organizationmanager_organization_iam_binding` resource, the roles can be deleted from other users within the organization as well. Be careful!

## Example usage

```terraform
//
// Create a new OrganizationManager Organization IAM Member.
//
resource "yandex_organizationmanager_organization_iam_member" "editor" {
  organization_id = "some_organization_id"
  role            = "editor"
  member          = "userAccount:user_id"
}
```

## Argument Reference

The following arguments are supported:

* `organization_id` - (Required) ID of the organization to attach a policy to.

* `role` - (Required) The role that should be assigned.

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field. This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
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
# terraform import yandex_organizationmanager_organization_iam_member.<resource Name> "<resource Id> <resource Role> <subject>"
terraform import yandex_organizationmanager_organization_iam_member.editor "abjjf**********p3gp8 editor foo@example.com"
```
