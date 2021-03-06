package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// test encode and decode
func TestValueStruct(t *testing.T) {
	v := ValueStruct{
		Value:     []byte("SimpleKV"),
		Meta:      2,
		ExpiresAt: 213123123123,
	}
	data := make([]byte, v.EncodedSize())
	v.EncodeValue(data)
	var vv ValueStruct
	vv.DecodeValue(data)
	assert.Equal(t, vv, v)
}
