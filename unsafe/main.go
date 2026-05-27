package main

type MemoryArena[T any] struct {
	buffer []T
	offset int
}

func NewMemoryArena[T any](size int) *MemoryArena[T] {
	return &MemoryArena[T]{buffer: make([]T, size)}
}

func (a *MemoryArena[T]) Alloc(obj T) *T {
	if a.offset >= len(a.buffer) {
		panic("out of memory")
	}
	a.buffer[a.offset] = obj
	p := &a.buffer[a.offset]
	a.offset++
	return p
}

func (a *MemoryArena[T]) Reset() {
	a.offset = 0
	// Optional: zero entries so the GC can collect anything they referenced.
	// Only matters if T contains pointers.
	var zero T
	for i := range a.buffer {
		a.buffer[i] = zero
	}
}

func main() {
	arena := NewMemoryArena[int](10)
	num := arena.Alloc(8)
	println(*num) // 8
	arena.Reset()
	println(*num) // 0 (after reset, the memory is zeroed out)
}
