// Interface to the LevelDB database.
package database

// glog levels:
//	1 = changes to registry
//  2 = changes to database
//  3 = database access

import (
	"encoding/json"
	"strings"
	"sync"
	"github.com/golang/glog"
	"github.com/laughlinez/flow"
	"github.com/laughlinez/flow/api"
	"github.com/syndtr/goleveldb/leveldb"
	dbutil "github.com/syndtr/goleveldb/leveldb/util"
)

var (
	once sync.Once
	db   *leveldb.DB
)

func init() {
	flow.Registry["LevelDB"] = func() flow.Circuitry { return &LevelDB{Settings:new(LevelDBSettingsAPI), DBAPI:new(LevelDBAPI)} }
	flow.Registry["DataSub"] = func() flow.Circuitry { return new(DataSub) }
}

func dbIterateOverKeys(from, to string, fun func(string, []byte)) {
	slice := &dbutil.Range{[]byte(from), []byte(to)}
	if len(to) == 0 {
		slice.Limit = append(slice.Start, 0xFF)
	}

	iter := db.NewIterator(slice, nil)
	defer iter.Release()

	for iter.Next() {
		fun(string(iter.Key()), iter.Value())
	}
}

func dbKeys(prefix string) (results []string) {
	glog.V(3).Infoln("keys", prefix)
	// TODO: decide whether this key logic is the most useful & least confusing
	// TODO: should use skips and reverse iterators once the db gets larger!
	skip := len(prefix)
	prev := "/" // impossible value, this never matches actual results

	dbIterateOverKeys(prefix, "", func(k string, v []byte) {
		i := strings.IndexRune(k[skip:], '/') + skip
		if i < skip {
			i = len(k)
		}
		if prev != k[skip:i] {
			// need to make a copy of the key, since it's owned by iter
			prev = k[skip:i]
			results = append(results, string(prev))
		}
	})
	return
}

