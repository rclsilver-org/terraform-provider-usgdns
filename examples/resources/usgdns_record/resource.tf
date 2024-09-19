# Manage example record.
resource "usgdns_record" "example" {
  name   = "example.com"
  target = "127.0.0.1"
}
