package memory

import (
	"arena" // requires export GOEXPERIMENT=arenas to be set
	"sync"
)

type GoArenaAllocator struct {
	mem *arena.Arena
	// Keep track on all the allocations, when all use then we can call free.
	// map with the allocations which we need, I think this would be awesome.
	addrs map[int]bool
	sync.Mutex
}

func NewGoArenaAllocator() *GoArenaAllocator {
	return &GoArenaAllocator{arena.NewArena(), map[int]bool{}, sync.Mutex{}}
}

func (a *GoArenaAllocator) Allocate(size int) []byte {
	buf := arena.MakeSlice[byte](a.mem, size+alignment, size+alignment) // padding for 64-byte alignment, I dont think this is needed in the arena since we make all 64 bit aligned
	addr := int(addressOf(buf))
	// So data will be loaded based upon division with 64, here we check the address pointer.
	// If the data is even division with 64 we can load it to the cache way more efficient and gain speed ups
	// What we do here is move ths buffer around so the address is has a start that is even with 64 so we can load it faster.
	//
	next := roundUpToMultipleOf64(addr)
	a.Lock()
	defer a.Unlock()
	if addr != next {
		shift := next - addr
		out := buf[shift : size+shift : size+shift]
		addr := int(addressOf(out))
		a.addrs[addr] = true
		return out
	}
	a.addrs[addr] = true
	return buf
}

func (a *GoArenaAllocator) CheckSize() int {
	return len(a.addrs)
}

func (a *GoArenaAllocator) Reallocate(size int, b []byte) []byte {
	if size == len(b) {
		return b
	}
	newBuf := a.Allocate(size)
	copy(newBuf, b)
	return newBuf
}

func (a *GoArenaAllocator) Free(b []byte) {
	addr := int(addressOf(b))
	a.Lock()
	delete(a.addrs, addr)
	a.Unlock()
	if len(a.addrs) > 0 {
		return
	}
	a.mem.Free()
}

var _ Allocator = &GoArenaAllocator{}

// Next step is to dig down in to memory allocations and see how we can use it.
