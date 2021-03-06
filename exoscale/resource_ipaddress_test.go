package exoscale

import (
	"fmt"
	"testing"

	"github.com/exoscale/egoscale"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccElasticIP(t *testing.T) {
	eip := new(egoscale.IPAddress)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckElasticIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticIPCreate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticIPExists("exoscale_ipaddress.eip", eip),
					testAccCheckElasticIPAttributes(eip),
					testAccCheckElasticIPCreateAttributes(EXOSCALE_ZONE),
				),
			},
			{
				Config: testAccElasticIPUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticIPExists("exoscale_ipaddress.eip", eip),
					testAccCheckElasticIPAttributes(eip),
					testAccCheckElasticIPCreateAttributes(EXOSCALE_ZONE),
				),
			},
		},
	})
}

func testAccCheckElasticIPExists(n string, eip *egoscale.IPAddress) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No elastic IP ID is set")
		}

		id, err := egoscale.ParseUUID(rs.Primary.ID)
		if err != nil {
			return err
		}

		client := GetComputeClient(testAccProvider.Meta())
		eip.ID = id
		if err := client.Get(eip); err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckElasticIPAttributes(eip *egoscale.IPAddress) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if eip.IPAddress == nil {
			return fmt.Errorf("eip IP address is nil")
		}

		return nil
	}
}

func testAccCheckElasticIPCreateAttributes(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "exoscale_ipaddress" {
				continue
			}

			if rs.Primary.Attributes["zone"] != name {
				continue
			}

			if rs.Primary.Attributes["ip_address"] == "" {
				return fmt.Errorf("Elastic IP: expected ip address to be set")
			}

			return nil
		}

		return fmt.Errorf("Could not find elastic ip %s", name)
	}
}

func testAccCheckElasticIPDestroy(s *terraform.State) error {
	client := GetComputeClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "exoscale_ipaddress" {
			continue
		}

		id, err := egoscale.ParseUUID(rs.Primary.ID)
		if err != nil {
			return err
		}

		key := &egoscale.IPAddress{
			ID:        id,
			IsElastic: true,
		}
		if err := client.Get(key); err != nil {
			if r, ok := err.(*egoscale.ErrorResponse); ok {
				if r.ErrorCode == egoscale.ParamError {
					return nil
				}
			}
			return err
		}
		return fmt.Errorf("ipAddress: %#v still exists", key)
	}
	return nil
}

var testAccElasticIPCreate = fmt.Sprintf(`
resource "exoscale_ipaddress" "eip" {
  zone = %q
  tags {
    test = "acceptance"
  }
}
`,
	EXOSCALE_ZONE,
)

var testAccElasticIPUpdate = fmt.Sprintf(`
resource "exoscale_ipaddress" "eip" {
  zone = %q
}
`,
	EXOSCALE_ZONE,
)
