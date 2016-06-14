// Copyright ©2016 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package f64

import (
	"math"
	"runtime"
	"testing"
)

var (
	nan = math.NaN()
	inf = math.Inf(1)
)

func guardVector(v []float64, g float64, g_ln int) (guarded []float64) {
	guarded = make([]float64, len(v)+g_ln*2)
	copy(guarded[g_ln:], v)
	for i := 0; i < g_ln; i++ {
		guarded[i] = g
		guarded[len(guarded)-1-i] = g
	}
	return guarded
}

func validGuard(v []float64, g float64, g_ln int) bool {
	for i := 0; i < g_ln; i++ {
		if !same(v[i], g) || v[len(v)-1-i] != g {
			return false
		}
	}
	return true
}

func guardIncVector(v []float64, g float64, inc, g_ln int) (guarded []float64) {
	s_ln := len(v) * inc
	guarded = make([]float64, s_ln+g_ln*2)
	for i, j := 0, 0; i < len(guarded); i++ {
		switch {
		case i < g_ln, i > g_ln+s_ln:
			guarded[i] = g
		case (i-g_ln)%(inc) == 0 && j < len(v):
			guarded[i] = v[j]
			j++
		default:
			guarded[i] = g
		}
	}
	return guarded
}

func validIncGuard(t *testing.T, v []float64, g float64, inc, g_ln int) {
	s_ln := len(v) - 2*g_ln
	for i := range v {
		switch {
		case same(v[i], g):
			// Correct value
		case i < g_ln:
			t.Error("Front guard violated at", i, v[:g_ln])
		case i > g_ln+s_ln:
			t.Error("Back guard violated at", i-g_ln-s_ln, v[g_ln+s_ln:])
		case (i-g_ln)%inc == 0 && (i-g_ln)/inc < len(v):
			// Ignore input values
		default:
			t.Error("Internal guard violated at", i-g_ln, v[g_ln:g_ln+s_ln])
		}
	}
}

func same(a, b float64) bool {
	return a == b || (math.IsNaN(a) && math.IsNaN(b))
}

func TestAbsSum(t *testing.T) {
	var src_gd float64 = 1
	for j, v := range []struct {
		ex  float64
		src []float64
	}{
		{0, []float64{}},
		{2, []float64{2}},
		{6, []float64{1, 2, 3}},
		{6, []float64{-1, -2, -3}},
		{nan, []float64{nan}},
		{40, []float64{8, -8, 8, -8, 8}},
		{5, []float64{0, 1, 0, -1, 0, 1, 0, -1, 0, 1}},
	} {
		g_ln := 4 + j%2
		v.src = guardVector(v.src, src_gd, g_ln)
		src := v.src[g_ln : len(v.src)-g_ln]
		ret := AbsSum(src)
		if !same(ret, v.ex) {
			t.Error("Test", j, "AbsSum error Got:", ret, "Expected:", v.ex)
			t.Error(src)
		}
		if !validGuard(v.src, src_gd, g_ln) {
			t.Error("Test", j, "Guard violated in x vector", v.src[:g_ln], v.src[len(v.src)-g_ln:])
		}
		runtime.GC()
	}
}

func TestAbsSumInc(t *testing.T) {
	var src_gd float64 = 1
	for j, v := range []struct {
		inc int
		ex  float64
		src []float64
	}{
		{2, 0, []float64{}},
		{3, 2, []float64{2}},
		{10, 6, []float64{1, 2, 3}},
		{5, 6, []float64{-1, -2, -3}},
		{3, nan, []float64{nan}},
		{15, 40, []float64{8, -8, 8, -8, 8}},
		{1, 5, []float64{0, 1, 0, -1, 0, 1, 0, -1, 0, 1}},
	} {
		g_ln, ln := 4+j%2, len(v.src)
		v.src = guardIncVector(v.src, src_gd, v.inc, g_ln)
		src := v.src[g_ln : len(v.src)-g_ln]
		ret := AbsSumInc(src, ln, v.inc)
		if !same(ret, v.ex) {
			t.Error("Test", j, "AbsSum error Got:", ret, "Expected:", v.ex)
			t.Error(src)
		}
		validIncGuard(t, v.src, src_gd, v.inc, g_ln)
		runtime.GC()
	}
}

