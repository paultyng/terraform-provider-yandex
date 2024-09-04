---
layout: "yandex"
page_title: "Yandex: yandex_compute_instance"
sidebar_current: "docs-yandex-compute-instance-x"
description: |-
  Manages a VM instance resource.
---

# yandex\_compute\_instance

A VM instance resource. For more information, see
[the official documentation](https://cloud.yandex.com/docs/compute/concepts/vm).

## Example Usage

```hcl
resource "yandex_compute_instance" "default" {
  name        = "test"
  platform_id = "standard-v1"
  zone        = "ru-central1-a"

  resources {
    cores  = 2
    memory = 4
  }

  boot_disk {
    disk_id = yandex_compute_disk.boot-disk.id
  }

  network_interface {
    index  = 1
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }

  metadata = {
    foo      = "bar"
    ssh-keys = "ubuntu:${file("~/.ssh/id_rsa.pub")}"
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

## Argument Reference

The following arguments are supported:

* `resources` - (Required) Compute resources that are allocated for the instance. The structure is documented below.

* `boot_disk` - (Required) The boot disk for the instance. The structure is documented below.

* `network_interface` - (Required) Networks to attach to the instance. This can
    be specified multiple times. The structure is documented below.

- - -

* `name` - (Optional) Resource name.

* `description` - (Optional) Description of the instance.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the instance.

* `zone` - (Optional) The availability zone where the virtual machine will be created. If it is not provided,
    the default provider folder is used.

* `hostname` - (Optional) Host name for the instance. This field is used to generate the instance `fqdn` value. 
    The host name must be unique within the network and region. If not specified, the host name will be equal 
    to `id` of the instance and `fqdn` will be `<id>.auto.internal`. 
    Otherwise FQDN will be `<hostname>.<region_id>.internal`.                        

* `metadata` - (Optional) Metadata key/value pairs to make available from
    within the instance.

* `platform_id` - (Optional) The type of virtual machine to create. The default is 'standard-v1'.

* `secondary_disk` - (Optional) A set of disks to attach to the instance. The structure is documented below.
    **Note**: The [`allow_stopping_for_update`](#allow_stopping_for_update) property must be set to true in order to update this structure.

* `scheduling_policy` - (Optional) Scheduling policy configuration. The structure is documented below.

* `placement_policy` - (Optional) The placement policy configuration. The structure is documented below.

* `service_account_id` - (Optional) ID of the service account authorized for this instance.

* `allow_stopping_for_update` - (Optional) If true, allows Terraform to stop the instance in order to update its properties.
    If you try to update a property that requires stopping the instance without setting this field, the update will fail.
    
* `network_acceleration_type` - (Optional) Type of network acceleration. The default is `standard`. Values: `standard`, `software_accelerated`

* `local_disk` - (Optional) List of local disks that are attached to the instance. Structure is documented below.

* `filesystem` - (Optional) List of filesystems that are attached to the instance. Structure is documented below.

* `gpu_cluster_id` - (Optional) ID of the GPU cluster to attach this instance to. The GPU cluster must exist in the same zone as the instance.

* `metadata_options` - (Optional) Options allow user to configure access to instance's metadata

* `maintenance_policy` - (Optional) Behaviour on maintenance events. The default is `unspecified`. Values: `unspecified`, `migrate`, `restart`.

* `maintenance_grace_period` - (Optional) Time between notification via metadata service and maintenance. E.g., `60s`.

---

The `resources` block supports:

* `cores` - (Required) CPU cores for the instance.

* `memory` - (Required) Memory size in GB.

* `core_fraction` - (Optional) If provided, specifies baseline performance for a core as a percent.

* `gpus` - (Optional) If provided, specifies the number of GPU devices for the instance 

The `boot_disk` block supports:

* `auto_delete` - (Optional) Defines whether the disk will be auto-deleted when the instance
    is deleted. The default value is `True`.

* `device_name` - (Optional) Name that can be used to access an attached disk.

* `mode` - (Optional) Type of access to the disk resource. By default, a disk is attached in `READ_WRITE` mode.

* `disk_id` - (Optional) The ID of the existing disk (such as those managed by
    `yandex_compute_disk`) to attach as a boot disk.

* `initialize_params` - (Optional) Parameters for a new disk that will be created
    alongside the new instance. Either `initialize_params` or `disk_id` must be set. The structure is documented below.

~> **NOTE:** Either `initialize_params` or `disk_id` must be specified.

The `initialize_params` block supports:

* `name` - (Optional) Name of the boot disk.

* `description` - (Optional) Description of the boot disk.

* `size` - (Optional) Size of the disk in GB.

* `block_size` - (Optional) Block size of the disk, specified in bytes.

* `type` - (Optional) Disk type.

* `image_id` - (Optional) A disk image to initialize this disk from.

* `snapshot_id` - (Optional) A snapshot to initialize this disk from.

~> **NOTE:** Either `image_id` or `snapshot_id` must be specified.

The `network_interface` block supports:

* `subnet_id` - (Required) ID of the subnet to attach this
    interface to. The subnet must exist in the same zone where this instance will be
    created.

* `index` - (Optional) Index of network interface, will be calculated automatically for instance create or update operations
    if not specified. Required for attach/detach operations.

* `ipv4` - (Optional) Allocate an IPv4 address for the interface. The default value is `true`.

* `ip_address` - (Optional) The private IP address to assign to the instance. If
    empty, the address will be automatically assigned from the specified subnet.

* `ipv6` - (Optional) If true, allocate an IPv6 address for the interface.
    The address will be automatically assigned from the specified subnet.

* `ipv6_address` - (Optional) The private IPv6 address to assign to the instance.

* `nat` - (Optional) Provide a public address, for instance, to access the internet over NAT.

* `nat_ip_address` - (Optional) Provide a public address, for instance, to access the internet over NAT. You can use `yandex_vpc_address` resource to create and manage a public address or address should be already reserved in web UI.

* `security_group_ids` - (Optional) Security group ids for network interface.

* `dns_record` - (Optional) List of configurations for creating ipv4 DNS records. The structure is documented below.

* `ipv6_dns_record` - (Optional) List of configurations for creating ipv6 DNS records. The structure is documented below.

* `nat_dns_record` - (Optional) List of configurations for creating ipv4 NAT DNS records. The structure is documented below.

The `dns_record` block supports:

* `fqdn` - (Required) DNS record FQDN (must have a dot at the end).

* `dns_zone_id` - (Optional) DNS zone ID (if not set, private zone used).

* `ttl` - (Optional) DNS record TTL. in seconds

* `ptr` - (Optional) When set to true, also create a PTR DNS record.

The `ipv6_dns_record` block supports:

* `fqdn` - (Required) DNS record FQDN (must have a dot at the end).

* `dns_zone_id` - (Optional) DNS zone ID (if not set, private zone used).

* `ttl` - (Optional) DNS record TTL. in seconds

* `ptr` - (Optional) When set to true, also create a PTR DNS record.

The `nat_dns_record` block supports:

* `fqdn` - (Required) DNS record FQDN (must have a dot at the end).

* `dns_zone_id` - (Optional) DNS zone ID (if not set, private zone used).

* `ttl` - (Optional) DNS record TTL. in seconds

* `ptr` - (Optional) When set to true, also create a PTR DNS record.

The `secondary_disk` block supports:

* `disk_id` - (Required) ID of the disk that is attached to the instance.

* `auto_delete` - (Optional) Whether the disk is auto-deleted when the instance
    is deleted. The default value is false.

* `device_name` - (Optional) Name that can be used to access an attached disk
    under `/dev/disk/by-id/`.

* `mode` - (Optional) Type of access to the disk resource. By default, a disk is attached in `READ_WRITE` mode.

The `scheduling_policy` block supports:

* `preemptible` - (Optional) Specifies if the instance is preemptible. Defaults to false.

The `placement_policy` block supports:

* `placement_group_id` - (Optional) Specifies the id of the Placement Group to assign to the instance.

* `host_affinity_rules` - (Optional) List of host affinity rules. The structure is documented below.

~> **NOTE:** Due to terraform limitations, simply deleting the `placement_policy` fields does not work. To reset the values of these fields, you need to set them empty:
```
placement_policy {
    placement_group_id = ""
    host_affinity_rules = []
}
```

The `host_affinity_rules` block supports:

* `key` - (Required) Affinity label or one of reserved values - `yc.hostId`, `yc.hostGroupId`.

* `op` - (Required) Affinity action. The only value supported is `IN`.

* `value` - (Required) List of values (host IDs or host group IDs).

The `local_disk` block supports:

* `size_bytes` - (Required) Size of the disk, specified in bytes.

~> **NOTE:** Local disks are not available for all users by default.

The `filesystem` block supports:

* `filesystem_id` - (Required) ID of the filesystem that should be attached.

* `device_name` - (Optional) Name of the device representing the filesystem on the instance.

* `mode` - (Optional) Mode of access to the filesystem that should be attached. By default, filesystem is attached 
   in `READ_WRITE` mode.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `fqdn` - The fully qualified DNS name of this instance.

* `network_interface.0.ip_address` - The internal IP address of the instance.

* `network_interface.0.nat_ip_address` - The external IP address of the instance.

* `status` - The status of this instance.

* `created_at` - Creation timestamp of the instance.

* `local_disk.device_name` - The name of the local disk device.

## Timeouts

This resource provides the following configuration options for
[timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default 5 minutes
- `update` - Default 5 minutes
- `delete` - Default 5 minutes

## Import

Instances can be imported using the `ID` of an instance, e.g.

```
$ terraform import yandex_compute_instance.default instance_id
```
