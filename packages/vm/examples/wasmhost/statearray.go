package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

type StateArray struct {
	ArrayObject
	items  *kv.MustArray
	typeId int32
}

func NewStateArray(vm *wasmVMPocProcessor, items *kv.MustArray, typeId int32) HostObject {
	return &StateArray{ArrayObject: ArrayObject{vm: vm, name: "StateArray"}, items: items, typeId: typeId}
}

func (a *StateArray) GetBytes(keyId int32) []byte {
	if !a.valid(keyId, OBJTYPE_BYTES) {
		return []byte(nil)
	}
	return a.items.GetAt(uint16(keyId))
}

func (a *StateArray) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyLength:
		return int64(a.GetLength())
	}

	if !a.valid(keyId, OBJTYPE_INT) {
		return 0
	}
	value, _ := kv.DecodeInt64(a.items.GetAt(uint16(keyId)))
	return value
}

func (a *StateArray) GetLength() int32 {
	return int32(a.items.Len())
}

func (a *StateArray) GetString(keyId int32) string {
	if !a.valid(keyId, OBJTYPE_STRING) {
		return ""
	}
	return string(a.items.GetAt(uint16(keyId)))
}

func (a *StateArray) SetBytes(keyId int32, value []byte) {
	if !a.valid(keyId, OBJTYPE_BYTES) {
		return
	}
	a.items.SetAt(uint16(keyId), value)
}

func (a *StateArray) SetInt(keyId int32, value int64) {
	if keyId == KeyLength {
		a.items.Erase()
		return
	}
	if !a.valid(keyId, OBJTYPE_INT) {
		return
	}
	a.items.SetAt(uint16(keyId), util.Uint64To8Bytes(uint64(value)))
}

func (a *StateArray) SetString(keyId int32, value string) {
	if !a.valid(keyId, OBJTYPE_STRING) {
		return
	}
	a.items.SetAt(uint16(keyId), []byte(value))
}

func (a *StateArray) valid(keyId int32, typeId int32) bool {
	if a.typeId != typeId {
		a.error("valid: Invalid access")
		return false
	}
	max := a.GetLength()
	if keyId == max {
		switch typeId {
		case OBJTYPE_BYTES:
			a.items.Push([]byte(nil))
		case OBJTYPE_INT:
			a.items.Push(util.Uint64To8Bytes(0))
		case OBJTYPE_STRING:
			a.items.Push([]byte(""))
		default:
			a.error("valid: Invalid type id")
			return false
		}
		return true
	}
	if keyId < 0 || keyId >= max {
		a.error("valid: Invalid index")
		return false
	}
	return true
}