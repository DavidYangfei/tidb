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

package mysql

import (
	. "github.com/pingcap/check"
	"strings"
)

var _ = Suite(&testMyDecimalSuite{})

type testMyDecimalSuite struct {
}

func (s *testMyDecimalSuite) TestFromInt(c *C) {
	cases := []struct {
		input  int64
		output string
	}{
		{-12345, "-12345"},
		{-1, "-1"},
		{1, "1"},
		{-9223372036854775807, "-9223372036854775807"},
		{-9223372036854775808, "-9223372036854775808"},
	}
	for _, ca := range cases {
		var dec MyDecimal
		dec.FromInt(ca.input)
		str, err := dec.ToString(0, 0)
		c.Check(err, IsNil)
		c.Check(string(str), Equals, ca.output)
	}
}

func (s *testMyDecimalSuite) TestFromUint(c *C) {
	cases := []struct {
		input  uint64
		output string
	}{
		{12345, "12345"},
		{0, "0"},
		{18446744073709551615, "18446744073709551615"},
	}
	for _, ca := range cases {
		var dec MyDecimal
		dec.FromUint(ca.input)
		str, err := dec.ToString(0, 0)
		c.Check(err, IsNil)
		c.Check(string(str), Equals, ca.output)
	}
}

func (s *testMyDecimalSuite) TestToInt(c *C) {
	cases := []struct {
		input  string
		output int64
		err    error
	}{
		{"18446744073709551615", 9223372036854775807, ErrOverflow},
		{"-1", -1, nil},
		{"1", 1, nil},
		{"-1.23", -1, ErrTruncated},
		{"-9223372036854775807", -9223372036854775807, nil},
		{"-9223372036854775808", -9223372036854775808, nil},
		{"9223372036854775808", 9223372036854775807, ErrOverflow},
		{"-9223372036854775809", -9223372036854775808, ErrOverflow},
	}
	for _, ca := range cases {
		var dec MyDecimal
		dec.FromString([]byte(ca.input))
		result, ec := dec.ToInt()
		c.Check(ec, Equals, ca.err)
		c.Check(result, Equals, ca.output)
	}
}

func (s *testMyDecimalSuite) TestToUint(c *C) {
	cases := []struct {
		input  string
		output uint64
		err    error
	}{
		{"12345", 12345, nil},
		{"0", 0, nil},
		/* ULLONG_MAX = 18446744073709551615ULL */
		{"18446744073709551615", 18446744073709551615, nil},
		{"18446744073709551616", 18446744073709551615, ErrOverflow},
		{"-1", 0, ErrOverflow},
		{"1.23", 1, ErrTruncated},
		{"9999999999999999999999999.000", 18446744073709551615, ErrOverflow},
	}
	for _, ca := range cases {
		var dec MyDecimal
		dec.FromString([]byte(ca.input))
		result, ec := dec.ToUint()
		c.Check(ec, Equals, ca.err)
		c.Check(result, Equals, ca.output)
	}
}

func (s *testMyDecimalSuite) TestFromFloat(c *C) {
	cases := []struct {
		s string
		f float64
	}{
		{"12345", 12345},
		{"123.45", 123.45},
		{"-123.45", -123.45},
		{"0.00012345000098765", 0.00012345000098765},
		{"1234500009876.5", 1234500009876.5},
	}
	for _, ca := range cases {
		var dec MyDecimal
		dec.FromFloat64(ca.f)
		str, err := dec.ToString(0, 0)
		c.Check(err, IsNil)
		c.Check(string(str), Equals, ca.s)
	}
}

func (s *testMyDecimalSuite) TestToFloat(c *C) {
	cases := []struct {
		s string
		f float64
	}{
		{"12345", 12345},
		{"123.45", 123.45},
		{"-123.45", -123.45},
		{"0.00012345000098765", 0.00012345000098765},
		{"1234500009876.5", 1234500009876.5},
	}
	for _, ca := range cases {
		var dec MyDecimal
		dec.FromString([]byte(ca.s))
		f, err := dec.ToFloat64()
		c.Check(err, IsNil)
		c.Check(f, Equals, ca.f)
	}
}

