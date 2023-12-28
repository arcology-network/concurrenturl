package ccurltest

// import (
// 	"reflect"
// 	"testing"

// 	storage "github.com/arcology-network/common-lib/storage"
// 	"github.com/arcology-network/common-lib/common"
// 	ccurl "github.com/arcology-network/concurrenturl"
// 	committercommon "github.com/arcology-network/concurrenturl/common"
// 	importer "github.com/arcology-network/concurrenturl/importer"
// 	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
// 	storage "github.com/arcology-network/concurrenturl/storage"
// )

// func TestPartialCache(t *testing.T) {
// 	memDB := storage.NewMemDB()
// 	policy := storage.NewCachePolicy(10000000, 1.0)
// 	store := storage.NewDataStore(nil, policy, memDB, committercommon.Codec{}.Encode, committercommon.Codec{}.Decode)
// 		committer := ccurl.NewStorageCommitter(store)
// writeCache := committer.WriteCache()
// 	alice := AliceAccount()
// 	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	committer.Write(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
// 	acctTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
// 	committer.Import(importer.Univalues{}.Decode(importer.Univalues(acctTrans).Encode()).(importer.Univalues))
// 	committer.Sort()
// 	committer.Commit([]uint32{committercommon.SYSTEM})

// 	/* Filter persistent data source */
// 	excludeMemDB := func(db storage.PersistentStorageInterface) bool { // Do not access MemDB
// 		name := reflect.TypeOf(db).String()
// 		return name != "*storage.MemDB"
// 	}

// 	committer.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
// 	acctTrans = importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
// 	committer.Importer().Store().(*storage.DataStore).Cache().Clear()
// 	committer.Import(importer.Univalues{}.Decode(importer.Univalues(acctTrans).Encode()).(importer.Univalues), true, excludeMemDB) // The changes will be discarded.
// 	committer.Sort()
// 	committer.Commit([]uint32{1})

// 	// if v, _ := committer.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
// 	// 	t.Error("Error: The entry shouldn't be in the DB !")
// 	// } else {
// 	// 	if string(*(v.(*noncommutative.String))) != "1234" {
// 	// 		t.Error("Error: The entry shouldn't changed !")
// 	// 	}
// 	// }

// 	/* Don't filter persistent data source	*/
// 	committer.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
// 	committer.Importer().Store().(*storage.DataStore).Cache().Clear()                                 // Make sure only the persistent storage has the data.
// 	committer.Import(importer.Univalues{}.Decode(importer.Univalues(acctTrans).Encode()).(importer.Univalues)) // This should take effect
// 	committer.Sort()
// 	committer.Commit([]uint32{1})

// 	if v, _ := committer.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
// 		t.Error("Error: The entry shouldn't be in the DB !")
// 	} else {
// 		if v.(string) != "9999" {
// 			t.Error("Error: The entry should have been changed !")
// 		}
// 	}
// }

// func TestPartialCacheWithFilter(t *testing.T) {
// 	memDB := storage.NewMemDB()
// 	policy := storage.NewCachePolicy(10000000, 1.0)

// 	excludeMemDB := func(db storage.PersistentStorageInterface) bool { /* Filter persistent data source */
// 		name := reflect.TypeOf(db).String()
// 		return name == "*storage.MemDB"
// 	}

// 	store := storage.NewDataStore(nil, policy, memDB, committercommon.Codec{}.Encode, committercommon.Codec{}.Decode, excludeMemDB)
// 		committer := ccurl.NewStorageCommitter(store)
// writeCache := committer.WriteCache()
// 	alice := AliceAccount()
// 	if _, err := writeCache.CreateNewAccount(committercommon.SYSTEM, alice); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	committer.Write(committercommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
// 	acctTrans := importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
// 	committer.Import(importer.Univalues{}.Decode(importer.Univalues(acctTrans).Encode()).(importer.Univalues))
// 	committer.Sort()
// 	committer.Commit([]uint32{committercommon.SYSTEM})

// 	if _, err := committer.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999")); err != nil {
// 		t.Error(err)
// 	}

// 	acctTrans = importer.Univalues(common.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

// 	// 	committer := ccurl.NewStorageCommitter(store)
// writeCache := committer.WriteCache()

// 	committer.WriteCache().Clear()

// 	// ccmap2 := committer.Importer().Store().(*storage.DataStore).Cache()
// 	// fmt.Print(ccmap2)
// 	out := importer.Univalues{}.Decode(importer.Univalues(common.Clone(acctTrans)).Encode()).(importer.Univalues)
// 	committer.Import(out, true, excludeMemDB) // The changes will be discarded.
// 	committer.Sort()
// 	committer.Commit([]uint32{1})

// 	if v, _ := committer.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
// 		t.Error("Error: The entry shouldn't be in the DB as the persistent DB has been excluded !")
// 	} else {
// 		if v.(string) != "9999" {
// 			t.Error("Error: The entry shouldn't changed !")
// 		}
// 	}
// }
