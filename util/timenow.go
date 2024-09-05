package util

import (
	"strconv"
	"time"
)

func TimeNow() []byte {
	return []byte(strconv.FormatInt(time.Now().UnixNano(), 10))
}
