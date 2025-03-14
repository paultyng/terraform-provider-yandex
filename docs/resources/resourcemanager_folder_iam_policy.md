---
subcategory: "Resource Manager"
page_title: "Yandex: yandex_resourcemanager_folder_iam_policy"
description: |-
  Allows management of the IAM policy for a Yandex Resource Manager folder.
---

# yandex_resourcemanager_folder_iam_policy (Resource)

Allows creation and management of the IAM policy for an existing Yandex Resource Manager folder.

## Example usage

```terraform
//
// Create a new IAM Policy for existing Folder.
//
data "yandex_resourcemanager_folder" "project1" {
  folder_id = "my_folder_id"
}

data "yandex_iam_policy" "admin" {
  binding {
    role = "editor"

    members = [
      "userAccount:some_user_id",
    ]
  }
}

resource "yandex_resourcemanager_folder_iam_policy" "folder_admin_policy" {
  folder_id   = data.yandex_folder.project1.id
  policy_data = data.yandex_iam_policy.admin.policy_data
}
```

## Argument Reference

The following arguments are supported:

* `folder_id` - (Required) ID of the folder that the policy is attached to.
* `policy_data` - (Required) The `yandex_iam_policy` data source that represents the IAM policy that will be applied to the folder. This policy overrides any existing policy applied to the folder.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_resourcemanager_folder_iam_policy.<resource Name> <resource Id>
terraform import yandex_resourcemanager_folder_iam_policy.folder_admin_policy ...
```
