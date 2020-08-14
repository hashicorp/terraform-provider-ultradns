## 0.1.1 (Unreleased)

ENHANCEMENTS
* provider: Upgrade to Terraform 0.12. There should be no breaking changes from a practitioner's perspective.
* Ensured that the advanced pool are working which were previously not working

TESTING THE PLUGIN LOCALLY
* Download the plugin binary from the release.
* Unzip the plugin binary
``sh 
$> unzip terraform-provider-ultradns_0.1.1_RC.zip
``
* Rename the unzipped plugin to "terraform-provider-ultradns"
```sh
$> mv terraform-provider-ultradns_0.1.1_RC terraform-provider-ultradns
```
* Move the renamed plugin to appropriate directory
```
$> mv terraform-provider-ultradns ~/.terraform.d/plugins/
```
* Remove the older terraform plugin
```sh
$> rm -f .terraform/plugins/<OS>_<ARCH>/terraform-provider-ultradns_{Version}
```
* Comment the "version" parameter present in the "main.tf" terraform file.
* Change the working directory to where our "main.tf" terraform file lies.
* Finally, initialize the plugin using "terraform init" command.

ROLLBACK PPROCEDURE
* Uncomment  the "version" parameter present in the "main.tf" terraform file.
* Initialize the plugin using "terraform init".


## 0.1.0 (June 21, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
