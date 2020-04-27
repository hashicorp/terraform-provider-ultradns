<!-- archived-provider -->
This Terraform provider is archived, per our [provider archiving process](https://terraform.io/docs/internals/archiving.html). What does this mean?

1. The code repository and all commit history will still be available.
1. Existing released binaries will remain available on the releases site.
1. Documentation will remain on the Terraform website.
1. Issues and pull requests are not being monitored.
1. New releases will not be published.

If anyone from the community or an interested third party is willing to maintain it, they can either be granted access here or fork the repository for their own uses. If you are interested in maintaining this provider, please reach out to the [Terraform Provider Development Program](https://www.terraform.io/guides/terraform-provider-development-program.html) at *terraform-provider-dev@hashicorp.com*.

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
$ git clone https://github.com/terraform-providers/terraform-provider-ultradns.git 
```

Enter the provider directory and build the provider

```sh
$ go mod tidy
$ go build -o terraform-provider-ultradns
```

Using the provider
----------------------
## Fill in for each provider

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.14+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `go build -o terraform-provider-ultradns`. This will build the provider and put the provider binary in the `terraform-provider-ultradns/` directory.

```sh
...
$ terraform-provider-ultradns/terraform-provider-ultradns
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ export TF_VAR_ULTRADNS_USERNAME='***********'
$ export TF_VAR_ULTRADNS_PASSWORD='***********'
$ export TF_VAR_ULTRADNS_BASEURL='RestAPI path'
$ make testacc
```

In order to add the compiled plugin to terraform, you can simply run the following:
*Note:* "{terraform_project_directory}" is the directory where actual project is written to be applied by terraform.
```sh
$ cp terraform-provider-ultradns/terraform-provider-ultradns ~/.terraform.d/plugins
$ cd ${terraform_project_directory}/
$ terraform init
$ terraform validate
```
