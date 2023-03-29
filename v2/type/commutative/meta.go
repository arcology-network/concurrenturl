package commutative

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	codec "github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"

	// performance "github.com/arcology-network/common-lib/mhasher"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	orderedmap "github.com/elliotchance/orderedmap"
)

type Meta struct {
	keyView         *orderedmap.OrderedMap //The latest view of all the keys. keyView == keys + added - removed
	addedBuffer     *orderedmap.OrderedMap
	removedBuffer   map[string]ccurlcommon.UnivalueInterface
	finalized       bool
	iterator        *orderedmap.Element
	reverseIterator *orderedmap.Element
	cacheDirty      bool

	// Export only
	keys    []string
	added   []string // added keys
	removed []string // removed keys
}

func NewMeta(path string) (interface{}, error) {
	if !ccurlcommon.IsPath(path) {
		return nil, errors.New("Error: Wrong path format !")
	}

	if !ccurlcommon.CheckDepth(path) {
		return nil, errors.New("Error: Exceeded the maximum depth")
	}

	this := &Meta{
		keys:            []string{},
		added:           []string{},
		removed:         []string{},
		finalized:       false,
		keyView:         nil,
		addedBuffer:     orderedmap.NewOrderedMap(),
		removedBuffer:   make(map[string]ccurlcommon.UnivalueInterface),
		iterator:        nil,
		reverseIterator: nil,
		cacheDirty:      false,
	}
	return this, nil
}

func (this *Meta) Deepcopy() interface{} {
	var keyView *orderedmap.OrderedMap
	if this.keyView != nil {
		keyView = this.keyView.Copy()
	}

	return &Meta{
		keys:          this.keys,
		added:         ccurlcommon.Deepcopy(this.added),
		removed:       ccurlcommon.Deepcopy(this.removed),
		keyView:       keyView,
		addedBuffer:   orderedmap.NewOrderedMap(),
		removedBuffer: make(map[string]ccurlcommon.UnivalueInterface),
		finalized:     this.finalized,
		cacheDirty:    false,
	}
}

func (this *Meta) Equal(other *Meta) bool {
	return reflect.DeepEqual(this.keys, other.keys) &&
		reflect.DeepEqual(this.added, other.added) &&
		reflect.DeepEqual(this.removed, other.removed) &&
		this.finalized == other.finalized
}

func (this *Meta) ToAccess() interface{} {
	return nil
}

func (this *Meta) Get(path string, source interface{}) (interface{}, uint32, uint32) {
	this.finalized = true
	if !this.cacheDirty { // cache clean
		return this, 1, 0
	}

	return this, 1, 1
}

func (this *Meta) Delta(source interface{}) interface{} {
	return &Meta{
		keys:            []string{},
		added:           this.PeekAdded(),
		removed:         this.PeekRemoved(),
		finalized:       this.finalized,
		keyView:         this.keyView,
		addedBuffer:     this.addedBuffer,
		removedBuffer:   this.removedBuffer,
		iterator:        this.iterator,
		reverseIterator: this.reverseIterator,
		cacheDirty:      this.cacheDirty,
	}
}

func (this *Meta) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	keys := append(this.keys, this.added...)
	toRemove := this.removed
	vec := v.([]ccurlcommon.UnivalueInterface)
	for i := 0; i < len(vec); i++ {
		if vec[i].GetPath() == nil { // Not in the whitelist
			continue
		}

		v := vec[i].Value()
		if v == nil { // Deletion
			keys = keys[:0]
			toRemove = toRemove[:0]
			this = nil
			continue
		}

		if this == nil {
			this = this.Value().(*Meta)
		}

		keys = append(keys, v.(*Meta).added...)
		toRemove = append(toRemove, v.(*Meta).removed...)
	}

	if this != nil {
		if len(toRemove) > 0 {
			// t0 := time.Now()
			// keys, _ = performance.RemoveString(keys, toRemove)
			toRemoveDict := make(map[string]struct{})
			for _, v := range toRemove {
				toRemoveDict[v] = struct{}{}
			}
			next := 0
			for i := 0; i < len(keys); i++ {
				if _, ok := toRemoveDict[keys[i]]; ok {
					continue
				} else {
					keys[next] = keys[i]
					next++
				}
			}
			keys = keys[:next]
			// fmt.Println("RemoveBytes ", time.Since(t0))
		}

		this.keys = keys
		this.cacheDirty = false
	}
	//fmt.Println("ApplyDelta :", time.Since(t0))

	if this == nil {
		return nil
	}
	return this
}

