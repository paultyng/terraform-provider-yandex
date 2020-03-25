package yandex

import (
	"context"
	"fmt"
	"testing"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
)

func init() {
	resource.AddTestSweepers("yandex_compute_instance_group", &resource.Sweeper{
		Name: "yandex_compute_instance_group",
		F:    testSweepComputeInstanceGroups,
		Dependencies: []string{
			"yandex_kubernetes_node_group",
		},
	})
}

func testSweepComputeInstanceGroups(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	var serviceAccountID, networkID, subnetID string
	var depsCreated bool

	it := conf.sdk.InstanceGroup().InstanceGroup().InstanceGroupIterator(conf.Context(), conf.FolderID)
	result := &multierror.Error{}
	for it.Next() {
		if !depsCreated {
			depsCreated = true
			serviceAccountID, err = createIAMServiceAccountForSweeper(conf)
			if err != nil {
				result = multierror.Append(result, err)
				break
			}
			networkID, err = createVPCNetworkForSweeper(conf)
			if err != nil {
				result = multierror.Append(result, err)
				break
			}
			subnetID, err = createVPCSubnetForSweeper(conf, networkID)
			if err != nil {
				result = multierror.Append(result, err)
				break
			}
		}

		id := it.Value().GetId()
		if !updateComputeInstanceGroupWithSweeperDeps(conf, id, serviceAccountID, networkID, subnetID) {
			result = multierror.Append(result,
				fmt.Errorf("failed to sweep (update with dependencies) compute instance group %q", id))
			continue
		}

		if !sweepComputeInstanceGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep compute instance group %q", id))
		}
	}

	if serviceAccountID != "" {
		if !sweepIAMServiceAccount(conf, serviceAccountID) {
			result = multierror.Append(result,
				fmt.Errorf("failed to sweep IAM service account %q", serviceAccountID))
		}
	}
	if subnetID != "" {
		if !sweepVPCSubnet(conf, subnetID) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC subnet %q", subnetID))
		}
	}
	if networkID != "" {
		if !sweepVPCNetwork(conf, networkID) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC network %q", networkID))
		}
	}

	return result.ErrorOrNil()
}

func sweepComputeInstanceGroup(conf *Config, id string) bool {
	return sweepWithRetry(sweepComputeInstanceGroupOnce, conf, "Compute instance group", id)
}

func sweepComputeInstanceGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexVPCNetworkDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.InstanceGroup().InstanceGroup().Delete(ctx, &instancegroup.DeleteInstanceGroupRequest{
		InstanceGroupId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func updateComputeInstanceGroupWithSweeperDeps(conf *Config, instanceGroupID, serviceAccountID, networkID, subnetID string) bool {
	debugLog("started updating instance group %q", instanceGroupID)

	client := conf.sdk.InstanceGroup().InstanceGroup()
	for i := 1; i <= conf.MaxRetries; i++ {
		req := &instancegroup.UpdateInstanceGroupRequest{
			InstanceGroupId:  instanceGroupID,
			ServiceAccountId: serviceAccountID,
			AllocationPolicy: &instancegroup.AllocationPolicy{
				Zones: []*instancegroup.AllocationPolicy_Zone{
					{ZoneId: conf.Zone},
				},
			},
			InstanceTemplate: &instancegroup.InstanceTemplate{
				NetworkInterfaceSpecs: []*instancegroup.NetworkInterfaceSpec{
					{
						NetworkId:            networkID,
						SubnetIds:            []string{subnetID},
						PrimaryV4AddressSpec: &instancegroup.PrimaryAddressSpec{},
					},
				},
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{
					"allocation_policy",
					"service_account_id",
					"instance_template.network_interface_specs",
				},
			},
		}

		_, err := conf.sdk.WrapOperation(client.Update(conf.Context(), req))
		if err != nil {
			debugLog("[instance group %q] update try #%d: %v", instanceGroupID, i, err)
		} else {
			debugLog("[instance group %q] update try #%d: request was successfully sent", instanceGroupID, i)
			return true
		}
	}

	debugLog("instance group %q update failed", instanceGroupID)
	return false
}

func computeInstanceGroupImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      "yandex_compute_instance_group.group1",
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccComputeInstanceGroup_basic(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigMain(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})

}

func TestAccComputeInstanceGroup_Gpus(t *testing.T) {
	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigGpus(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupHasGpus(&ig, 1),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_NetworkSettings(t *testing.T) {
	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigNetworkSettings(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupNetworkSettings(&ig, "SOFTWARE_ACCELERATED"),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_Variables(t *testing.T) {
	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigVariables(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupVariables(&ig,
						append(make([]*instancegroup.Variable, 0),
							&instancegroup.Variable{Key: "test_key1", Value: "test_value1"},
							&instancegroup.Variable{Key: "test_key2", Value: "test_value2"})),
				),
			},
			{
				Config: testAccComputeInstanceGroupConfigVariables2(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupVariables(&ig,
						append(make([]*instancegroup.Variable, 0),
							&instancegroup.Variable{Key: "test_key1", Value: "test_value1_new"},
							&instancegroup.Variable{Key: "test_key2", Value: "test_value2"},
							&instancegroup.Variable{Key: "test_key3", Value: "test_value3"})),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_full(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigFull(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupDefaultValues(&ig),
					testAccCheckComputeInstanceGroupFixedScalePolicy(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}
func TestAccComputeInstanceGroup_autoscale(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigAutocsale(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupAutoScalePolicy(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_update(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigWithLabels(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupLabel(&ig, "label_key1", "label_value1"),
				),
			},
			{
				Config: testAccComputeInstanceGroupConfigWithLabels2(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupLabel(&ig, "label_key1", "label_value2"),
					testAccCheckComputeInstanceGroupLabel(&ig, "label_key_extra", "label_value_extra"),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_update2(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigWithTemplateLabels3(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupTemplateLabel(&ig, "label_key1", "label_value1"),
					testAccCheckComputeInstanceGroupTemplateMeta(&ig, "meta_key1", "meta_val1"),
				),
			},
			{
				Config: testAccComputeInstanceGroupConfigWithTemplateLabels4(name, saName),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupTemplateLabel(&ig, "label_key1", "label_value2"),
					testAccCheckComputeInstanceGroupTemplateLabel(&ig, "label_key_extra", "label_value_extra"),
					testAccCheckComputeInstanceGroupTemplateMeta(&ig, "meta_key1", "meta_val2"),
					testAccCheckComputeInstanceGroupTemplateMeta(&ig, "meta_key_extra", "meta_value_extra"),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func testAccCheckComputeInstanceGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_instance_group" {
			continue
		}

		_, err := config.sdk.InstanceGroup().InstanceGroup().Get(context.Background(), &instancegroup.GetInstanceGroupRequest{
			InstanceGroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Instance Group still exists")
		}
	}

	return nil
}

type Disk struct {
	Description string
	Mode        string
	Size        int
	Type        string
	Image       string
	Snapshot    string
}

func testAccComputeInstanceGroupConfigMain(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
    }

    boot_disk {
      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }
  }

  scale_policy {
    fixed_scale {
      size = 2
    }
  }

  allocation_policy {
    zones = ["ru-central1-a"]
  }

  deploy_policy {
    max_unavailable = 3
    max_creating    = 3
    max_expansion   = 3
    max_deleting    = 3
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  description    = "tf-test"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigWithLabels(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
    }

    boot_disk {
      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }
  }

  scale_policy {
    fixed_scale {
      size = 2
    }
  }

  allocation_policy {
    zones = ["ru-central1-a"]
  }

  deploy_policy {
    max_unavailable = 3
    max_creating    = 3
    max_expansion   = 3
    max_deleting    = 3
  }

  labels = {
    label_key1 = "label_value1"
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  description    = "tf-test"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigWithLabels2(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
    }

    boot_disk {
      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }
  }

  scale_policy {
    fixed_scale {
      size = 2
    }
  }

  allocation_policy {
    zones = ["ru-central1-a"]
  }

  deploy_policy {
    max_unavailable = 3
    max_creating    = 3
    max_expansion   = 3
    max_deleting    = 3
  }

  labels = {
    label_key1      = "label_value2"
    label_key_extra = "label_value_extra"
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  description    = "tf-test"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigWithTemplateLabels3(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
    }

    boot_disk {
      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }

    labels = {
      label_key1 = "label_value1"
    }

    metadata = {
      meta_key1 = "meta_val1"
    }
  }

  scale_policy {
    fixed_scale {
      size = 2
    }
  }

  allocation_policy {
    zones = ["ru-central1-a"]
  }

  deploy_policy {
    max_unavailable = 3
    max_creating    = 3
    max_expansion   = 3
    max_deleting    = 3
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  description    = "tf-test"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigWithTemplateLabels4(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
    }

    boot_disk {
      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }

    labels = {
      label_key1      = "label_value2"
      label_key_extra = "label_value_extra"
    }

    metadata = {
      meta_key1      = "meta_val2"
      meta_key_extra = "meta_value_extra"
    }
  }

  scale_policy {
    fixed_scale {
      size = 2
    }
  }

  allocation_policy {
    zones = ["ru-central1-a"]
  }

  deploy_policy {
    max_unavailable = 3
    max_creating    = 3
    max_expansion   = 3
    max_deleting    = 3
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  description    = "tf-test"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigFull(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory        = 2
      cores         = 1
      core_fraction = 20
    }

    boot_disk {
      mode = "READ_WRITE"

      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
        type     = "network-hdd"
      }
    }

    secondary_disk {
      initialize_params {
        description = "desc1"
        image_id    = "${data.yandex_compute_image.ubuntu.id}"
        size        = 3
        type        = "network-nvme"
      }
    }

    secondary_disk {
      initialize_params {
        description = "desc2"
        image_id    = "${data.yandex_compute_image.ubuntu.id}"
        size        = 3
        type        = "network-hdd"
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }

    scheduling_policy {
      preemptible = true
    }
  }

  scale_policy {
    fixed_scale {
      size = 2
    }
  }

  allocation_policy {
    zones = ["ru-central1-a"]
  }

  deploy_policy {
    max_unavailable  = 4
    max_creating     = 3
    max_expansion    = 2
    max_deleting     = 1
    startup_duration = 5
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  description    = "tf-test"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)

}

func testAccComputeInstanceGroupConfigAutocsale(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
    }

    boot_disk {
      mode = "READ_WRITE"

      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
        type     = "network-hdd"
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }

    scheduling_policy {
      preemptible = true
    }
  }

  scale_policy {
    auto_scale {
      initial_size           = 1
      max_size               = 2
      min_zone_size          = 1
      measurement_duration   = 120
      cpu_utilization_target = 80
      custom_rule {
        rule_type   = "WORKLOAD"
        metric_type = "GAUGE"
        metric_name = "metric1"
        target      = 50
      }
    }
  }

  allocation_policy {
    zones = ["ru-central1-a"]
  }

  deploy_policy {
    max_unavailable  = 4
    max_creating     = 3
    max_expansion    = 2
    max_deleting     = 1
    startup_duration = 5
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  description    = "tf-test"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)

}

func testAccComputeInstanceGroupConfigGpus(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "gpu-standard-v1"
    description = "template_description"

    resources {
      cores  = 8
      memory = 96
      gpus   = 1
    }

    boot_disk {
      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }
  }

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  allocation_policy {
    zones = ["ru-central1-b"]
  }

  deploy_policy {
    max_unavailable = 3
    max_creating    = 3
    max_expansion   = 3
    max_deleting    = 3
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  description    = "tf-test"
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigNetworkSettings(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
    }

    boot_disk {
      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }

    network_settings {
      type = "SOFTWARE_ACCELERATED"
    }
  }

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  allocation_policy {
    zones = ["ru-central1-b"]
  }

  deploy_policy {
    max_unavailable = 3
    max_creating    = 3
    max_expansion   = 3
    max_deleting    = 3
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  description    = "tf-test"
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigVariables(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
    }

    boot_disk {
      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }
  }

  variables {
    key   = "test_key1"
    value = "test_value1"
  }
  variables {
    key   = "test_key2"
    value = "test_value2"
  }

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  allocation_policy {
    zones = ["ru-central1-b"]
  }

  deploy_policy {
    max_unavailable = 3
    max_creating    = 3
    max_expansion   = 3
    max_deleting    = 3
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  description    = "tf-test"
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigVariables2(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
    }

    boot_disk {
      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }
  }

  variables {
    key   = "test_key1"
    value = "test_value1_new"
  }
  variables {
    key   = "test_key2"
    value = "test_value2"
  }
  variables {
    key   = "test_key3"
    value = "test_value3"
  }

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  allocation_policy {
    zones = ["ru-central1-b"]
  }

  deploy_policy {
    max_unavailable = 3
    max_creating    = 3
    max_expansion   = 3
    max_deleting    = 3
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  description    = "tf-test"
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccCheckComputeInstanceGroupExists(n string, instance *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.InstanceGroup().InstanceGroup().Get(context.Background(), &instancegroup.GetInstanceGroupRequest{
			InstanceGroupId: rs.Primary.ID,
			View:            instancegroup.InstanceGroupView_FULL,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("instancegroup is not found")
		}

		*instance = *found

		return nil
	}
}

func testAccCheckComputeInstanceGroupHasGpus(ig *instancegroup.InstanceGroup, gpus int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.GetInstanceTemplate().ResourcesSpec.Gpus != gpus {
			return fmt.Errorf("invalid resources.gpus value in instance_template in instance group %s", ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupVariables(ig *instancegroup.InstanceGroup, variables []*instancegroup.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for i, raw := range variables {
			if ig.GetVariables()[i].GetKey() != raw.GetKey() || ig.GetVariables()[i].GetValue() != raw.GetValue() {
				return fmt.Errorf("invalid variables value in instance group %s", ig.Name)
			}
		}
		return nil
	}
}

func testAccCheckComputeInstanceGroupNetworkSettings(ig *instancegroup.InstanceGroup, nst string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.GetInstanceTemplate().GetNetworkSettings().GetType().String() != nst {
			return fmt.Errorf("invalid network_settings.type value in instance_template in instance group %s", ig.Name)
		}
		return nil
	}
}

func testAccCheckComputeInstanceGroupLabel(ig *instancegroup.InstanceGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.Labels == nil {
			return fmt.Errorf("no labels found on instance group %s", ig.Name)
		}

		if v, ok := ig.Labels[key]; ok {
			if v != value {
				return fmt.Errorf("expected value '%s' but found value '%s' for label '%s' on instance group %s", value, v, key, ig.Name)
			}
		} else {
			return fmt.Errorf("no label found with key %s on instance group %s", key, ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupTemplateLabel(ig *instancegroup.InstanceGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.InstanceTemplate.GetLabels() == nil {
			return fmt.Errorf("no template labels found on instance group %s", ig.Name)
		}

		if v, ok := ig.InstanceTemplate.Labels[key]; ok {
			if v != value {
				return fmt.Errorf("expected value '%s' but found value '%s' for label '%s' on instance group %s template labels", value, v, key, ig.Name)
			}
		} else {
			return fmt.Errorf("no label found with key %s on instance group %s template labels", key, ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupTemplateMeta(ig *instancegroup.InstanceGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.InstanceTemplate.GetMetadata() == nil {
			return fmt.Errorf("no template labels found on instance group %s", ig.Name)
		}

		if v, ok := ig.InstanceTemplate.Metadata[key]; ok {
			if v != value {
				return fmt.Errorf("expected value '%s' but found value '%s' for label '%s' on instance group %s template labels", value, v, key, ig.Name)
			}
		} else {
			return fmt.Errorf("no label found with key %s on instance group %s template labels", key, ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupDefaultValues(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// InstanceTemplate
		if ig.GetInstanceTemplate() == nil {
			return fmt.Errorf("no InstanceTemplate in instance group %s", ig.Name)
		}
		if ig.GetInstanceTemplate().PlatformId != "standard-v1" {
			return fmt.Errorf("invalid PlatformId value in instance group %s", ig.Name)
		}
		if ig.GetInstanceTemplate().Description != "template_description" {
			return fmt.Errorf("invalid Description value in instance group %s", ig.Name)
		}
		// Resources
		if ig.GetInstanceTemplate().ResourcesSpec == nil {
			return fmt.Errorf("no ResourcesSpec in instance group %s", ig.Name)
		}
		if ig.GetInstanceTemplate().ResourcesSpec.Cores != 1 {
			return fmt.Errorf("invalid Cores value in instance group %s", ig.Name)
		}
		if ig.GetInstanceTemplate().ResourcesSpec.Memory != toBytes(2) {
			return fmt.Errorf("invalid Memory value in instance group %s", ig.Name)
		}
		if ig.GetInstanceTemplate().ResourcesSpec.CoreFraction != 20 {
			return fmt.Errorf("invalid CoreFraction value in instance group %s", ig.Name)
		}
		// SchedulingPolicy
		if !ig.GetInstanceTemplate().SchedulingPolicy.Preemptible {
			return fmt.Errorf("invalid Preemptible value in instance group %s", ig.Name)
		}
		// BootDisk
		bootDisk := &Disk{Mode: "READ_WRITE", Size: 4, Type: "network-hdd"}
		if err := checkDisk(fmt.Sprintf("instancegroup %s boot disk", ig.Name), ig.GetInstanceTemplate().BootDiskSpec, bootDisk); err != nil {
			return err
		}
		// SecondaryDisk
		if len(ig.InstanceTemplate.SecondaryDiskSpecs) != 2 {
			return fmt.Errorf("invalid number of secondary disks in instance group %s", ig.Name)
		}

		disk0 := &Disk{Size: 3, Type: "network-nvme", Description: "desc1"}
		if err := checkDisk(fmt.Sprintf("instancegroup %s secondary disk #0", ig.Name), ig.InstanceTemplate.SecondaryDiskSpecs[0], disk0); err != nil {
			return err
		}

		disk1 := &Disk{Size: 3, Type: "network-hdd", Description: "desc2"}
		if err := checkDisk(fmt.Sprintf("instancegroup %s secondary disk #1", ig.Name), ig.InstanceTemplate.SecondaryDiskSpecs[1], disk1); err != nil {
			return err
		}

		// AllocationPolicy
		if ig.AllocationPolicy == nil || len(ig.AllocationPolicy.Zones) != 1 || ig.AllocationPolicy.Zones[0].ZoneId != "ru-central1-a" {
			return fmt.Errorf("invalid allocation policy in instance group %s", ig.Name)
		}

		// Deploy policy
		if ig.GetDeployPolicy() == nil {
			return fmt.Errorf("no deploy policy in instance group %s", ig.Name)
		}

		if ig.GetDeployPolicy().MaxUnavailable != 4 {
			return fmt.Errorf("invalid MaxUnavailable in instance group %s", ig.Name)
		}
		if ig.GetDeployPolicy().MaxCreating != 3 {
			return fmt.Errorf("invalid MaxCreating in instance group %s", ig.Name)
		}
		if ig.GetDeployPolicy().MaxExpansion != 2 {
			return fmt.Errorf("invalid MaxExpansion in instance group %s", ig.Name)
		}
		if ig.GetDeployPolicy().MaxDeleting != 1 {
			return fmt.Errorf("invalid MaxDeleting in instance group %s", ig.Name)
		}
		if ig.GetDeployPolicy().StartupDuration.Seconds != 5 {
			return fmt.Errorf("invalid StartupDuration in instance group %s", ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupFixedScalePolicy(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.ScalePolicy.GetFixedScale() == nil || ig.ScalePolicy.GetFixedScale().Size != 2 {
			return fmt.Errorf("invalid fixed scale policy on instance group %s", ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupAutoScalePolicy(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.ScalePolicy.GetAutoScale() == nil {
			return fmt.Errorf("no auto scale policy on instance group %s", ig.Name)
		}

		sp := ig.ScalePolicy.GetAutoScale()
		if sp.InitialSize != 1 {
			return fmt.Errorf("wrong initialsize on instance group %s", ig.Name)
		}
		if sp.MaxSize != 2 {
			return fmt.Errorf("wrong max_size on instance group %s", ig.Name)
		}
		if sp.MeasurementDuration == nil || sp.MeasurementDuration.Seconds != 120 {
			return fmt.Errorf("wrong measurement_duration on instance group %s", ig.Name)
		}
		if sp.CpuUtilizationRule == nil || sp.CpuUtilizationRule.UtilizationTarget != 80. {
			return fmt.Errorf("wrong cpu_utilization_target on instance group %s", ig.Name)
		}
		return nil
	}
}

func checkDisk(name string, a *instancegroup.AttachedDiskSpec, d *Disk) error {
	if d.Mode != "" && a.Mode.String() != d.Mode {
		return fmt.Errorf("invalid Mode value in %s", name)
	}
	if a.DiskSpec.Description != d.Description {
		return fmt.Errorf("invalid Description value in %s", name)
	}
	if d.Type != "" && a.DiskSpec.TypeId != d.Type {
		return fmt.Errorf("invalid Type value in %s", name)
	}
	if a.DiskSpec.Size != toBytes(d.Size) {
		return fmt.Errorf("invalid Size value in %s", name)
	}
	if a.DiskSpec.GetSnapshotId() != d.Snapshot {
		return fmt.Errorf("invalid Snapshot value in %s", name)
	}
	if d.Image != "" && a.DiskSpec.GetImageId() != d.Image {
		return fmt.Errorf("invalid Image value in %s", name)
	}
	return nil
}
