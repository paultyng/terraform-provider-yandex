---
subcategory: "Compute Cloud"
page_title: "Yandex: yandex_compute_filesystem"
description: |-
  Get information about a Yandex Compute filesystem.
---

# yandex_compute_filesystem (Data Source)

Get information about a Yandex Compute filesystem. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/filesystem).

## Example usage

```terraform
//
// Get information about existing Compute Filesystem.
//
data "yandex_compute_filesystem" "my_fs" {
  filesystem_id = "some_fs_id"
}

// You can use "data.yandex_compute_filesystem.my_fs.id" identifier 
// as reference to the existing resource.
resource "yandex_compute_instance" "default" {
  # ...

  filesystem {
    filesystem_id = data.yandex_compute_filesystem.my_fs.id
  }
}
```

## Argument Reference

The following arguments are supported:

* `filsystem_id` - (Optional) ID of the filesystem.

* `name` - (Optional) Name of the filesystem.

~> One of `filesystem_id` or `name` should be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `description` - Optional description of the filesystem.
* `folder_id` - ID of the folder that the filesystem belongs to.
* `zone` - ID of the zone where the filesystem resides.
* `size` - Size of the filesystem, specified in Gb.
* `block_size` - The block size of the filesystem in bytes.
* `type` - ID of the filesystem type.
* `status` - Current status of the filesystem.
* `labels` - Filesystem labels as `key:value` pairs. For details about the concept, see [documentation](https://yandex.cloud/docs/overview/concepts/services#labels).
* `created_at` - Creation timestamp.
