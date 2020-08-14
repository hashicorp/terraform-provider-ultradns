## 0.1.1 (Unreleased)

ENHANCEMENTS:
* Upgrade to Terraform 0.12. There should be no breaking changes from a practitioner's perspective.

BUG FIXES:
* resource/ultradns_tcpool: Removed unsupported keys `availableToServe`.
* resource/ultradns_probe_ping: Resolved compatibility errors for the specified resource. 
* resource/ultradns_probe_http: Resolved compatibility errors for the specified resource.

## 0.1.0 (June 21, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
