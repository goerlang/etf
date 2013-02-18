package etf

import (
	"fmt"
)

type cacheFlag struct {
	isNew      bool
	segmentIdx uint8
}

type atomCacheRef struct {
	idx  uint8
	text *string
}

type Context struct {
	atomCache    [2048]*string
	currentCache []*string
}

type Term interface{}
type Tuple []Term
type List []Term
type Atom string

type Pid struct {
	Node     Atom
	Id       uint32
	Serial   uint32
	Creation byte
}

type Port struct {
	Node     Atom
	Id       uint32
	Creation byte
}

type Ref struct {
	Node     Atom
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

// Erlang external term tags.
const (
	ettAtom          = 'd'
	ettAtomUTF8      = 'v' // this is beyond retarded
	ettBinary        = 'm'
	ettBitBinary     = 'M'
	ettCachedAtom    = 'C'
	ettCacheRef      = 'R'
	ettExport        = 'q'
	ettFloat         = 'c'
	ettFun           = 'u'
	ettInteger       = 'b'
	ettLargeBig      = 'o'
	ettLargeTuple    = 'i'
	ettList          = 'l'
	ettNewCache      = 'N'
	ettNewFloat      = 'F'
	ettNewFun        = 'p'
	ettNewRef        = 'r'
	ettNil           = 'j'
	ettPid           = 'g'
	ettPort          = 'f'
	ettRef           = 'e'
	ettSmallAtom     = 's'
	ettSmallAtomUTF8 = 'w' // this is beyond retarded
	ettSmallBig      = 'n'
	ettSmallInteger  = 'a'
	ettSmallTuple    = 'h'
	ettString        = 'k'
)

const (
	// Erlang external term format version
	EtVersion = byte(131)
)

const (
	// Erlang distribution header
	EtDist = byte('D')
)

var tagNames = map[byte]string{
	ettAtom:          "ATOM_EXT",
	ettAtomUTF8:      "ATOM_UTF8_EXT",
	ettBinary:        "BINARY_EXT",
	ettBitBinary:     "BIT_BINARY_EXT",
	ettCachedAtom:    "ATOM_CACHE_REF",
	ettExport:        "EXPORT_EXT",
	ettFloat:         "FLOAT_EXT",
	ettFun:           "FUN_EXT",
	ettInteger:       "INTEGER_EXT",
	ettLargeBig:      "LARGE_BIG_EXT",
	ettLargeTuple:    "LARGE_TUPLE_EXT",
	ettList:          "LIST_EXT",
	ettNewCache:      "NEW_CACHE_EXT",
	ettNewFloat:      "NEW_FLOAT_EXT",
	ettNewFun:        "NEW_FUN_EXT",
	ettNewRef:        "NEW_REFERENCE_EXT",
	ettNil:           "NIL_EXT",
	ettPid:           "PID_EXT",
	ettPort:          "PORT_EXT",
	ettRef:           "REFERENCE_EXT",
	ettSmallAtom:     "SMALL_ATOM_EXT",
	ettSmallAtomUTF8: "SMALL_ATOM_UTF8_EXT",
	ettSmallBig:      "SMALL_BIG_EXT",
	ettSmallInteger:  "SMALL_INTEGER_EXT",
	ettSmallTuple:    "SMALL_TUPLE_EXT",
	ettString:        "STRING_EXT",
}

func (t Tuple) Element(i int) Term {
	return t[i-1]
}

func tagName(t byte) (name string) {
	name = tagNames[t]
	if name == "" {
		name = fmt.Sprintf("%d", t)
	}
	return
}