func dbGet(key string) (any interface{}) {
	glog.V(3).Infoln("get", key)
	data, err := db.Get([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		return nil
	}
	flow.Check(err)
	err = json.Unmarshal(data, &any)
	flow.Check(err)
	return
}

func dbPut(key string, value interface{}) {
	glog.V(2).Infoln("put", key, value)
	if value != nil {
		data, err := json.Marshal(value)
		flow.Check(err)
		db.Put([]byte(key), data, nil)
	} else {
		db.Delete([]byte(key), nil)
	}
}

func dbRegister(key string) {
	data, err := db.Get([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		glog.Warningln("cannot register:", key)
		return
	}
	name := key[strings.LastIndex(key, "/")+1:]
	glog.V(1).Infof("register %s: %d bytes (%s)", name, len(data), key)
	flow.Registry[key] = func() flow.Circuitry {
		c := flow.NewCircuit()
		c.LoadJSON(data)
		return c
	}
}

func openDatabase() {
	// opening the database takes time, make sure we don't re-enter this code
	once.Do(func() {
		dbPath := flow.Config["DATA_DIR"]
		if dbPath == "" {
			glog.Fatalln("cannot open database, DATA_DIR not set")
		}
		ldb, err := leveldb.OpenFile(dbPath, nil)
		flow.Check(err)
		db = ldb
	})
}

// // Get a list of keys from the database, given a prefix.
// func Keys(prefix string) []string {
// 	openDatabase()
// 	return dbKeys(prefix)
// }
// 
// // Get an entry from the database, returns nil if not found.
// func Get(key string) interface{} {
// 	openDatabase()
// 	return dbGet(key)
// }
// 
// // Store or delete an entry in the database.
// func Put(key string, value interface{}) {
// 	openDatabase()
// 	dbPut(key, value)
// }

// LevelDB is a multi-purpose gadget to get, put, and scan keys in a database.
// Acts on tags received on the input port. Registers itself as "LevelDB".
type LevelDB struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
	DBAPI api.IDBReadWriteAPI `flowapi:"DBReadWriteAPI"`
	Settings *LevelDBSettingsAPI `flowapi:"SettingsAPI,new"`
}

type LevelDBAPI struct {}  //because we need more than *db

func (d *LevelDBAPI) Keys(prefix string) ([]string, error) {
	openDatabase()
	return dbKeys(prefix),nil
}

func (d *LevelDBAPI) Get(key string) (interface{},error) {
	openDatabase()
	return dbGet(key),nil
}

func (d *LevelDBAPI) Put(key string, value interface{}) (error) {
	openDatabase()
	dbPut(key, value)
	return nil
}

//FlowAPI SettingsAPI
type LevelDBSettingsAPI struct {
	name string
	path string
	namespace string
}

func (d *LevelDBSettingsAPI) InitAPI(a...interface{}) {

	for k,v := range a {
		switch k {
		case 0:
			d.name = v.(string)
		case 1:
			d.path = v.(string)
		}
	}

	d.namespace = "/settings" + d.path + d.name + "/"

	return
}


func (d *LevelDBSettingsAPI) Keys(prefix string) ([]string,error) {
	openDatabase()
	return dbKeys(d.namespace + prefix),nil
}


func (d *LevelDBSettingsAPI) Get(key string) (interface{},error) {
	openDatabase()
	return dbGet(d.namespace + key),nil
}

func (d *LevelDBSettingsAPI) Put(key string, value interface{}) (error) {
	openDatabase()
	dbPut(d.namespace + key, value)
	return nil
}




// Open the database and start listening to incoming get/put/keys requests.
func (w *LevelDB) Run() {
	openDatabase()
	for m := range w.In {
		if tag, ok := m.(flow.Tag); ok {
			switch tag.Tag {
			case "<keys>":
				w.Out.Send(m)
				for _, s := range dbKeys(tag.Msg.(string)) {
					w.Out.Send(s)
				}
			case "<get>":
				w.Out.Send(m)
				w.Out.Send(dbGet(tag.Msg.(string)))
			case "<clear>":
				prefix := tag.Msg.(string)
				glog.V(2).Infoln("clear", prefix)
				dbIterateOverKeys(prefix, "", func(k string, v []byte) {
					db.Delete([]byte(k), nil)
					publishChange(flow.Tag{k, nil})
				})
			case "<range>":
				prefix := tag.Msg.(string)
				glog.V(3).Infoln("range", prefix)
				w.Out.Send(m)
				dbIterateOverKeys(prefix, "", func(k string, v []byte) {
					var any interface{}
					err := json.Unmarshal(v, &any)
					flow.Check(err)
					w.Out.Send(flow.Tag{k, any})
				})
			case "<register>":
				dbRegister(tag.Msg.(string))
				// publishChange(tag) // TODO: why was this being sent out?
			default:
				if strings.HasPrefix(tag.Tag, "<") {
					w.Out.Send(m) // pass on other tags without processing
				} else {
					dbPut(tag.Tag, tag.Msg)
					publishChange(tag)
				}
			}
		} else {
			w.Out.Send(m)
		}
	}
}

// use a map of tag channels to publish changes to all DataSub gadgets
// TODO: this will not clean up, all subscriptions will stay running forever
//	cleanup could be done by closing the change channel(s), somehow...

var (
	mutex       sync.RWMutex //this allows RLock on publishChange which would benefit reentrant publishes
	subscribers = map[*DataSub]chan flow.Tag{}
)

func publishChange(tag flow.Tag) {
	mutex.RLock()
	defer mutex.RUnlock()
	// all channels are buffered (-capacity one-), so  this loop will run to completion
	// this is essential to release the lock again for the next iteration
	// TODO: investigate whether RWMutex would make any difference here
	// lightbulb: we only need a reader lock so others dont modify subscribers whilst we range
	for _, c := range subscribers {
		c <- tag
	}
}

// Generate database change messages based on one or more subscription prefixes.
type DataSub struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Subscribe to database changes to pick up and publish all the matching ones.
func (g *DataSub) Run() {
	// collect all subscription prefixes
	subs := []string{}
	for m := range g.In {
		prefix := m.(string)
		glog.V(2).Infoln("data-sub", prefix)
		subs = append(subs, prefix)
	}
	// no subscriptions is treated as a subscription to all changes
	if len(subs) == 0 {
		subs = append(subs, "")
	}
	// set up a change channel
	changes := g.subscribe()
	defer g.unsubscribe()
	// listen for changes and emit those which match
	for t := range changes {
		for _, s := range subs {
			if strings.HasPrefix(t.Tag, s) {
				g.Out.Send(t)
				break
			}
		}
	}
}

func (g *DataSub) subscribe() chan flow.Tag {
	//lightbulb: TODO: this buffer should be calculated more effectively! do we have data to derive??
	//5+1 seems ok with aggregator added (more stable than 1) but will still fail on 'busy' systems.
	bufSize := 5 //a single subscriber may get swamped (e.g multiple sensor subscriptions) that cause further writes.
	changes := make(chan flow.Tag, bufSize+1) // don't block publishChange
	mutex.Lock() //full W lock to modify
	defer mutex.Unlock()
	subscribers[g] = changes
	return changes
}

func (g *DataSub) unsubscribe() {
	mutex.Lock() //full W lock to modify
	defer mutex.Unlock()
	delete(subscribers, g)
}
