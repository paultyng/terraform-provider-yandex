---
subcategory: "Managed Service for YDB"
page_title: "Yandex: yandex_ydb_table"
description: |-
  Manages Yandex Database dedicated cluster.
---

# yandex_ydb_table (Resource)

Yandex Database table resource.

## Example usage

```terraform
//
// Create a new YDB Table.
//
resource "yandex_ydb_table" "test_table" {
  path              = "test_dir/test_table_3_col"
  connection_string = yandex_ydb_database_serverless.database1.ydb_full_endpoint

  column {
    name     = "a"
    type     = "Utf8"
    not_null = true
  }
  column {
    name     = "b"
    type     = "Uint32"
    not_null = true
  }
  column {
    name     = "c"
    type     = "Int32"
    not_null = false
  }
  column {
    name = "d"
    type = "Timestamp"
  }

  primary_key = ["a", "b"]
}
```

## Argument Reference

The following arguments are supported:

* `path` - (Required) Table path.

* `connection_string` - (Required) Connection string for database.

* `primary_key` - (Required) A list of table columns to be uased as primary key.

* `column` - (Required) A list of column configuration options. The structure is documented below.

* `family` - (Optional) A list of column group configuration options. The structure is documented below.

* `ttl` - (Optional) ttl TTL settings The structure is documented below.

* `attributes` - (Optional) A map of table attributes.

* `partitioning_settings` - (Optional) Table partiotioning settings The structure is documented below.

* `key_bloom_filter` - (Optional) Use the Bloom filter for the primary key

* `read_replicas_settings` - (Optional) Read replication settings

---

The `column` block supports:

* `name` - (Required) Column name

* `type` - (Required) Column data type. YQL data types are used.

* `family` - (Optional) Column group

* `not_null` - (Optional) A column cannot have the NULL data type. ( Default: false )

---

The `family` block may be used to group columns into [families](https://ydb.tech/en/docs/yql/reference/syntax/create_table#column-family) to set shared parameters for them, such as:

* `name` - (Required) Column family name

* `data` - (Optional) Type of storage device for column data in this group (acceptable values: ssd, rot (from HDD spindle rotation)).

* `compression` - (Optional) Data codec (acceptable values: off, lz4).

---

The `ttl` block supports allow you to create a special column type, [TTL column](https://ydb.tech/en/docs/concepts/ttl), whose values determine the time-to-live for rows:

* `column_name` - (Required) Column name for TTL

* `expire_interval` - (Required) Interval in the ISO 8601 format

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_ydb_table.<resource Name> <resource Id>
terraform import yandex_ydb_table.test_table ...
```
