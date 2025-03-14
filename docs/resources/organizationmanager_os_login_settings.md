---
subcategory: "Cloud Organization"
page_title: "Yandex: yandex_organizationmanager_os_login_settings"
description: |-
  Allows management of OsLogin Settings within an existing Yandex Cloud Organization.
---

# yandex_organizationmanager_os_login_settings (Resource)

## Example usage

```terraform
//
// Create a new OrganizationManager OS Login Settings.
//
resource "yandex_organizationmanager_os_login_settings" "my_settings" {
  organization_id = "sdf4*********3fr"
  user_ssh_key_settings {
    enabled               = true
    allow_manage_own_keys = true
  }
  ssh_certificate_settings {
    enabled = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `organization_id` - (Required) The organization to manage it's OsLogin Settings.
* `user_ssh_key_settings` - (Optional) The structure is documented below.
* `ssh_certificate_settings` - (Optional) The structure is documented below.

The `user_ssh_key_settings` block supports:
* `enabled` - Enables or disables usage of ssh keys assigned to a specific subject.
* `allow_manage_own_keys` - If set to true subject is allowed to manage own ssh keys without having to be assigned specific permissions.

The `ssh_certificate_settings` block supports:
* `enabled` - Enables or disables usage of ssh certificates signed by trusted CA.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_organizationmanager_os_login_settings.<resource Name> <resource Id>
terraform import yandex_organizationmanager_os_login_settings.my_settings ...
```
