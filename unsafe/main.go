package main

import "sync"

type MemoryArenaInterface[T any] interface {
	Alloc(obj T) *T
	Reset()
}

type MemoryArena[T any] struct {
	offset uintptr
	buffer []T
	mu     sync.Mutex
}

func NewMemoryArena[T any](size int) MemoryArenaInterface[T] {
	return &MemoryArena[T]{
		offset: 0,
		buffer: make([]T, size),
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
		panic("MemoryArena: Out of memory")
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
	arena := NewMemoryArena[int](10)

	allocated := arena.Alloc(10)
	println(*allocated)

	arena.Reset()
	println(*allocated)

}
