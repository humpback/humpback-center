package storage

import "gopkg.in/mgo.v2"
import "github.com/humpback/humpback-center/models"

func (datastorage *DataStorage) GetGroups() []*models.Group {

	groups := []*models.Group{}
	datastorage.M(C_GROUPS, func(c *mgo.Collection) error {
		return c.Find(M{"IsCluster": true}).All(&groups)
	})
	return groups
}

func (datastorage *DataStorage) GetGroup(groupid string) *models.Group {

	group := models.Group{}
	if err := datastorage.M(C_GROUPS, func(c *mgo.Collection) error {
		return c.Find(M{"IsCluster": true, "ID": groupid}).One(&group)
	}); err != nil {
		return nil
	}
	return &group
}
