package go_utils

"gopkg.in/mgo.v2"

// MgoClient : mgo struct , Session and Collection
type MgoClient struct {
	Session    *mgo.Session
	Collection *mgo.Collection
}

// Init : init session and set mode
func (m *MgoClient) Init(uri string) {
	session, err := mgo.Dial(uri)
	if err != nil {
		log.Panicf("MongoDb Connection errors : %s", err.Error())
	}
	m.Session = session
	m.Session.SetMode(mgo.Monotonic, true)
}
