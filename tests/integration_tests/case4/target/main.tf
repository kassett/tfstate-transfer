
resource "local_file" "this" {
  filename = "/tmp/temporaryfile.txt"
  content = "CONTENT"
}