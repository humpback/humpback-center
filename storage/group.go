package storage

import "gopkg.in/mgo.v2"
import "github.com/humpback/humpback-center/models"

func (datastorage *DataStorage) GetGroups() ([]*models.Group, error) {

	groups := []*models.Group{}
	if err := datastorage.M(C_GROUPS, func(c *mgo.Collection) error {
		return c.Find(M{"IsCluster": true}).All(&groups)
	}); err != nil {
		return groups, err
	}
	return groups, nil
}

func (datastorage *DataStorage) GetGroup(groupid string) (*models.Group, error) {

	group := models.Group{}
	if err := datastorage.M(C_GROUPS, func(c *mgo.Collection) error {
		return c.Find(M{"IsCluster": true, "ID": groupid}).One(&group)
	}); err != nil {
		return nil, err
	}
	return &group, nil
}
