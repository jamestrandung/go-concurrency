package helper

import "reflect"

// IsComparable returns whether v is not nil and has an underlying
// type that is comparable.
func IsComparable(v interface{}) bool {
	return v != nil && reflect.TypeOf(v).Comparable()
}
