package common

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// MakeTimestampMilli returns current ms timestamp
func MakeTimestampMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func TryParseInt(val string, def int) int {
	i, err := strconv.Atoi(val)

	if err != nil {
		return def
	}

	return i
}

func Abs(val int64) int64 {
	if val < 0 {
		return -val
	}

	return val
}

func GetOrDefaultMultipleString(variable string, def []string) []string {
	var vars []string
	for _, x := range strings.Split(variable, ",") {
		y := strings.TrimSpace(x)
		if len(y) != 0 {
			vars = append(vars, y)
		}
	}

	if len(vars) == 0 {
		return def
	}

	return vars
}

// RemoveUInt64Duplicates removes duplicate values from uint64 slice
func RemoveStringDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	var result []string

	for v := range elements {
		if !encountered[elements[v]] == true {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}

// GetRandomFromStringSlice returns a random string from a slice
func GetRandomFromStringSlice(s []string) string {
	if len(s) == 0 {
		return ""
	}

	rand.Seed(time.Now().UnixNano())
	return s[rand.Intn(len(s))]
}

// MemoryUsage returns some memory usage stats which can be useful in output
func MemoryUsage() string {
	bToMb := func(b uint64) uint64 {
		return b / 1024 / 1024
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	result := fmt.Sprintf("memoryusage::Alloc = %v MB::TotalAlloc = %v MB::Sys = %v MB::tNumGC = %v", bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC)
	return result
}

// MemoryAllocatedMb returns how much memory is allocated
func MemoryAllocatedMb() uint64 {
	bToMb := func(b uint64) uint64 {
		return b / 1024 / 1024
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return bToMb(m.Alloc)
}

func SieveOfEratosthenes(N int) (primes []int) {
	b := make([]bool, N)
	for i := 2; i < N; i++ {
		if b[i] == true {
			continue
		}
		primes = append(primes, i)
		for k := i * i; k < N; k += i {
			b[k] = true
		}
	}
	return
}

// Call with wait
//
//		if value,ok := WithTimeout(func()interface{}{return <- inbox}, time.Second); ok {
//	   // returned
//	} else {
//
//	   // didn't return
//	}
//
// Send with wait
// _,ok = WithTimeout(func()interface{}{outbox <- myValue; return nil}, time.Second)
//
//	if !ok{...
func WithTimeout(delegate func() interface{}, timeout time.Duration) (ret interface{}, ok bool) {
	ch := make(chan interface{}, 1) // buffered
	go func() { ch <- delegate() }()
	select {
	case ret = <-ch:
		return ret, true
	case <-time.After(timeout):
	}
	return nil, false
}

// https://searchcode.com/file/623770908/registry/api/v2/routes_test.go/
// -------------- START LICENSED CODE --------------
// The following code is derivative of https://github.com/google/gofuzz
// gofuzz is licensed under the Apache License, Version 2.0, January 2004,
// a copy of which can be found in the LICENSE file at the root of this
// repository.

// These functions allow us to generate strings containing only multibyte
// characters that are invalid in our URLs. They are used above for fuzzing
// to ensure we always get 404s on these invalid strings
type charRange struct {
	first, last rune
}

// choose returns a random unicode character from the given range, using the
// given randomness source.
func (r *charRange) choose() rune {
	count := int64(r.last - r.first)
	return r.first + rune(rand.Int63n(count))
}

var unicodeRanges = []charRange{
	{'\u00a0', '\u02af'}, // Multi-byte encoded characters
	{'\u4e00', '\u9fff'}, // Common CJK (even longer encodings)
}

func RandomString(length int) string {
	runes := make([]rune, length)
	for i := range runes {
		runes[i] = unicodeRanges[rand.Intn(len(unicodeRanges))].choose()
	}
	return string(runes)
}

// -------------- END LICENSED CODE --------------
