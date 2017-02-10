package storage

import "gopkg.in/mgo.v2"

func (datastorage *DataStorage) GetGroups() ([]*Group, error) {

	groups := []*Group{}
	if err := datastorage.M(C_GROUPS, func(c *mgo.Collection) error {
		return c.Find(M{"IsCluster": true}).All(&groups)
	}); err != nil {
		return groups, err
	}
	return groups, nil
}

func (datastorage *DataStorage) GetGroup(groupid string) (*Group, error) {

	group := Group{}
	if err := datastorage.M(C_GROUPS, func(c *mgo.Collection) error {
		return c.Find(M{"IsCluster": true, "ID": groupid}).One(&group)
	}); err != nil {
		return nil, err
	}
	return &group, nil
}
