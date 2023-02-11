// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package helper

import "reflect"

// IsComparable returns whether v is not nil and has an underlying
// type that is comparable.
func IsComparable(v interface{}) bool {
	return v != nil && reflect.TypeOf(v).Comparable()
}
