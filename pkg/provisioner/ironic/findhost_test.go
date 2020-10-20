package ironic

import (
	"testing"

	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"

	"github.com/metal3-io/baremetal-operator/pkg/bmc"
	"github.com/metal3-io/baremetal-operator/pkg/provisioner/ironic/clients"
	"github.com/metal3-io/baremetal-operator/pkg/provisioner/ironic/testserver"
)

func TestFindExistingHost(t *testing.T) {
	cases := []struct {
		name   string
		ironic *testserver.IronicMock

		hostName       string
		provisioningID string
		nodeName       string
	}{
		{
			name:           "no-node",
			hostName:       "name",
			provisioningID: "uuid",
			ironic:         testserver.NewIronic(t).NoNode("name").NoNode("uuid"),
		},
		{
			name:           "by-name",
			hostName:       "name",
			provisioningID: "uuid",
			ironic: testserver.NewIronic(t).NoNode("uuid").
				WithNode(nodes.Node{
					Name: "name",
					UUID: "different-uuid",
				}),
			nodeName: "name",
		},
		{
			name:           "by-uuid",
			hostName:       "name",
			provisioningID: "uuid",
			ironic: testserver.NewIronic(t).NoNode("name").
				WithNode(nodes.Node{
					Name: "different-name",
					UUID: "uuid",
				}),
			nodeName: "different-name",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.ironic != nil {
				tc.ironic.Start()
				defer tc.ironic.Stop()
			}

			auth := clients.AuthConfig{Type: clients.NoAuth}

			// Update the default host to match the test settings
			host := makeHost()
			host.ObjectMeta.Name = tc.hostName
			host.Status.Provisioning.ID = tc.provisioningID

			prov, err := newProvisionerWithSettings(host, bmc.Credentials{}, nil,
				tc.ironic.Endpoint(), auth, "https://inspector.test/", auth,
			)
			if err != nil {
				t.Fatalf("could not create provisioner: %s", err)
			}

			node, err := prov.findExistingHost()
			t.Logf("requests: %s", tc.ironic.Requests)
			if err != nil {
				t.Fatalf("could not look up host: %s", err)
			}

			if tc.nodeName == "" && node != nil {
				t.Fatalf("found unexpected node %s (%s)", node.Name, node.UUID)
			}
		})
	}
}