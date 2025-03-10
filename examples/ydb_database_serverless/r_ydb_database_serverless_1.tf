//
// Create a new YDB Serverless Database.
//
resource "yandex_ydb_database_serverless" "my_ydb" {
  name      = "test-ydb-serverless"
  folder_id = data.yandex_resourcemanager_folder.test_folder.id

  deletion_protection = true
}
