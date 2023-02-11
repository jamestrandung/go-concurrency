// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package promising

import "errors"

var (
	ErrNoMoreResultIncoming = errors.New("no more result is coming for the given key")
	ErrNoResultForGivenKey  = errors.New("no result is available for the given key")
	ErrKeyIsNotComparable   = errors.New("the given key is not comparable")
)
