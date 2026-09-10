package teamcity

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"regexp"
	"testing"
)

func TestAccAuth_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_auth_settings" "test" {
  allow_guest         = false
  guest_username      = "guest"
  welcome_text        = ""
  collapse_login_form = false
  # two_factor_mode     = "OPTIONAL"
  per_project_permissions = false
  email_verification  = false

  modules = {
    token = {}
	basic_http = {}
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_auth_settings.test", "allow_guest", "false"),
				),
			},
		},
	})
}

// TestAccAuth_otherModule covers the 'modules.others' escape hatch used for
// auth modules with no dedicated typed attribute (e.g. plugin-provided modules
// such as the SAML plugin's "HTTP-SAML.v2"). Without it, any module not modeled
// by a typed attribute is silently dropped from the server on every apply,
// since SetAuthSettings() PUTs the whole module list.
func TestAccAuth_otherModule(t *testing.T) {
	config := providerConfig + `
resource "teamcity_auth_settings" "test" {
  allow_guest              = false
  guest_username           = "guest"
  welcome_text             = ""
  collapse_login_form      = false
  per_project_permissions  = false
  email_verification       = false

  modules = {
    token      = {}
    basic_http = {}

    others = [
      {
        name = "HTTP-SAML.v2"
        properties = {
          someProperty = "someValue"
        }
      }
    ]
  }
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_auth_settings.test", "modules.others.0.name", "HTTP-SAML.v2"),
					resource.TestCheckResourceAttr("teamcity_auth_settings.test", "modules.others.0.properties.someProperty", "someValue"),
				),
			},
			// A second plan against the same config must be empty: the module
			// added via 'others' must not keep drifting/re-appearing/disappearing.
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccAuth_otherModuleNoProperties covers the common case where an 'others'
// entry has no properties at all (e.g. the real-world "HTTP-SAML.v2" module,
// which has no auth-config.xml properties of its own — its actual settings
// live in the SAML plugin's own storage, not here). This must produce an empty
// plan on a second run: an omitted 'properties' (null) must not perpetually
// diff against what the server returns for a module with no properties
// (empty map) — otherwise every apply looks like there's a real change.
func TestAccAuth_otherModuleNoProperties(t *testing.T) {
	config := providerConfig + `
resource "teamcity_auth_settings" "test" {
  allow_guest              = false
  guest_username           = "guest"
  welcome_text             = ""
  collapse_login_form      = false
  per_project_permissions  = false
  email_verification       = false

  modules = {
    token      = {}
    basic_http = {}

    others = [
      {
        name = "HTTP-SAML.v2"
      }
    ]
  }
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_auth_settings.test", "modules.others.0.name", "HTTP-SAML.v2"),
				),
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccAuth_otherModuleCollision ensures a module already covered by a
// dedicated typed attribute cannot also be declared under 'others', to avoid
// two config paths fighting over the same TeamCity auth-module entry.
func TestAccAuth_otherModuleCollision(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_auth_settings" "test" {
  allow_guest              = false
  guest_username           = "guest"
  welcome_text             = ""
  collapse_login_form      = false
  per_project_permissions  = false
  email_verification       = false

  modules = {
    token = {}

    others = [
      {
        name = "Token-Auth"
      }
    ]
  }
}
`,
				ExpectError: regexp.MustCompile("already has a dedicated attribute"),
			},
		},
	})
}