func (s *testMyDecimalSuite) TestShift(c *C) {
	type tcase struct {
		input  string
		shift  int
		output string
		err    error
	}
	var dotest = func(c *C, cases []tcase) {
		for _, ca := range cases {
			var dec MyDecimal
			err := dec.FromString([]byte(ca.input))
			c.Check(err, IsNil)
			origin := dec
			err = dec.Shift(ca.shift)
			c.Check(err, Equals, ca.err)
			result, err := dec.ToString(0, 0)
			c.Check(err, IsNil)
			c.Check(string(result), Equals, ca.output, Commentf("origin:%s\ndec:%s", origin.String(), origin.String()))
		}
	}
	wordBufLen = maxWordBufLen
	cases := []tcase{
		{"123.123", 1, "1231.23", nil},
		{"123457189.123123456789000", 1, "1234571891.23123456789", nil},
		{"123457189.123123456789000", 8, "12345718912312345.6789", nil},
		{"123457189.123123456789000", 9, "123457189123123456.789", nil},
		{"123457189.123123456789000", 10, "1234571891231234567.89", nil},
		{"123457189.123123456789000", 17, "12345718912312345678900000", nil},
		{"123457189.123123456789000", 18, "123457189123123456789000000", nil},
		{"123457189.123123456789000", 19, "1234571891231234567890000000", nil},
		{"123457189.123123456789000", 26, "12345718912312345678900000000000000", nil},
		{"123457189.123123456789000", 27, "123457189123123456789000000000000000", nil},
		{"123457189.123123456789000", 28, "1234571891231234567890000000000000000", nil},
		{"000000000000000000000000123457189.123123456789000", 26, "12345718912312345678900000000000000", nil},
		{"00000000123457189.123123456789000", 27, "123457189123123456789000000000000000", nil},
		{"00000000000000000123457189.123123456789000", 28, "1234571891231234567890000000000000000", nil},
		{"123", 1, "1230", nil},
		{"123", 10, "1230000000000", nil},
		{".123", 1, "1.23", nil},
		{".123", 10, "1230000000", nil},
		{".123", 14, "12300000000000", nil},
		{"000.000", 1000, "0", nil},
		{"000.", 1000, "0", nil},
		{".000", 1000, "0", nil},
		{"1", 1000, "1", ErrOverflow},
		{"123.123", -1, "12.3123", nil},
		{"123987654321.123456789000", -1, "12398765432.1123456789", nil},
		{"123987654321.123456789000", -2, "1239876543.21123456789", nil},
		{"123987654321.123456789000", -3, "123987654.321123456789", nil},
		{"123987654321.123456789000", -8, "1239.87654321123456789", nil},
		{"123987654321.123456789000", -9, "123.987654321123456789", nil},
		{"123987654321.123456789000", -10, "12.3987654321123456789", nil},
		{"123987654321.123456789000", -11, "1.23987654321123456789", nil},
		{"123987654321.123456789000", -12, "0.123987654321123456789", nil},
		{"123987654321.123456789000", -13, "0.0123987654321123456789", nil},
		{"123987654321.123456789000", -14, "0.00123987654321123456789", nil},
		{"00000087654321.123456789000", -14, "0.00000087654321123456789", nil},
	}
	dotest(c, cases)
	wordBufLen = 2
	cases = []tcase{
		{"123.123", -2, "1.23123", nil},
		{"123.123", -3, "0.123123", nil},
		{"123.123", -6, "0.000123123", nil},
		{"123.123", -7, "0.0000123123", nil},
		{"123.123", -15, "0.000000000000123123", nil},
		{"123.123", -16, "0.000000000000012312", ErrTruncated},
		{"123.123", -17, "0.000000000000001231", ErrTruncated},
		{"123.123", -18, "0.000000000000000123", ErrTruncated},
		{"123.123", -19, "0.000000000000000012", ErrTruncated},
		{"123.123", -20, "0.000000000000000001", ErrTruncated},
		{"123.123", -21, "0", ErrTruncated},
		{".000000000123", -1, "0.0000000000123", nil},
		{".000000000123", -6, "0.000000000000000123", nil},
		{".000000000123", -7, "0.000000000000000012", ErrTruncated},
		{".000000000123", -8, "0.000000000000000001", ErrTruncated},
		{".000000000123", -9, "0", ErrTruncated},
		{".000000000123", 1, "0.00000000123", nil},
		{".000000000123", 8, "0.0123", nil},
		{".000000000123", 9, "0.123", nil},
		{".000000000123", 10, "1.23", nil},
		{".000000000123", 17, "12300000", nil},
		{".000000000123", 18, "123000000", nil},
		{".000000000123", 19, "1230000000", nil},
		{".000000000123", 20, "12300000000", nil},
		{".000000000123", 21, "123000000000", nil},
		{".000000000123", 22, "1230000000000", nil},
		{".000000000123", 23, "12300000000000", nil},
		{".000000000123", 24, "123000000000000", nil},
		{".000000000123", 25, "1230000000000000", nil},
		{".000000000123", 26, "12300000000000000", nil},
		{".000000000123", 27, "123000000000000000", nil},
		{".000000000123", 28, "0.000000000123", ErrOverflow},
		{"123456789.987654321", -1, "12345678.998765432", ErrTruncated},
		{"123456789.987654321", -2, "1234567.899876543", ErrTruncated},
		{"123456789.987654321", -8, "1.234567900", ErrTruncated},
		{"123456789.987654321", -9, "0.123456789987654321", nil},
		{"123456789.987654321", -10, "0.012345678998765432", ErrTruncated},
		{"123456789.987654321", -17, "0.000000001234567900", ErrTruncated},
		{"123456789.987654321", -18, "0.000000000123456790", ErrTruncated},
		{"123456789.987654321", -19, "0.000000000012345679", ErrTruncated},
		{"123456789.987654321", -26, "0.000000000000000001", ErrTruncated},
		{"123456789.987654321", -27, "0", ErrTruncated},
		{"123456789.987654321", 1, "1234567900", ErrTruncated},
		{"123456789.987654321", 2, "12345678999", ErrTruncated},
		{"123456789.987654321", 4, "1234567899877", ErrTruncated},
		{"123456789.987654321", 8, "12345678998765432", ErrTruncated},
		{"123456789.987654321", 9, "123456789987654321", nil},
		{"123456789.987654321", 10, "123456789.987654321", ErrOverflow},
		{"123456789.987654321", 0, "123456789.987654321", nil},
	}
	dotest(c, cases)
	wordBufLen = maxWordBufLen
}

