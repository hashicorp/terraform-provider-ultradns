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
## Fill in for each provider

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.14+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `go build -o terraform-provider-ultradns`. This will build the provider and put the provider binary in current directory.


In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

- *Note:* Acceptance tests create real resources, and often cost money to run.

- *Note:* "{terraform_plugin_directory}" is the `terraform.d` directory where we will place the binaries

- *Note:*  The test domain specified in TF_VAR_ULTRADNS_DOMAINNAME must already be present at UltraDNS before running Acceptance Test Suite

```sh
$ cp terraform-provider-ultradns ${terraform_plugin_directory}/plugins
$ export TF_VAR_ULTRADNS_USERNAME='***********'
$ export TF_VAR_ULTRADNS_PASSWORD='***********'
$ export TF_VAR_ULTRADNS_BASEURL='https://api.ultradns.com'
$ export TF_VAR_ULTRADNS_DOMAINNAME='Domain Name'
$ make testacc
```

In order to add the compiled plugin to terraform, you can simply run the following:

- *Note:* "{terraform_project_directory}" is the directory where actual project is written to be applied by terraform.


- *Note:* "{terraform_plugin_directory}" is the `terraform.d` directory where we will place the binaries
```sh
$ cp terraform-provider-ultradns ${terraform_plugin_directory}/plugins
$ cd ${terraform_project_directory}/
$ terraform init
$ terraform validate
```
