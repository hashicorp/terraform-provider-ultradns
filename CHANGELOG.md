## 0.2.0 (Unreleased)

ENHANCEMENTS:
* Added feature for Terraform Import in all of the resources. There should be no breaking changes from a practitioner's perspective. ([#16](https://github.com/terraform-providers/terraform-provider-ultradns/issues/16))
* Added Unit Testcases in all the resources which weren't present before.
* Changed the delimiter in resource ID generator from `.` to `:`
* It is now compatible with latest ultradns-sdk-go plugin.
* Enhanced ".travis.yml" file to support code coverage.
* Updated "README" file to support current changes.
* Updated "GNUMake" file for code coverage and testing.

BUG FIXES:
* resource/ultradns_rdpool: Resolved the acceptance test which was failing due to unsupported fields.
* resource/ultradns_dirpool: Resolved the acceptance failing due to no `TTL` field in DirPoolProfile DTO.
* provider: Removed static "BASEURL" used while running acceptance test.
* Added "domain" variable in acceptance tests of all resources to use custom domains.

## 0.1.1 (Unreleased)

ENHANCEMENTS:
* Upgrade to Terraform 0.12. There should be no breaking changes from a practitioner's perspective.

BUG FIXES:
* resource/ultradns_tcpool: Removed unsupported key `availableToServe`.
* resource/ultradns_probe_ping: Resolved compatibility errors for the specified resource. 
* resource/ultradns_probe_http: Resolved compatibility errors for the specified resource.

## 0.1.0 (June 21, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