// Check new keys
func (this *Meta) PeekAdded() []string {
	added := []string{}
	for iter := this.addedBuffer.Front(); iter != nil; iter = iter.Next() {
		added = append(added, iter.Key.(string))
	}
	return added
}

// Check removed keys
func (this *Meta) PeekRemoved() []string {
	removed := []string{}
	for k := range this.removedBuffer {
		removed = append(removed, k)
	}
	return removed
}

// committed keys + added - removed
func (this *Meta) PeekKeys() []string {
	this.LoadKeys()
	newKeys := make([]string, 0, this.keyView.Len())
	for iter := this.keyView.Front(); iter != nil; iter = iter.Next() {
		if iter.Value != nil {
			newKeys = append(newKeys, (iter.Key.(string)))
		}
	}
	return newKeys
}

func (this *Meta) Value() interface{} {
	return this.PeekKeys()
}

// Load keys into an orderedmap for quick access, only happens at once
func (this *Meta) LoadKeys() {
	if this.keyView != nil { // Keys have been loaded.
		return
	}

	this.keyView = orderedmap.NewOrderedMap()
	for _, k := range this.keys {
		if _, ok := this.removedBuffer[k]; !ok {
			this.keyView.Set(k, true) // Not in the removed set
		}
	}

	for iter := this.addedBuffer.Front(); iter != nil; iter = iter.Next() {
		this.keyView.Set(iter.Key, true)
	}
	this.iterator = this.addedBuffer.Front()
	this.reverseIterator = this.addedBuffer.Back()
}

// func (this *Meta) Length() uint32 {
// 	return uint32(this.keyView.Len())
// }

func (this *Meta) ResetIterator()        { this.iterator = this.keyView.Front() }
func (this *Meta) ResetReverseIterator() { this.reverseIterator = this.keyView.Back() }

func (this *Meta) Next() string {
	this.LoadKeys()
	if this.iterator == nil {
		return ""
	}

	key := this.iterator.Key.(string)
	this.iterator = this.iterator.Next()
	return key
}

func (this *Meta) Previous() string {
	this.LoadKeys()
	if this.reverseIterator == nil {
		return ""
	}

	key := this.reverseIterator.Key.(string)
	this.reverseIterator = this.reverseIterator.Prev()
	return key
}

func (this *Meta) This(source interface{}) interface{} {
	return this
}

func (this *Meta) Set(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value == nil { // Remove the path completely
		indexer := source.([2]interface{})[1].(ccurlcommon.IndexerInterface)

		tx := source.([2]interface{})[0].(uint32)
		value := indexer.Read(tx, path)
		for _, subpath := range value.(*Meta).PeekKeys() {
			indexer.Write(tx, path+subpath, nil, false) // Remove all the sub paths.
		}
		return 0, 1, nil
	}
	return 0, 1, errors.New("Error: A path can only be created or deleted, it cannot be rewritten!")
}

func (this *Meta) Reset(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	panic("Error: This function should never be called!!")
	// return 0, 1, nil
}

// Debugging interfaces
// func (this *Meta) SetKeys(keys []string)    { this.keys = keys }
// func (this *Meta) SetAdded(keys []string)   { this.added = keys }
// func (this *Meta) SetRemoved(keys []string) { this.removed = keys }

func (this *Meta) Composite() bool { return !this.finalized }
func (this *Meta) TypeID() uint8   { return ccurlcommon.CommutativeMeta }

func (this *Meta) UpdateCaches(child ccurlcommon.UnivalueInterface, source interface{}) bool {
	if this.keyView != nil {
		key := *child.GetPath()
		subkey := key[strings.LastIndex(key[:len(key)-1], "/")+1:] // Extract the sub key

		if child.Value() == nil {
			this.keyView.Delete(subkey) // Delete a key
		} else {
			this.keyView.Set(subkey, child) // Add a new one
		}
	}
	this.cacheDirty = this.toAddedBuffer(child) || this.toRemovedBuffer(child) // Either is dirty
	return this.cacheDirty
}

func (this *Meta) toAddedBuffer(child ccurlcommon.UnivalueInterface) bool {
	key := *child.GetPath()
	subkey := key[strings.LastIndex(key[:len(key)-1], "/")+1:] // Extract the sub key

	dirty := false
	if !child.Preexist() && child.Value() != nil { // A new Elemnet
		if _, ok := this.addedBuffer.Get(subkey); !ok { // Not in the added cache yet
			this.addedBuffer.Set(subkey, child)
			dirty = true
		}
	}

	if child.Value() == nil { // Delete an Element, it is possible the element is in the added cache
		if _, ok := this.addedBuffer.Get(subkey); ok {
			this.addedBuffer.Delete(subkey)
		}
	}
	return dirty
}

