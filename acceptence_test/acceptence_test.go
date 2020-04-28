package ultradns

import (
	"github.com/gruntwork-io/terratest/modules/terraform"
	"testing"
)

// Testing complete ultradns by creating resources using terraform project using Terratest.
func TestUltraDNSPlugin(t *testing.T) {
	t.Parallel()

	terraformOptions := &terraform.Options{
		// website::tag::1::Set the path to the Terraform code that will be tested.
		// The path to where our Terraform code is located
		TerraformDir: "../acceptence_test/",

		// Disable colors in Terraform commands so its easier to parse stdout/stderr
		NoColor: true,
	}

	// website::tag::4::Clean up resources with "terraform destroy". Using "defer" runs the command at the end of the test, whether the test succeeds or fails.
	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	// defer terraform.Destroy(t, terraformOptions)
	defer terraform.RunTerraformCommandE(t, terraformOptions, "destroy", "-input=false", "-force")
	// website::tag::2::Run "terraform init" and "terraform apply".
	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

}
