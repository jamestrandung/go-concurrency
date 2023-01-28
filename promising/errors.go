package promising

import "errors"

var (
	ErrNoMoreResultIncoming = errors.New("no more result is coming for the given key")
	ErrNoResultForGivenKey  = errors.New("no result is available for the given key")
	ErrKeyIsNotComparable   = errors.New("the given key is not comparable")
)
