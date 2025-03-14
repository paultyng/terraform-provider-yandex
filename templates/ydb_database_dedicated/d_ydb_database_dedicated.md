---
subcategory: "Managed Service for YDB"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Database dedicated cluster.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Database (dedicated) cluster. For more information, see [the official documentation](https://yandex.cloud/docs/ydb/concepts/serverless_and_dedicated).

## Example usage

{{ tffile "examples/ydb_database_dedicated/d_ydb_database_dedicated_1.tf" }}

## Argument Reference

The following arguments are supported:

* `database_id` - (Optional) ID of the Yandex Database cluster.

* `name` - (Optional) Name of the Yandex Database cluster.

* `folder_id` - (Optional) ID of the folder that the Yandex Database cluster belongs to. It will be deduced from provider configuration if not set explicitly.

~> If `database_id` is not specified `name` and `folder_id` will be used to designate Yandex Database cluster.

## Attributes Reference

* `network_id` - ID of the network the Yandex Database cluster is attached to.

* `subnet_ids` - List of subnet IDs the Yandex Database cluster is attached to.

* `resource_preset_id` - The Yandex Database cluster preset.

* `scale_policy` - Scaling policy of the Yandex Database cluster. The structure is documented below.

* `storage_config` - A list of storage configuration options of the Yandex Database cluster. The structure is documented below.

* `location` - Location of the Yandex Database cluster. The structure is documented below.

* `location_id` - Location ID of the Yandex Database cluster.

* `assign_public_ips` - Whether public IP addresses are assigned to the Yandex Database cluster.

* `description` - A description of the Yandex Database cluster.

* `labels` - A set of key/value label pairs assigned to the Yandex Database cluster.

* `ydb_full_endpoint` - Full endpoint of the Yandex Database cluster.

* `ydb_api_endpoint` - API endpoint of the Yandex Database cluster. Useful for SDK configuration.

* `database_path` - Full database path of the Yandex Database cluster. Useful for SDK configuration.

* `tls_enabled` - Whether TLS is enabled for the Yandex Database cluster. Useful for SDK configuration.

* `persistence_mode` - Persistence mode of the Yandex Database cluster. Useful for SDK configuration.

* `announce_hostnames` - Announce fqdn instead of ip address for the Yandex Database cluster. Useful for SDK configuration.

* `created_at` - The Yandex Database cluster creation timestamp.

* `status` - Status of the Yandex Database cluster.

* `deletion_protection` - Inhibits deletion of the database. Can be either `true` or `false`

---

The `scale_policy` block supports:

* `fixed_scale` - Fixed scaling policy of the Yandex Database cluster. The structure is documented below.

~> Currently, only `fixed_scale` is supported.

---

The `fixed_scale` block supports:

* `size` - Number of instances in the Yandex Database cluster.

---

The `storage_config` block supports:

* `storage_type_id` - Storage type ID of the Yandex Database cluster.

* `group_count` - Amount of storage groups of selected type in the Yandex Database cluster.

---

The `location` block supports:

* `region` - Region of the Yandex Database cluster. The structure is documented below.

---

The `region` block supports:

* `id` - Region ID of the Yandex Database cluster.