func TestAdd(t *testing.T) {
	var src_gd, dst_gd float64 = 1, 0
	for j, v := range []struct {
		dst, src, expect []float64
	}{
		{[]float64{1}, []float64{0}, []float64{1}},
		{[]float64{1, 2, 3}, []float64{1}, []float64{2, 2, 3}},
		{[]float64{}, []float64{}, []float64{}},
		{[]float64{1}, []float64{nan}, []float64{nan}},
		{[]float64{8, 8, 8, 8, 8},
			[]float64{2, 4, nan, 8, 9},
			[]float64{10, 12, nan, 16, 17}},
		{[]float64{0, 1, 2, 3, 4},
			[]float64{-inf, 4, nan, 8, 9},
			[]float64{-inf, 5, nan, 11, 13}},
		{make([]float64, 50)[1:49],
			make([]float64, 50)[1:49],
			make([]float64, 50)[1:49]},
	} {
		g_ln := 4 + j%2
		v.src, v.dst = guardVector(v.src, src_gd, g_ln), guardVector(v.dst, dst_gd, g_ln)
		src, dst := v.src[g_ln:len(v.src)-g_ln], v.dst[g_ln:len(v.dst)-g_ln]
		Add(dst, src)
		for i := range v.expect {
			if !same(dst[i], v.expect[i]) {
				t.Error("Test", j, "Add error at", i, "Got:", dst[i], "Expected:", v.expect[i])
			}
		}
		if !validGuard(v.src, src_gd, g_ln) {
			t.Error("Test", j, "Guard violated in x vector", v.src[:g_ln], v.src[len(v.src)-g_ln:])
		}
		if !validGuard(v.dst, dst_gd, g_ln) {
			t.Error("Test", j, "Guard violated in y vector", v.dst[:g_ln], v.dst[len(v.dst)-g_ln:])
		}
	}
	runtime.GC()
}

func TestAddConst(t *testing.T) {
	var src_gd float64 = 0
	for j, v := range []struct {
		alpha       float64
		src, expect []float64
	}{
		{1, []float64{0}, []float64{1}},
		{5, []float64{}, []float64{}},
		{1, []float64{nan}, []float64{nan}},
		{8, []float64{2, 4, nan, 8, 9}, []float64{10, 12, nan, 16, 17}},
		{inf, []float64{-inf, 4, nan, 8, 9}, []float64{nan, inf, nan, inf, inf}},
	} {
		g_ln := 4 + j%2
		v.src = guardVector(v.src, src_gd, g_ln)
		src := v.src[g_ln : len(v.src)-g_ln]
		AddConst(v.alpha, src)
		for i := range v.expect {
			if !same(src[i], v.expect[i]) {
				t.Error("Test", j, "AddConst error at", i, "Got:", src[i], "Expected:", v.expect[i])
			}
		}
		if !validGuard(v.src, src_gd, g_ln) {
			t.Error("Test", j, "Guard violated in x vector", v.src[:g_ln], v.src[len(v.src)-g_ln:])
		}
	}
	runtime.GC()
}

