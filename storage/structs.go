package storage

type Group struct {
	ID      string   `bson:"ID"`
	Servers []string `bson:"Servers"`
	Owners  []string `bson:"Owners"`
}
