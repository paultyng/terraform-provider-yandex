---
subcategory: "Managed Service for Apache Kafka"
page_title: "Yandex: yandex_mdb_kafka_connector"
description: |-
  Manages a connectors of a Kafka cluster within Yandex Cloud.
---

# yandex_mdb_kafka_connector (Resource)

Manages a connector of a Kafka cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).

## Example usage

```terraform
//
// Create a new MDB Kafka Connector.
//
resource "yandex_mdb_kafka_connector" "my_conn" {
  cluster_id = yandex_mdb_kafka_cluster.my_cluster.id
  name       = "replication"
  tasks_max  = 3
  properties = {
    refresh.topics.enabled = "true"
  }
  connector_config_mirrormaker {
    topics             = "data.*"
    replication_factor = 1
    source_cluster {
      alias = "source"
      external_cluster {
        bootstrap_servers = "somebroker1:9091,somebroker2:9091"
        sasl_username     = "someuser"
        sasl_password     = "somepassword"
        sasl_mechanism    = "SCRAM-SHA-512"
        security_protocol = "SASL_SSL"
      }
    }
    target_cluster {
      alias = "target"
      this_cluster {}
    }
  }
}

resource "yandex_mdb_kafka_connector" "connector" {
  cluster_id = yandex_mdb_kafka_cluster.my_cluster.id
  name       = "s3-sink"
  tasks_max  = 3
  properties = {
    "key.converter"                  = "org.apache.kafka.connect.storage.StringConverter"
    "value.converter"                = "org.apache.kafka.connect.json.JsonConverter"
    "value.converter.schemas.enable" = "false"
    "format.output.type"             = "jsonl"
    "file.name.template"             = "dir1/dir2/{{topic}}-{{partition:padding=true}}-{{start_offset:padding=true}}.gz"
    "timestamp.timezone"             = "Europe/Moscow"
  }
  connector_config_s3_sink {
    topics                = "data.*"
    file_compression_type = "gzip"
    file_max_records      = 100
    s3_connection {
      bucket_name = "somebucket"
      external_s3 {
        endpoint          = "storage.yandexcloud.net"
        access_key_id     = "some_access_key_id"
        secret_access_key = "some_secret_access_key"
      }
    }
  }
}

resource "yandex_mdb_kafka_cluster" "my_cluster" {
  name       = "foo"
  network_id = "c64vs98keiqc7f24pvkd"

  config {
    version = "2.8"
    zones   = ["ru-central1-a"]
    kafka {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-hdd"
        disk_size          = 16
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:
* `name` - (Required) The name of the connector.
* `tasks_max` - (Optional) The number of the connector's parallel working tasks. Default is the number of brokers
* `properties` - (Optional) Additional properties for connector.
* `connector_config_mirrormaker` - (Optional) Params for MirrorMaker2 connector. The structure is documented below.
* `connector_config_s3_sink` - (Optional) Params for S3 Sink connector. The structure is documented below.

The `connector_config_mirrormaker` block supports:
* `topics` - (Required) The pattern for topic names to be replicated.
* `replication_factor` - (Optional) Replication factor for topics created in target cluster
* `source_cluster` - (Required) Settings for source cluster. The structure is documented below.
* `target_cluster` - (Required) Settings for target cluster. The structure is documented below.

The `source_cluster` and `target_cluster` block supports:
* `alias` - (Optional) Name of the cluster. Used also as a topic prefix
* `external_cluster` - (Optional) Connection params for external cluster
* `this_cluster` - (Optional) Using this section in the cluster definition (source or target) means it's this cluster

The `external_cluster` blocks support:
* `bootstrap_servers` - (Required) List of bootstrap servers to connect to cluster
* `sasl_username` - (Optional) Username to use in SASL authentification mechanism
* `sasl_password` - (Optional) Password to use in SASL authentification mechanism
* `sasl_mechanism` - (Optional) Type of SASL authentification mechanism to use
* `security_protocol` - (Optional) Security protocol to use

The `connector_config_s3_sink` block supports:
* `topics` - (Required) The pattern for topic names to be copied to s3 bucket.
* `file_compression_type` - (Required) Сompression type for messages. Cannot be changed.
* `file_max_records` - (Optional) Max records per file.
* `s3_connection` - (Required) Settings for connection to s3-compatible storage. The structure is documented below.

The `s3_connection` block supports:
* `bucket_name` - (Required) Name of the bucket in s3-compatible storage.
* `external_s3` - (Required) Connection params for external s3-compatible storage. The structure is documented below.

The `external_s3` blocks support:
* `endpoint` - (Required) URL of s3-compatible storage.
* `access_key_id` - (Optional) ID of aws-compatible static key.
* `secret_access_key` - (Optional) Secret key of aws-compatible static key.
* `region` - (Optional) region of s3-compatible storage. [Available region list](https://docs.aws.amazon.com/AWSJavaSDK/latest/javadoc/com/amazonaws/regions/Regions.html).


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_mdb_kafka_connector.<resource Name> <resource Id>
terraform import yandex_mdb_kafka_connector.my_conn ...
```
