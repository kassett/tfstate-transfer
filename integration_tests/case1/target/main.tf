
provider "aws" {
  region = "us-east-1"
}

resource "aws_secretsmanager_secret" "this" {
  name_prefix = "key-case1"
}

resource "aws_secretsmanager_secret_version" "this" {
  secret_id = aws_secretsmanager_secret.this.id
  secret_string = "value"
}

resource "aws_dynamodb_table" "this" {
  name = "table-case1"
  hash_key = "key"

  attribute {
    name = "key"
    type = "S"
  }

  read_capacity = 90
  write_capacity = 90
}