package aadgraph_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/terraform-providers/terraform-provider-azuread/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/services/aadgraph/graph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
)

const testCertificateServicePrincipal string = `-----BEGIN CERTIFICATE-----
MIIDGjCCAgICCQDAQlCA1jw1BjANBgkqhkiG9w0BAQsFADBPMQswCQYDVQQGEwJV
UzELMAkGA1UECAwCQ0ExFzAVBgNVBAoMDkhhc2hpQ29ycCwgSW5jMRowGAYDVQQD
DBFoYXNoaWNvcnB0ZXN0LmNvbTAeFw0yMDA1MzEyMDI2MTFaFw0yMTA1MzEyMDI2
MTFaME8xCzAJBgNVBAYTAlVTMQswCQYDVQQIDAJDQTEXMBUGA1UECgwOSGFzaGlD
b3JwLCBJbmMxGjAYBgNVBAMMEWhhc2hpY29ycHRlc3QuY29tMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAmVsolhGq5SKLB4OJrQqlBmqOMTZLNKeAved5
f0pNmr9Sdb7VapA8vVALENJq0XWDsY5blrsDevYtPVh3ZXKfZpmJwbtcy5ZJ+B6B
HVXZHO2Ep3RzZ/bFOxXXvmtUGmOpJxJyHlXao4mz2LuNHUa6GJqDjVaZ3ZL2EjFa
Q2nRIorIYoERHxEcogpQqF0csL70hi8Ho/dFpfsKhooYfzWRgj4KncMbjWYb82L6
qPhnonETOGHgpqXP7FhGLkm0pfWL1oz21hZRWRVh6+RjBQ4gm3iRrzffyjvvdVL8
molsNguVYF5Km7D/oRTILImPgXLlKnB/XRpORvhDK4zw8ewiewIDAQABMA0GCSqG
SIb3DQEBCwUAA4IBAQB7XGxsoU1DF7ECvuR5T9cPrW60DxjPP7uEp2UQeEZgkCDD
WpAqmhXrtn2fRpKVBCFzTVZZKW3/f8W3ivGR9hnVTmPtvjG9n9wLq997k8h2GfD8
z4YaA8vEwqmrMYX8azeTGL/JQHFIleA6YPokdybgP6aC9hIWdSVXdV4G5kgN3/GV
lWOie5Wf9IZotaDzExG7P38mGzQOLtZVCnIxGyy6/Q1E5n5PUxc/9i/iY6xlC2Ih
HraQzsK7BNxC5NSwwirT95JH+Xd8rvWu+bCveJz3mnZ3sgolCoxL6Hv1uD2UOZb5
rCHdW31vp5PYNJaSkYL0j259Ogb8crkIzDr3Z8YF
-----END CERTIFICATE-----`

func testCheckServicePrincipalKeyExists(name string) resource.TestCheckFunc { //nolint unparam
	return func(s *terraform.State) error {
		client := acceptance.AzureADProvider.Meta().(*clients.AadClient).AadGraph.ServicePrincipalsClient
		ctx := acceptance.AzureADProvider.Meta().(*clients.AadClient).StopContext

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		id, err := graph.ParseCredentialId(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("parsing Service Principal Key Credential ID: %v", err)
		}
		resp, err := client.Get(ctx, id.ObjectId)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Service Principal %q does not exist", id.ObjectId)
			}
			return fmt.Errorf("Bad: Get on ServicePrincipalsClient: %+v", err)
		}

		credentials, err := client.ListKeyCredentials(ctx, id.ObjectId)
		if err != nil {
			return fmt.Errorf("listing Key Credentials for Service Principal %q: %+v", id.ObjectId, err)
		}

		cred := graph.KeyCredentialResultFindByKeyId(credentials, id.KeyId)
		if cred != nil {
			return nil
		}

		return fmt.Errorf("Key Credential %q was not found in Service Principal %q", id.KeyId, id.ObjectId)
	}
}

func testCheckServicePrincipalKeyCheckDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		client := acceptance.AzureADProvider.Meta().(*clients.AadClient).AadGraph.ServicePrincipalsClient
		ctx := acceptance.AzureADProvider.Meta().(*clients.AadClient).StopContext

		if rs.Type != "azuread_service_principal_certificate" {
			continue
		}

		id, err := graph.ParseCredentialId(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("parsing Service Principal Credential ID: %v", err)
		}

		resp, err := client.Get(ctx, id.ObjectId)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return nil
			}

			return err
		}

		return fmt.Errorf("Service Principal Key Credential still exists:\n%#v", resp)
	}

	return nil
}

