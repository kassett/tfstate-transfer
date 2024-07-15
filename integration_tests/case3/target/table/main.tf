
resource "aws_dynamodb_table" "this" {
  name = var.name
  hash_key = "key"

  attribute {
    name = "key"
    type = "S"
  }

  read_capacity = 90
  write_capacity = 90
}