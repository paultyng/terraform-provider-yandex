---
subcategory: "Managed Service for MongoDB"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a MongoDB cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a MongoDB cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts).

## Example usage

{{ tffile "examples/mdb_mongodb_cluster/r_mdb_mongodb_cluster_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the MongoDB cluster. Provided by the client when the cluster is created.

* `network_id` - (Required, ForceNew) ID of the network, to which the MongoDB cluster belongs.

* `environment` - (Required, ForceNew) Deployment environment of the MongoDB cluster. Can be either `PRESTABLE` or `PRODUCTION`.

* `cluster_config` - (Required) Configuration of the MongoDB subcluster. The structure is documented below.

* `user` - (Optional) A user of the MongoDB cluster. The structure is documented below.

* `database` - (Optional) A database of the MongoDB cluster. The structure is documented below.

* `host` - (Required) A host of the MongoDB cluster. The structure is documented below.

* `resources_mongod` - (Optional) Resources allocated to `mongod` hosts of the MongoDB cluster. The structure is documented below.

* `resources_mongocfg` - (Optional) Resources allocated to `mongocfg` hosts of the MongoDB cluster. The structure is documented below.

* `resources_mongos` - (Optional) Resources allocated to `mongos` hosts of the MongoDB cluster. The structure is documented below.

* `resources_mongoinfra` - (Optional) Resources allocated to `mongoinfra` hosts of the MongoDB cluster. The structure is documented below.

* `resources` - (**DEPRECATED**, use `resources_*` instead) Resources allocated to hosts of the MongoDB cluster. The structure is documented below.

---

* `description` - (Optional) Description of the MongoDB cluster.

* `labels` - (Optional) A set of key/value label pairs to assign to the MongoDB cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `security_group_ids` - (Optional) A set of ids of security groups assigned to hosts of the cluster.

* `maintenance_window` - (Optional) Maintenance window settings of the MongoDB cluster. The structure is documented below.

* `deletion_protection` - (Optional) Inhibits deletion of the cluster. Can be either `true` or `false`.

---

* `restore` - (Optional, ForceNew) The cluster will be created from the specified backup. The structure is documented below.

---

The `cluster_config` block supports:

* `version` - (Required) Version of the MongoDB server software. Can be either `4.2`, `4.4`, `4.4-enterprise`, `5.0`, `5.0-enterprise`, `6.0` and `6.0-enterprise`.

* `feature_compatibility_version` - (Optional) Feature compatibility version of MongoDB. If not provided version is taken. Can be either `6.0`, `5.0`, `4.4` and `4.2`.

* `backup_window_start` - (Optional) Time to start the daily backup, in the UTC timezone. The structure is documented below.

* `backup_retain_period_days` - (Optional) Retain period of automatically created backup in days.

* `performance_diagnostics` - (Optional) Performance diagnostics to the MongoDB cluster. The structure is documented below.

* `access` - (Optional) Access policy to the MongoDB cluster. The structure is documented below.

* `mongod` - (Optional) Configuration of the mongod service. The structure is documented below.

* `mongocfg` - (Optional) Configuration of the mongocfg service. The structure is documented below.

* `mongos` - (Optional) Configuration of the mongos service. The structure is documented below.

The `backup_window_start` block supports:

* `hours` - (Optional) The hour at which backup will be started.

* `minutes` - (Optional) The minute at which backup will be started.

The `resources`, `resources_mongod`, `resources_mongos`, `resources_mongocfg`, `resources_mongoinfra`, blocks supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a MongoDB host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts).

* `disk_size` - (Required) Volume of the storage available to a MongoDB host, in gigabytes.

* `disk_type_id` - (Required) Type of the storage of MongoDB hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).

The `disk_size_autoscaling_mongod`, `disk_size_autoscaling_mongos`, `disk_size_autoscaling_mongoinfra`, `disk_size_autoscaling_mongocfg` blocks support:

* `disk_size_limit` - Limit of disk size after autoscaling (GiB).
* `planned_usage_threshold` - Maintenance window autoscaling disk usage (percent).
* `emergency_usage_threshold` - Immediate autoscaling disk usage (percent).

The `user` block supports:

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `database_name` - (Required) The name of the database that the permission grants access to.

* `roles` - (Optional) The roles of the user in this database. For more information see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts/users-and-roles).

The `database` block supports:

* `name` - (Required) The name of the database.

The `host` block supports:

* `name` - (Computed) The fully qualified domain name of the host. Computed on server side.

