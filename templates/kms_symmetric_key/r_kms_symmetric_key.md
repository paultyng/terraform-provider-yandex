---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Creates a Yandex KMS symmetric key that can be used for cryptographic operation.
---

# {{.Name}} ({{.Type}})

Creates a Yandex KMS symmetric key that can be used for cryptographic operation.

~> When Terraform destroys this key, any data previously encrypted with these key will be irrecoverable. For this reason, it is strongly recommended that you add lifecycle hooks to the resource to prevent accidental destruction.

For more information, see [the official documentation](https://yandex.cloud/docs/kms/concepts/).

## Example usage

{{ tffile "examples/kms_symmetric_key/r_kms_symmetric_key_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the key.

* `description` - (Optional) An optional description of the key.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the key.

* `default_algorithm` - (Optional) Encryption algorithm to be used with a new key version, generated with the next rotation. The default value is `AES_128`.

* `rotation_period` - (Optional) Interval between automatic rotations. To disable automatic rotation, omit this parameter.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - The status of the key.
* `rotated_at` - Last rotation timestamp of the key.
* `created_at` - Creation timestamp of the key.

## Timeouts

`yandex_kms_symmetric_key` provides the following configuration options for [timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default 1 minute
- `update` - Default 1 minute
- `delete` - Default 1 minute


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/kms_symmetric_key/import.sh" }}
