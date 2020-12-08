<!-- archived-provider -->
Please note: This Terraform provider is archived, per our [provider archiving process](https://terraform.io/docs/internals/archiving.html). What does this mean?
1. The code repository and all commit history will still be available.
1. Existing released binaries will remain available on the releases site.
1. Issues and pull requests are not being monitored.
1. New releases will not be published.

If anyone from the community or an interested third party is willing to maintain it, they can fork the repository and [publish it](https://www.terraform.io/docs/registry/providers/publishing.html) to the Terraform Registry. If you are interested in maintaining this provider, please reach out to the [Terraform Provider Development Program](https://www.terraform.io/guides/terraform-provider-development-program.html) at *terraform-provider-dev@hashicorp.com*.

---

<!-- /archived-provider -->

Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.11.x or higher
-	[Go](https://golang.org/doc/install) 1.14 or higher (to build the provider plugin)

Building The Provider
---------------------

Clone repository in HOME directory

```sh
$ git clone https://github.com/terraform-providers/terraform-provider-ultradns.git terraform-provider-ultradns
```

Enter the provider directory and build the provider

```sh
$ go mod tidy
$ go mod vendor
$ make fmt
$ go build -o terraform-provider-ultradns
```
Using the provider
----------------------
*Note:* This will be utilized as a third-party plugin, until officially released by Hashicorp 
- Download the plugin binary (terraform-provider-ultradns_v0.X.X.zip) from the release assets
- Unzip the plugin binary
```$> unzip terraform-provider-ultradns_v0.X.X.zip```
- Move the plugin to appropriate (third-party plugin) directory
```$> mv terraform-provider-ultradns_v0.X.X ~/.terraform.d/plugins/```
- Remove the older terraform plugin, if it exists
```$> rm -f .terraform/plugins/<OS>_<ARCH>/terraform-provider-ultradns*```
- Update main.tf to use the provider plugin as intended WITH the desired ultradns provider "version"

	```
	provider "ultradns" {
	  version  = "~>0.X.X"
	  username = "${var.ULTRADNS_USERNAME}"
	  password = "${var.ULTRADNS_PASSWORD}"
	  baseurl  = "${var.ULTRADNS_BASEURL}"
	}
	```
- Initialize the plugin using `terraform init` command

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.14+ is **required**). You'll also need to setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`

To compile the provider, run `go build -o terraform-provider-ultradns`. This will build the provider and put the provider binary in current directory


In order to test the provider, you can simply run `make test`

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`

- *Note:* Acceptance tests create real resources, and often cost money to run

- *Note:* "{terraform_plugin_directory}" is the `terraform.d` directory where we will place the binaries

- *Note:*  The test domain specified in ULTRADNS_DOMAINNAME must already be present at UltraDNS before running Acceptance Test Suite

```sh
$ cp terraform-provider-ultradns_v0.X.X ${terraform_plugin_directory}/plugins
$ export ULTRADNS_USERNAME='***********'
$ export ULTRADNS_PASSWORD='***********'
$ export ULTRADNS_BASEURL='https://api.ultradns.com'
$ export ULTRADNS_DOMAINNAME='Domain Name'
$ make testacc
```

In order to add the compiled plugin to terraform, you can simply run the following:

- *Note:* "{terraform_project_directory}" is the directory where the actual project is written to be applied by terraform


- *Note:* "{terraform_plugin_directory}" is the `terraform.d` directory where we will place the binaries
```sh
$ cp terraform-provider-ultradns_v0.X.X ${terraform_plugin_directory}/plugins
$ cd ${terraform_project_directory}/
$ terraform init
$ terraform validate
```