func (s *testMyDecimalSuite) TestRound(c *C) {
	type tcase struct {
		input  string
		scale  int
		output string
		err    error
	}
	var doTest = func(c *C, cases []tcase) {
		for _, ca := range cases {
			var dec MyDecimal
			dec.FromString([]byte(ca.input))
			var rounded MyDecimal
			err := dec.Round(&rounded, ca.scale)
			c.Check(err, Equals, ca.err)
			result, err := rounded.ToString(0, 0)
			c.Check(string(result), Equals, ca.output)
		}
	}
	cases := []tcase{
		{"123456789.987654321", 1, "123456790.0", nil},
		{"15.1", 0, "15", nil},
		{"15.5", 0, "16", nil},
		{"15.9", 0, "16", nil},
		{"-15.1", 0, "-15", nil},
		{"-15.5", 0, "-16", nil},
		{"-15.9", 0, "-16", nil},
		{"15.1", 1, "15.1", nil},
		{"-15.1", 1, "-15.1", nil},
		{"15.17", 1, "15.2", nil},
		{"15.4", -1, "20", nil},
		{"-15.4", -1, "-20", nil},
		{"5.4", -1, "10", nil},
		{".999", 0, "1", nil},
		{"999999999", -9, "1000000000", nil},
	}
	doTest(c, cases)
}

