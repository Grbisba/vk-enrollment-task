package subpub

import (
	"errors"
)

var (
	errNoSubscribers  = errors.New("no subscribers available")
	errNoSubscriber   = errors.New("this partition does not have a subscriber")
	errTimeoutToWrite = errors.New("timeout to write into partition")
)
