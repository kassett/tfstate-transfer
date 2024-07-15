
provider "aws" {
  region = "us-east-1"
}

locals {
  secret_names = ["1", "2", "3"]
  foreach_secrets = {
    "one" = "1",
    "two" = "2"
  }
}


resource "aws_secretsmanager_secret" "iterate_count" {
  count = length(local.secret_names)
  name_prefix = "key-case2-${local.secret_names[count.index]}"
}

resource "aws_secretsmanager_secret_version" "iterate_count" {
  count = length(local.secret_names)

  secret_id = aws_secretsmanager_secret.iterate_count[count.index].id
  secret_string = "value"
}

resource "aws_secretsmanager_secret" "iterate_foreach" {
  for_each = local.foreach_secrets
  name_prefix = "key-case2-${each.value}"
}

resource "aws_secretsmanager_secret_version" "iterate_foreach" {
  for_each = local.foreach_secrets
  secret_id = aws_secretsmanager_secret.iterate_foreach[each.key].id
  secret_string = "value"
}

resource "aws_dynamodb_table" "this" {
  name = "table-case2"
  hash_key = "key"

  attribute {
    name = "key"
    type = "S"
  }

  read_capacity = 90
  write_capacity = 90
}