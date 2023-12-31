// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package expression

import (
	"fmt"
	"math"

	"github.com/pingcap/errors"

	"github.com/secretflow/scql/pkg/parser/mysql"
	"github.com/secretflow/scql/pkg/parser/opcode"
	"github.com/secretflow/scql/pkg/sessionctx"
	"github.com/secretflow/scql/pkg/types"
	"github.com/secretflow/scql/pkg/util/chunk"
)

var (
	_ functionClass = &logicAndFunctionClass{}
	_ functionClass = &logicOrFunctionClass{}
	_ functionClass = &logicXorFunctionClass{}
	_ functionClass = &unaryMinusFunctionClass{}
	_ functionClass = &isNullFunctionClass{}
	_ functionClass = &unaryNotFunctionClass{}
)

var (
	_ builtinFunc = &builtinLogicAndSig{}
	_ builtinFunc = &builtinLogicOrSig{}
	_ builtinFunc = &builtinLogicXorSig{}
	_ builtinFunc = &builtinUnaryMinusIntSig{}
	_ builtinFunc = &builtinUnaryMinusDecimalSig{}
	_ builtinFunc = &builtinUnaryMinusRealSig{}
	_ builtinFunc = &builtinDecimalIsNullSig{}
	_ builtinFunc = &builtinIntIsNullSig{}
	_ builtinFunc = &builtinRealIsNullSig{}
	_ builtinFunc = &builtinStringIsNullSig{}
	_ builtinFunc = &builtinUnaryNotRealSig{}
	_ builtinFunc = &builtinUnaryNotDecimalSig{}
	_ builtinFunc = &builtinUnaryNotIntSig{}
)

type logicAndFunctionClass struct {
	baseFunctionClass
}

func (c *logicAndFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	err := c.verifyArgs(args)
	if err != nil {
		return nil, err
	}
	args[0], err = wrapWithIsTrue(ctx, true, args[0])
	if err != nil {
		return nil, errors.Trace(err)
	}
	args[1], err = wrapWithIsTrue(ctx, true, args[1])
	if err != nil {
		return nil, errors.Trace(err)
	}

	bf := newBaseBuiltinFuncWithTp(ctx, args, types.ETInt, types.ETInt, types.ETInt)
	sig := &builtinLogicAndSig{bf}
	sig.tp.Flen = 1
	bf.tp.Flag |= mysql.IsBooleanFlag
	return sig, nil
}

type builtinLogicAndSig struct {
	baseBuiltinFunc
}

