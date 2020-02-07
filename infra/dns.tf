resource "google_dns_managed_zone" "gcp_norix" {
  name        = "gcp-norix-zone"
  dns_name    = "gcp.norix.co.uk."
  description = "DNS Zone for gcp.norix.co.uk"
}