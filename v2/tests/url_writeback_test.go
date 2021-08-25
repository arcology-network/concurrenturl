package ccurltest

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	ccurl "github.com/HPISTechnologies/concurrenturl/v2"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
	ccurltype "github.com/HPISTechnologies/concurrenturl/v2/type"
	commutative "github.com/HPISTechnologies/concurrenturl/v2/type/commutative"
	noncommutative "github.com/HPISTechnologies/concurrenturl/v2/type/noncommutative"
)

func SimulatedTx0() []byte {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	url.Write(0, "blcc://eth1.0/account/alice/storage/ctrn-0/", path)
	url.Write(0, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-00", noncommutative.NewString("tx0-elem-00")) /* The first Element */
	url.Write(0, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-01", noncommutative.NewString("tx0-elem-01")) /* The second Element */

	_, transitions := url.Export(true)
	return ccurltype.Univalues(transitions).Encode()
}

func SimulatedTx1() []byte {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-1/") // create a path
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-1/", path)
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-1/elem-00", noncommutative.NewString("tx1-elem-00")) /* The first Element */
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-1/elem-01", noncommutative.NewString("tx1-elem-00")) /* The second Element */

	_, transitions := url.Export(true)
	return ccurltype.Univalues(transitions).Encode()
}

func CheckPaths(url *ccurl.ConcurrentUrl) error {
	v, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-00")
	if v.(ccurlcommon.TypeInterface).Value() == "tx0-elem-00" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-01")
	if v.(ccurlcommon.TypeInterface).Value() == "tx0-elem-01" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-1/elem-00")
	if v.(ccurlcommon.TypeInterface).Value() == "tx1-elem-00" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-1/elem-01")
	if v.(ccurlcommon.TypeInterface).Value() == "tx1-elem-01" {
		return errors.New("Error: Not match")
	}

	//Read the path
	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	if !reflect.DeepEqual(v.(ccurlcommon.TypeInterface).Value(), []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Keys don't match !")
	}

	// Read the path again
	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	if !reflect.DeepEqual(v.(ccurlcommon.TypeInterface).Value(), []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Keys don't match !")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-1/")
	if !reflect.DeepEqual(v.(ccurlcommon.TypeInterface).Value(), []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Keys don't match !")
	}
	return nil
}

func TestStateUpdate(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	tx0bytes := SimulatedTx0()
	tx0Out := ccurltype.Univalues{}.Decode(tx0bytes).(ccurltype.Univalues)

	tx1bytes := SimulatedTx1()
	tx1Out := ccurltype.Univalues{}.Decode(tx1bytes).(ccurltype.Univalues)

	url.Indexer().Import(tx0Out)
	url.Indexer().Import(tx1Out)

	_, _, errs := url.Indexer().Commit([]uint32{0, 1})
	if len(errs) != 0 || CheckPaths(url) != nil {
		t.Error(errs)
	}

	// Delete an nonexistent entry, should fail !
	if err := url.Write(9, "blcc://eth1.0/account/alice/storage/ctrn-0", nil); err == nil {
		t.Error("Error: Writing Should fail !")
	}

	v, _ := url.Read(9, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	if !reflect.DeepEqual(v.(ccurlcommon.TypeInterface).Value(), []string{"elem-00", "elem-01"}) {
		t.Error("Error: Keys don't match !")
	}

	// Delete the container-0
	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", nil); err != nil {
		t.Error("Error: Failed to delete the path !")
	}

	if v, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/"); v != nil {
		t.Error("Error: The element should be gone already !")
	}

	_, transitions := url.Export(true)

	url.Indexer().Import(ccurltype.Univalues{}.Decode(ccurltype.Univalues(transitions).Encode()).(ccurltype.Univalues))
	_, _, errs = url.Indexer().Commit([]uint32{0, 1})

	if v, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/alice/storage/ctrn-0/"); v != nil {
		t.Error("Error: Should be gone already !")
	}
}

func TestMultipleTxStateUpdate(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	tx0bytes := SimulatedTx0()
	tx0Out := ccurltype.Univalues{}.Decode(tx0bytes).(ccurltype.Univalues)

	tx1bytes := SimulatedTx1()
	tx1Out := ccurltype.Univalues{}.Decode(tx1bytes).(ccurltype.Univalues)

	url.Indexer().Import(tx0Out)
	url.Indexer().Import(tx1Out)

	_, _, errs := url.Indexer().Commit([]uint32{0, 1})
	if len(errs) != 0 {
		t.Error(errs)
	}

	/* Check Paths */
	CheckPaths(url)

	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-111", noncommutative.NewString("tx0-elem-111")); err != nil {
		t.Error("Error: Failed to delete the path !")
	}

	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-222", noncommutative.NewString("tx0-elem-222")); err != nil {
		t.Error("Error: Failed to delete the path !")
	}

	_, transitions := url.Export(true)
	url.Indexer().Import(transitions)

	if _, _, errs = url.Indexer().Commit([]uint32{1}); len(errs) > 0 {
		t.Error(errs)
	}

	v, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	keys := v.(ccurlcommon.TypeInterface).Value()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01", "elem-111", "elem-222"}) {
		t.Error("Error: Keys don't match !")
	}
	//url.Indexer().Store().Print()
}

