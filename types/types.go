package types

type ErlType byte

// Erlang types.
const (
	ErlTypeAtom         = ErlType('d')
	ErlTypeBinary       = 'm'
	ErlTypeCachedAtom   = 'C'
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

type Term interface{}
type Tuple []Term
type Array []Term

type ErlAtom string

type Node ErlAtom

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
	Pid       Pid
	FreeVars  []Term
}

type Export struct {
	Module   ErlAtom
	Function ErlAtom
	Arity    byte
}
