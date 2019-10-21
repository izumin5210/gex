package main

import (
	"io"
	"strings"
	"testing"
	"unicode"
)

func NewTestWriter(t *testing.T) io.Writer {
	return &TestWriter{t: t}
}

type TestWriter struct {
	t *testing.T
}

func (w *TestWriter) Write(p []byte) (n int, err error) {
	w.t.Helper()
	s := string(p)
	n = len(s)
	s = strings.TrimRightFunc(string(p), unicode.IsSpace)
	w.t.Log(s)
	return
}
