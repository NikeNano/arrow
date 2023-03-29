package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requires export GOEXPERIMENT=arenas to be set

func TestMemory(t *testing.T) {
	a := NewGoArenaAllocator()
	require.Equal(t, 0, a.CheckSize())
	s1 := a.Allocate(10)
	require.Equal(t, 1, a.CheckSize())
	s2 := a.Allocate(11)
	require.Equal(t, 2, a.CheckSize())
	a.Free(s1)
	require.Equal(t, 1, a.CheckSize())
	a.Free(s2)
	require.Equal(t, 0, a.CheckSize())
}

func TestNewGoArenaAllocator_Allocate(t *testing.T) {
	tests := []struct {
		name string
		sz   int
	}{
		{"lt alignment", 33},
		{"gt alignment unaligned", 65},
		{"eq alignment", 64},
		{"large unaligned", 4097},
		{"large aligned", 8192},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alloc := NewGoArenaAllocator()
			buf := alloc.Allocate(test.sz)
			assert.NotNil(t, buf)
			assert.Len(t, buf, test.sz)
			defer alloc.Free(buf)
		})
	}
}

func TestGoArenaAllocator_Reallocate(t *testing.T) {
	tests := []struct {
		name     string
		sz1, sz2 int
	}{
		{"smaller", 200, 100},
		{"same", 200, 200},
		{"larger", 200, 300},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alloc := NewGoArenaAllocator()
			buf := alloc.Allocate(test.sz1)
			for i := range buf {
				buf[i] = byte(i & 0xFF)
			}

			exp := make([]byte, test.sz2)
			copy(exp, buf)

			newBuf := alloc.Reallocate(test.sz2, buf)
			assert.Equal(t, exp, newBuf)
			defer alloc.Free(newBuf)
		})
	}
}