func TestCumSum(t *testing.T) {
	var src_gd, dst_gd float64 = -1, 0
	for j, v := range []struct {
		dst, src, expect []float64
	}{
		{[]float64{}, []float64{}, []float64{}},
		{[]float64{0}, []float64{1}, []float64{1}},
		{[]float64{nan}, []float64{nan}, []float64{nan}},
		{[]float64{0, 0, 0}, []float64{1, 2, 3}, []float64{1, 3, 6}},
		{[]float64{0, 0, 0, 0}, []float64{1, 2, 3}, []float64{1, 3, 6}},
		{[]float64{0, 0, 0, 0}, []float64{1, 2, 3, 4}, []float64{1, 3, 6, 10}},
		{[]float64{1, nan, nan, 1, 1},
			[]float64{1, 1, nan, 1, 1},
			[]float64{1, 2, nan, nan, nan}},
		{[]float64{nan, 4, inf, -inf, 9},
			[]float64{inf, 4, nan, -inf, 9},
			[]float64{inf, inf, nan, nan, nan}},
		{make([]float64, 16),
			[]float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}},
	} {
		g_ln := 4 + j%2
		v.src, v.dst = guardVector(v.src, src_gd, g_ln), guardVector(v.dst, dst_gd, g_ln)
		src, dst := v.src[g_ln:len(v.src)-g_ln], v.dst[g_ln:len(v.dst)-g_ln]
		ret := CumSum(dst, src)
		for i := range v.expect {
			if !same(ret[i], v.expect[i]) {
				t.Error("Test", j, "CumSum error at", i, "Got:", ret[i], "Expected:", v.expect[i])
				t.Error(ret, v.expect)
			}
			if !same(ret[i], dst[i]) {
				t.Error("Test", j, "CumSum ret/dst mismatch", i, "Ret:", ret[i], "Dst:", dst[i])
			}
		}
		if !validGuard(v.src, src_gd, g_ln) {
			t.Error("Test", j, "Guard violated in x vector", v.src[:g_ln], v.src[len(v.src)-g_ln:])
		}
		if !validGuard(v.dst, dst_gd, g_ln) {
			t.Error("Test", j, "Guard violated in y vector", v.dst[:g_ln], v.dst[len(v.dst)-g_ln:])
		}
	}
	runtime.GC()
}

