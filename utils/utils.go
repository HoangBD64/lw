// package utils contains library-dependent helper and ultility function
package utils

import (
	"crypto/rand"
	"math"
	"math/big"
	"time"

	"github.com/gofiber/fiber/v2"
)

// defArgs takes in an array of optional arguments, with optional fallback variable
//
// Return the fallback variable of array of optional arguments is empty
// Return the first element of array of optional arguments is not empty
// Return the zero value from type of provided array of optional arguments is empty and no fallback option given
func DefArgs[T any](args []T, fallback ...T) (T, bool) {
	var out T
	if len(fallback) > 0 {
		out = fallback[0]
	}
	if len(args) > 0 {
		return args[0], true
	}
	return out, false
}

func IsInRange[T any](index int, arr []T) bool {
	if index >= 0 && index < len(arr) {
		return false
	}
	return true
}

func InterfaceToStringSlice(in any) ([]string, bool) {
	out := []string{}
	if in == nil {
		return nil, false
	}
	switch v := in.(type) {
	case []any:
		for _, inner := range v {
			str, ok := inner.(string)
			if !ok {
				return nil, false
			}
			out = append(out, str)
		}
		return out, true
	default:
		return nil, false
	}
}

func IP(c *fiber.Ctx) string {
	fw := c.IPs()
	if len(fw) == 0 {
		return c.IP()
	}
	return fw[0]
}

// StrDef : [Str]ing [Def]ault
//
// returns given string if not empty, otherwise returns fallback string
func StrDef(check string, fallback string) string {
	if check == "" {
		return fallback
	}
	return check
}

// Returns a Map with key of type K and value of type struct{}, used for existence checks
func IndexLookup[T any, K comparable](arr []T, callback func(iter T) K) map[K]struct{} {
	result := make(map[K]struct{}, len(arr))

	for _, t := range arr {
		result[callback(t)] = struct{}{}
	}

	return result
}

// Returns a Map with key of type K and value of type T = oginal array type, used for quick lookups
func IndexMap[T any, K comparable](arr []T, callback func(iter T) K) map[K]T {
	result := make(map[K]T, len(arr))

	for _, t := range arr {
		result[callback(t)] = t
	}

	return result
}

func MathRoundPrice(number float64) float64 {
	number = math.Round(number*1000) / 1000
	return number
}

func Retry[K any](attempts int, sleep time.Duration, f func() (K, error)) (rs K, err error) {
	rs, err = f()
	if err == nil {
		return
	}
	for i := 0; i < attempts-1; i++ {
		jitter := time.Duration(RandomInt(int(sleep)))
		sleep = sleep + jitter/2
		time.Sleep(sleep)
		rs, err = f()
		if err == nil {
			return
		}
	}

	return
}

func RandomInt(max int) int {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(nBig.Int64())
}
