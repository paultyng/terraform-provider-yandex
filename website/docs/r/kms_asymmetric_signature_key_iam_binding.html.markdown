---
layout: "yandex"
page_title: "Yandex: yandex_kms_asymmetric_signature_key_iam_binding"
sidebar_current: "docs-yandex-kms-asymmetric-signature-key-iam-binding"
description: |-
Allows management of a single IAM binding for a [Yandex Key Management Service](https://cloud.yandex.com/docs/kms/).
---

## yandex\_kms\_asymmetric\_signature\_key\_iam\_binding

Allows creation and management of a single binding within IAM policy for
an existing Yandex KMS Asymmetric Signature Key.

## Example Usage

```hcl
resource "yandex_kms_asymmetric_signature_key" "your-key" {
  folder_id = "your-folder-id"
  name      = "asymmetric-signature-key-name"
}

resource "yandex_kms_asymmetric_signature_key_iam_binding" "viewer" {
  asymmetric_signature_key_id = yandex_kms_asymmetric_signature_key.your-key.id
  role             = "viewer"

  members = [
    "userAccount:foo_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `asymmetric_signature_key_id` - (Required) The [Yandex Key Management Service](https://cloud.yandex.com/docs/kms/) Asymmetric Signature Key ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/kms/security/).

* `members` - (Required) Identities that will be granted the privilege in `role`.
  Each entry can have one of the following values:
    * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
    * **serviceAccount:{service_account_id}**: A unique service account ID.
    * **system:{allUsers|allAuthenticatedUsers}**: see [system groups](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group)

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role.
These bindings can be imported using the `asymmetric_signature_key_id` and role, e.g.

```
$ terraform import yandex_kms_asymmetric_signature_key_iam_binding.viewer "asymmetric_signature_key_id viewer"
```
