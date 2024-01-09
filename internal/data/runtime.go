package data

import (
	"fmt"
	"strconv"
)

//can customize the json behavior with Marshaller interface, the marshal looks if the struct satisfies the marshaler interface, if it supports it
// it uses the MarshalJSON method of the interface to convert the struct to json if not fallback to it's own implementation

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}
