package main

import "fmt"

const chunkSize = 1024

type arenaChunk[T any] struct {
	data []T
	used int
}

type MemoryArena[T any] struct {
	chunks  []arenaChunk[T]
	current int
}

func NewMemoryArena[T any]() *MemoryArena[T] {
	return &MemoryArena[T]{
		chunks: []arenaChunk[T]{
			{data: make([]T, chunkSize)},
		},
	}
}

func (a *MemoryArena[T]) Alloc(data T) *T {
	chunk := &a.chunks[a.current]
	if chunk.used == len(chunk.data) {
		a.chunks = append(a.chunks, arenaChunk[T]{data: make([]T, chunkSize)})
		a.current++
		chunk = &a.chunks[a.current]
	}
	chunk.data[chunk.used] = data
	ptr := &chunk.data[chunk.used]
	chunk.used++
	return ptr
}

func main() {
	arena := NewMemoryArena[int]()
	ptr := arena.Alloc(42)
	fmt.Println(*ptr) // 42
}
