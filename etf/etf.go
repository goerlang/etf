package etf

/*
Copyright © 2012 Serge Zirukin

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

type erlType byte

// Erlang types.
const (
  erlAtom         = erlType('d')
  erlBinary       = erlType('m')
  erlCachedAtom   = erlType('C')
  erlFloat        = erlType('c')
  erlFun          = erlType('u')
  erlInteger      = erlType('b')
  erlLargeBig     = erlType('o')
  erlLargeTuple   = erlType('i')
  erlList         = erlType('l')
  erlNewCache     = erlType('N')
  erlNewFloat     = erlType('F')
  erlNewFun       = erlType('p')
  erlNewReference = erlType('r')
  erlNil          = erlType('j')
  erlPid          = erlType('g')
  erlPort         = erlType('f')
  erlReference    = erlType('e')
  erlSmallBig     = erlType('n')
  erlSmallInteger = erlType('a')
  erlSmallTuple   = erlType('h')
  erlString       = erlType('k')
)

type Atom string

func reverseBytes(b []byte) []byte {
  size := len(b)
  r := make([]byte, size)

  for i := 0; i < size; i++ {
    r[i] = b[size - i - 1]
  }

  return r
}

// Local Variables:
// indent-tabs-mode: nil
// tab-width: 2
// End:
// ex: set tabstop=2 shiftwidth=2 expandtab:
