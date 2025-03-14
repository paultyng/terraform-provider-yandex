---
subcategory: "Compute Cloud"
page_title: "Yandex: yandex_compute_gpu_cluster"
description: |-
  Get information about a Yandex Compute GPU cluster.
---

# yandex_compute_gpu_cluster (Data Source)

Get information about a Yandex Compute GPU cluster. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/gpu-cluster).

## Example usage

```terraform
//
// Get information about existing GPU Cluster.
//
data "yandex_compute_gpu_cluster" "my_gpu_cluster" {
  gpu_cluster_id = "some_gpu_cluster_id"
}

// You can use "data.yandex_compute_gpu_cluster.my_gpu_cluster.id" identifier 
// as reference to the existing resource.
resource "yandex_compute_instance" "default" {
  # ...

  gpu_cluster_id = data.yandex_compute_gpu_cluster.my_gpu_cluster.id

}
```

## Argument Reference

The following arguments are supported:

* `gpu_cluster_id` - (Optional) ID of the GPU cluster.

* `name` - (Optional) Name of the GPU cluster.

~> One of `gpu_cluster_id` or `name` should be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `description` - Optional description of the GPU cluster.
* `folder_id` - ID of the folder that the GPU cluster belongs to.
* `zone` - ID of the zone where the GPU cluster resides.
* `interconnect_type` - type of interconnect used between nodes in GPU cluster.
* `status` - Current status of the GPU cluster.
* `labels` - GPU cluster labels as `key:value` pairs. For details about the concept, see [documentation](https://yandex.cloud/docs/overview/concepts/services#labels).
* `created_at` - Creation timestamp.
