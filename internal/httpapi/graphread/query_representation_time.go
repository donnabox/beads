package graphread

// The date/time checks below adapt ONLY date, getTime(true) and getDateTime(true)
// from ajv-formats 3.0.1 dist/formats.js, as selected by BDP's wrapper at
// ba8d18c6355cbd6dd3499befe03641874bb01c73:
// packages/protocol/src/schema-formats.ts. That wrapper additionally requires
// the date/time separator to be T/t. Two RFC 3339 guards correct known AJV
// overacceptance: numeric offsets require HH:MM, and local hours/minutes remain
// bounded even for leap seconds. Preserve valid offset-adjusted leap seconds;
// time.Parse has a different acceptance set. This input validator permits t/z;
// the mapper separately enforces uppercase T/Z on emitted recorded content.
// No accepted timestamp is rewritten.
//
// Exact source SHA-256:
// formats.js b61693a26869970a92cc1fe832c931ccd3afe8b6d66802b19c48306349e84880
// BDP wrapper ff5e19649d2d4b375eeebd0232199cd8c203566d4a7a8c04f9f6346bd27979c3
// ajv-formats LICENSE 9df3bb69929a3b650ed73b3bfa1756725aaff0ac296461605753547004eafeaf
//
// MIT License
// Copyright (c) 2020 Evgeny Poberezkin
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import (
	"regexp"
	"strconv"
)

// AJV splits the whole string on T/t or whitespace; the BDP prefix guard requires
// T/t at position 10. A single anchored expression enforces that distinction;
// numeric offset punctuation additionally follows RFC 3339, correcting AJV's
// acceptance of hour-only and colonless offsets.
var queryDateTimePattern = regexp.MustCompile(`^([0-9]{4})-([0-9]{2})-([0-9]{2})[Tt]([0-9]{2}):([0-9]{2}):([0-9]{2}(?:\.[0-9]+)?)([Zz]|([+-])([0-9]{2}):([0-9]{2}))$`)

func queryRepresentationInputDateTime(value string) bool {
	matches := queryDateTimePattern.FindStringSubmatch(value)
	if matches == nil {
		return false
	}
	number := func(index int) int {
		if matches[index] == "" {
			return 0
		}
		n, _ := strconv.Atoi(matches[index])
		return n
	}
	year, month, day := number(1), number(2), number(3)
	days := [...]int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if month < 1 || month > 12 || day < 1 {
		return false
	}
	maximum := days[month]
	if month == 2 && year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		maximum = 29
	}
	if day > maximum {
		return false
	}
	hour, minute := number(4), number(5)
	if hour > 23 || minute > 59 {
		return false
	}
	seconds, err := strconv.ParseFloat(matches[6], 64)
	if err != nil {
		return false
	}
	offsetHour, offsetMinute := number(9), number(10)
	if offsetHour > 23 || offsetMinute > 59 {
		return false
	}
	if seconds < 60 {
		return true
	}
	sign := 1
	if matches[8] == "-" {
		sign = -1
	}
	utcMinute := minute - offsetMinute*sign
	utcHour := hour - offsetHour*sign
	if utcMinute < 0 {
		utcHour--
	}
	return (utcHour == 23 || utcHour == -1) && (utcMinute == 59 || utcMinute == -1) && seconds < 61
}
