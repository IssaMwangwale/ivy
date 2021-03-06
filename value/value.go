// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package value // import "robpike.io/ivy/value"

import (
	"fmt"
	"math/big"
	"strings"

	"robpike.io/ivy/config"
)

var conf *config.Config

func SetConfig(c *config.Config) {
	conf = c
}

type Value interface {
	String() string
	Eval(Context) Value

	toType(valueType) Value
}

// Error is the type we recognize as a recoverable run-time error.
type Error string

func (err Error) Error() string {
	return string(err)
}

// Errorf panics with the formatted string, with type Error.
func Errorf(format string, args ...interface{}) {
	panic(Error(fmt.Sprintf(format, args...)))
}

func Parse(s string) (Value, error) {
	// Is it a rational? If so, it's tricky.
	if strings.ContainsRune(s, '/') {
		elems := strings.Split(s, "/")
		if len(elems) != 2 {
			panic("bad rat")
		}
		num, err := Parse(elems[0])
		if err != nil {
			return nil, err
		}
		den, err := Parse(elems[1])
		if err != nil {
			return nil, err
		}
		// Common simple case.
		if whichType(num) == intType && whichType(den) == intType {
			return bigRatTwoInt64s(int64(num.(Int)), int64(den.(Int))).shrink(), nil
		}
		// General mix-em-up.
		rden := den.toType(bigRatType)
		if rden.(BigRat).Sign() == 0 {
			Errorf("zero denominator in rational")
		}
		return binaryBigRatOp(num.toType(bigRatType), (*big.Rat).Quo, rden), nil
	}
	// Not a rational, but might be something like 1.3e-2 and therefore
	// become a rational.
	i, err := setIntString(s)
	if err == nil {
		return i, nil
	}
	b, err := setBigIntString(s)
	if err == nil {
		return b.shrink(), nil
	}
	r, err := setBigRatFromFloatString(s) // We know there is no slash.
	if err == nil {
		return r.shrink(), nil
	}
	return nil, err
}

func bigInt64(x int64) BigInt {
	return BigInt{big.NewInt(x)}
}

func bigRatInt64(x int64) BigRat {
	return bigRatTwoInt64s(x, 1)
}

func bigRatTwoInt64s(x, y int64) BigRat {
	if y == 0 {
		Errorf("zero denominator in rational")
	}
	return BigRat{big.NewRat(x, y)}
}
