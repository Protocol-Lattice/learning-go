
# 🚀 Go Interview Preparation Guide

A comprehensive guide for Go developers preparing for interviews, covering core concepts, language fundamentals, and system design.

---

## 📋 Table of Contents
1. [Core Concurrency](#1-core-concurrency)
2. [Memory Management (GC)](#2-memory-management-gc)
3. [Language Fundamentals](#3-language-fundamentals)
4. [Junior Q&A Cheat Sheet](#4-junior-qa-cheat-sheet)
5. [Networking (HTTP & gRPC)](#5-networking-http--grpc)
6. [System Design & Infrastructure](#6-system-design--infrastructure)
7. [Code Examples](#7-code-examples)

---

## 1. Core Concurrency

### Concurrency vs. Parallelism
- **Concurrency**: A technique for managing multiple tasks that are executed in an interleaved manner (often on a single processor).
- **Parallelism**: Executing multiple tasks truly at the same time (on multiple cores or processors).

### Goroutines & Channels
- **Goroutines**:
  - Lightweight "threads" managed by the Go runtime.
  - Initial stack size is only **2 KB** (OS threads are several MB).
  - Stacks are dynamic; they can grow and shrink as needed.
- **Channels**: The primary way goroutines communicate.
  - **Unbuffered**: Enforces a synchronous hand-off. Both sender and receiver block until the other is ready.
  - **Buffered**: Queues up to `cap(ch)` values. Sends block only when full; receives block only when empty.

### The M:N Scheduler
Go uses an **M:N scheduler** to map **M goroutines** onto **N OS threads**.
- Each thread has a processor (**P**) with a **local run queue**.
- **Work Stealing**: If a processor's local queue is empty, the scheduler steals work from other processors' queues to balance the load.
- **Global Queue**: If local queues are full, goroutines are moved to a global queue.

### Synchronization
- **Waitgroup**: Used to wait for a collection of goroutines to finish.
- **Mutex**: Prevents race conditions by ensuring only one goroutine can access a "critical section" of code at a time.
- **RWMutex**: A read-write lock. Multiple goroutines can hold a read lock (`RLock`), but the write lock (`Lock`) is exclusive.

---

## 2. Memory Management (GC)

Go uses a **Tri-color Mark and Sweep** garbage collector:

1. **White**: Objects that are candidates for deletion.
2. **Gray**: Objects reachable from roots but whose references haven't been scanned yet.
3. **Black**: Reachable objects that have been fully scanned.

**The Process:**
- Starts with all objects marked as **white**.
- **Roots** (global variables, stack pointers) are marked **gray**.
- The collector scans gray objects, marking their children as gray and moving the parent to **black**.
- Once no gray objects remain, all remaining **white** objects are considered unreachable and are **freed**.

---

## 3. Language Fundamentals

### Types & Data Structures
- **Interfaces**: Define a set of method signatures. A type implements an interface by implementing all its methods.
  - Interfaces can **embed** other interfaces (the result is the union of their method sets).
- **Slices**: A dynamic view over an underlying array. Contains `len`, `cap`, and a pointer to the data.
- **Maps**: Key-value dictionaries. **Not concurrency-safe** (use `sync.Mutex` or `sync.Map` for concurrent access).
- **Struct Embedding**: Go's way of composition (not classical inheritance). Fields and methods of embedded structs are "promoted" to the outer struct.

### Essential Concepts
- **Pointers**: Refer to a specific location in memory.
- **Pass by Value**: Everything in Go is passed by value.
  - When passing a "handle" (pointer, slice, map, channel), you copy the handle, but it still points to the same underlying memory.
- **Variable Declaration**:
  - `:=` (Short declaration): Use inside functions for concise, type-inferred local variables.
  - `var`: Use for zero-value initialization, explicit types, or package-level variables.
- **Generics (Go 1.18+)**: Allows writing functions and structs that operate on different types without code duplication.
- **Defer**: Delays execution until the surrounding function returns. Operates in **LIFO** (Last In, First Out) order.
- **Panic & Recover**:
  - `panic`: Stops normal execution.
  - `recover`: Used inside a `defer` to catch a panic and regain control.
- **Context**: Used to propagate timeouts, deadlines, cancellation signals, and metadata across goroutines.

---

## 4. Junior Q&A Cheat Sheet

| Question | Answer |
| :--- | :--- |
| **When to use a pointer receiver?** | When the method needs to mutate the receiver, to avoid copying large structs, or for consistency. |
| **What happens if you write to a nil map?** | It **panics**. You must initialize it with `make(map[K]V)` first. |
| **How to avoid goroutine leaks?** | Use `context` for timeouts or signals, and ensure every channel send has a matching receive. |
| **How to handle JSON?** | Use `json.Marshal` and `json.Unmarshal`. Use struct tags like ``json:"field_name,omitempty"``. |
| **How to set a timeout on HTTP?** | Create a `context.WithTimeout`, defer its `cancel()`, and pass it to the HTTP request. |
| **Common loop/goroutine gotcha?** | Capturing the loop variable in a closure. Fix by shadowing: `v := v` inside the loop. |
| **How to detect races?** | Run tests with the flag: `go test -race`. |

---

## 5. Networking (HTTP & gRPC)

### HTTP Methods
- **GET / HEAD**: Read-only. `GET` returns body; `HEAD` returns headers only.
- **POST**: Create a resource or trigger an action (non-idempotent).
- **PUT**: Replace/Upsert a resource at a specific URI.
- **PATCH**: Partially modify a resource.
- **DELETE**: Remove a resource (idempotent).
- **OPTIONS**: Used for CORS preflight and discovering supported methods.

### gRPC Overview
- **Protocol Buffers**: IDL used to define services and messages.
- **HTTP/2**: The underlying transport, supporting multiplexing and streaming.
- **Interceptors**: Middleware for gRPC (Auth, Logging, Metrics, etc.).

#### gRPC Streaming Types
1. **Unary**: 1 Request ➡️ 1 Response.
2. **Server-Streaming**: 1 Request ➡️ Stream of Responses.
3. **Client-Streaming**: Stream of Requests ➡️ 1 Response.
4. **Bidirectional**: Stream of Requests ↔️ Stream of Responses (Full-duplex).

---

## 6. System Design & Infrastructure

### Scalability
- **Horizontal Pod Autoscaling (HPA)**: Adding/removing pods based on metrics.
- **Vertical Pod Autoscaling (VPA)**: Resizing pod CPU/Memory limits.
- **Load Balancer**: Distributes traffic across multiple application instances.
- **Ingress**: A K8s API object that manages external access to services, typically via HTTP/HTTPS routing rules.

### Rate Limiting
- Acts as a "bouncer" to protect downstream services.
- Policies typically define "R requests per second with a burst of B."

---

## 7. Code Examples

### Generic Thread-Safe Cache with TTL
A basic in-memory key-value cache implementation using generics and a background cleanup janitor.

<details>
<summary><b>Click to view full implementation</b></summary>

```go
package cache

import (
	"sync"
	"time"
)

// Option configures the Cache at construction time.
type Option func(*cacheConfig)

type cacheConfig struct {
	defaultTTL     time.Duration // 0 => no default expiry
	cleanupEvery   time.Duration // 0 => disable background janitor
}

func WithDefaultTTL(ttl time.Duration) Option {
	return func(c *cacheConfig) { c.defaultTTL = ttl }
}

func WithCleanupInterval(every time.Duration) Option {
	return func(c *cacheConfig) { c.cleanupEvery = every }
}

type entry[V any] struct {
	value    V
	expireAt time.Time // zero => no expiry
}

// Cache is a basic in-memory key-value cache with optional TTL.
type Cache[K comparable, V any] struct {
	mu           sync.RWMutex
	items        map[K]entry[V]
	defaultTTL   time.Duration
	cleanupEvery time.Duration
	stopCh       chan struct{}
	doneCh       chan struct{}
}

// New creates a new Cache instance.
func New[K comparable, V any](opts ...Option) *Cache[K, V] {
	cfg := cacheConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	c := &Cache[K, V]{
		items:        make(map[K]entry[V]),
		defaultTTL:   cfg.defaultTTL,
		cleanupEvery: cfg.cleanupEvery,
	}
	if c.cleanupEvery > 0 {
		c.startJanitor()
	}
	return c
}

// Close stops the background janitor (if enabled).
func (c *Cache[K, V]) Close() {
	if c.stopCh == nil {
		return
	}
	close(c.stopCh)
	<-c.doneCh
}

// Set adds or updates a key-value pair using the cache's default TTL.
func (c *Cache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL adds or updates a key-value pair with a specific TTL.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	c.mu.Lock()
	c.items[key] = entry[V]{value: value, expireAt: exp}
	c.mu.Unlock()
}

// Get retrieves the value by key.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	e, ok := c.items[key]
	if !ok {
		c.mu.RUnlock()
		var zero V
		return zero, false
	}
	expired := !e.expireAt.IsZero() && time.Now().After(e.expireAt)
	value := e.value
	c.mu.RUnlock()

	if !expired {
		return value, true
	}

	// Double-check expiration under write lock
	c.mu.Lock()
	if e2, ok2 := c.items[key]; ok2 && !e2.expireAt.IsZero() && time.Now().After(e2.expireAt) {
		delete(c.items, key)
	}
	c.mu.Unlock()

	var zero V
	return zero, false
}

// Pop removes and returns the value.
func (c *Cache[K, V]) Pop(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}

	if !e.expireAt.IsZero() && time.Now().After(e.expireAt) {
		delete(c.items, key)
		var zero V
		return zero, false
	}

	delete(c.items, key)
	return e.value, true
}

func (c *Cache[K, V]) startJanitor() {
	c.stopCh = make(chan struct{})
	c.doneCh = make(chan struct{})

	go func() {
		defer close(c.doneCh)
		t := time.NewTicker(c.cleanupEvery)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				c.removeExpired()
			case <-c.stopCh:
				return
			}
		}
	}()
}

func (c *Cache[K, V]) removeExpired() {
	now := time.Now()
	c.mu.Lock()
	for k, e := range c.items {
		if !e.expireAt.IsZero() && now.After(e.expireAt) {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}
```
</details>

---

### External Resources
- [AWS Free Tier](https://aws.amazon.com/free/)
- [Go Documentation](https://golang.org/doc/)

---
*Maintained with ❤️ for the Go community.*
