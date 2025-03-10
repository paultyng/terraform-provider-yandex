---
subcategory: "IoT Core"
page_title: "Yandex: yandex_iot_core_device"
description: |-
  Get information about a Yandex Cloud IoT Core Device.
---

# yandex_iot_core_device (Data Source)

Get information about a Yandex IoT Core device. For more information about IoT Core, see [Yandex Cloud IoT Device](https://yandex.cloud/docs/iot-core/quickstart).

## Example usage

```terraform
//
// Get information about existing IoT Core Device.
//
data "yandex_iot_core_device" "my_device" {
  device_id = "are1sampleregistry11"
}
```

This data source is used to define [Yandex Cloud IoT Device](https://yandex.cloud/docs/iot-core/quickstart) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `device_id` (Optional) - IoT Core Device id used to define device

* `name` (Optional) - IoT Core Device name used to define device

~> Either `device_id` or `name` must be specified.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the IoT Core Device
* `registry_id` - IoT Core Registry ID for the IoT Core Device
* `aliases` - A set of key/value aliases pairs to assign to the IoT Core Device
* `created_at` - Creation timestamp of the IoT Core Device
* `certificates` - A set of certificate's fingerprints for the IoT Core Device
* `passwords` - A set of passwords's id for the IoT Core Device