func (s *testMyDecimalSuite) TestFromString(c *C) {
	type tcase struct {
		input  string
		output string
		err    error
	}
	cases := []tcase{
		{"12345", "12345", nil},
		{"12345.", "12345", nil},
		{"123.45.", "123.45", nil},
		{"-123.45.", "-123.45", nil},
		{".00012345000098765", "0.00012345000098765", nil},
		{".12345000098765", "0.12345000098765", nil},
		{"-.000000012345000098765", "-0.000000012345000098765", nil},
		{"1234500009876.5", "1234500009876.5", nil},
		{"123E5", "12300000", nil},
		{"123E-2", "1.23", nil},
	}
	for _, ca := range cases {
		var dec MyDecimal
		err := dec.FromString([]byte(ca.input))
		c.Check(err, Equals, ca.err)
		result, err := dec.ToString(0, 0)
		c.Check(err, IsNil)
		c.Check(string(result), Equals, ca.output, Commentf("dec:%s", dec.String()))
	}
	wordBufLen = 1
	cases = []tcase{
		{"123450000098765", "98765", ErrOverflow},
		{"123450.000098765", "123450", ErrTruncated},
	}
	for _, ca := range cases {
		var dec MyDecimal
		err := dec.FromString([]byte(ca.input))
		c.Check(err, Equals, ca.err)
		result, err := dec.ToString(0, 0)
		c.Check(err, IsNil)
		c.Check(string(result), Equals, ca.output, Commentf("dec:%s", dec.String()))
	}
	wordBufLen = maxWordBufLen
}

func (s *testMyDecimalSuite) TestToString(c *C) {
	type tcase struct {
		input  string
		prec   int
		frac   int
		output string
		err    error
	}
	cases := []tcase{
		{"123.123", 0, 0, "123.123", nil},
		/* For fixed precision, we no longer count the '.' here. */
		{"123.123", 6, 3, "123.123", nil},
		{"123.123", 8, 3, "00123.123", nil},
		{"123.123", 8, 4, "0123.1230", nil},
		{"123.123", 8, 5, "123.12300", nil},
		{"123.123", 8, 2, "000123.12", ErrTruncated},
		{"123.123", 8, 6, "23.123000", ErrOverflow},
	}
	for _, ca := range cases {
		var dec MyDecimal
		dec.FromString([]byte(ca.input))
		result, err := dec.ToString(ca.prec, ca.frac)
		c.Check(err, Equals, ca.err)
		c.Check(string(result), Equals, ca.output)
	}
}

func (s *testMyDecimalSuite) TestToBinFromBin(c *C) {
	type tcase struct {
		input     string
		precision int
		frac      int
		output    string
		err       error
	}
	cases := []tcase{
		{"-10.55", 4, 2, "-10.55", nil},
		{"-10.55", 0, 0, "-10.55", nil},
		{"0.0123456789012345678912345", 30, 25, "0.0123456789012345678912345", nil},
		{"12345", 5, 0, "12345", nil},
		{"12345", 10, 3, "12345.000", nil},
		{"123.45", 10, 3, "123.450", nil},
		{"-123.45", 20, 10, "-123.4500000000", nil},
		{".00012345000098765", 15, 14, "0.00012345000098", ErrTruncated},
		{".00012345000098765", 22, 20, "0.00012345000098765000", nil},
		{".12345000098765", 30, 20, "0.12345000098765000000", nil},
		{"-.000000012345000098765", 30, 20, "-0.00000001234500009876", ErrTruncated},
		{"1234500009876.5", 30, 5, "1234500009876.50000", nil},
		{"111111111.11", 10, 2, "11111111.11", ErrOverflow},
		{"000000000.01", 7, 3, "0.010", nil},
		{"000000000.01", 0, 0, "0.01", nil},
		{"000000001.10000", 0, 0, "1.10000", nil},
		{"123.4", 10, 2, "123.40", nil},
		{"1000", 3, 0, "0", ErrOverflow},
	}
	for _, ca := range cases {
		var dec MyDecimal
		err := dec.FromString([]byte(ca.input))
		c.Assert(err, IsNil)
		buf, err := dec.ToBin(ca.precision, ca.frac)
		c.Assert(err, Equals, ca.err, Commentf(ca.input))
		var dec2 MyDecimal
		err = dec2.FromBin(buf)
		c.Assert(err, IsNil)
		str, _ := dec2.ToString(0, 0)
		c.Assert(string(str), Equals, ca.output)
	}
}

