// SPDX-License-Identifier: MIT
//
// The indexByteTwo / lastIndexByteTwo functions in this file are adapted from
// fzf (https://github.com/junegunn/fzf).
//
// Copyright (c) 2013-2026 Junegunn Choi
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

//go:build !arm64 && !amd64

package str

import "bytes"

// indexByteTwo returns the index of the first occurrence of b1 or b2 in s,
// or -1 if neither is present.
func indexByteTwo(s []byte, b1, b2 byte) int {
	i1 := bytes.IndexByte(s, b1)
	if i1 == 0 {
		return 0
	}
	scope := s
	if i1 > 0 {
		scope = s[:i1]
	}
	if i2 := bytes.IndexByte(scope, b2); i2 >= 0 {
		return i2
	}
	return i1
}

// lastIndexByteTwo returns the index of the last occurrence of b1 or b2 in s,
// or -1 if neither is present.
func lastIndexByteTwo(s []byte, b1, b2 byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == b1 || s[i] == b2 {
			return i
		}
	}
	return -1
}
