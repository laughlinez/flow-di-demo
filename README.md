flow-di-demo
============

Demonstrates DI for a flowapi


Copy a 'data' (leveldb) folder from another housemon installation to the apps folder.

Assumes that this leveldb data is in './data' when run.

main() imports some local gadgets Foo & Bar. It also imports gadget LevelDB which is a 'clone' of jeebus LevelDB gadget 
with a few API's added.

Note that Foo & Bar (see gadgets/gadgets.go) both have the following struct members:

```
	Settings flowapi.ISettingsAPI `gadget:"SettingsAPI"`  //gadget requests ISettingsAPI
	DB flowapi.IDBReadWriteAPI `gadget:"DBReadWriteAPI"`  //gadget requests IDBReadWriteAPI
```

They are 'requesting' if the flow API can provide these services.

Now note in gadgets/database/database.go:

```
type LevelDB struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
	DBAPI api.IDBReadWriteAPI `flowapi:"DBReadWriteAPI"`
	Settings *LevelDBSettingsAPI `flowapi:"SettingsAPI,new"`
}
```
Here the database gadget is 'providing' services.
 
Note how flow has no direct access to the gadgets (no imports), and the gadgets themselves do not reference each other.

Happy to discuss more via issuelist/forum.



