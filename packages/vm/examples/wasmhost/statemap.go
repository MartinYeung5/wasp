package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

type StateMap struct {
	MapObject
	items *kv.MustDictionary
	types map[int32]int32
}

func NewStateMap(vm *wasmVMPocProcessor, items *kv.MustDictionary) HostObject {
	return &StateMap{MapObject: MapObject{vm: vm, name: "StateMap"}, items: items, types: make(map[int32]int32)}
}

func (m *StateMap) GetBytes(keyId int32) []byte {
	if !m.valid(keyId, OBJTYPE_BYTES) {
		return []byte(nil)
	}
	key := []byte(m.vm.GetKey(keyId))
	return m.items.GetAt(key)
}

func (m *StateMap) GetInt(keyId int32) int64 {
	if !m.valid(keyId, OBJTYPE_INT) {
		return 0
	}
	key := []byte(m.vm.GetKey(keyId))
	value, _ := kv.DecodeInt64(m.items.GetAt(key))
	return value
}

func (m *StateMap) GetLength() int32 {
	m.error("GetLength: Invalid length")
	return 0
}

func (m *StateMap) GetObjectId(keyId int32, typeId int32) int32 {
	m.error("GetObjectId: Invalid access")
	return 0
}

func (m *StateMap) GetString(keyId int32) string {
	if !m.valid(keyId, OBJTYPE_STRING) {
		return ""
	}
	key := []byte(m.vm.GetKey(keyId))
	return string(m.items.GetAt(key))
}

func (m *StateMap) SetBytes(keyId int32, value []byte) {
	if !m.valid(keyId, OBJTYPE_BYTES) {
		return
	}
	key := []byte(m.vm.GetKey(keyId))
	m.items.SetAt(key, value)
}

func (m *StateMap) SetInt(keyId int32, value int64) {
	if keyId == KeyLength {
		m.error("SetInt: Invalid clear")
		return
	}
	if !m.valid(keyId, OBJTYPE_INT) {
		return
	}
	key := []byte(m.vm.GetKey(keyId))
	m.items.SetAt(key, util.Uint64To8Bytes(uint64(value)))
}

func (m *StateMap) SetString(keyId int32, value string) {
	if !m.valid(keyId, OBJTYPE_STRING) {
		return
	}
	key := []byte(m.vm.GetKey(keyId))
	m.items.SetAt(key, []byte(value))
}

func (m *StateMap) valid(keyId int32, typeId int32) bool {
	fieldType, ok := m.types[keyId]
	if !ok {
		m.types[keyId] = typeId
		return true
	}
	if fieldType != typeId {
		m.error("valid: Invalid access")
		return false
	}
	return true
}