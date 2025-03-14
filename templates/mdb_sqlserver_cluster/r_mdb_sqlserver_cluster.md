---
subcategory: "Managed Service for SQLServer"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Microsoft SQLServer cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a SQLServer cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-sqlserver/).

Please read [Pricing for Managed Service for SQL Server](https://yandex.cloud/docs/managed-sqlserver/pricing#prices) before using SQLServer cluster.

## Example usage

{{ tffile "examples/mdb_sqlserver_cluster/r_mdb_sqlserver_cluster_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the SQLServer cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the SQLServer cluster uses.

* `environment` - (Required) Deployment environment of the SQLServer cluster. (PRODUCTION, PRESTABLE)

* `version` - (Required) Version of the SQLServer cluster. (2016sp2std, 2016sp2ent)

* `resources` - (Required) Resources allocated to hosts of the SQLServer cluster. The structure is documented below.

* `user` - (Required) A user of the SQLServer cluster. The structure is documented below.

* `database` - (Required) A database of the SQLServer cluster. The structure is documented below.

* `host` - (Required) A host of the SQLServer cluster. The structure is documented below.

* `sqlserver_config` - (Optional) SQLServer cluster config. Detail info in "SQLServer config" section (documented below).

---

* `description` - (Optional) Description of the SQLServer cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the SQLServer cluster.

* `backup_window_start` - (Optional) Time to start the daily backup, in the UTC. The structure is documented below.

* `security_group_ids` - (Optional) A set of ids of security groups assigned to hosts of the cluster.

* `deletion_protection` - (Optional) Inhibits deletion of the cluster. Can be either `true` or `false`.

* `host_group_ids` - (Optional) A list of IDs of the host groups hosting VMs of the cluster.

* `sqlcollation` - (Optional) SQL Collation cluster will be created with. This attribute cannot be changed when cluster is created!

---

The `resources` block supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a SQLServer host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-sqlserver/concepts/instance-types).

* `disk_size` - (Required) Volume of the storage available to a SQLServer host, in gigabytes.

* `disk_type_id` - (Required) Type of the storage of SQLServer hosts.

The `backup_window_start` block supports:

* `hours` - (Optional) The hour at which backup will be started.

* `minutes` - (Optional) The minute at which backup will be started.

The `user` block supports:

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `database_name` - (Required) The name of the database that the permission grants access to.

* `roles` - (Optional) List user's roles in the database. Allowed roles: `OWNER`, `SECURITYADMIN`, `ACCESSADMIN`, `BACKUPOPERATOR`, `DDLADMIN`, `DATAWRITER`, `DATAREADER`, `DENYDATAWRITER`, `DENYDATAREADER`.

The `database` block supports:

* `name` - (Required) The name of the database.

The `host` block supports:

* `fqdn` - (Computed) The fully qualified domain name of the host.

* `zone` - (Required) The availability zone where the SQLServer host will be created.

* `subnet_id` - (Optional) The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.

* `assign_public_ip` - (Optional) Sets whether the host should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the cluster.

* `health` - Aggregated health of the cluster.

* `status` - Status of the cluster.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/mdb_sqlserver_cluster/import.sh" }}


## SQLServer config

If not specified `sqlserver_config` then does not make any changes.

* max_degree_of_parallelism - Limits the number of processors to use in parallel plan execution per task. See in-depth description in [SQL Server documentation](https://docs.microsoft.com/en-us/sql/database-engine/configure-windows/configure-the-max-degree-of-parallelism-server-configuration-option?view=sql-server-2016).

* cost_threshold_for_parallelism - Specifies the threshold at which SQL Server creates and runs parallel plans for queries. SQL Server creates and runs a parallel plan for a query only when the estimated cost to run a serial plan for the same query is higher than the value of the option. See in-depth description in [SQL Server documentation](https://docs.microsoft.com/en-us/sql/database-engine/configure-windows/configure-the-cost-threshold-for-parallelism-server-configuration-option?view=sql-server-2016).

* audit_level - Describes how to configure login auditing to monitor SQL Server Database Engine login activity. Possible values:
  - 0 — do not log login attempts,˚√
  - 1 — log only failed login attempts,
  - 2 — log only successful login attempts (not recommended),
  - 3 — log all login attempts (not recommended). See in-depth description in [SQL Server documentation](https://docs.microsoft.com/en-us/sql/ssms/configure-login-auditing-sql-server-management-studio?view=sql-server-2016).

* fill_factor_percent - Manages the fill factor server configuration option. When an index is created or rebuilt the fill factor determines the percentage of space on each index leaf-level page to be filled with data, reserving the rest as free space for future growth. Values 0 and 100 mean full page usage (no space reserved). See in-depth description in [SQL Server documentation](https://docs.microsoft.com/en-us/sql/database-engine/configure-windows/configure-the-fill-factor-server-configuration-option?view=sql-server-2016).
* optimize_for_ad_hoc_workloads - Determines whether plans should be cached only after second execution. Allows to avoid SQL cache bloat because of single-use plans. See in-depth description in [SQL Server documentation](https://docs.microsoft.com/en-us/sql/database-engine/configure-windows/optimize-for-ad-hoc-workloads-server-configuration-option?view=sql-server-2016).
