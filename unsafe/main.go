package main

import (
	"fmt"
	"sync"
)

type MemoryArena[T any] struct {
	buffer []T
	mu     sync.Mutex
	offset uintptr
}

func NewMemoryArena[T any](size int) *MemoryArena[T] {
	return &MemoryArena[T]{
		buffer: make([]T, size),
		offset: 0,
		mu:     sync.Mutex{},
	}
}

func (arena *MemoryArena[T]) EnoughSpace() bool {
	return arena.offset >= uintptr(len(arena.buffer))
}

func (arena *MemoryArena[T]) Alloc(obj T) *T {
	arena.mu.Lock()
	defer arena.mu.Unlock()
	if arena.EnoughSpace() {
		panic("Not enough space")
	}
	arena.buffer[arena.offset] = obj
	allocated := &arena.buffer[arena.offset]

	arena.offset++
	return allocated
}

func (arena *MemoryArena[T]) Reset() {
	arena.mu.Lock()
	defer arena.mu.Unlock()
	arena.offset = 0
	var zero T
	for i := range arena.buffer {
		arena.buffer[i] = zero
	}
}

func main() {
	mem := NewMemoryArena[int](10)
	num := mem.Alloc(15)
	fmt.Println(*num)
	mem.Reset()
	fmt.Println(*num)

}
