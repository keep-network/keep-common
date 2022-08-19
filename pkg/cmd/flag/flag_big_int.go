package flag

import (
	"fmt"
	"math/big"

	"github.com/spf13/pflag"
)

// BigIntVarFlag is a custom flag to handle `big.Int` type, that is not supported
// by `pflag.FlagSet`.
func BigIntVarFlag(f *pflag.FlagSet, p *BigIntFlagValue, name string, defaultValue *big.Int, usage string) {
	BigIntVarPFlag(f, p, name, "", defaultValue, usage)
}

// BigIntVarPFlag is a custom flag to handle `big.Int` type, that is not supported
// by `pflag.FlagSet`.
func BigIntVarPFlag(f *pflag.FlagSet, p *BigIntFlagValue, name string, short string, defaultValue *big.Int, usage string) {
	f.VarP(newBigIntValue(defaultValue, p), name, short, usage)
}

// BigIntFlagValue is a wrapper for big.Int to use as a flag value. The flag value
// supports setting `nil` as a default value.
type BigIntFlagValue struct {
	*big.Int
}

func newBigIntValue(val *big.Int, p *BigIntFlagValue) *BigIntFlagValue {
	if p == nil {
		p = &BigIntFlagValue{}
	}
	*p = BigIntFlagValue{val}
	return p
}

func (b *BigIntFlagValue) Set(s string) error {
	v := &big.Int{}
	v, ok := v.SetString(s, 0)
	if !ok {
		return fmt.Errorf("failed to parse as big.Int: %s", s)
	}
	b.Int = v

	return nil
}

func (b *BigIntFlagValue) Type() string {
	return "big.Int"
}

func (b *BigIntFlagValue) String() string {
	if b.Int == nil {
		return ""
	}
	return b.Int.String()
}
