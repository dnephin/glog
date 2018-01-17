// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package glog

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
)

func setupBuffer() (*bytes.Buffer, func()) {
	buf := new(bytes.Buffer)
	oldOut := logging.out
	Init(Options{Output: buf})
	return buf, func() { Init(Options{Output: oldOut}) }
}

func hasPrefix(out, prefix string) func() (bool, string) {
	return func() (bool, string) {
		msg := fmt.Sprintf("%q does not have prefix %q", out, prefix)
		return strings.HasPrefix(out, prefix), msg
	}
}

func hasSuffix(out, suffix string) func() (bool, string) {
	return func() (bool, string) {
		msg := fmt.Sprintf("%q does not have suffix %q", out, suffix)
		return strings.HasSuffix(out, suffix), msg
	}
}

// Test that Info works as advertised.
func TestInfo(t *testing.T) {
	buf, teardown := setupBuffer()
	defer teardown()

	Info("test")
	out := buf.String()
	assert.Check(t, hasPrefix(out, "I"))
	assert.Check(t, hasSuffix(out, "] test\n"))
	assert.Check(t, is.Contains(out, "glog_test.go"))
}

func TestInfoDepth(t *testing.T) {
	buf, teardown := setupBuffer()
	defer teardown()

	f := func() { InfoDepth(1, "depth-test1") }

	// The next three lines must stay together
	line := nextLineNum()
	InfoDepth(0, "depth-test0")
	f()

	msgs := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	assert.Assert(t, is.Len(msgs, 2))

	for i, out := range msgs {
		assert.Check(t, hasPrefix(out, "I"))
		assert.Check(t, is.Contains(out, fmt.Sprintf("depth-test%d", i)))
		assert.Check(t, is.Contains(out, fmt.Sprintf("glog_test.go:%d", line+i)))
	}
}

func nextLineNum() int {
	_, _, line, _ := runtime.Caller(1)
	return line + 1
}

// Test that using the standard log package logs to INFO.
func TestStandardLog(t *testing.T) {
	buf, teardown := setupBuffer()
	defer teardown()
	CopyStandardLogTo("INFO")
	defer func() { log.SetOutput(os.Stderr) }()

	log.Print("test")
	out := buf.String()
	assert.Check(t, hasPrefix(out, "I"))
	assert.Check(t, hasSuffix(out, "] test\n"))
	assert.Check(t, is.Contains(out, "glog_test.go"))
}

func patchTimeNow() func() {
	old := timeNow
	timeNow = func() time.Time {
		return time.Date(2006, 1, 2, 15, 4, 5, .067890e9, time.Local)
	}
	return func() { timeNow = old }
}

// Test that the header has the correct format.
func TestHeader(t *testing.T) {
	buf, teardown := setupBuffer()
	defer teardown()
	defer patchTimeNow()()

	pid = 1234
	line := nextLineNum()
	Info("test")

	expected := fmt.Sprintf("I0102 15:04:05.067890    1234 glog_test.go:%d] test\n", line)
	assert.Equal(t, expected, buf.String())
}

func TestError(t *testing.T) {
	buf, teardown := setupBuffer()
	defer teardown()

	Error("test")
	out := buf.String()
	assert.Check(t, hasPrefix(out, "E"))
	assert.Check(t, hasSuffix(out, "] test\n"))
	assert.Check(t, is.Contains(out, "glog_test.go"))
}

func BenchmarkHeader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf, _, _ := logging.header(infoLog, 0)
		logging.putBuffer(buf)
	}
}
