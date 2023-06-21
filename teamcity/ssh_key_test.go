package teamcity

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccSshKey_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "test"
}

resource "teamcity_ssh_key" "test" {
	project_id = teamcity_project.test.id
	name = "test"
	private_key = <<-EOT
		-----BEGIN OPENSSH PRIVATE KEY-----
		b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtz
		c2gtZWQyNTUxOQAAACC2jINkdjGhM5xQFQGeRxD9d5DL6t45U/fUAQ6q+a8N3QAA
		AIhyMPQocjD0KAAAAAtzc2gtZWQyNTUxOQAAACC2jINkdjGhM5xQFQGeRxD9d5DL
		6t45U/fUAQ6q+a8N3QAAAEBhWHQbULxpi62cazCQoePYoXcP6MWZKXyT+66B5W9/
		GraMg2R2MaEznFAVAZ5HEP13kMvq3jlT99QBDqr5rw3dAAAAAAECAwQF
		-----END OPENSSH PRIVATE KEY-----
	EOT
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_project.test", "name", "test"),
				),
			},
		},
	})
}
