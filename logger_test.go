package logger

import (
	"testing"
)

func Benchmark_Log(b *testing.B) {
	l := NewLogger("log8080", true, true, false)

	for i := 0; i < b.N; i++ {

		l.Log("%s %s", "hi", "my boy")
	}
}
