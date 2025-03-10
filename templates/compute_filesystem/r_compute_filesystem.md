---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  File storage is a virtual file system that can be attached to multiple Compute Cloud VMs in the same availability zone.
---

# {{.Name}} ({{.Type}})

File storage is a virtual file system that can be attached to multiple Compute Cloud VMs in the same availability zone.

Users can share files in storage and use them from different VMs.

Storage is attached to a VM through the [Filesystem in Userspace](https://en.wikipedia.org/wiki/Filesystem_in_Userspace) (FUSE) interface as a [virtiofs](https://www.kernel.org/doc/html/latest/filesystems/virtiofs.html) device that is not linked to the host file system directly.

For more information about filesystems in Yandex Cloud, see:

* [Documentation](https://yandex.cloud/docs/compute/concepts/filesystem)
* How-to Guides
  * [Attach filesystem to a VM](https://yandex.cloud/docs/compute/operations/filesystem/attach-to-vm)
  * [Detach filesystem from VM](https://yandex.cloud/docs/compute/operations/filesystem/detach-from-vm)

## Example usage

{{ tffile "examples/compute_filesystem/r_compute_filesystem_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the filesystem. Provide this property when you create a resource.

* `description` - (Optional) Description of the filesystem. Provide this property when you create a resource.

* `folder_id` - (Optional) The ID of the folder that the filesystem belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) Labels to assign to this filesystem. A list of key/value pairs. For details about the concept, see [documentation](https://yandex.cloud/docs/overview/concepts/services#labels).

* `zone` - (Optional) Availability zone where the filesystem will reside.

* `size` - (Optional) Size of the filesystem, specified in GB.

* `block_size` - (Optional) Block size of the filesystem, specified in bytes.

* `type` - (Optional) Type of filesystem to create. Type `network-hdd` is set by default.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - The status of the filesystem.
* `created_at` - Creation timestamp of the filesystem.

## Timeouts

This resource provides the following configuration options for [timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default is 5 minutes.
- `update` - Default is 5 minutes.
- `delete` - Default is 5 minutes.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/compute_filesystem/import.sh" }}
