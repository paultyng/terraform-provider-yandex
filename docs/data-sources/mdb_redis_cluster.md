---
subcategory: "Managed Service for Redis"
page_title: "Yandex: yandex_mdb_redis_cluster"
description: |-
  Get information about a Yandex Managed Redis cluster.
---


# yandex_mdb_redis_cluster




Get information about a Yandex Managed Redis cluster. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-redis/concepts).

## Example usage

```terraform
data "yandex_mdb_redis_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_redis_cluster.foo.network_id
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the Redis cluster.
* `name` - (Optional) The name of the Redis cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `network_id` - ID of the network, to which the Redis cluster belongs.
* `created_at` - Creation timestamp of the key.
* `description` - Description of the Redis cluster.
* `labels` - A set of key/value label pairs to assign to the Redis cluster.
* `environment` - Deployment environment of the Redis cluster.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.
* `config` - Configuration of the Redis cluster. The structure is documented below.
* `resources` - Resources allocated to hosts of the Redis cluster. The structure is documented below.
* `host` - A host of the Redis cluster. The structure is documented below.
* `sharded` - Redis Cluster mode enabled/disabled.
* `tls_enabled` - TLS support mode enabled/disabled.
* `persistence_mode` - Persistence mode. Possible values: `ON`, `OFF`.
* `announce_hostnames` - Announce fqdn instead of ip address.
* `security_group_ids` - A set of ids of security groups assigned to hosts of the cluster.

The `config` block supports:

* `timeout` - Close the connection after a client is idle for N seconds.
* `maxmemory_policy` - Redis key eviction policy for a dataset that reaches maximum memory.
* `maxmemory_percent` - Redis maxmemory usage in percent
* `notify_keyspace_events` - Select the events that Redis will notify among a set of classes.
* `slowlog_log_slower_than` - Log slow queries below this number in microseconds.
* `slowlog_max_len` - Slow queries log length.
* `databases` - Number of databases (changing requires redis-server restart).
* `version` - Version of Redis (6.2).
* `client_output_buffer_limit_normal` - Normal clients output buffer limits.
* `client_output_buffer_limit_pubsub` - Pubsub clients output buffer limits.
* `lua_time_limit` - Maximum time in milliseconds for Lua scripts.
* `repl_backlog_size_percent` - Replication backlog size as a percentage of flavor maxmemory.
* `cluster_require_full_coverage` - Controls whether all hash slots must be covered by nodes.
* `cluster_allow_reads_when_down` - Allows read operations when cluster is down.
* `cluster_allow_pubsubshard_when_down` - Permits Pub/Sub shard operations when cluster is down.
* `lfu_decay_time` - The time, in minutes, that must elapse in order for the key counter to be divided by two (or decremented if it has a value less <= 10).
* `lfu_log_factor` - Determines how the frequency counter represents key hits.
* `turn_before_switchover` - Allows to turn before switchover in RDSync.
* `allow_data_loss` - Allows some data to be lost in favor of faster switchover/restart by RDSync.
* `use_luajit` - Use JIT for lua scripts and functions.
* `io_threads_allowed` - Allow Redis to use io-threads.
* `backup_window_start` - Time to start the daily backup, in the UTC timezone. The structure is documented below.

The `backup_window_start` block supports:

* `hours` - The hour at which backup will be started.
* `minutes` - The minute at which backup will be started.

The `resources` block supports:

* `resources_preset_id` - The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-redis/concepts/instance-types).
* `disk_size` - Volume of the storage available to a host, in gigabytes.
* `disk_type_id` - Type of the storage of a host.

The `host` block supports:

* `zone` - The availability zone where the Redis host will be created.
* `subnet_id` - The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.
* `shard_name` - The name of the shard to which the host belongs.
* `fqdn` - The fully qualified domain name of the host.
* `replica_priority` - Replica priority of a current replica (usable for non-sharded only).
* `assign_public_ip` - Sets whether the host should get a public IP address or not.

The `maintenance_window` block supports:

* `type` - Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.
* `hour` - Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.
* `day` - Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.

The `disk_size_autoscaling` block supports:

* `disk_size_limit` - Limit of disk size after autoscaling (GiB).
* `planned_usage_threshold` - Maintenance window autoscaling disk usage (percent).
* `emergency_usage_threshold` - Immediate autoscaling disk usage (percent).
