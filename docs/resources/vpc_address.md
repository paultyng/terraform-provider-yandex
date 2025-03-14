---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: yandex_vpc_address"
description: |-
  Manages a VPC address within Yandex Cloud.
---

# yandex_vpc_address (Resource)

Manages a address within the Yandex Cloud. You can only create a reserved (static) address via this resource. An ephemeral address could be obtained via implicit creation at a compute instance creation only. For more information, see [the official documentation](https://yandex.cloud/docs/vpc/concepts/address).

* How-to Guides
  * [Cloud Networking](https://yandex.cloud/docs/vpc/)
  * [VPC Addressing](https://yandex.cloud/docs/vpc/concepts/address)

## Example usage

```terraform
//
// Create a new VPC regular IPv4 Address.
//
resource "yandex_vpc_address" "addr" {
  name = "exampleAddress"

  external_ipv4_address {
    zone_id = "ru-central1-a"
  }
}
```

### Address with DDoS protection

```terraform
//
// Create a new VPC IPv4 Address with DDoS Protection.
//
resource "yandex_vpc_address" "vpnaddr" {
  name = "vpnaddr"

  external_ipv4_address {
    zone_id                  = "ru-central1-a"
    ddos_protection_provider = "qrator"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the address. Provided by the client when the address is created.
* `description` - (Optional) An optional description of this resource. Provide this property when you create the resource.
* `folder_id` - (Optional) ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.
* `labels` - (Optional) Labels to apply to this resource. A list of key/value pairs.
* `deletion_protection` - (Optional) Flag that protects the address from accidental deletion.

---

* `external_ipv4_address` - (Optional) spec of IP v4 address

---

The `external_ipv4_address` block supports:

* `zone_id` - Zone for allocating address.
* `ddos_protection_provider` - (Optional) Enable DDOS protection. Possible values are: "qrator"
* `outgoing_smtp_capability` - (Optional) Wanted outgoing smtp capability.

~> Either one `address` or `zone_id` arguments can be specified. ~> Either one `ddos_protection_provider` or `outgoing_smtp_capability` arguments can be specified. ~> Change any argument in `external_ipv4_address` will cause an address recreate

---

* `dns_record` - (Optional) DNS record specification of address

---

The `dns_record` block supports:

* `fqdn` - (Required) FQDN for record to address
* `dns_zone_id` - (Required) DNS zone id to create record at.
* `ttl` - (Optional) TTL of DNS record
* `ptr` - (Optional) If PTR record is needed

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `address` - Allocated IP address.
* `created_at` - Creation timestamp of the key.
* `reserved` - `false` means that address is ephemeral.
* `used` - `true` if address is used.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_vpc_address.<resource Name> <resource Id>
terraform import yandex_vpc_address.addr ...
```
