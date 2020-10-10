package common

import (
    "bytes"
    "crypto/ecdsa"
    "crypto/elliptic"
    "go-serverRPC/common/errors"
)

type AddressType byte

const (
  AddressLen=20
    AddressTypeExternal = AddressType(1)
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

func PubKeyToAddress(pubKey *ecdsa.PublicKey,hashFunc func(interface{}) Hash)  Address {
    buf := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)
    hash := hashFunc(buf[1:]).Bytes()

    var addr Address
    copy(addr[:], hash[12:]) // use last 20 bytes of public key hash

    // set address type in the last 4 bits
    addr[19] &= 0xF0
    addr[19] |= byte(AddressTypeExternal)

    return addr

}