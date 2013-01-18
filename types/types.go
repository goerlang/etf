// Package types defines Erlang-specific types in Go.
package types

import (
	"fmt"
)

type Term interface{}
type Tuple []Term
type Array []Term

type ErlAtom string

type Node ErlAtom

type ErlPid struct {
	Node     Node
	Id       uint32
	Serial   uint32
	Creation byte
}

type Port struct {
	Node     Node
	Id       uint32
	Creation byte
}

type Reference struct {
	Node     Node
	Creation byte
	Id       []uint32
}

type Function struct {
	Arity     byte
	Unique    [16]byte
	Index     uint32
	Free      uint32
	Module    ErlAtom
	OldIndex  uint32
	OldUnique uint32
	Pid       ErlPid
	FreeVars  []Term
}

type Export struct {
	Module   ErlAtom
	Function ErlAtom
	Arity    byte
}

// Erlang types.
const (
	ErlTypeAtom         = 'd'
	ErlTypeBinary       = 'm'
	ErlTypeBitBinary    = 'M'
	ErlTypeCachedAtom   = 'C'
	ErlTypeExport       = 'q'
	ErlTypeFloat        = 'c'
	ErlTypeFun          = 'u'
	ErlTypeInteger      = 'b'
	ErlTypeLargeBig     = 'o'
	ErlTypeLargeTuple   = 'i'
	ErlTypeList         = 'l'
	ErlTypeNewCache     = 'N'
	ErlTypeNewFloat     = 'F'
	ErlTypeNewFun       = 'p'
	ErlTypeNewReference = 'r'
	ErlTypeNil          = 'j'
	ErlTypePid          = 'g'
	ErlTypePort         = 'f'
	ErlTypeReference    = 'e'
	ErlTypeSmallAtom    = 's'
	ErlTypeSmallBig     = 'n'
	ErlTypeSmallInteger = 'a'
	ErlTypeSmallTuple   = 'h'
	ErlTypeString       = 'k'
)

const (
	// Erlang external format version number.
	ErlFormatVersion = byte(131)
)

var typeNames = map[byte]string{
	ErlTypeAtom:         "ATOM_EXT",
	ErlTypeBinary:       "BINARY_EXT",
	ErlTypeBitBinary:    "BIT_BINARY_EXT",
	ErlTypeCachedAtom:   "ATOM_CACHE_REF",
	ErlTypeExport:       "EXPORT_EXT",
	ErlTypeFloat:        "FLOAT_EXT",
	ErlTypeFun:          "FUN_EXT",
	ErlTypeInteger:      "INTEGER_EXT",
	ErlTypeLargeBig:     "LARGE_BIG_EXT",
	ErlTypeLargeTuple:   "LARGE_TUPLE_EXT",
	ErlTypeList:         "LIST_EXT",
	ErlTypeNewCache:     "NewCache",
	ErlTypeNewFloat:     "NEW_FLOAT_EXT",
	ErlTypeNewFun:       "NEW_FUN_EXT",
	ErlTypeNewReference: "NEW_REFERENCE_EXT",
	ErlTypeNil:          "NIL_EXT",
	ErlTypePid:          "PID_EXT",
	ErlTypePort:         "PORT_EXT",
	ErlTypeReference:    "REFERENCE_EXT",
	ErlTypeSmallAtom:    "SMALL_ATOM_EXT",
	ErlTypeSmallBig:     "SMALL_BIG_EXT",
	ErlTypeSmallInteger: "SMALL_INTEGER_EXT",
	ErlTypeSmallTuple:   "SMALL_TUPLE_EXT",
	ErlTypeString:       "STRING_EXT",
}

func TypeName(t byte) (name string) {
	name = typeNames[t]
	if name == "" {
		name = fmt.Sprintf("%d", t)
	}
	return
}
