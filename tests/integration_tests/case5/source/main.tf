
provider "aws" {
  region = "us-east-1"
}

resource "aws_kinesis_stream" "this" {
  name = "case5"

  shard_count = 1
}