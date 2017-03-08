package auth

import (
	"chatterbox/mqtt/connections"

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

func (r MongoDriver) MakeID() string {
	return uuid.New()
}

func (r MongoDriver) Register(u *User) {
	u.Password = u.PassHash()
	u.ID = bson.NewObjectId().Hex()

	r.authCollection.Insert(u)

	return
}

func (r MongoDriver) CheckUsername(username string) (exist bool, err error) {
	c, err := r.authCollection.Find(bson.M{"username": username}).Count()
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
	r.authCollection.Find(bson.M{"username": username}).One(&u)

	return u, nil
}

func (r MongoDriver) DeleteByUsername(username string) (deleted bool, err error) {
	r.authCollection.Remove(bson.M{"username": username})

	return true, nil
}

func (r MongoDriver) CheckACL(clientID, topic, acltype string) bool {
	var u User

	r.aclCollection.Find(bson.M{"username": clientID}).One(&u)

	switch acltype {
	case AclPub:
		for _, m := range u.Publish {
			if m == topic {
				return aclAllow
			}
		}
	case AclSub:
		for _, m := range u.Subscribe {
			if m == topic {
				return aclAllow
			}
		}
	case AclPubSub:
		for _, m := range u.PubSub {
			if m == topic {
				return aclAllow
			}
		}
	}
	return aclDeny
}

func (r MongoDriver) SetACL(clientID, topic, acltype string) {
	var key string
	switch acltype {
	case AclPub:
		key = "publish"
	case AclSub:
		key = "subscribe"
	case AclPubSub:
		key = "pubsub"
	}

	selector := bson.M{"username": clientID}
	update := bson.M{"$addToSet": bson.M{key: topic}}
	r.aclCollection.Upsert(selector, update)

}

func (r MongoDriver) RemoveACL(clientID, topic string) {
	selector := bson.M{"username": clientID}
	r.aclCollection.Remove(selector)
}
