### glb
`glb` stands for `glb Load Balancer`; no, it's not `Generics Load Balancer`.

I decided to practice generics in Go 1.18 and developed this in the process.
The idea behind it was hugely influence by [bag](https://github.com/toeydevelopment/bag).
Kudos to the author...

##### Cool Little Escape Analysis
There are two implementations of `LeastConns` strategy. They are only different in
how they internally choose the least used backend. `LeastConns` holds `backends []backend[T]`
and returns the index of least used backend through `leastUsedBackendIdx`. `LeastConnsHeap`,
on the other hand, contains `backends []*backend[T]` and returns the least used backend
as `*backend[T]` through `leastUsedBackend`. These algorithms are safe for concurrent
usage, meaning that instances of these structs can be used by several goroutines at the same time.
One ramification of this for `LeastConnsHeap` is that each `*backend` will be allocated on the
heap, and as a result in situations where many instances of the structs are being created, each load balancing amongst hundreds of backends (e.g. think of modern cloud native apps which are comprised of hundreds of pods)
`LeastConns` is more preferable.

Following is the output of escape analysis for examples of `LeastConns` and `LeastConnsHeap` where I've marked
heap allocation lines with arrows:

```bash
go build -gcflags="-m" ./examples/leastconns/...
examples/leastconns/main.go:28:41: []*url.URL{...} does not escape
examples/leastconns/main.go:28:30: make([]glb.backend[go.shape.*uint8_0], len(glb.backends)) escapes to heap
examples/leastconns/main.go:28:30: &glb.LeastConns[go.shape.*uint8_0]{...} escapes to heap
glb/leastconns.go:25:12: make([]glb.backend[go.shape.*uint8_0], len(glb.backends)) escapes to heap
glb/leastconns.go:32:8: &glb.LeastConns[go.shape.*uint8_0]{...} escapes to heap
```

```bash
go build -gcflags="-m" ./examples/leastconnsheap/...
examples/leastconnsheap/main.go:28:45: []*url.URL{...} does not escape
examples/leastconnsheap/main.go:28:34: make([]*glb.backend[go.shape.*uint8_0], len(glb.backends)) escapes to heap
examples/leastconnsheap/main.go:28:34: &glb.backend[go.shape.*uint8_0]{...} escapes to heap <== **allocs**
examples/leastconnsheap/main.go:28:34: &glb.LeastConnsHeap[go.shape.*uint8_0]{...} escapes to heap
glb/leastconnsheap.go:20:12: make([]*glb.backend[go.shape.*uint8_0], len(glb.backends)) escapes to heap
glb/leastconnsheap.go:22:11: &glb.backend[go.shape.*uint8_0]{...} escapes to heap  <== **allocs**
glb/leastconnsheap.go:27:8: &glb.LeastConnsHeap[go.shape.*uint8_0]{...} escapes to heap
```