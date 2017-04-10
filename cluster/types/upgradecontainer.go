package types

import "github.com/humpback/humpback-agent/models"

// UpgradeContainerResponse is exported
type UpgradeContainerResponse struct {
	ID string `json:"Id"`
}

// UpgradeContainer is exported
type UpgradeContainer struct {
	IP       string `json:"IP"`
	HostName string `json:"HostName"`
	models.Container
}

// UpgradeContainers is exported
type UpgradeContainers []*UpgradeContainer

// SetUpgradePair is exported
func (upgrade UpgradeContainers) SetUpgradePair(ip string, hostname string, container models.Container) UpgradeContainers {

	upgradeContainer := &UpgradeContainer{
		IP:        ip,
		HostName:  hostname,
		Container: container,
	}
	upgrade = append(upgrade, upgradeContainer)
	return upgrade
}
