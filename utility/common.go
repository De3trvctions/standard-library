package utility

import (
	"math/rand"
	"strconv"
	"time"
)

func GetRandomNumber(length int) (result string) {
	min := pow(10, length-1)
	max := pow(10, length) - 1

	a := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNumber := a.Intn(max - min + 1)
	result = strconv.Itoa(randomNumber)

	return result
}

// pow calculates the power of a number
func pow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}
