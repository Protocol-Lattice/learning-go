# 🚀 The Ultimate Go (Golang) Interview Preparation Guide

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Awesome](https://img.shields.io/badge/Awesome-Go-green?style=for-the-badge&logo=awesome)](https://github.com/avelino/awesome-go)
[![Interview Ready](https://img.shields.io/badge/Interview-Ready-orange?style=for-the-badge)](https://github.com/Protocol-Lattice/golang-interview-prep-in-english)

Welcome to the **Ultimate Go Interview Preparation Guide**. This document has been meticulously designed and engineered to serve as a high-density, easily skimmable, and visually rich reference during live interviews, technical screens, or rapid pre-interview revision. 

It covers everything from low-level runtime internals (the scheduler, channels, GC) to Go-specific language behaviors, concurrency patterns, and production-ready system architectures.

---

> [!TIP]
> **How to use this guide during an interview:**
> - Keep this guide open in a browser or markdown viewer side-by-side.
> - Use the [📋 Table of Contents](#-table-of-contents) below for instant navigation.
> - Refer to **[Section 6: The Ultimate Cheat Sheets](#-six-the-ultimate-cheat-sheets)** for rapid, bulletproof answers to standard trick questions.

---

## 📋 Table of Contents
1. [⚡ Core Concurrency & Runtime (The Engine)](#-one-core-concurrency--runtime-the-engine)
2. [🧹 Memory Management & GC Internals](#-two-memory-management--gc-internals)
3. [🧩 Go Language Deep-Dives](#-three-go-language-deep-dives)
4. [🌐 Networking & APIs (HTTP vs gRPC)](#-four-networking--apis-http-vs-grpc)
5. [🏗️ System Design & Infrastructure](#-five-system-design--infrastructure)
6. [🔥 The Ultimate Cheat Sheets](#-six-the-ultimate-cheat-sheets)
7. [💻 Production-Ready Code Exercises](#-seven-production-ready-code-exercises)
8. [📚 High-Quality Curated Resources](#-eight-high-quality-curated-resources)

---

## ⚡ One: Core Concurrency & Runtime (The Engine)

### Concurrency vs. Parallelism
- **Concurrency** is about **structure**. It is the composition of independently executing processes (managing multiple tasks at once).
- **Parallelism** is about **execution**. It is the simultaneous execution of multiple things at once (requires multiple CPU cores).
- *Analogy:* Concurrency is a single queue at a grocery store being handled by a single cashier who switches between processing items and bagging. Parallelism is opening a second register with another cashier.

---

### The Go M:N Scheduler
Go uses a highly efficient **M:N scheduler** in user-space to map $M$ goroutines onto $N$ OS threads. 

```mermaid
graph TD
    subgraph Go Runtime Scheduler
        G1[Goroutine G1] -->|scheduled on| P1[Processor P1]
        G2[Goroutine G2] -->|waiting in LRQ| P1
        P1 -->|bound to| M1[OS Thread M1]
        
        G3[Goroutine G3] -->|scheduled on| P2[Processor P2]
        P2 -->|bound to| M2[OS Thread M2]
        
        GQ[Global Run Queue] -->|steal fallback| P1
        GQ -->|steal fallback| P2
    end
    
    style G1 fill:#00ADD8,stroke:#333,stroke-width:2px,color:#fff
    style G2 fill:#e0f7fa,stroke:#333,stroke-width:1px,color:#000
    style G3 fill:#00ADD8,stroke:#333,stroke-width:2px,color:#fff
    style P1 fill:#4caf50,stroke:#333,stroke-width:2px,color:#fff
    style P2 fill:#4caf50,stroke:#333,stroke-width:2px,color:#fff
    style M1 fill:#ff9800,stroke:#333,stroke-width:2px,color:#fff
    style M2 fill:#ff9800,stroke:#333,stroke-width:2px,color:#fff
    style GQ fill:#9c27b0,stroke:#333,stroke-width:2px,color:#fff
```

#### The Core Entities (G, M, P)
- **G (Goroutine):** Lightweight green thread. Holds execution stack (starts at **2 KB**, grows dynamically up to 1 GB), program counter, and scheduling metadata.
- **M (OS Thread / Machine):** A real OS thread managed by the host OS kernel scheduler.
- **P (Processor / Logical Context):** Represents resources required to execute Go code. The number of Ps is governed by `$GOMAXPROCS` (defaults to CPU cores). Each P maintains a **Local Run Queue (LRQ)** of up to 256 runnable Goroutines.

#### Scheduling Algorithms & Core Mechanics
1. **Work Stealing:** When a P finishes its LRQ:
   - It checks the **Global Run Queue (GRQ)** (polled 1 out of every 61 scheduler ticks to prevent starvation).
   - It attempts to steal **half** of another P's local run queue.
   - If still empty, it checks network pollers.
2. **Syscall Handoff (Network Poller vs. Syscall):**
   - **Blocking Syscall:** When G blocks on a file I/O syscall, the scheduler detaches the running OS thread (M) from its P. The P continues executing other Gs by acquiring or creating a new OS thread (M).
   - **Network I/O:** Handled out-of-band by the Network Poller (using OS abstractions like `epoll` or `kqueue`). The G registers its interest, detaches from P, and goes to sleep, freeing the P and M to run other Gs immediately.
3. **Preemption:**
   - **Cooperative (Pre-Go 1.14):** Goroutines could only be preempted at function call boundaries where compiler-injected stack checks (`morestack`) occurred. A tight loop `for {}` without function calls could freeze a thread.
   - **Asynchronous (Go 1.14+):** Uses OS signals (`SIGURG` on Unix systems) to interrupt and preempt running goroutines every 10ms, preventing long-running goroutines from hogging CPUs.

---

### Goroutines vs. OS Threads

| Feature | Goroutine (Go Green Thread) | OS Thread (Kernel Thread) |
| :--- | :--- | :--- |
| **Memory Footprint** | Dynamic stack starts at **2 KB** (grows/shrinks up to 1GB). | Fixed stack size set by OS (typically **1 MB to 8 MB**). |
| **Creation Cost** | Extremely low (~nanoseconds). Allocated purely in user-space heap. | High (~microseconds). Demands a system call to the OS kernel. |
| **Context Switch Cost**| Very fast (~10ns - 100ns). Saves ~14 registers. | Slow (~1µs - 2µs). Causes CPU cache line misses, TLB flushes. |
| **Scheduling Model** | User-space M:N scheduler (cooperative & async signal). | Kernel-space 1:1 scheduler (preemptive). |

---

### Channels & the `hchan` Struct
Under the hood, a Go channel is not a magical pipe; it is a pointer to an `hchan` struct (defined in `runtime/chan.go`):

```go
type hchan struct {
    qcount   uint           // Total data in the circular queue
    dataqsiz uint           // Capacity of the circular queue (buffer size)
    buf      unsafe.Pointer // Pointer to an underlying circular array
    elemsize uint16
    closed   uint32         // 1 if closed, 0 otherwise
    elemtype *_type         // Element type
    sendx    uint           // Send index in the circular buffer
    recvx    uint           // Receive index in the circular buffer
    recvq    waitq          // Linked list of blocked receivers (waitq of sudog)
    sendq    waitq          // Linked list of blocked senders (waitq of sudog)
    lock     mutex          // Protects all fields in hchan
}
```

#### Channel State & Operations Matrix
This table is critical for live coding and rapid technical questions:

| Channel State | Send (`ch <- x`) | Receive (`<-ch`) | Close (`close(ch)`) |
| :--- | :--- | :--- | :--- |
| **Nil** (`var ch chan T`) | **Blocks forever** | **Blocks forever** | **Panics** (`panic: close of nil channel`) |
| **Open & Empty** | Succeeds (blocks if unbuffered) | **Blocks** | Succeeds (receivers get zero-value, `ok == false`) |
| **Open & Full** | **Blocks** | Succeeds | Succeeds (receivers get zero-value, `ok == false`) |
| **Closed** | **Panics** (`panic: send on closed channel`) | Succeeds immediately (returns zero-value, `ok == false`) | **Panics** (`panic: close of closed channel`) |

---

### Synchronization Primitives
- **`sync.Mutex` Starvation Mode:**
  - **Normal Mode:** Waiters are kept in a LIFO/FIFO queue, but newly active CPU goroutines also compete for the lock. New goroutines usually win because they are already on the CPU.
  - **Starvation Mode:** If a waiter fails to acquire the Mutex for > **1ms**, the Mutex enters starvation mode. The lock is transferred **directly** to the front waiter. New arrivals do not spin or attempt to steal the lock; they immediately enqueue. This mitigates tail-latency spikes.
- **`sync.RWMutex`:** A reader-writer lock. Multiple readers can hold the read lock (`RLock`), but the write lock (`Lock`) is completely exclusive. To prevent writer starvation, new readers are blocked if a writer is already waiting.
- **`sync.WaitGroup`:** Atomic uint64 counter. Two 32-bit halves (on 32-bit platforms) or a uint64 representing the wait count and the waiter count. Manipulated atomically.
- **`sync.Map`:** Specialized concurrent map optimized for two cases:
  1. Read-heavy workloads where keys don't change frequently.
  2. Disjoint concurrent writes (different keys written by different goroutines).
  - *How it works:* Uses two maps: a lockless `read` map (updated atomically) and a locked `dirty` map. Lookups check `read` first. If it misses repeatedly (above a threshold), the `dirty` map is promoted to the `read` map under a lock.

---

## 🧹 Two: Memory Management & GC Internals

### The Tri-Color Mark & Sweep GC
Go uses a concurrent, tri-color mark-and-sweep garbage collector designed for low latency.

```
       [ Roots ] (Globals, Stack Pointers)
          │
          ▼
   ┌─────────────┐
   │ GRAY STAGE  │ ──► Reachable but children unscanned
   └─────────────┘
          │
          ├─────────────────────────┐
          ▼                         ▼
   ┌─────────────┐           ┌─────────────┐
   │ BLACK STAGE │           │ WHITE STAGE │
   │  Reachable  │           │ Unreachable │
   │ (Preserved) │           │  (Swept)    │
   └─────────────┘           └─────────────┘
```

1. **White Set (Unvisited):** Candidates for deletion. At the start of a GC cycle, all objects are White.
2. **Gray Set (Visited, Unexplored):** Reachable from roots, but their children have not been scanned yet.
3. **Black Set (Visited & Explored):** Reachable, and all child pointers have been fully scanned. Black objects contain no pointers directly to White objects.

#### The GC Process:
- **Phase 1: Sweep Termination (STW - Stop The World):** Prepares for marking, activates write barriers.
- **Phase 2: Concurrent Mark:** Scans root pointers (stacks, globals) and pushes them onto the gray queue. Goroutines traverse gray objects, mark children gray, and mark parents black.
- **Phase 3: Mark Termination (STW):** Flushes local cache buffers, completes the marking phase.
- **Phase 4: Concurrent Sweep:** Reclaims memory occupied by remaining White objects and returns it to the allocator.

#### Why do we need the Write Barrier?
Since GC runs concurrently with application threads (mutators), a mutator could hide a white object by assigning it to a black object and breaking the pointer chain from gray objects. The **Write Barrier** intercepts write operations at runtime: if a pointer to a white object is written, it is forced into the gray set, preserving the GC invariants.

---

### Stack vs. Heap & Escape Analysis
- **Stack:** Very fast, local allocation. Follows LIFO structure, managed at thread-level, does not require garbage collection.
- **Heap:** Shared memory pool. Slower allocation, managed via a TCMalloc-inspired allocator (`mcache` -> `mcentral` -> `mheap`), reclaimed by the Garbage Collector.

#### Escape Analysis Rules
The Go compiler uses Escape Analysis at compile-time to decide if a variable can reside on the stack or must escape to the heap.

> [!IMPORTANT]
> **Common Causes of Heap Escape:**
> 1. **Returning a Pointer:** A pointer to a local variable is returned from a function. The variable's lifetime exceeds the function's stack frame.
> 2. **Interface Dynamic Dispatch:** Storing a concrete type in an interface (e.g., calling `fmt.Println(x)` or `json.Marshal(x)`).
> 3. **Dynamic / Large Allocations:** Creating a slice with a size only known at runtime (e.g., `make([]byte, size)`) or exceeding the compiler's size threshold (~64KB).
> 4. **Channel Transmission:** Sending a pointer to a channel (the compiler cannot guarantee which goroutine receives it or when).

#### How to Analyze Escape Decisions:
```bash
go build -gcflags="-m -l" main.go
```
*Note: The `-l` flag disables function inlining, making escape analysis decisions easier to trace.*

---

## 🧩 Three: Go Language Deep-Dives

### Slices Under the Hood
A slice is not an array; it is a **descriptor header** containing metadata that references a backing array. It is defined in `reflect.SliceHeader`:

```go
type SliceHeader struct {
    Data uintptr // Pointer to the underlying array element
    Len  int     // Length: number of elements in the slice
    Cap  int     // Capacity: maximum elements the backing array can hold from this start pointer
}
```

- **Pass-By-Value:** Go passes everything by value. Passing a slice into a function copies the 24-byte `SliceHeader` (pointer, length, capacity). Modifying elements inside the function alters the shared backing array, but appending to the slice within the function may reallocate a new backing array, leaving the caller's slice header unchanged.

#### Slice Capacity Growth (Go 1.18+):
1. If the new capacity is greater than double the old capacity, the new capacity is set to the requested capacity.
2. Otherwise, if the old capacity is less than 256, it doubles.
3. For larger capacities, it transitions to a scale factor:
   $$\text{newcap} = \text{oldcap} + \frac{\text{oldcap} + 3 \times 256}{4}$$
   *This ensures a smooth transition from $2\times$ growth to $1.25\times$ growth.*

---

### Maps Under the Hood
A Go map is a pointer to an `hmap` struct. Maps are implemented as a collection of buckets:

- **Structure:** Each bucket (`bmap` struct) holds up to 8 key-value pairs.
- **Accessing a Key:** Go hashes the key. The low-order bits determine which bucket contains the key, and the high-order bits (`tophash`) identify the specific key within the bucket.
- **Why maps are NOT thread-safe:** Maps are optimized for speed. Reading/writing maps concurrently sets a `flags` bit in `hmap`. If the runtime detects a write while another operation is in progress, it triggers a non-recoverable runtime crash: `fatal error: concurrent map writes`.
- **How to make them safe:** Wrap the map in a struct with a `sync.RWMutex`, or use `sync.Map`.

---

### The Interface Nil Trap
This is one of the most common senior-level Go trick questions.

```go
var p *int = nil
var i any = p

fmt.Println(i == nil) // Outputs: FALSE!
```

#### Why?
An interface variable contains two fields internally:
1. `type` (dynamic type information, e.g., `*int`).
2. `data` (pointer to the concrete value, e.g., `nil`).

An interface value is only considered `nil` if **both** the `type` and `data` fields are `nil`. In the example above, `i` has a valid type pointer (`*int`), so `i == nil` evaluates to `false`.

---

### Defer, Panic, and Recover
- **`defer` LIFO Ordering:** Deferred calls are executed in a Last-In, First-Out (LIFO) stack order.
- **Argument Evaluation:** Arguments to a deferred function are evaluated **immediately** when the `defer` statement is reached, not when the surrounding function exits.
- **`recover()` Placement:** `recover` only works when called **directly** inside a deferred function. Calling it inside a nested helper function will not catch the panic.

---

## 🌐 Four: Networking & APIs (HTTP vs gRPC)

### HTTP Methods & Idempotency

| Method | Safe | Idempotent | Request Body | Response Body |
| :--- | :---: | :---: | :---: | :---: |
| **GET** | ✅ Yes | ✅ Yes | ❌ No | ✅ Yes |
| **POST** | ❌ No | ❌ No | ✅ Yes | ✅ Yes |
| **PUT** | ❌ No | ✅ Yes | ✅ Yes | ✅ Yes/No |
| **PATCH**| ❌ No | ❌ No | ✅ Yes | ✅ Yes |
| **DELETE**| ❌ No | ✅ Yes | ❌ No | ✅ Yes/No |

- **Safe:** Does not modify resource state on the server.
- **Idempotent:** Making multiple identical requests yields the same server state as a single request.

---

### HTTP/1.1 vs. HTTP/2

| Feature | HTTP/1.1 | HTTP/2 |
| :--- | :--- | :--- |
| **Transport Format** | Plain Text | Binary Framing (Stream/Frame structure) |
| **Multiplexing** | ❌ No (Requires multiple TCP connections) | ✅ Yes (Concurrent streams over a single TCP connection) |
| **Header Compression**| ❌ No (Text headers sent repeatedly) | ✅ Yes (HPACK compression) |
| **Server Push** | ❌ No | ✅ Yes (Server can preemptively push resources) |

---

### gRPC Architecture & Streaming
gRPC uses **Protocol Buffers** for high-performance, compact binary serialization, running on top of **HTTP/2** for multiplexed transport.

#### The 4 Streaming Patterns:
1. **Unary:** Standard request-response pattern (1 Request ➡️ 1 Response).
2. **Server Streaming:** 1 Request ➡️ Stream of Responses (e.g., real-time notifications).
3. **Client Streaming:** Stream of Requests ➡️ 1 Response (e.g., heavy file upload).
4. **Bidirectional Streaming:** Stream of Requests ↔️ Stream of Responses (full duplex).

---

## 🏗️ Five: System Design & Infrastructure

### Autoscaling: HPA vs. VPA
- **Horizontal Pod Autoscaler (HPA):** Scales the **number of pods** in response to metrics like CPU usage, Memory limits, or custom Prometheus metrics (e.g., HTTP request rate).
- **Vertical Pod Autoscaler (VPA):** Adjusts the **CPU and Memory limits** of existing pods. Useful for stateful services.
- *Production Warning:* Do not run HPA and VPA concurrently on the exact same resource metrics (like CPU/Memory) to avoid race conditions.

---

### Rate Limiting Algorithms
1. **Token Bucket:** A bucket is filled with tokens at a constant rate up to a max capacity. Every request consumes a token. Allows handling bursts up to the bucket capacity.
2. **Leaky Bucket:** Requests enter a queue and are processed at a constant, fixed rate. Smoothes out traffic spikes but adds latency to bursts.
3. **Sliding Window Counter:** Divides time into windows and keeps track of requests. Prevents sudden traffic bursts at window boundaries.

---

### Microservice Resiliency
- **Circuit Breaker:** Prevents cascading failures.
  - **Closed:** Normal traffic flows.
  - **Open:** Fails fast immediately without calling the struggling downstream service.
  - **Half-Open:** Periodically sends a small fraction of traffic to test if the downstream service has recovered.
- **Exponential Backoff and Jitter:** When retrying failed requests, double the wait time with each retry (exponential) and add a random variance (jitter) to prevent the "thundering herd" problem from overwhelming downstream databases.

---

## 🔥 Six: The Ultimate Cheat Sheets

### Junior & Mid-Level Cheat Sheet

| Question | Core Technical Answer |
| :--- | :--- |
| **When should you use a pointer receiver?** | 1. When the method needs to modify the receiver's state.<br>2. To avoid copying large struct values on every call.<br>3. For consistency across the entire method set. |
| **What happens if you write to a nil map?** | It **panics** (`panic: assignment to entry in nil map`). You must initialize it using `make(map[K]V)` first. |
| **How do you avoid goroutine leaks?** | Always ensure every goroutine has a guaranteed exit path (e.g., using `context.Context` cancellation, close signals on control channels, or setting active timeouts). |
| **Why does Go use `make` vs. `new`?** | - `new(T)` allocates zero-initialized memory and returns a pointer `*T`.<br>- `make(T, args)` initializes and returns completed complex internal headers for **slices, maps, and channels** only. |
| **How does Go 1.22 fix the loop variable bug?** | Prior to 1.22, loop variables shared the same memory address across iterations, requiring manual shadowing (`v := v`). Since 1.22, loop variables are freshly allocated per iteration. |
| **How do you safely detect race conditions?** | Run your test suite or binary with the race detector enabled using the flags: `go test -race` or `go run -race`. |

---

### Senior & Staff-Level Cheat Sheet

| Question | Core Technical Answer |
| :--- | :--- |
| **How do you tune Go's GC?** | 1. **`GOGC` (default 100):** Governs target heap growth ratio. A value of 100 means GC runs when live heap size doubles.<br>2. **`GOMEMLIMIT` (Go 1.19+):** Defines a hard memory limit. Prevents OOM kills in containerized environments by triggering aggressive GC sweeps as memory usage approaches the limit. |
| **Explain Mutex Starvation mode.** | When a Mutex waiter fails to acquire the lock for > 1ms, the Mutex enters starvation mode. The lock is transferred directly to the first waiter. New CPU-active goroutines do not spin or try to steal the lock, mitigating tail-latency spikes. |
| **How do you profile a Go service in production?** | Integrate the `net/http/pprof` standard library endpoint. It allows low-overhead collection of CPU, heap, threadcreate, and blocking profiles. You can analyze flamegraphs using `go tool pprof`. |
| **What is Profile-Guided Optimization (PGO)?** | Introduced in Go 1.20, PGO allows the compiler to optimize code generation (e.g., devirtualizing interface calls, aggressive function inlining) using real performance profiles collected from production, boosting CPU efficiency by 2% to 14%. |
| **What are contiguous stacks in Go?** | Go uses contiguous stacks. If a goroutine requires more stack space than its current frame provides, Go allocates a new, double-sized contiguous memory block, copies the old stack, updates all pointers, and frees the old block. |
| **Why are maps not concurrent safe?** | To maximize execution speed. Concurrent map writes set a flag; if Go detects simultaneous read/write or write/write operations, it calls `throw()` to trigger an immediate, non-recoverable runtime crash. |

---

## 💻 Seven: Production-Ready Code Exercises

### Exercise 1: Generic Thread-Safe Cache with TTL
This is a standard interview question requiring generics, concurrent read/write locks, and active background janitor cleanup.

<details>
<summary><b>Click to view Cache Implementation</b></summary>

```go
package cache

import (
	"sync"
	"time"
)

// Option configures the Cache at construction time.
type Option func(*cacheConfig)

type cacheConfig struct {
	defaultTTL   time.Duration // 0 means no default expiration
	cleanupEvery time.Duration // 0 means background cleanup is disabled
}

// WithDefaultTTL configures the default lifetime of items in the cache.
func WithDefaultTTL(ttl time.Duration) Option {
	return func(c *cacheConfig) { c.defaultTTL = ttl }
}

// WithCleanupInterval configures how often the background cleanup janitor runs.
func WithCleanupInterval(every time.Duration) Option {
	return func(c *cacheConfig) { c.cleanupEvery = every }
}

type entry[V any] struct {
	value    V
	expireAt time.Time // Zero value means no expiration
}

// Cache is a high-performance, concurrent-safe in-memory key-value store.
type Cache[K comparable, V any] struct {
	mu           sync.RWMutex
	items        map[K]entry[V]
	defaultTTL   time.Duration
	cleanupEvery time.Duration
	stopCh       chan struct{}
	doneCh       chan struct{}
}

// New creates a new, optimized Cache instance.
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

// Close stops the background janitor cleanly, blocking until it exits.
func (c *Cache[K, V]) Close() {
	if c.stopCh == nil {
		return
	}
	close(c.stopCh)
	<-c.doneCh
}

// Set inserts or updates a key-value pair using the default TTL.
func (c *Cache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL inserts or updates a key-value pair with a specific custom TTL.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	c.mu.Lock()
	c.items[key] = entry[V]{value: value, expireAt: exp}
	c.mu.Unlock()
}

// Get retrieves a value by key. Handles lazy deletion on cache misses.
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

	// Double-check expiration under a write lock to perform lazy deletion safely
	c.mu.Lock()
	if e2, ok2 := c.items[key]; ok2 && !e2.expireAt.IsZero() && time.Now().After(e2.expireAt) {
		delete(c.items, key)
	}
	c.mu.Unlock()

	var zero V
	return zero, false
}

// Pop removes and returns an item from the cache.
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
		ticker := time.NewTicker(c.cleanupEvery)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
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

### Exercise 2: Graceful Worker Pool with Context Cancellation
A crucial live coding template that demonstrates advanced channel coordination, error propagation, and resource cleanup.

<details>
<summary><b>Click to view Worker Pool Implementation</b></summary>

```go
package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Task represents a unit of concurrent work.
type Task struct {
	ID   int
	Data string
}

// Result represents the outcome of a processed Task.
type Result struct {
	TaskID int
	Output string
	Err    error
}

// WorkerPool manages the concurrent processing of tasks.
type WorkerPool struct {
	numWorkers  int
	tasksChan   chan Task
	resultsChan chan Result
	wg          sync.WaitGroup
}

// NewWorkerPool initializes a new WorkerPool.
func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		numWorkers:  numWorkers,
		tasksChan:   make(chan Task),
		resultsChan: make(chan Result),
	}
}

// Start spawns the workers and prepares them to listen for tasks.
func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 1; i <= wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

func (wp *WorkerPool) worker(ctx context.Context, workerID int) {
	defer wp.wg.Done()
	
	for {
		select {
		case <-ctx.Done():
			// Exit worker immediately on context cancellation
			return
		case task, ok := <-wp.tasksChan:
			if !ok {
				// Tasks channel closed, exit gracefully
				return
			}

			output, err := wp.process(ctx, task)
			
			select {
			case <-ctx.Done():
				return
			case wp.resultsChan <- Result{TaskID: task.ID, Output: output, Err: err}:
			}
		}
	}
}

func (wp *WorkerPool) process(ctx context.Context, t Task) (string, error) {
	// Simulate work or network request
	select {
	case <-time.After(50 * time.Millisecond):
	case <-ctx.Done():
		return "", ctx.Err()
	}

	if t.ID%5 == 0 {
		return "", fmt.Errorf("simulated error for task %d", t.ID)
	}
	return fmt.Sprintf("Success processing data: %s", t.Data), nil
}

// Submit sends a task into the queue.
func (wp *WorkerPool) Submit(task Task) {
	wp.tasksChan <- task
}

// Results returns a read-only channel for collecting outputs.
func (wp *WorkerPool) Results() <-chan Result {
	return wp.resultsChan
}

// Stop cleanly terminates all workers and closes open channels.
func (wp *WorkerPool) Stop() {
	close(wp.tasksChan)
	wp.wg.Wait()
	close(wp.resultsChan)
}
```
</details>

---

## 📚 Eight: High-Quality Curated Resources

- **Official Guides:**
  - [Effective Go](https://golang.org/doc/effective_go) — The definitive style guide.
  - [Go Memory Model](https://golang.org/ref/mem) — Details on synchronization invariants.
- **Deep-Dive Reading:**
  - [Go 101](https://go101.org/) — In-depth look at language semantics, runtime, and structures.
  - [High Performance Go (Dave Cheney)](https://dave.cheney.net/) — Essential profiling and optimization tips.
- **Tools & Profiling:**
  - `go tool pprof` — Native Go profiling suite.
  - `go tool trace` — Advanced execution tracer for scheduler bottlenecks.

---
*Maintained with ❤️ for the Go engineering community.*
