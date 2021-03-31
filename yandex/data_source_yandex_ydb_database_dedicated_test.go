package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
)

const ydbDatabaseDedicatedDataSource = "data.yandex_ydb_database_dedicated.test-ydb-database-dedicated"

func TestAccDataSourceYandexYDBDatabaseDedicated_byID(t *testing.T) {
	t.Parallel()

	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-dedicated")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseDedicatedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexYDBDatabaseDedicatedByID(databaseName, databaseDesc),
				Check: resource.ComposeTestCheckFunc(
					testYandexYDBDatabaseDedicatedExists(ydbDatabaseDedicatedDataSource, &database),
					resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedDataSource, "database_id"),
					resource.TestCheckResourceAttr(ydbDatabaseDedicatedDataSource, "name", databaseName),
					resource.TestCheckResourceAttr(ydbDatabaseDedicatedDataSource, "description", databaseDesc),
					resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(ydbDatabaseDedicatedDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexYDBDatabaseDedicated_byName(t *testing.T) {
	t.Parallel()

	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-dedicated")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseDedicatedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexYDBDatabaseDedicatedByName(databaseName, databaseDesc),
				Check: resource.ComposeTestCheckFunc(
					testYandexYDBDatabaseDedicatedExists(ydbDatabaseDedicatedDataSource, &database),
					resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedDataSource, "database_id"),
					resource.TestCheckResourceAttr(ydbDatabaseDedicatedDataSource, "name", databaseName),
					resource.TestCheckResourceAttr(ydbDatabaseDedicatedDataSource, "description", databaseDesc),
					resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(ydbDatabaseDedicatedDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexYDBDatabaseDedicated_full(t *testing.T) {
	t.Parallel()

	var database ydb.Database
	params := testYandexYDBDatabaseDedicatedParameters{}
	params.name = acctest.RandomWithPrefix("tf-ydb-database-dedicated")
	params.desc = acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-ydb-database-dedicated-label")
	params.labelValue = acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseDedicatedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexYDBDatabaseDedicatedDataSource(params),
				Check: resource.ComposeTestCheckFunc(
					testYandexYDBDatabaseDedicatedExists(ydbDatabaseDedicatedDataSource, &database),
					resource.TestCheckResourceAttr(ydbDatabaseDedicatedDataSource, "name", params.name),
					resource.TestCheckResourceAttr(ydbDatabaseDedicatedDataSource, "description", params.desc),
					resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedDataSource, "folder_id"),
					testYandexYDBDatabaseDedicatedContainsLabel(&database, params.labelKey, params.labelValue),
					testAccCheckCreatedAtAttr(ydbDatabaseDedicatedDataSource),
				),
			},
		},
	})
}

func testYandexYDBDatabaseDedicatedByID(name string, desc string) string {
	return fmt.Sprintf(ydbDatabaseDedicatedDependencies+`
data "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  database_id = "${yandex_ydb_database_dedicated.test-ydb-database-dedicated.id}"
}

resource "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  name               = "%s"
  description        = "%s"
  resource_preset_id = "medium"
  scale_policy {
    fixed_scale {
      size = 1
    }
  }
  storage_config {
    group_count     = 1
    storage_type_id = "ssd"
  }
  network_id = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  subnet_ids = [
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-a.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-b.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-c.id}",
  ]
}`, name, desc)
}

func testYandexYDBDatabaseDedicatedByName(name string, desc string) string {
	return fmt.Sprintf(ydbDatabaseDedicatedDependencies+`
data "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  name = "${yandex_ydb_database_dedicated.test-ydb-database-dedicated.name}"
}

resource "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  name               = "%s"
  description        = "%s"
  resource_preset_id = "medium"
  scale_policy {
    fixed_scale {
      size = 1
    }
  }
  storage_config {
    group_count     = 1
    storage_type_id = "ssd"
  }
  network_id = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  subnet_ids = [
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-a.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-b.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-c.id}",
  ]
}
`, name, desc)
}

func testYandexYDBDatabaseDedicatedDataSource(params testYandexYDBDatabaseDedicatedParameters) string {
	return fmt.Sprintf(ydbDatabaseDedicatedDependencies+`
data "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  database_id = "${yandex_ydb_database_dedicated.test-ydb-database-dedicated.id}"
}

resource "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  resource_preset_id = "medium"
  scale_policy {
    fixed_scale {
      size = 1
    }
  }
  storage_config {
    group_count     = 1
    storage_type_id = "ssd"
  }
  network_id = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  subnet_ids = [
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-a.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-b.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-c.id}",
  ]
}
`,
		params.name,
		params.desc,
		params.labelKey,
		params.labelValue)
}
