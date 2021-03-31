package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
)

const ydbDatabaseServerlessDataSource = "data.yandex_ydb_database_serverless.test-ydb-database-serverless"

func TestAccDataSourceYandexYDBDatabaseServerless_byID(t *testing.T) {
	t.Parallel()

	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-serverless")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-serverless-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexYDBDatabaseServerlessByID(databaseName, databaseDesc),
				Check: resource.ComposeTestCheckFunc(
					testYandexYDBDatabaseServerlessExists(ydbDatabaseServerlessDataSource, &database),
					resource.TestCheckResourceAttrSet(ydbDatabaseServerlessDataSource, "database_id"),
					resource.TestCheckResourceAttr(ydbDatabaseServerlessDataSource, "name", databaseName),
					resource.TestCheckResourceAttr(ydbDatabaseServerlessDataSource, "description", databaseDesc),
					resource.TestCheckResourceAttrSet(ydbDatabaseServerlessDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(ydbDatabaseServerlessDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexYDBDatabaseServerless_byName(t *testing.T) {
	t.Parallel()

	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-serverless")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-serverless-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexYDBDatabaseServerlessByName(databaseName, databaseDesc),
				Check: resource.ComposeTestCheckFunc(
					testYandexYDBDatabaseServerlessExists(ydbDatabaseServerlessDataSource, &database),
					resource.TestCheckResourceAttrSet(ydbDatabaseServerlessDataSource, "database_id"),
					resource.TestCheckResourceAttr(ydbDatabaseServerlessDataSource, "name", databaseName),
					resource.TestCheckResourceAttr(ydbDatabaseServerlessDataSource, "description", databaseDesc),
					resource.TestCheckResourceAttrSet(ydbDatabaseServerlessDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(ydbDatabaseServerlessDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexYDBDatabaseServerless_full(t *testing.T) {
	t.Parallel()

	var database ydb.Database
	params := testYandexYDBDatabaseServerlessParameters{}
	params.name = acctest.RandomWithPrefix("tf-ydb-database-serverless")
	params.desc = acctest.RandomWithPrefix("tf-ydb-database-serverless-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-ydb-database-serverless-label")
	params.labelValue = acctest.RandomWithPrefix("tf-ydb-database-serverless-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexYDBDatabaseServerlessDataSource(params),
				Check: resource.ComposeTestCheckFunc(
					testYandexYDBDatabaseServerlessExists(ydbDatabaseServerlessDataSource, &database),
					resource.TestCheckResourceAttr(ydbDatabaseServerlessDataSource, "name", params.name),
					resource.TestCheckResourceAttr(ydbDatabaseServerlessDataSource, "description", params.desc),
					resource.TestCheckResourceAttrSet(ydbDatabaseServerlessDataSource, "folder_id"),
					testYandexYDBDatabaseServerlessContainsLabel(&database, params.labelKey, params.labelValue),
					testAccCheckCreatedAtAttr(ydbDatabaseServerlessDataSource),
				),
			},
		},
	})
}

func testYandexYDBDatabaseServerlessByID(name string, desc string) string {
	return fmt.Sprintf(`
data "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  database_id = "${yandex_ydb_database_serverless.test-ydb-database-serverless.id}"
}

resource "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  name        = "%s"
  description = "%s"
}`, name, desc)
}

func testYandexYDBDatabaseServerlessByName(name string, desc string) string {
	return fmt.Sprintf(`
data "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  name = "${yandex_ydb_database_serverless.test-ydb-database-serverless.name}"
}

resource "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  name        = "%s"
  description = "%s"
}
`, name, desc)
}

func testYandexYDBDatabaseServerlessDataSource(params testYandexYDBDatabaseServerlessParameters) string {
	return fmt.Sprintf(`
data "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  database_id = "${yandex_ydb_database_serverless.test-ydb-database-serverless.id}"
}

resource "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
}
`,
		params.name,
		params.desc,
		params.labelKey,
		params.labelValue)
}