func TestAccessControl(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	tx0bytes := SimulatedTx0()
	tx0Out := ccurltype.Univalues{}.Decode(tx0bytes).(ccurltype.Univalues)

	tx1bytes := SimulatedTx1()
	tx1Out := ccurltype.Univalues{}.Decode(tx1bytes).(ccurltype.Univalues)

	url.Indexer().Import(tx0Out)
	url.Indexer().Import(tx1Out)

	_, _, errs := url.Indexer().Commit([]uint32{0, 1})
	if len(errs) != 0 {
		t.Error(errs)
	}

	// Account root Path
	v, err := url.Read(1, "blcc://eth1.0/account/alice/")
	if v == nil {
		t.Error(err) // Users shouldn't be able to read any of the system paths
	}

	v, err = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/alice/")
	if v == nil {
		t.Error(err)
	}

	/* Code */
	v, err = url.Read(1, "blcc://eth1.0/account/alice/code")
	if err != nil { // Shouldn't be able to read
		t.Error(err)
	}

	err = url.Write(1, "blcc://eth1.0/account/alice/code", noncommutative.NewString("New code"))
	if err == nil {
		t.Error("Error: Users shouldn't be updated blcc://eth1.0/account/alice/code")
	}

	/* Balance */
	v, err = url.Read(1, "blcc://eth1.0/account/alice/balance")
	if err != nil {
		t.Error(err)
	}

	err = url.Write(1, "blcc://eth1.0/account/alice/balance", commutative.NewBalance(big.NewInt(100), big.NewInt(100)))
	if err != nil {
		t.Error("Error: Failed to write the balance")
	}

	err = url.Write(1, "blcc://eth1.0/account/alice/balance", commutative.NewBalance(big.NewInt(100), big.NewInt(0)))
	if err != nil {
		t.Error("Error: Failed to initialize balance")
	}

	err = url.Write(1, "blcc://eth1.0/account/alice/balance", commutative.NewBalance(big.NewInt(100), big.NewInt(100)))
	if err != nil {
		t.Error("Error: Failed to initialize balance")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/alice/balance")
	if v.(*commutative.Balance).Value().(*big.Int).Cmp(big.NewInt(200)) != 0 {
		t.Error("Error: blcc://eth1.0/account/alice/balance, should be 300")
	}

	/* Nonce */
	err = url.Write(1, "blcc://eth1.0/account/alice/nonce", commutative.NewInt64(0, 10))
	if err != nil {
		t.Error("Error: Failed to read the nonce value !")
	}

	/* Storage */
	meta, _ := commutative.NewMeta("")
	err = url.Write(1, "blcc://eth1.0/account/alice/storage/", meta)
	if err == nil {
		t.Error("Error: Users shouldn't be able to change the storage path !")
	}

	_, err = url.Read(1, "blcc://eth1.0/account/alice/storage/")
	if err != nil {
		t.Error("Error: Failed to read the storage !")
	}

	err = url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/alice/storage/", meta)
	if err != nil {
		t.Error("Error: The system should be able to change the storage path !")
	}

	_, err = url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/alice/storage/")
	if err != nil {
		t.Error(err)
	}
}
