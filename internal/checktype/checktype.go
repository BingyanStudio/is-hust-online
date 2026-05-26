package checktype

import (
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
)

// Bit returns the bitmask value for a given CheckType.
func Bit(ct myproto.CheckType) int32 {
	switch ct {
	case myproto.CheckType_CHECK_TYPE_HTTP:
		return 1
	case myproto.CheckType_CHECK_TYPE_PING:
		return 2
	case myproto.CheckType_CHECK_TYPE_TCP:
		return 4
	case myproto.CheckType_CHECK_TYPE_OTHER:
		return 8
	default:
		return 0
	}
}
