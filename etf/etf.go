package etf

type erlType byte

// Erlang types.
const (
	erlAtom         = erlType('d')
	erlBinary       = 'm'
	erlCachedAtom   = 'C'
	erlFloat        = 'c'
	erlFun          = 'u'
	erlInteger      = 'b'
	erlLargeBig     = 'o'
	erlLargeTuple   = 'i'
	erlList         = 'l'
	erlNewCache     = 'N'
	erlNewFloat     = 'F'
	erlNewFun       = 'p'
	erlNewReference = 'r'
	erlNil          = 'j'
	erlPid          = 'g'
	erlPort         = 'f'
	erlReference    = 'e'
	erlSmallAtom    = 's'
	erlSmallBig     = 'n'
	erlSmallInteger = 'a'
	erlSmallTuple   = 'h'
	erlString       = 'k'
)

const (
	// Erlang external format version number.
	erlFormatVersion = byte(131)
)

type Term interface{}

type Atom string

type Node Atom

type Pid struct {
	node     Node
	id       uint32
	serial   uint32
	creation byte
}

type Port struct {
	node     Node
	id       uint32
	creation byte
}

type Reference struct {
	node     Node
	creation byte
	id       []uint32
}

type Function struct {
	arity     byte
	unique    [16]byte
	index     uint32
	free      uint32
	module    Atom
	oldIndex  uint32
	oldUnique uint32
	pid       Pid
	freeVars  []Term
}

type Export struct {
	module   Atom
	function Atom
	arity    byte
}

func reverseBytes(b []byte) []byte {
	size := len(b)
	r := make([]byte, size)

	for i := 0; i < size; i++ {
		r[i] = b[size-i-1]
	}

	return r
}
