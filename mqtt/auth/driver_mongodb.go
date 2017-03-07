package auth

import (
	"chatterbox/mqtt/connections"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	log "github.com/cihub/seelog"
	"github.com/pborman/uuid"
)

type MongoDriver struct {
	collection *mgo.Collection
}

func newMongoDriver() MongoDriver {
	rd := MongoDriver{}

	session, err := connections.GetMongoSession()
	if err != nil {
		log.Debugf("error in connection to mongo")
	}

	rd.collection = session.DB("chatterbox").C("user")

	return rd
}

func (r MongoDriver) MakeID() string {
	return uuid.New()
}

func (r MongoDriver) Register(u *User) {
	nu := u
	nu.Password = u.PassHash()
	r.collection.Insert(u)

	return
}

func (r MongoDriver) CheckUsername(username string) (exist bool, err error) {
	c, err := r.collection.Find(bson.M{"username": username}).Count()
	if err != nil {
		log.Debugf("%s", err.Error())
		return false, err
	}
	if c > 0 {
		return true, nil
	}
	return false, nil
}

func (r MongoDriver) GetByUsername(username string) (u *User, err error) {
	u = NewUserHandler()
	r.collection.Find(bson.M{"username": username}).One(&u)

	return u, nil
}

func (r MongoDriver) DeleteByUsername(username string) (deleted bool, err error) {
	r.collection.Remove(bson.M{"username": username})

	return true, nil
}