func TestAccServicePrincipalCertificate_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_service_principal_certificate", "test")
	keyType := "AsymmetricX509Cert"
	endDate := time.Now().AddDate(0, 6, 0).UTC().Format(time.RFC3339)
	value := testCertificateServicePrincipal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckServicePrincipalKeyCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServicePrincipalCertificate_basic(data.RandomInteger, keyType, endDate, value),
				Check: resource.ComposeTestCheckFunc(
					testCheckServicePrincipalKeyExists(data.ResourceName),
				),
			},
			data.ImportStep("value"),
		},
	})
}

func TestAccServicePrincipalCertificate_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_service_principal_certificate", "test")
	keyType := "AsymmetricX509Cert"
	startDate := time.Now().AddDate(0, 0, 7).UTC().Format(time.RFC3339)
	endDate := time.Now().AddDate(0, 6, 0).UTC().Format(time.RFC3339)
	value := testCertificateServicePrincipal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckServicePrincipalKeyCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServicePrincipalCertificate_complete(data.RandomInteger, data.RandomID, keyType, startDate, endDate, value),
				Check: resource.ComposeTestCheckFunc(
					testCheckServicePrincipalKeyExists(data.ResourceName),
				),
			},
			data.ImportStep("end_date_relative", "value"),
		},
	})
}

func TestAccServicePrincipalCertificate_relativeEndDate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_service_principal_certificate", "test")
	keyType := "AsymmetricX509Cert"
	value := testCertificateServicePrincipal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckServicePrincipalKeyCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServicePrincipalCertificate_relativeEndDate(data.RandomInteger, keyType, value),
				Check: resource.ComposeTestCheckFunc(
					// can't assert on Value since it's not returned
					testCheckServicePrincipalKeyExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "end_date"),
				),
			},
			data.ImportStep("end_date_relative", "value"),
		},
	})
}

func TestAccServicePrincipalCertificate_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_service_principal_certificate", "test")
	keyType := "AsymmetricX509Cert"
	endDate := time.Now().AddDate(0, 6, 0).UTC().Format(time.RFC3339)
	value := testCertificateServicePrincipal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckServicePrincipalKeyCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServicePrincipalCertificate_basic(data.RandomInteger, keyType, endDate, value),
				Check: resource.ComposeTestCheckFunc(
					testCheckServicePrincipalKeyExists(data.ResourceName),
				),
			},
			data.RequiresImportErrorStep(testAccServicePrincipalCertificate_requiresImport(data.RandomInteger, keyType, endDate, value)),
		},
	})
}

func testAccServicePrincipalCertificate_template(ri int) string {
	return fmt.Sprintf(`
resource "azuread_application" "test" {
  name = "acctestApp-%d"
}

resource "azuread_service_principal" "test" {
  application_id = azuread_application.test.application_id
}
`, ri)
}

func testAccServicePrincipalCertificate_basic(ri int, keyType, endDate, value string) string {
	return fmt.Sprintf(`
%s

resource "azuread_service_principal_certificate" "test" {
  service_principal_id = azuread_service_principal.test.id
  type                 = "%s"
  end_date             = "%s"
  value                = <<EOT
%s
EOT
}
`, testAccServicePrincipalCertificate_template(ri), keyType, endDate, value)
}

func testAccServicePrincipalCertificate_complete(ri int, keyId, keyType, startDate, endDate, value string) string {
	return fmt.Sprintf(`
%s

resource "azuread_service_principal_certificate" "test" {
  service_principal_id = azuread_service_principal.test.id
  key_id               = "%s"
  type                 = "%s"
  start_date           = "%s"
  end_date             = "%s"
  value                = <<EOT
%s
EOT
}
`, testAccServicePrincipalCertificate_template(ri), keyId, keyType, startDate, endDate, value)
}

func testAccServicePrincipalCertificate_relativeEndDate(ri int, keyType, value string) string {
	return fmt.Sprintf(`
%s

resource "azuread_service_principal_certificate" "test" {
  service_principal_id = azuread_service_principal.test.id
  end_date_relative    = "4320h"
  type                 = "%s"
  value                = <<EOT
%s
EOT
}
`, testAccServicePrincipalCertificate_template(ri), keyType, value)
}

func testAccServicePrincipalCertificate_requiresImport(ri int, keyType, endDate, value string) string {
	template := testAccServicePrincipalCertificate_basic(ri, keyType, endDate, value)
	return fmt.Sprintf(`
%s

resource "azuread_service_principal_certificate" "import" {
  service_principal_id = azuread_service_principal_certificate.test.service_principal_id
  key_id               = azuread_service_principal_certificate.test.key_id
  type                 = azuread_service_principal_certificate.test.type
  end_date             = azuread_service_principal_certificate.test.end_date
  value                = azuread_service_principal_certificate.test.value
}
`, template)
}
