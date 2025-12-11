// package utils contains library-dependent helper and ultility function
package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/unicode/norm"
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

var (
	reSpecial = regexp.MustCompile(`[^a-z0-9\s-]+`)
	reHyphens = regexp.MustCompile(`[-\s]+`)
)

func GenerateSlug(title string) string {
	handle := strings.ToLower(title)

	t := norm.NFD.String(handle)
	handle = strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		if r == 'đ' {
			return 'd'
		}
		return r
	}, t)

	handle = reSpecial.ReplaceAllString(handle, "")

	handle = reHyphens.ReplaceAllString(handle, "-")

	handle = strings.Trim(handle, "-")

	return handle
}

func SplitSliceEvenly[T any](items []T, maxPerGroup int) [][]T {
	n := len(items)

	if n == 0 || maxPerGroup <= 0 {
		return nil
	}

	// Step 1: number of packages required
	numPackages := int(math.Ceil(float64(n) / float64(maxPerGroup)))

	// Step 2: base size & remainder for even distribution
	base := n / numPackages
	extra := n % numPackages

	// Step 3: distribute items evenly
	result := make([][]T, numPackages)
	start := 0
	for i := range numPackages {
		size := base
		if extra > 0 {
			size++
			extra--
		}

		end := min(start+size, n)

		result[i] = items[start:end]
		start = end
	}

	return result
}

type ElementWithCapacity[T any] struct {
	Src      T
	Capacity float64
}

// SplitSliceWithCapacity splits a slice of ElementWithCapacity into multiple slices,
//
// *not guaranteeing even distribution*
//
// return error if a single item exceeds the maxPerGroup capacity

// SplitSliceWithCapacity splits a slice of ElementWithCapacity into multiple slices,
// ensuring that the total capacity of each group is < maxPerGroup.
//
// Returns an error if any single element exceeds maxPerGroup.
func SplitSliceWithCapacity[T any](items []ElementWithCapacity[T], maxPerGroup float64) ([][]ElementWithCapacity[T], error) {
	if len(items) == 0 {
		return nil, nil
	}

	if maxPerGroup <= 0 {
		return nil, errors.New("invalid max capacity per group limit")
	}

	// Validate that no item exceeds maxPerGroup
	for _, item := range items {
		if item.Capacity >= maxPerGroup {
			return nil, fmt.Errorf("an element exceeds the max capacity per group limit (%v >= %v)", item.Capacity, maxPerGroup)
		}
	}

	// Sort descending by capacity
	slices.SortFunc(items, func(a, b ElementWithCapacity[T]) int {
		switch {
		case a.Capacity > b.Capacity:
			return -1
		case a.Capacity < b.Capacity:
			return 1
		default:
			return 0
		}
	})

	var result [][]ElementWithCapacity[T]
	var group []ElementWithCapacity[T]
	currentCap := 0.0

	for _, item := range items {
		if currentCap+item.Capacity >= maxPerGroup && len(group) > 0 {
			result = append(result, group)
			group = []ElementWithCapacity[T]{}
			currentCap = 0
		}

		group = append(group, item)
		currentCap += item.Capacity
	}

	if len(group) > 0 {
		result = append(result, group)
	}

	return result, nil
}

// return tracking code string with 8 digit padding with zero
func MakePackageInhouseTrackingCode(packageId int) string {
	return fmt.Sprintf("JCW0000%07d", packageId)
}

type jsonData interface {
	~[]byte
}

func AppendJsonSliceEntry[T any, K jsonData](ctx context.Context, src *K, entry T) error {
	var logs []any

	if src != nil && len(*src) > 0 {
		if err := json.Unmarshal(*src, &logs); err != nil {
			slog.ErrorContext(ctx, "AppendJsonSliceEntry: failed to unmarshal slice entry", "error", err)
			return err
		}
	}

	logs = append(logs, entry)

	newLogs, err := json.Marshal(logs)
	if err != nil {
		slog.ErrorContext(ctx, "AppendJsonSliceEntry: failed to marshal logs slice", "error", err)
		return err
	}

	*src = newLogs

	return nil
}

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

func NewSetFromSlice[T comparable, K any](src []K, callback func(src K) T) Set[T] {
	out := make(Set[T])

	for _, elem := range src {
		newKey := callback(elem)
		out[newKey] = struct{}{}
	}

	return out
}

type Set[T comparable] map[T]struct{}

func (s Set[T]) Append(cmon ...T) {
	for _, elem := range cmon {
		s[elem] = struct{}{}
	}
}

func (s Set[T]) Members() []T {
	var out []T
	for key := range s {
		out = append(out, key)
	}

	return out
}

func NewIndexedSet[T any, K comparable](src []T, callback func(src T) K) IndexedSet[T, K] {
	return IndexMap(src, callback)
}

// index set is a map with key of type K and value of type T, used for quick lookups
type IndexedSet[T any, K comparable] map[K]T

// GetMany returns array of T for given array of K keys
func (i IndexedSet[T, K]) GetMany(reqs ...K) []T {
	out := []T{}
	for _, req := range reqs {
		elem := i[req]
		out = append(out, elem)
	}

	return out
}