func (b *builtinLogicAndSig) Clone() builtinFunc {
	newSig := &builtinLogicAndSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

type isTrueOrFalseFunctionClass struct {
	baseFunctionClass
	op opcode.Op

	// keepNull indicates how this function treats a null input parameter.
	// If keepNull is true and the input parameter is null, the function will return null.
	// If keepNull is false, the null input parameter will be cast to 0.
	keepNull bool
}

func (c *isTrueOrFalseFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	if err := c.verifyArgs(args); err != nil {
		return nil, err
	}

	argTp := args[0].GetType().EvalType()
	if argTp == types.ETTimestamp || argTp == types.ETDatetime || argTp == types.ETDuration {
		argTp = types.ETInt
	} else if argTp == types.ETJson || argTp == types.ETString {
		argTp = types.ETReal
	}

	bf := newBaseBuiltinFuncWithTp(ctx, args, types.ETInt, argTp)
	bf.tp.Flen = 1
	bf.tp.Flag |= mysql.IsBooleanFlag
	var sig builtinFunc
	switch c.op {
	case opcode.IsTruth:
		switch argTp {
		case types.ETInt:
			sig = &builtinIntIsTrueSig{bf, c.keepNull}
		default:
			return nil, errors.Errorf("unexpected types.EvalType %v", argTp)
		}
	case opcode.IsFalsity:
		switch argTp {
		case types.ETInt:
			sig = &builtinIntIsFalseSig{bf, c.keepNull}
		default:
			return nil, errors.Errorf("unexpected types.EvalType %v", argTp)
		}
	}
	return sig, nil
}

type builtinIntIsFalseSig struct {
	baseBuiltinFunc
	keepNull bool
}

func (b *builtinIntIsFalseSig) Clone() builtinFunc {
	newSig := &builtinIntIsFalseSig{keepNull: b.keepNull}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

type builtinIntIsTrueSig struct {
	baseBuiltinFunc
	keepNull bool
}

func (b *builtinIntIsTrueSig) Clone() builtinFunc {
	newSig := &builtinIntIsTrueSig{keepNull: b.keepNull}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

type logicOrFunctionClass struct {
	baseFunctionClass
}

func (c *logicOrFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	err := c.verifyArgs(args)
	if err != nil {
		return nil, err
	}
	args[0], err = wrapWithIsTrue(ctx, true, args[0])
	if err != nil {
		return nil, errors.Trace(err)
	}
	args[1], err = wrapWithIsTrue(ctx, true, args[1])
	if err != nil {
		return nil, errors.Trace(err)
	}

	bf := newBaseBuiltinFuncWithTp(ctx, args, types.ETInt, types.ETInt, types.ETInt)
	bf.tp.Flen = 1
	bf.tp.Flag |= mysql.IsBooleanFlag
	sig := &builtinLogicOrSig{bf}
	return sig, nil
}

type builtinLogicOrSig struct {
	baseBuiltinFunc
}

func (b *builtinLogicOrSig) Clone() builtinFunc {
	newSig := &builtinLogicOrSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

type logicXorFunctionClass struct {
	baseFunctionClass
}

func (c *logicXorFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	err := c.verifyArgs(args)
	if err != nil {
		return nil, err
	}

	bf := newBaseBuiltinFuncWithTp(ctx, args, types.ETInt, types.ETInt, types.ETInt)
	sig := &builtinLogicXorSig{bf}
	sig.tp.Flen = 1
	bf.tp.Flag |= mysql.IsBooleanFlag
	return sig, nil
}

type builtinLogicXorSig struct {
	baseBuiltinFunc
}

func (b *builtinLogicXorSig) Clone() builtinFunc {
	newSig := &builtinLogicXorSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

type unaryMinusFunctionClass struct {
	baseFunctionClass
}

func (c *unaryMinusFunctionClass) handleIntOverflow(arg *Constant) (overflow bool) {
	if mysql.HasUnsignedFlag(arg.GetType().Flag) {
		uval := arg.Value.GetUint64()
		// -math.MinInt64 is 9223372036854775808, so if uval is more than 9223372036854775808, like
		// 9223372036854775809, -9223372036854775809 is less than math.MinInt64, overflow occurs.
		if uval > uint64(-math.MinInt64) {
			return true
		}
	} else {
		val := arg.Value.GetInt64()
		// The math.MinInt64 is -9223372036854775808, the math.MaxInt64 is 9223372036854775807,
		// which is less than abs(-9223372036854775808). When val == math.MinInt64, overflow occurs.
		if val == math.MinInt64 {
			return true
		}
	}
	return false
}

// typeInfer infers unaryMinus function return type. when the arg is an int constant and overflow,
// typerInfer will infers the return type as types.ETDecimal, not types.ETInt.
func (c *unaryMinusFunctionClass) typeInfer(argExpr Expression) (types.EvalType, bool) {
	tp := argExpr.GetType().EvalType()
	if tp != types.ETInt && tp != types.ETDecimal {
		tp = types.ETReal
	}

	overflow := false
	// TODO: Handle float overflow.
	if arg, ok := argExpr.(*Constant); ok && tp == types.ETInt {
		overflow = c.handleIntOverflow(arg)
		if overflow {
			tp = types.ETDecimal
		}
	}
	return tp, overflow
}

func (c *unaryMinusFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (sig builtinFunc, err error) {
	if err = c.verifyArgs(args); err != nil {
		return nil, err
	}

	argExpr, argExprTp := args[0], args[0].GetType()
	_, intOverflow := c.typeInfer(argExpr)

	var bf baseBuiltinFunc
	switch argExprTp.EvalType() {
	case types.ETInt:
		if intOverflow {
			bf = newBaseBuiltinFuncWithTp(ctx, args, types.ETDecimal, types.ETDecimal)
			sig = &builtinUnaryMinusDecimalSig{bf, true}
			//sig.setPbCode(tipb.ScalarFuncSig_UnaryMinusDecimal)
		} else {
			bf = newBaseBuiltinFuncWithTp(ctx, args, types.ETInt, types.ETInt)
			sig = &builtinUnaryMinusIntSig{bf}
			//sig.setPbCode(tipb.ScalarFuncSig_UnaryMinusInt)
		}
		bf.tp.Decimal = 0
	case types.ETDecimal:
		bf = newBaseBuiltinFuncWithTp(ctx, args, types.ETDecimal, types.ETDecimal)
		bf.tp.Decimal = argExprTp.Decimal
		sig = &builtinUnaryMinusDecimalSig{bf, false}
		//sig.setPbCode(tipb.ScalarFuncSig_UnaryMinusDecimal)
	case types.ETReal:
		bf = newBaseBuiltinFuncWithTp(ctx, args, types.ETReal, types.ETReal)
		sig = &builtinUnaryMinusRealSig{bf}
		//sig.setPbCode(tipb.ScalarFuncSig_UnaryMinusReal)
	default:
		tp := argExpr.GetType().Tp
		if types.IsTypeTime(tp) || tp == mysql.TypeDuration {
			bf = newBaseBuiltinFuncWithTp(ctx, args, types.ETDecimal, types.ETDecimal)
			sig = &builtinUnaryMinusDecimalSig{bf, false}
			//sig.setPbCode(tipb.ScalarFuncSig_UnaryMinusDecimal)
		} else {
			bf = newBaseBuiltinFuncWithTp(ctx, args, types.ETReal, types.ETReal)
			sig = &builtinUnaryMinusRealSig{bf}
			//sig.setPbCode(tipb.ScalarFuncSig_UnaryMinusReal)
		}
	}
	bf.tp.Flen = argExprTp.Flen + 1
	return sig, err
}

type builtinUnaryMinusIntSig struct {
	baseBuiltinFunc
}

func (b *builtinUnaryMinusIntSig) Clone() builtinFunc {
	newSig := &builtinUnaryMinusIntSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryMinusIntSig) evalInt(row chunk.Row) (res int64, isNull bool, err error) {
	var val int64
	val, isNull, err = b.args[0].EvalInt(b.ctx, row)
	if err != nil || isNull {
		return val, isNull, err
	}

	if mysql.HasUnsignedFlag(b.args[0].GetType().Flag) {
		uval := uint64(val)
		if uval > uint64(-math.MinInt64) {
			return 0, false, types.ErrOverflow.GenWithStackByArgs("BIGINT", fmt.Sprintf("-%v", uval))
		} else if uval == uint64(-math.MinInt64) {
			return math.MinInt64, false, nil
		}
	} else if val == math.MinInt64 {
		return 0, false, types.ErrOverflow.GenWithStackByArgs("BIGINT", fmt.Sprintf("-%v", val))
	}
	return -val, false, nil
}

type builtinUnaryMinusDecimalSig struct {
	baseBuiltinFunc

	constantArgOverflow bool
}

func (b *builtinUnaryMinusDecimalSig) Clone() builtinFunc {
	newSig := &builtinUnaryMinusDecimalSig{constantArgOverflow: b.constantArgOverflow}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryMinusDecimalSig) evalDecimal(row chunk.Row) (*types.MyDecimal, bool, error) {
	dec, isNull, err := b.args[0].EvalDecimal(b.ctx, row)
	if err != nil || isNull {
		return dec, isNull, err
	}
	return types.DecimalNeg(dec), false, nil
}

type builtinUnaryMinusRealSig struct {
	baseBuiltinFunc
}

func (b *builtinUnaryMinusRealSig) Clone() builtinFunc {
	newSig := &builtinUnaryMinusRealSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryMinusRealSig) evalReal(row chunk.Row) (float64, bool, error) {
	val, isNull, err := b.args[0].EvalReal(b.ctx, row)
	return -val, isNull, err
}

type isNullFunctionClass struct {
	baseFunctionClass
}

func (c *isNullFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	if err := c.verifyArgs(args); err != nil {
		return nil, err
	}
	argTp := args[0].GetType().EvalType()
	if argTp == types.ETTimestamp {
		argTp = types.ETDatetime
	} else if argTp == types.ETJson {
		argTp = types.ETString
	}
	bf := newBaseBuiltinFuncWithTp(ctx, args, types.ETInt, argTp)
	bf.tp.Flen = 1
	bf.tp.Flag |= mysql.IsBooleanFlag
	var sig builtinFunc
	switch argTp {
	case types.ETInt:
		sig = &builtinIntIsNullSig{bf}
		//sig.setPbCode(tipb.ScalarFuncSig_IntIsNull)
	case types.ETDecimal:
		sig = &builtinDecimalIsNullSig{bf}
		//sig.setPbCode(tipb.ScalarFuncSig_DecimalIsNull)
	case types.ETReal:
		sig = &builtinRealIsNullSig{bf}
		//sig.setPbCode(tipb.ScalarFuncSig_RealIsNull)
	case types.ETString:
		sig = &builtinStringIsNullSig{bf}
		//sig.setPbCode(tipb.ScalarFuncSig_StringIsNull)
	default:
		panic("unexpected types.EvalType")
	}
	return sig, nil
}

type builtinDecimalIsNullSig struct {
	baseBuiltinFunc
}

func (b *builtinDecimalIsNullSig) Clone() builtinFunc {
	newSig := &builtinDecimalIsNullSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func evalIsNull(isNull bool, err error) (int64, bool, error) {
	if err != nil {
		return 0, true, err
	}
	if isNull {
		return 1, false, nil
	}
	return 0, false, nil
}

func (b *builtinDecimalIsNullSig) evalInt(row chunk.Row) (int64, bool, error) {
	_, isNull, err := b.args[0].EvalDecimal(b.ctx, row)
	return evalIsNull(isNull, err)
}

type builtinIntIsNullSig struct {
	baseBuiltinFunc
}

func (b *builtinIntIsNullSig) Clone() builtinFunc {
	newSig := &builtinIntIsNullSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinIntIsNullSig) evalInt(row chunk.Row) (int64, bool, error) {
	_, isNull, err := b.args[0].EvalInt(b.ctx, row)
	return evalIsNull(isNull, err)
}

type builtinRealIsNullSig struct {
	baseBuiltinFunc
}

func (b *builtinRealIsNullSig) Clone() builtinFunc {
	newSig := &builtinRealIsNullSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinRealIsNullSig) evalInt(row chunk.Row) (int64, bool, error) {
	_, isNull, err := b.args[0].EvalReal(b.ctx, row)
	return evalIsNull(isNull, err)
}

type builtinStringIsNullSig struct {
	baseBuiltinFunc
}

func (b *builtinStringIsNullSig) Clone() builtinFunc {
	newSig := &builtinStringIsNullSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinStringIsNullSig) evalInt(row chunk.Row) (int64, bool, error) {
	_, isNull, err := b.args[0].EvalString(b.ctx, row)
	return evalIsNull(isNull, err)
}

type unaryNotFunctionClass struct {
	baseFunctionClass
}

func (c *unaryNotFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	if err := c.verifyArgs(args); err != nil {
		return nil, err
	}

	argTp := args[0].GetType().EvalType()
	if argTp == types.ETTimestamp || argTp == types.ETDatetime || argTp == types.ETDuration {
		argTp = types.ETInt
	} else if argTp == types.ETJson || argTp == types.ETString {
		argTp = types.ETReal
	}

	bf := newBaseBuiltinFuncWithTp(ctx, args, types.ETInt, argTp)
	bf.tp.Flen = 1
	bf.tp.Flag |= mysql.IsBooleanFlag

	var sig builtinFunc
	switch argTp {
	case types.ETReal:
		sig = &builtinUnaryNotRealSig{bf}
		//sig.setPbCode(tipb.ScalarFuncSig_UnaryNotReal)
	case types.ETDecimal:
		sig = &builtinUnaryNotDecimalSig{bf}
		//sig.setPbCode(tipb.ScalarFuncSig_UnaryNotDecimal)
	case types.ETInt:
		sig = &builtinUnaryNotIntSig{bf}
		//sig.setPbCode(tipb.ScalarFuncSig_UnaryNotInt)
	default:
		return nil, errors.Errorf("unexpected types.EvalType %v", argTp)
	}
	return sig, nil
}

type builtinUnaryNotRealSig struct {
	baseBuiltinFunc
}

func (b *builtinUnaryNotRealSig) Clone() builtinFunc {
	newSig := &builtinUnaryNotRealSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryNotRealSig) evalInt(row chunk.Row) (int64, bool, error) {
	arg, isNull, err := b.args[0].EvalReal(b.ctx, row)
	if isNull || err != nil {
		return 0, true, err
	}
	if arg == 0 {
		return 1, false, nil
	}
	return 0, false, nil
}

type builtinUnaryNotDecimalSig struct {
	baseBuiltinFunc
}

func (b *builtinUnaryNotDecimalSig) Clone() builtinFunc {
	newSig := &builtinUnaryNotDecimalSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryNotDecimalSig) evalInt(row chunk.Row) (int64, bool, error) {
	arg, isNull, err := b.args[0].EvalDecimal(b.ctx, row)
	if isNull || err != nil {
		return 0, true, err
	}
	if arg.IsZero() {
		return 1, false, nil
	}
	return 0, false, nil
}

type builtinUnaryNotIntSig struct {
	baseBuiltinFunc
}

func (b *builtinUnaryNotIntSig) Clone() builtinFunc {
	newSig := &builtinUnaryNotIntSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUnaryNotIntSig) evalInt(row chunk.Row) (int64, bool, error) {
	arg, isNull, err := b.args[0].EvalInt(b.ctx, row)
	if isNull || err != nil {
		return 0, true, err
	}
	if arg == 0 {
		return 1, false, nil
	}
	return 0, false, nil
}
