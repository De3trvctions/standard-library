package utility

import (
	"log"
	"strconv"
)

func StringToInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Panicln(err.Error())
	}
	return i
}