func (this *Meta) toRemovedBuffer(child ccurlcommon.UnivalueInterface) bool {
	key := *child.GetPath()
	subkey := key[strings.LastIndex(key[:len(key)-1], "/")+1:] // Extract the sub key

	dirtyFlag := false
	if child.Preexist() && child.Value() == nil {
		if _, ok := this.removedBuffer[subkey]; !ok {
			this.removedBuffer[subkey] = child // Add to the deleteion list
			dirtyFlag = true
		}
	}

	if child.Value() != nil { // Possible the element has been added back, remove it from the cache in this case
		if _, ok := this.addedBuffer.Get(subkey); ok {
			delete(this.removedBuffer, subkey)
		}
	}
	return dirtyFlag
}

// data cleaning before saving to storage
func (this *Meta) Purge() {
	this.added = []string{}
	this.removed = []string{}
	this.finalized = false
	this.keyView = nil
	this.addedBuffer = orderedmap.NewOrderedMap()
	this.removedBuffer = make(map[string]ccurlcommon.UnivalueInterface)
}

func (this *Meta) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

func (this *Meta) Encode() []byte {
	this.keys = this.keys[:0] // Clear keys, no need to send

	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Meta) HeaderSize() uint32 {
	return 5 * codec.UINT32_LEN
}

func (this *Meta) Size() uint32 {
	if this == nil {
		return 0
	}

	total := this.HeaderSize() +
		codec.Strings(this.keys).Size() +
		codec.Strings(this.added).Size() +
		codec.Strings(this.removed).Size() +
		uint32(codec.Bool(this.finalized).Size())
	return total
}

func (this *Meta) FillHeader(buffer []byte) {
	total := uint32(0)
	codec.Uint32(4).EncodeToBuffer(buffer[codec.UINT32_LEN*0:])

	codec.Uint32(total).EncodeToBuffer(buffer[codec.UINT32_LEN*1:])
	total += codec.Strings(this.keys).Size()

	codec.Uint32(total).EncodeToBuffer(buffer[codec.UINT32_LEN*2:])
	total += codec.Strings(this.added).Size()

	codec.Uint32(total).EncodeToBuffer(buffer[codec.UINT32_LEN*3:])
	total += codec.Strings(this.removed).Size()

	codec.Uint32(total).EncodeToBuffer(buffer[codec.UINT32_LEN*4:])
}

func (this *Meta) EncodeToBuffer(buffer []byte) int {
	this.FillHeader(buffer)
	offset := int(this.HeaderSize())

	offset += codec.Strings(this.keys).EncodeToBuffer(buffer[offset:])
	offset += codec.Strings(this.added).EncodeToBuffer(buffer[offset:])
	offset += codec.Strings(this.removed).EncodeToBuffer(buffer[offset:])

	return int(offset) + codec.Bool(this.finalized).EncodeToBuffer(buffer[offset:])
}

func (this *Meta) Decode(bytes []byte) interface{} {
	buffers := codec.Byteset{}.Decode(bytes).(codec.Byteset)
	this = &Meta{
		keys:          codec.Strings([]string{}).Decode(common.ArrayCopy(buffers[0])).(codec.Strings),
		added:         codec.Strings([]string{}).Decode(common.ArrayCopy(buffers[1])).(codec.Strings),
		removed:       codec.Strings([]string{}).Decode(buffers[2]).(codec.Strings),
		finalized:     bool(codec.Bool(true).Decode(buffers[3]).(codec.Bool)),
		keyView:       nil,
		addedBuffer:   orderedmap.NewOrderedMap(),
		removedBuffer: make(map[string]ccurlcommon.UnivalueInterface),
		cacheDirty:    false,
	}
	return this
}

func (this *Meta) EncodeCompact() []byte {
	byteset := [][]byte{
		codec.Strings(this.keys).Encode(),
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Meta) DecodeCompact(bytes []byte) interface{} {
	buffers := codec.Byteset{}.Decode(bytes).(codec.Byteset)
	return &Meta{
		keys:          codec.Strings([]string{}).Decode(buffers[0]).(codec.Strings),
		added:         []string{},
		removed:       []string{},
		finalized:     false,
		keyView:       nil,
		addedBuffer:   orderedmap.NewOrderedMap(),
		removedBuffer: make(map[string]ccurlcommon.UnivalueInterface),
		cacheDirty:    false,
	}
}

func (this *Meta) Print() {
	fmt.Println("Keys: ", this.keys)
	fmt.Println("Added: ", this.added)
	fmt.Println("Removed: ", this.removed)
	fmt.Println()
}
