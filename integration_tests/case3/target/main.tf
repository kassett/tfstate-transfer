
provider "aws" {
  region = "us-east-1"
}

locals {
  secret_names = ["1", "2", "3"]
  foreach_secrets = {
    "1" = "one",
    "2" = "two"
  }
}

module "table_simple" {
  source = "./table"
  name = "case3"
}

module "table_count" {
  count = length(local.secret_names)
  source = "./table"
  name = "case3-${local.secret_names[count.index]}"
}

module "table_foreach" {
  for_each = local.foreach_secrets
  source = "./table"
  name = "case3-${each.value}"
}