func (s *testMyDecimalSuite) TestCompare(c *C) {
	type tcase struct {
		a   string
		b   string
		cmp int
	}
	cases := []tcase{
		{"12", "13", -1},
		{"13", "12", 1},
		{"-10", "10", -1},
		{"10", "-10", 1},
		{"-12", "-13", 1},
		{"0", "12", -1},
		{"-10", "0", -1},
		{"4", "4", 0},
		{"-1.1", "-1.2", 1},
		{"1.2", "1.1", 1},
		{"1.1", "1.2", -1},
	}
	for _, ca := range cases {
		var a, b MyDecimal
		a.FromString([]byte(ca.a))
		b.FromString([]byte(ca.b))
		c.Assert(a.Compare(&b), Equals, ca.cmp)
	}
}

func (s *testMyDecimalSuite) TestMaxDecimal(c *C) {
	type tcase struct {
		prec   int
		frac   int
		result string
	}
	cases := []tcase{
		{1, 1, "0.9"},
		{1, 0, "9"},
		{2, 1, "9.9"},
		{4, 2, "99.99"},
		{6, 3, "999.999"},
		{8, 4, "9999.9999"},
		{10, 5, "99999.99999"},
		{12, 6, "999999.999999"},
		{14, 7, "9999999.9999999"},
		{16, 8, "99999999.99999999"},
		{18, 9, "999999999.999999999"},
		{20, 10, "9999999999.9999999999"},
		{20, 20, "0.99999999999999999999"},
		{20, 0, "99999999999999999999"},
		{40, 20, "99999999999999999999.99999999999999999999"},
	}
	for _, ca := range cases {
		var dec MyDecimal
		maxDecimal(ca.prec, ca.frac, &dec)
		str, _ := dec.ToString(0, 0)
		c.Assert(string(str), Equals, ca.result)
	}
}

func (s *testMyDecimalSuite) TestAdd(c *C) {
	type tcase struct {
		a      string
		b      string
		result string
		err    error
	}
	cases := []tcase{
		{".00012345000098765", "123.45", "123.45012345000098765", nil},
		{".1", ".45", "0.55", nil},
		{"1234500009876.5", ".00012345000098765", "1234500009876.50012345000098765", nil},
		{"9999909999999.5", ".555", "9999910000000.055", nil},
		{"99999999", "1", "100000000", nil},
		{"989999999", "1", "990000000", nil},
		{"999999999", "1", "1000000000", nil},
		{"12345", "123.45", "12468.45", nil},
		{"-12345", "-123.45", "-12468.45", nil},
		{"-12345", "123.45", "-12221.55", nil},
		{"12345", "-123.45", "12221.55", nil},
		{"123.45", "-12345", "-12221.55", nil},
		{"-123.45", "12345", "12221.55", nil},
		{"5", "-6.0", "-1.0", nil},
	}
	for _, ca := range cases {
		var a, b, sum MyDecimal
		a.FromString([]byte(ca.a))
		b.FromString([]byte(ca.b))
		err := DecimalAdd(&a, &b, &sum)
		c.Assert(err, Equals, ca.err)
		result, _ := sum.ToString(0, 0)
		c.Assert(string(result), Equals, ca.result)
	}
}

