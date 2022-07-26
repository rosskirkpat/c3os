package role

import (
	"github.com/c3os-io/c3os/pkg/config"

	providerConfig "github.com/c3os-io/c3os/internal/provider/config"
	utils "github.com/mudler/edgevpn/pkg/utils"

	service "github.com/mudler/edgevpn/api/client/service"
)

func contains(slice []string, elem string) bool {
	for _, s := range slice {
		if elem == s {
			return true
		}
	}
	return false
}
func Auto(cc *config.Config, pconfig *providerConfig.Config) Role {
	return func(c *service.RoleConfig) error {
		advertizing, _ := c.Client.AdvertizingNodes()
		actives, _ := c.Client.ActiveNodes()

		c.Logger.Info("Active nodes:", actives)
		c.Logger.Info("Advertizing nodes:", advertizing)

		if len(advertizing) < 2 {
			c.Logger.Info("Not enough nodes")
			return nil
		}

		// first get available nodes
		nodes := advertizing
		shouldBeLeader := utils.Leader(advertizing)

		lead, _ := c.Client.Get("auto", "leader")

		// From now on, only the leader keeps processing
		// TODO: Make this more reliable with consensus
		if shouldBeLeader != c.UUID && lead != c.UUID {
			c.Logger.Infof("<%s> not a leader, leader is '%s', sleeping", c.UUID, shouldBeLeader)
			return nil
		}

		if shouldBeLeader == c.UUID && (lead == "" || !contains(nodes, lead)) {
			if err := c.Client.Set("auto", "leader", c.UUID); err != nil {
				c.Logger.Error(err)
				return err
			}
			c.Logger.Info("Announcing ourselves as leader, backing off")
			return nil
		}

		if lead != c.UUID {
			c.Logger.Info("Backing off, as we are not currently flagged as leader")
			return nil
		}

		return scheduleRoles(nodes, c, cc, pconfig)
	}
}
