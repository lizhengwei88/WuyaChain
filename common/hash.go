package common

const (
	HashLength =32
)
type Hash [HashLength]byte

func BytesToHash(b []byte) Hash  {
   h:=&Hash{}
  if len(b)>HashLength{
  	b=b[len(b)-HashLength:]
  }
  copy(h[:],b)
   return *h
}

// Bytes returns its actual bits
func (a Hash) Bytes() []byte {
	return a[:]
}