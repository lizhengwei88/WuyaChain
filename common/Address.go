package common

import (
    "bytes"
    "go-serverRPC/common/errors"
)

const (
  AddressLen=20
)

var EmptyAddress=Address{}
type  Address [AddressLen]byte

func NewAddress(b []byte) (Address,error)  {
    if len(b)!=AddressLen{
        return EmptyAddress,errors.Create(errors.ErrAddressLenInvalid,len(b),AddressLen)
    }
    var id Address
    copy(id[:],b)
   if err:=id.Validate();err!=nil{
       return EmptyAddress,err
    }
    return id,nil
}

func (id *Address) Validate() error {
    if Equal(id[:], EmptyAddress){
     return nil
    }

    return nil
}


func Equal(b []byte, id Address) bool {
    return bytes.Equal(id[:], b)
}

func HexToAddress(id string) (Address,error)  {
    byte,err:=HexToBytes(id)
   if err!=nil{
       return Address{}, err
   }
   nid,err:=NewAddress(byte)
   if err!=nil{
       return Address{}, err
   }
   return nid,nil
}
