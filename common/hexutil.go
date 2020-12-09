package common

import "encoding/hex"

var(
	ErrEmptyString=&decError{"empty hex string"}
	ErrMissingPrefix=&decError{"hex string without 0x prefix"}
)

type decError struct {
	msg string
}

func (err *decError) Error() string  {
	return err.msg
}

func HexToBytes(input string)([]byte,error)  {
     if len(input)==0{
     	return nil,ErrEmptyString
	 }
	if !Has0xPrefix(input){
		return nil,ErrMissingPrefix
	}
	b,err:=hex.DecodeString(input[2:])
	if err!=nil{
		err=err
	}
	return b, err
}

func Has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && input[1] == 'x'
}

// BytesToHex encodes b as a hex string with 0x prefix.
func BytesToHex(b []byte) string {
	enc := make([]byte, len(b)*2+2)
	copy(enc, "0x")
	hex.Encode(enc[2:], b)
	return string(enc)
}