* `zone_id` - (Required) The availability zone where the MongoDB host will be created. For more information see [the official documentation](https://yandex.cloud/docs/overview/concepts/geo-scope).

* `role` - (Optional) The role of the cluster (either PRIMARY or SECONDARY).

* `health` - (Computed) The health of the host.

* `subnet_id` - (Required) The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.

* `assign_public_ip` -(Optional) Should this host have assigned public IP assigned. Can be either `true` or `false`.

* `shard_name` - (Optional) The name of the shard to which the host belongs. Only for sharded cluster.

* `type` - (Optional) type of mongo daemon which runs on this host (mongod, mongos, mongocfg, mongoinfra). Defaults to mongod.

* `host_parameters` - (Optional) The parameters of mongod host in replicaset.
  - `hidden` - (Optional) Should this host be hidden in replicaset. Can be either `true` of `false`. For more information see [the official documentation](https://www.mongodb.com/docs/current/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.hidden)
  - `priority` - (Optional) A floating point number that indicates the relative likelihood of a replica set member to become the primary. For more information see [the official documentation](https://www.mongodb.com/docs/current/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.priority)
  - `secondary_delay_secs` - (Optional) The number of seconds "behind" the primary that this replica set member should "lag". For more information see [the official documentation](https://www.mongodb.com/docs/current/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.secondaryDelaySecs)
  - `tags` - (Optional) A set of key/value pairs to assign for the replica set member. For more information see [the official documentation](https://www.mongodb.com/docs/current/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.tags)

The `performance_diagnostics` block supports:

* `enabled` - (Optional) Enable or disable performance diagnostics.

The `access` block supports:

* `data_lens` - (Optional) Allow access for [Yandex DataLens](https://yandex.cloud/services/datalens)
* `data_transfer` - (Optional) Allow access for [DataTransfer](https://yandex.cloud/services/data-transfer)
* `web_sql` - (Optional) Allow access for [WebSQL](https://yandex.cloud/ru/docs/websql/)

The `maintenance_window` block supports:

* `type` - (Required) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.
* `hour` - (Optional) Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.
* `day` - (Optional) Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.

The `mongod` block supports:

* `security` - (Optional) A set of MongoDB Security settings (see the [security](https://www.mongodb.com/docs/manual/reference/configuration-options/#security-options) option). The structure is documented below. Available only in enterprise edition.

* `audit_log` - (Optional) A set of audit log settings (see the [auditLog](https://www.mongodb.com/docs/manual/reference/configuration-options/#auditlog-options) option). The structure is documented below. Available only in enterprise edition.

* `set_parameter` - (Optional) A set of MongoDB Server Parameters (see the [setParameter](https://www.mongodb.com/docs/manual/reference/configuration-options/#setparameter-option) option). The structure is documented below.

* `operation_profiling` - (Optional) A set of profiling settings (see the [operationProfiling](https://www.mongodb.com/docs/manual/reference/configuration-options/#operationprofiling-options) option). The structure is documented below.

* `net` - (Optional) A set of network settings (see the [net](https://www.mongodb.com/docs/manual/reference/configuration-options/#net-options) option). The structure is documented below.

* `storage` - (Optional) A set of storage settings (see the [storage](https://www.mongodb.com/docs/manual/reference/configuration-options/#storage-options) option). The structure is documented below.

The `mongocfg` block supports:

* `operation_profiling` - (Optional) A set of profiling settings (see the [operationProfiling](https://www.mongodb.com/docs/manual/reference/configuration-options/#operationprofiling-options) option). The structure is documented below.

* `net` - (Optional) A set of network settings (see the [net](https://www.mongodb.com/docs/manual/reference/configuration-options/#net-options) option). The structure is documented below.

* `storage` - (Optional) A set of storage settings (see the [storage](https://www.mongodb.com/docs/manual/reference/configuration-options/#storage-options) option). The structure is documented below.

The `mongos` block supports:

* `net` - (Optional) A set of network settings (see the [net](https://www.mongodb.com/docs/manual/reference/configuration-options/#net-options) option). The structure is documented below.

The `security` block supports:

* `enable_encryption` - (Optional) Enables the encryption for the WiredTiger storage engine. Can be either true or false. For more information see [security.enableEncryption](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.enableEncryption) description in the official documentation. Available only in enterprise edition.

* `kmip` - (Optional) Configuration of the third party key management appliance via the Key Management Interoperability Protocol (KMIP) (see [Encryption tutorial](https://www.mongodb.com/docs/rapid/tutorial/configure-encryption) ). Requires `enable_encryption` to be true. The structure is documented below. Available only in enterprise edition.

The `audit_log` block supports:

* `filter` - (Optional) Configuration of the audit log filter in JSON format. For more information see [auditLog.filter](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-auditLog.filter) description in the official documentation. Available only in enterprise edition.

* `runtime_configuration` - (Optional) Specifies if a node allows runtime configuration of audit filters and the auditAuthorizationSuccess variable. For more information see [auditLog.runtimeConfiguration](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-auditLog.runtimeConfiguration) description in the official documentation. Available only in enterprise edition.

The `set_parameter` block supports:

* `audit_authorization_success` - (Optional) Enables the auditing of authorization successes. Can be either true or false. For more information, see the [auditAuthorizationSuccess](https://www.mongodb.com/docs/manual/reference/parameters/#mongodb-parameter-param.auditAuthorizationSuccess) description in the official documentation. Available only in enterprise edition.

* `enable_flow_control` - (Optional) Enables the flow control. Can be either true or false. For more information, see the [enableFlowControl](https://www.mongodb.com/docs/rapid/reference/parameters/#mongodb-parameter-param.enableFlowControl) description in the official documentation.

* `min_snapshot_history_window_in_seconds` - (Optional) The minimum time window in seconds for which the storage engine keeps the snapshot history. For more information, see the [minSnapshotHistoryWindowInSeconds](https://www.mongodb.com/docs/manual/reference/parameters/#mongodb-parameter-param.minSnapshotHistoryWindowInSeconds) description in the official documentation.

The `operation_profiling` block supports:

* `mode` - (Optional) Specifies which operations should be profiled. The following profiler levels are available: off, slow_op, all. For more information, see the [operationProfiling.mode](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-operationProfiling.mode) description in the official documentation.

* `slow_op_threshold` - (Optional) The slow operation time threshold, in milliseconds. Operations that run for longer than this threshold are considered slow. For more information, see the [operationProfiling.slowOpThresholdMs](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-operationProfiling.slowOpThresholdMs) description in the official documentation.

* `slow_op_sample_rate` - (Optional) The fraction of slow operations that should be profiled or logged. Accepts values between 0 and 1, inclusive. For more information, see the [operationProfiling.slowOpSampleRate](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-operationProfiling.slowOpSampleRate) description in the official documentation.

The `net` block supports:

* `max_incoming_connections` - (Optional) The maximum number of simultaneous connections that host will accept. For more information, see the [net.maxIncomingConnections](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-net.maxIncomingConnections) description in the official documentation.

* `compressors` - (Optional) Specifies the default compressor(s) to use for communication between this mongod or mongos. Accepts array of compressors. Order matters. Available compressors: snappy, zlib, zstd, disabled. To disable network compression, make "disabled" the only value. For more information, see the [net.Compression.Compressors](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-net.compression.compressors) description in the official documentation.

The `storage` block supports:

* `wired_tiger` - (Optional) The WiredTiger engine settings. (see the [storage.wiredTiger](https://www.mongodb.com/docs/manual/reference/configuration-options/#storage.wiredtiger-options) option). These settings available only on `mongod` hosts. The structure is documented below.

* `journal` - (Optional) The durability journal to ensure data files remain valid and recoverable. The structure is documented below.

The `wired_tiger` block supports:

* `cache_size_gb` - (Optional) Defines the maximum size of the internal cache that WiredTiger will use for all data. For more information, see the [storage.wiredTiger.engineConfig.cacheSizeGB](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-storage.wiredTiger.engineConfig.cacheSizeGB) description in the official documentation.

* `block_compressor` - (Optional) Specifies the default compression for collection data. You can override this on a per-collection basis when creating collections. Available compressors are: none, snappy, zlib, zstd. This setting available only on `mongod` hosts. For more information, see the [storage.wiredTiger.collectionConfig.blockCompressor](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-storage.wiredTiger.collectionConfig.blockCompressor) description in the official documentation.

* `prefix_compression` - (Optional) Enables or disables prefix compression for index data. Сan be either true or false. For more information, see the [storage.wiredTiger.indexConfig.prefixCompression](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-storage.wiredTiger.indexConfig.prefixCompression) description in the official documentation.

The `journal` block supports:

* `commit_interval` - (Optional) The maximum amount of time in milliseconds that the mongod process allows between journal operations. For more information, see the [storage.journal.commitIntervalMs](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-storage.journal.commitIntervalMs) description in the official documentation.

The `kmip` block supports:

* `server_name` - (Required) Hostname or IP address of the KMIP server to connect to. For more information see [security.kmip.serverName](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.serverName) description in the official documentation.

* `port` - (Optional) Port number to use to communicate with the KMIP server. Default: 5696 For more information see [security.kmip.port](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.port) description in the official documentation.

* `server_ca` - (Required) Path to CA File. Used for validating secure client connection to KMIP server. For more information see [security.kmip.serverCAFile](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.serverCAFile) description in the official documentation.

* `client_certificate` - (Required) String containing the client certificate used for authenticating MongoDB to the KMIP server. For more information see [security.kmip.clientCertificateFile](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.clientCertificateFile) description in the official documentation.

* `key_identifier` - (Optional) Unique KMIP identifier for an existing key within the KMIP server. For more information see [security.kmip.keyIdentifier](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.keyIdentifier) description in the official documentation.

The `restore` block supports:

* `backup_id` - (Required, ForceNew) Backup ID. The cluster will be created from the specified backup. [How to get a list of PostgreSQL backups](https://yandex.cloud/docs/managed-mongodb/operations/cluster-backups)

* `time` - (Optional, ForceNew) Timestamp of the moment to which the MongoDB cluster should be restored. (Format: "2006-01-02T15:04:05" - UTC). When not set, current time is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the key.

* `health` - Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`. For more information see `health` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-mongodb/api-ref/Cluster/).

* `status` - Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`. For more information see `status` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-mongodb/api-ref/Cluster/).

* `cluster_id` - The ID of the cluster.

* `sharded` - MongoDB Cluster mode enabled/disabled.

## Timeouts

`yandex_audit_trails_trail` provides the following configuration options for [timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default 30 minutes.
- `update` - Default 60 minutes.
- `delete` - Default 30 minutes.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/mdb_mongodb_cluster/import.sh" }}