func (s *testMyDecimalSuite) TestSub(c *C) {
	type tcase struct {
		a      string
		b      string
		result string
		err    error
	}
	cases := []tcase{
		{".00012345000098765", "123.45", "-123.44987654999901235", nil},
		{"1234500009876.5", ".00012345000098765", "1234500009876.49987654999901235", nil},
		{"9999900000000.5", ".555", "9999899999999.945", nil},
		{"1111.5551", "1111.555", "0.0001", nil},
		{".555", ".555", "0", nil},
		{"10000000", "1", "9999999", nil},
		{"1000001000", ".1", "1000000999.9", nil},
		{"1000000000", ".1", "999999999.9", nil},
		{"12345", "123.45", "12221.55", nil},
		{"-12345", "-123.45", "-12221.55", nil},
		{"123.45", "12345", "-12221.55", nil},
		{"-123.45", "-12345", "12221.55", nil},
		{"-12345", "123.45", "-12468.45", nil},
		{"12345", "-123.45", "12468.45", nil},
	}
	for _, ca := range cases {
		var a, b, sum MyDecimal
		a.FromString([]byte(ca.a))
		b.FromString([]byte(ca.b))
		err := DecimalSub(&a, &b, &sum)
		c.Assert(err, Equals, ca.err)
		result, _ := sum.ToString(0, 0)
		c.Assert(string(result), Equals, ca.result)
	}
}

func (s *testMyDecimalSuite) TestMul(c *C) {
	type tcase struct {
		a      string
		b      string
		result string
		err    error
	}
	cases := []tcase{
		{"12", "10", "120", nil},
		{"-123.456", "98765.4321", "-12193185.1853376", nil},
		{"-123456000000", "98765432100000", "-12193185185337600000000000", nil},
		{"123456", "987654321", "121931851853376", nil},
		{"123456", "9876543210", "1219318518533760", nil},
		{"123", "0.01", "1.23", nil},
		{"123", "0", "0", nil},
		{"1" + strings.Repeat("0", 60), "1" + strings.Repeat("0", 60), "0", ErrOverflow},
	}
	for _, ca := range cases {
		var a, b, product MyDecimal
		a.FromString([]byte(ca.a))
		b.FromString([]byte(ca.b))
		err := DecimalMul(&a, &b, &product)
		c.Check(err, Equals, ca.err)
		result, _ := product.ToString(0, 0)
		c.Assert(string(result), Equals, ca.result)
	}
}

func (s *testMyDecimalSuite) TestDivMod(c *C) {
	type tcase struct {
		a      string
		b      string
		result string
		err    error
	}
	cases := []tcase{
		{"120", "10", "12.000000000", nil},
		{"123", "0.01", "12300.000000000", nil},
		{"120", "100000000000.00000", "0.000000001200000000", nil},
		{"123", "0", "", ErrDivByZero},
		{"0", "0", "", ErrDivByZero},
		{"-12193185.1853376", "98765.4321", "-123.456000000000000000", nil},
		{"121931851853376", "987654321", "123456.000000000", nil},
		{"0", "987", "0", nil},
		{"1", "3", "0.333333333", nil},
		{"1.000000000000", "3", "0.333333333333333333", nil},
		{"1", "1", "1.000000000", nil},
		{"0.0123456789012345678912345", "9999999999", "0.000000000001234567890246913578148141", nil},
		{"10.333000000", "12.34500", "0.837019036046982584042122316", nil},
		{"10.000000000060", "2", "5.000000000030000000", nil},
	}
	for _, ca := range cases {
		var a, b, to MyDecimal
		a.FromString([]byte(ca.a))
		b.FromString([]byte(ca.b))
		err := DecimalDiv(&a, &b, &to, 5)
		c.Check(err, Equals, ca.err)
		if ca.err == ErrDivByZero {
			continue
		}
		result, _ := to.ToString(0, 0)
		c.Assert(string(result), Equals, ca.result)
	}

	cases = []tcase{
		{"234", "10", "4", nil},
		{"234.567", "10.555", "2.357", nil},
		{"-234.567", "10.555", "-2.357", nil},
		{"234.567", "-10.555", "2.357", nil},
		{"99999999999999999999999999999999999999", "3", "0", nil},
	}
	for _, ca := range cases {
		var a, b, to MyDecimal
		a.FromString([]byte(ca.a))
		b.FromString([]byte(ca.b))
		ec := DecimalMod(&a, &b, &to)
		c.Check(ec, Equals, ca.err)
		if ca.err == ErrDivByZero {
			continue
		}
		result, _ := to.ToString(0, 0)
		c.Assert(string(result), Equals, ca.result)
	}
}