func TestCumProd(t *testing.T) {
	var src_gd, dst_gd float64 = -1, 1
	for j, v := range []struct {
		dst, src, expect []float64
	}{
		{[]float64{}, []float64{}, []float64{}},
		{[]float64{1}, []float64{1}, []float64{1}},
		{[]float64{nan}, []float64{nan}, []float64{nan}},
		{[]float64{0, 0, 0, 0}, []float64{1, 2, 3, 4}, []float64{1, 2, 6, 24}},
		{[]float64{0, 0, 0}, []float64{1, 2, 3}, []float64{1, 2, 6}},
		{[]float64{0, 0, 0, 0}, []float64{1, 2, 3}, []float64{1, 2, 6}},
		{[]float64{nan, 1, nan, 1, 0},
			[]float64{1, 1, nan, 1, 1},
			[]float64{1, 1, nan, nan, nan}},
		{[]float64{nan, 4, nan, -inf, 9},
			[]float64{inf, 4, nan, -inf, 9},
			[]float64{inf, inf, nan, nan, nan}},
		{make([]float64, 18),
			[]float64{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			[]float64{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536}},
	} {
		g_ln := 4 + j%2
		v.src, v.dst = guardVector(v.src, src_gd, g_ln), guardVector(v.dst, dst_gd, g_ln)
		src, dst := v.src[g_ln:len(v.src)-g_ln], v.dst[g_ln:len(v.dst)-g_ln]
		ret := CumProd(dst, src)
		for i := range v.expect {
			if !same(ret[i], v.expect[i]) {
				t.Error("Test", j, "CumProd error at", i, "Got:", ret[i], "Expected:", v.expect[i])
			}
			if !same(ret[i], dst[i]) {
				t.Error("Test", j, "CumProd ret/dst mismatch", i, "Ret:", ret[i], "Dst:", dst[i])
			}
		}
		if !validGuard(v.src, src_gd, g_ln) {
			t.Error("Test", j, "Guard violated in x vector", v.src[:g_ln], v.src[len(v.src)-g_ln:])
		}
		if !validGuard(v.dst, dst_gd, g_ln) {
			t.Error("Test", j, "Guard violated in y vector", v.dst[:g_ln], v.dst[len(v.dst)-g_ln:])
		}
	}
	runtime.GC()
}

func TestDiv(t *testing.T) {
	var src_gd, dst_gd float64 = -1, 0.5
	for j, v := range []struct {
		dst, src, expect []float64
	}{
		{[]float64{1}, []float64{1}, []float64{1}},
		{[]float64{nan}, []float64{nan}, []float64{nan}},
		{[]float64{1, 2, 3, 4}, []float64{1, 2, 3, 4}, []float64{1, 1, 1, 1}},
		{[]float64{2, 4, 6}, []float64{1, 2, 3}, []float64{2, 2, 2}},
		{[]float64{0, 0, 0, 0}, []float64{1, 2, 3}, []float64{0, 0, 0}},
		{[]float64{nan, 1, nan, 1, 0},
			[]float64{1, 1, nan, 1, 1},
			[]float64{nan, 1, nan, 1, 0}},
		{[]float64{inf, 4, nan, -inf, 9},
			[]float64{inf, 4, nan, -inf, 3},
			[]float64{nan, 1, nan, nan, 3}},
	} {
		g_ln := 4 + j%2
		v.src, v.dst = guardVector(v.src, src_gd, g_ln), guardVector(v.dst, dst_gd, g_ln)
		src, dst := v.src[g_ln:len(v.src)-g_ln], v.dst[g_ln:len(v.dst)-g_ln]
		Div(dst, src)
		for i := range v.expect {
			if !same(dst[i], v.expect[i]) {
				t.Error("Test", j, "Div error at", i, "Got:", dst[i], "Expected:", v.expect[i])
			}
		}
		if !validGuard(v.src, src_gd, g_ln) {
			t.Error("Test", j, "Guard violated in x vector", v.src[:g_ln], v.src[len(v.src)-g_ln:])
		}
		if !validGuard(v.dst, dst_gd, g_ln) {
			t.Error("Test", j, "Guard violated in y vector", v.dst[:g_ln], v.dst[len(v.dst)-g_ln:])
		}
		runtime.GC()
	}
}

func TestDivTo(t *testing.T) {
	var dst_gd, x_gd, y_gd float64 = -1, 0.5, 0.25
	for j, v := range []struct {
		dst, x, y, expect []float64
	}{
		{[]float64{1}, []float64{1}, []float64{1}, []float64{1}},
		{[]float64{1}, []float64{nan}, []float64{nan}, []float64{nan}},
		{[]float64{-2, -2, -2}, []float64{1, 2, 3},
			[]float64{1, 2, 3}, []float64{1, 1, 1}},
		{[]float64{0, 0, 0}, []float64{2, 4, 6},
			[]float64{1, 2, 3, 4}, []float64{2, 2, 2}},
		{[]float64{-1, -1, -1}, []float64{0, 0, 0},
			[]float64{1, 2, 3}, []float64{0, 0, 0}},
		{[]float64{inf, inf, inf, inf, inf}, []float64{nan, 1, nan, 1, 0},
			[]float64{1, 1, nan, 1, 1}, []float64{nan, 1, nan, 1, 0}},
		{[]float64{0, 0, 0, 0, 0}, []float64{inf, 4, nan, -inf, 9},
			[]float64{inf, 4, nan, -inf, 3}, []float64{nan, 1, nan, nan, 3}},
	} {
		g_ln := 4 + j%2
		v.y, v.x = guardVector(v.y, y_gd, g_ln), guardVector(v.x, x_gd, g_ln)
		y, x := v.y[g_ln:len(v.y)-g_ln], v.x[g_ln:len(v.x)-g_ln]
		v.dst = guardVector(v.dst, dst_gd, g_ln)
		dst := v.dst[g_ln : len(v.dst)-g_ln]
		ret := DivTo(dst, x, y)
		for i := range v.expect {
			if !same(ret[i], v.expect[i]) {
				t.Error("Test", j, "DivTo error at", i, "Got:", x[i], "Expected:", v.expect[i])
			}
			if !same(ret[i], dst[i]) {
				t.Error("Test", j, "DivTo ret/dst mismatch", i, "Ret:", ret[i], "X:", dst[i])
			}
		}
		if !validGuard(v.y, y_gd, g_ln) {
			t.Error("Test", j, "Guard violated in x vector", v.y[:g_ln], v.y[len(v.y)-g_ln:])
		}
		if !validGuard(v.x, x_gd, g_ln) {
			t.Error("Test", j, "Guard violated in y vector", v.x[:g_ln], v.x[len(v.x)-g_ln:])
		}
		if !validGuard(v.dst, dst_gd, g_ln) {
			t.Error("Test", j, "Guard violated in dst vector", v.dst[:g_ln], v.dst[len(v.dst)-g_ln:])
		}
	}
	runtime.GC()
}

func TestL1Norm(t *testing.T) {
	var t_gd, s_gd float64 = -inf, inf
	for j, v := range []struct {
		s, t   []float64
		expect float64
	}{
		{[]float64{1}, []float64{1}, 0},
		{[]float64{nan}, []float64{nan}, nan},
		{[]float64{1, 2, 3, 4}, []float64{1, 2, 3, 4}, 0},
		{[]float64{2, 4, 6}, []float64{1, 2, 3, 4}, 6},
		{[]float64{0, 0, 0}, []float64{1, 2, 3}, 6},
		{[]float64{0, -4, -10}, []float64{1, 2, 3}, 20},
		{[]float64{0, 1, 0, 1, 0}, []float64{1, 1, inf, 1, 1}, inf},
		{[]float64{inf, 4, nan, -inf, 9}, []float64{inf, 4, nan, -inf, 3}, nan},
	} {
		g_ln := 4 + j%2
		v.s, v.t = guardVector(v.s, s_gd, g_ln), guardVector(v.t, t_gd, g_ln)
		s_lc, t_lc := v.s[g_ln:len(v.s)-g_ln], v.t[g_ln:len(v.t)-g_ln]
		ret := L1Norm(s_lc, t_lc)
		if !same(ret, v.expect) {
			t.Error("Test", j, "L1Norm error. Got:", ret, "Expected:", v.expect)
		}
		if !validGuard(v.s, s_gd, g_ln) {
			t.Error("Test", j, "Guard violated in s vector", v.s[:g_ln], v.s[len(v.s)-g_ln:])
		}
		if !validGuard(v.t, t_gd, g_ln) {
			t.Error("Test", j, "Guard violated in t vector", v.t[:g_ln], v.t[len(v.t)-g_ln:])
		}
	}
	runtime.GC()
}

func TestLinfNorm(t *testing.T) {
	var t_gd, s_gd float64 = 0, inf
	for j, v := range []struct {
		s, t   []float64
		expect float64
	}{
		{[]float64{}, []float64{}, 0},
		{[]float64{1}, []float64{1}, 0},
		{[]float64{nan}, []float64{nan}, nan},
		{[]float64{1, 2, 3, 4}, []float64{1, 2, 3, 4}, 0},
		{[]float64{2, 4, 6}, []float64{1, 2, 3, 4}, 3},
		{[]float64{0, 0, 0}, []float64{1, 2, 3}, 3},
		{[]float64{0, 1, 0, 1, 0}, []float64{1, 1, inf, 1, 1}, inf},
		{[]float64{inf, 4, nan, -inf, 9}, []float64{inf, 4, nan, -inf, 3}, 6},
	} {
		g_ln := 4 + j%2
		v.s, v.t = guardVector(v.s, s_gd, g_ln), guardVector(v.t, t_gd, g_ln)
		s_lc, t_lc := v.s[g_ln:len(v.s)-g_ln], v.t[g_ln:len(v.t)-g_ln]
		ret := LinfNorm(s_lc, t_lc)
		if !same(ret, v.expect) {
			t.Error("Test", j, "LinfNorm error. Got:", ret, "Expected:", v.expect)
		}
		if !validGuard(v.s, s_gd, g_ln) {
			t.Error("Test", j, "Guard violated in s vector", v.s[:g_ln], v.s[len(v.s)-g_ln:])
		}
		if !validGuard(v.t, t_gd, g_ln) {
			t.Error("Test", j, "Guard violated in t vector", v.t[:g_ln], v.t[len(v.t)-g_ln:])
		}
	}
	runtime.GC()
}
