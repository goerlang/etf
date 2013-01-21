// Package types defines Erlang-specific types in Go.
package types

import (
	"fmt"
)

type Term interface{}
type Tuple []Term
type List []Term

type Atom string
type Node Atom

type Pid struct {
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

type Ref struct {
	Node     Node
	Creation byte
	Id       []uint32
}

type Function struct {
	Arity     byte
	Unique    [16]byte
	Index     uint32
	Free      uint32
	Module    Atom
	OldIndex  uint32
	OldUnique uint32
	Pid       Pid
	FreeVars  []Term
}

type Export struct {
	Module   Atom
	Function Atom
	Arity    byte
}

// Erlang types.
const (
	EttAtom         = 'd'
	EttBinary       = 'm'
	EttBitBinary    = 'M'
	EttCachedAtom   = 'C'
	EttExport       = 'q'
	EttFloat        = 'c'
	EttFun          = 'u'
	EttInteger      = 'b'
	EttLargeBig     = 'o'
	EttLargeTuple   = 'i'
	EttList         = 'l'
	EttNewCache     = 'N'
	EttNewFloat     = 'F'
	EttNewFun       = 'p'
	EttNewReference = 'r'
	EttNil          = 'j'
	EttPid          = 'g'
	EttPort         = 'f'
	EttReference    = 'e'
	EttSmallAtom    = 's'
	EttSmallBig     = 'n'
	EttSmallInteger = 'a'
	EttSmallTuple   = 'h'
	EttString       = 'k'
)

const (
	// Erlang external term format version
	EtVersion = byte(131)
)

var typeNames = map[byte]string{
	EttAtom:         "ATOM_EXT",
	EttBinary:       "BINARY_EXT",
	EttBitBinary:    "BIT_BINARY_EXT",
	EttCachedAtom:   "ATOM_CACHE_REF",
	EttExport:       "EXPORT_EXT",
	EttFloat:        "FLOAT_EXT",
	EttFun:          "FUN_EXT",
	EttInteger:      "INTEGER_EXT",
	EttLargeBig:     "LARGE_BIG_EXT",
	EttLargeTuple:   "LARGE_TUPLE_EXT",
	EttList:         "LIST_EXT",
	EttNewCache:     "NewCache",
	EttNewFloat:     "NEW_FLOAT_EXT",
	EttNewFun:       "NEW_FUN_EXT",
	EttNewReference: "NEW_REFERENCE_EXT",
	EttNil:          "NIL_EXT",
	EttPid:          "PID_EXT",
	EttPort:         "PORT_EXT",
	EttReference:    "REFERENCE_EXT",
	EttSmallAtom:    "SMALL_ATOM_EXT",
	EttSmallBig:     "SMALL_BIG_EXT",
	EttSmallInteger: "SMALL_INTEGER_EXT",
	EttSmallTuple:   "SMALL_TUPLE_EXT",
	EttString:       "STRING_EXT",
}

func TypeName(t byte) (name string) {
	name = typeNames[t]
	if name == "" {
		name = fmt.Sprintf("%d", t)
	}
	return
}
