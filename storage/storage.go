package storage

import "gopkg.in/mgo.v2"
import "gopkg.in/mgo.v2/bson"

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

// database and some collections.
const (
	//default database name
	DB_DATABASE = "humpback"
	//groups collection
	C_GROUPS = "groups"
)

var mgoConnectTimeout = time.Second * 10

// bson map
type M bson.M

// bson element
type D bson.D

type DataStorage struct {
	sync.RWMutex
	rawuris     string
	database    string
	maxpollsize int
	session     *mgo.Session
	stopCh      chan struct{}
}

func readArgs(rawuris string) (string, map[string]string) {

	var dbname string
	pstr := strings.SplitN(rawuris, "?", 2)
	if len(pstr) > 0 {
		str := strings.TrimPrefix(pstr[0], "mongodb://")
		prestr := strings.SplitN(str, "/", 2)
		if len(prestr) == 2 {
			dbname = prestr[1]
		}
	}

	if dbname == "" {
		dbname = DB_DATABASE
	}

	opts := make(map[string]string)
	if len(pstr) == 2 {
		pairs := strings.SplitN(pstr[1], "&", -1)
		if len(pairs) > 0 {
			for _, it := range pairs {
				pair := strings.Split(it, "=")
				if len(pair) == 2 {
					opts[pair[0]] = pair[1]
				}
			}
		}
	}
	return dbname, opts
}

func NewDataStorage(rawuris string) (*DataStorage, error) {

	datastorage := &DataStorage{
		rawuris: rawuris,
		stopCh:  make(chan struct{}),
	}

	database, opts := readArgs(rawuris)
	datastorage.database = database
	if value, ret := opts["maxPoolSize"]; ret {
		maxpoolsize, err := strconv.Atoi(value)
		if err == nil && maxpoolsize > 0 {
			datastorage.maxpollsize = maxpoolsize
		}
	}
	return datastorage, nil
}

func (datastorage *DataStorage) Session() (*mgo.Session, error) {

	if datastorage.session == nil {
		session, err := mgo.DialWithTimeout(datastorage.rawuris, mgoConnectTimeout)
		if err != nil {
			return nil, err
		}
		datastorage.session = session
	}
	return datastorage.session.Clone(), nil
}

func (datastorage *DataStorage) Close() {

	close(datastorage.stopCh)
	if datastorage.session != nil {
		datastorage.session.Close()
	}
}

func (datastorage *DataStorage) M(collection string, fn func(*mgo.Collection) error) error {

	session, err := datastorage.Session()
	if err != nil {
		return err
	}
	coll := session.DB(datastorage.database).C(collection)
	err = fn(coll)
	session.Close()
	return err
}
