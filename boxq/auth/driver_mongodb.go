package auth

import (
	"chatterbox/boxq/connections"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	log "github.com/cihub/seelog"
	"github.com/pborman/uuid"
)

var authDb = "chatterbox"
var authColUser = "user"
var authColACL = "acl"

type MongoDriver struct {
	authCollection *mgo.Collection
	aclCollection  *mgo.Collection
}

func newMongoDriver() MongoDriver {
	rd := MongoDriver{}

	session, err := connections.GetMongoSession()
	if err != nil {
		log.Debugf("error in connection to mongo")
	}

	rd.authCollection = session.DB(authDb).C(authColUser)
	rd.aclCollection = session.DB(authDb).C(authColACL)

	return rd
}

func (m MongoDriver) MakeID() string {
	return uuid.New()
}

func (m MongoDriver) Register(u *User) {
	u.Password = u.PassHash()
	u.ID = bson.NewObjectId().Hex()

	m.authCollection.Insert(u)

	return
}

func (m MongoDriver) CheckUsername(username string) (exist bool, err error) {
	c, err := m.authCollection.Find(bson.M{"username": username}).Count()
	if err != nil {
		log.Debugf("%s", err.Error())
		return false, err
	}
	if c > 0 {
		return true, nil
	}
	return false, nil
}

func (m MongoDriver) GetByUsername(username string) (u *User, err error) {
	u = NewUserHandler()
	m.authCollection.Find(bson.M{"username": username}).One(&u)

	return u, nil
}

func (m MongoDriver) DeleteByUsername(username string) (deleted bool, err error) {
	m.authCollection.Remove(bson.M{"username": username})

	return true, nil
}

func (m MongoDriver) CheckACL(clientID, topic, acltype string) (actType bool, noMatch bool) {
	var u User

	m.aclCollection.Find(bson.M{"username": clientID}).One(&u)

	for _, m := range u.PubSub {
		if m == topic {
			actType = aclAllow
			return actType, false
		}
	}

	switch acltype {
	case ACLPub:
		for _, m := range u.Publish {
			if m == topic {
				actType = aclAllow
				return actType, false
			}
		}
	case ACLSub:
		for _, m := range u.Subscribe {
			if m == topic {
				actType = aclAllow
				return actType, false
			}
		}
	}
	actType = aclAllow
	return actType, false
}

func (m MongoDriver) SetACL(clientID, topic, acltype string) {
	var key string
	switch acltype {
	case ACLPub:
		key = "publish"
	case ACLSub:
		key = "subscribe"
	case ACLPubSub:
		key = "pubsub"
	}

	selector := bson.M{"username": clientID}
	update := bson.M{"$addToSet": bson.M{key: topic}}
	m.aclCollection.Upsert(selector, update)

}

func (m MongoDriver) RemoveACL(clientID, topic string) {
	selector := bson.M{"username": clientID}
	m.aclCollection.Remove(selector)
}
