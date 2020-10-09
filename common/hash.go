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