---
subcategory: "Managed Service for Kubernetes (MK8S)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of Yandex Kubernetes Cluster.
---

# {{.Name}} ({{.Type}})

Creates a Yandex Cloud Managed Kubernetes Cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kubernetes/concepts/#kubernetes-cluster).

## Example usage

{{ tffile "examples/kubernetes_cluster/r_kubernetes_cluster_1.tf" }}

{{ tffile "examples/kubernetes_cluster/r_kubernetes_cluster_2.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of a specific Kubernetes cluster.
* `description` - (Optional) A description of the Kubernetes cluster.
* `folder_id` - (Optional) The ID of the folder that the Kubernetes cluster belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the Kubernetes cluster.
* `network_id` - (Optional) The ID of the cluster network.

* `cluster_ipv4_range` - (Optional) CIDR block. IP range for allocating pod addresses. It should not overlap with any subnet in the network the Kubernetes cluster located in. Static routes will be set up for this CIDR blocks in node subnets.
* `cluster_ipv6_range` - (Optional) Identical to cluster_ipv4_range but for IPv6 protocol.

* `node_ipv4_cidr_mask_size` - (Optional) Size of the masks that are assigned to each node in the cluster. Effectively limits maximum number of pods for each node.

* `service_ipv4_range` - (Optional) CIDR block. IP range Kubernetes service Kubernetes cluster IP addresses will be allocated from. It should not overlap with any subnet in the network the Kubernetes cluster located in.
* `service_ipv6_range` - (Optional) Identical to service_ipv4_range but for IPv6 protocol.

* `service_account_id` - Service account to be used for provisioning Compute Cloud and VPC resources for Kubernetes cluster. Selected service account should have `edit` role on the folder where the Kubernetes cluster will be located and on the folder where selected network resides.

* `node_service_account_id` - Service account to be used by the worker nodes of the Kubernetes cluster to access Container Registry or to push node logs and metrics.

**Note**: When access rights for `service_account_id` or `node_service_account_id` are provided using terraform resources, it is necessary to add dependency on these access resources to cluster config:

{{ tffile "examples/kubernetes_cluster/r_kubernetes_cluster_3.tf" }}

Without it, on destroy, terraform will delete cluster and remove access rights for service account(s) simultaneously, that will cause problems for cluster and related node group deletion.

* `release_channel` - Cluster release channel.

* `network_policy_provider` - (Optional) Network policy provider for the cluster. Possible values: `CALICO`.

* `kms_provider` - (Optional) cluster KMS provider parameters.

* `master` - Kubernetes master configuration options. The structure is documented below.

## Attributes Reference

* `id` - (Computed) ID of a new Kubernetes cluster.
* `status` - (Computed)Status of the Kubernetes cluster.
* `health` - (Computed) Health of the Kubernetes cluster.
* `created_at` - (Computed) The Kubernetes cluster creation timestamp.
* `log_group_id` - Log group where cluster stores cluster system logs, like audit, events, or controlplane logs.
* `network_implementation` - (Optional) Network Implementation options. The structure is documented below.

---

The `master` block supports:

* `version` - (Optional) (Computed) Version of Kubernetes that will be used for master.
* `public_ip` - (Optional) (Computed) Boolean flag. When `true`, Kubernetes master will have visible ipv4 address.
* `security_group_ids` - (Optional) List of security group IDs to which the Kubernetes cluster belongs.

* `maintenance_policy` - (Optional) (Computed) Maintenance policy for Kubernetes master. If policy is omitted, automatic revision upgrades of the kubernetes master are enabled and could happen at any time. Revision upgrades are performed only within the same minor version, e.g. 1.29. Minor version upgrades (e.g. 1.29->1.30) should be performed manually. The structure is documented below.

* `zonal` - (Optional) Initialize parameters for Zonal Master (single node master). The structure is documented below.

* `regional` - (Optional) Initialize parameters for Regional Master (highly available master). The structure is documented below.

* `master_location` - (Optional) Cluster master's instances locations array (zone and subnet). Cannot be used together with `zonal` or `regional`. Currently, supports either one, for zonal master, or three instances of `master_location`. Can be updated inplace. When creating regional cluster (three master instances), its `region` will be evaluated automatically by backend. The structure is documented below.

* `version_info` - (Computed) Information about cluster version. The structure is documented below.

* `internal_v4_address` - (Computed) An IPv4 internal network address that is assigned to the master.
* `external_v4_address` - (Computed) An IPv4 external network address that is assigned to the master.
* `internal_v4_endpoint` - (Computed) Internal endpoint that can be used to connect to the master from cloud networks.
* `external_v4_endpoint` - (Computed) External endpoint that can be used to access Kubernetes cluster API from the internet (outside of the cloud).
* `cluster_ca_certificate` - (Computed) PEM-encoded public certificate that is the root of trust for the Kubernetes cluster.

* `master_logging` - (Optional) Master Logging options. The structure is documented below.

---

The `maintenance_policy` block supports:

* `auto_upgrade` - (Required) Boolean flag that specifies if master can be upgraded automatically. When omitted, default value is TRUE.
* `maintenance_window` - (Optional) (Computed) This structure specifies maintenance window, when update for master is allowed. When omitted, it defaults to any time. To specify time of day interval, for all days, one element should be provided, with two fields set, `start_time` and `duration`. Please see `zonal_cluster_resource_name` config example.

To allow maintenance only on specific days of week, please provide list of elements, with all fields set. Only one time interval (`duration`) is allowed for each day of week. Please see `regional_cluster_resource_name` config example.

---

The `zonal` block supports:

* `zone` - (Optional) ID of the availability zone.
* `subnet_id` - (Optional) ID of the subnet. If no ID is specified, and there only one subnet in specified zone, an address in this subnet will be allocated.

---

The `regional` block supports:

* `region` - (Required) Name of availability region (e.g. "ru-central1"), where master instances will be allocated.
* `location` - Array of locations, where master instances will be allocated. The structure is documented below.

---

The `location` block supports repeated values:

* `zone` - (Optional) ID of the availability zone.
* `subnet_id` - (Optional) ID of the subnet.

---

The `master_location` block supports repeated values:

* `zone` - (Optional) ID of the availability zone.
* `subnet_id` - (Optional) ID of the subnet.

---

The `version_info` block supports:

* `current_version` - Current Kubernetes version, major.minor (e.g. 1.30).
* `new_revision_available` - Boolean flag. Newer revisions may include Kubernetes patches (e.g 1.30.1 -> 1.30.2) as well as some internal component updates - new features or bug fixes in yandex-specific components either on the master or nodes.

* `new_revision_summary` - Human readable description of the changes to be applied when updating to the latest revision. Empty if new_revision_available is false.
* `version_deprecated` - Boolean flag. The current version is on the deprecation schedule, component (master or node group) should be upgraded.

---

The `kms_provider` block contains:

* `key_id` - KMS key ID.

---

The `network_implementation` block can contain one of:

* `cilium` - (Optional) Cilium network implementation configuration. No options exist.

---

The `master_logging` block supports:

* `enabled` - (Optional) Boolean flag that specifies if master components logs should be sent to [Yandex Cloud Logging](https://yandex.cloud/docs/logging/). The exact components that will send their logs must be configured via the options described below.
* `log_group_id` - (Optional) ID of the Yandex Cloud Logging [Log group](https://yandex.cloud/docs/logging/concepts/log-group).
* `folder_id` - (Optional) ID of the folder default Log group of which should be used to collect logs.
* `kube_apiserver_enabled` - (Optional) Boolean flag that specifies if kube-apiserver logs should be sent to Yandex Cloud Logging.
* `cluster_autoscaler_enabled` - (Optional) Boolean flag that specifies if cluster-autoscaler logs should be sent to Yandex Cloud Logging.
* `events_enabled` - (Optional) Boolean flag that specifies if kubernetes cluster events should be sent to Yandex Cloud Logging.
* `audit_enabled` - (Optional) Boolean flag that specifies if kube-apiserver audit logs should be sent to Yandex Cloud Logging.

~> Only one of `log_group_id` or `folder_id` (or none) may be specified. If `log_group_id` is specified, logs will be sent to this specific Log group. If `folder_id` is specified, logs will be sent to **default** Log group of this folder. If none of two is specified, logs will be sent to **default** Log group of the **same** folder as Kubernetes cluster.

## Timeouts

This resource provides the following configuration options for [timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default is 30 minute.
- `update` - Default is 20 minute.
- `delete` - Default is 20 minute.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/kubernetes_cluster/import.sh" }}
