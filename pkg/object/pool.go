package object

import (
	"sync"
)

// SmallIntegerCache caches small integer objects to reduce allocations
// Similar to Java's Integer cache for -128 to 127
const (
	SmallIntCacheMin = -128
	SmallIntCacheMax = 256
	SmallIntCacheSize = SmallIntCacheMax - SmallIntCacheMin + 1
)

var smallIntCache [SmallIntCacheSize]*Integer
var intCacheOnce sync.Once

// initIntCache initializes the small integer cache
func initIntCache() {
	for i := SmallIntCacheMin; i <= SmallIntCacheMax; i++ {
		smallIntCache[i-SmallIntCacheMin] = &Integer{Value: int64(i)}
	}
}

// GetCachedInteger returns a cached integer or creates a new one
// This significantly reduces GC pressure for common integer values
func GetCachedInteger(value int64) *Integer {
	intCacheOnce.Do(initIntCache)
	
	if value >= SmallIntCacheMin && value <= SmallIntCacheMax {
		return smallIntCache[value-SmallIntCacheMin]
	}
	return &Integer{Value: value}
}

// String pool for interning common strings
var stringPool = sync.Map{}

// GetInternedString returns an interned string object
// Strings with the same value share the same object
func GetInternedString(value string) *String {
	if len(value) > 64 {
		// Don't intern long strings
		return &String{Value: value}
	}
	
	if cached, ok := stringPool.Load(value); ok {
		return cached.(*String)
	}
	
	str := &String{Value: value}
	// Use LoadOrStore to avoid race conditions
	actual, _ := stringPool.LoadOrStore(value, str)
	return actual.(*String)
}

// Object pools for reducing allocations
var arrayPool = sync.Pool{
	New: func() interface{} {
		return &Array{Elements: make([]Object, 0, 8)}
	},
}

// GetPooledArray gets an array from the pool
func GetPooledArray(size int) *Array {
	arr := arrayPool.Get().(*Array)
	if cap(arr.Elements) < size {
		arr.Elements = make([]Object, 0, size)
	} else {
		arr.Elements = arr.Elements[:0]
	}
	return arr
}

// ReturnPooledArray returns an array to the pool
func ReturnPooledArray(arr *Array) {
	if arr != nil && cap(arr.Elements) <= 1024 {
		arr.Elements = arr.Elements[:0]
		arrayPool.Put(arr)
	}
}

// Hash pool for reducing hash map allocations
var hashPool = sync.Pool{
	New: func() interface{} {
		return &Hash{Pairs: make(map[HashKey]HashPair, 8)}
	},
}

// GetPooledHash gets a hash from the pool
func GetPooledHash() *Hash {
	h := hashPool.Get().(*Hash)
	// Clear existing entries
	for k := range h.Pairs {
		delete(h.Pairs, k)
	}
	return h
}

// ReturnPooledHash returns a hash to the pool
func ReturnPooledHash(h *Hash) {
	if h != nil && len(h.Pairs) <= 256 {
		hashPool.Put(h)
	}
}

// Null singleton - avoid creating new Null objects
var NullObj = &Null{}

// Boolean singletons
var (
	TrueObj  = &Boolean{Value: true}
	FalseObj = &Boolean{Value: false}
)

// GetBoolean returns the singleton boolean object
func GetBoolean(value bool) *Boolean {
	if value {
		return TrueObj
	}
	return FalseObj
}

// Closure pool for reducing closure allocations
var closurePool = sync.Pool{
	New: func() interface{} {
		return &Closure{}
	},
}

// GetPooledClosure gets a closure from the pool
func GetPooledClosure(fn *CompiledFunction, numFree int) *Closure {
	cl := closurePool.Get().(*Closure)
	cl.Fn = fn
	if cap(cl.Free) < numFree {
		cl.Free = make([]Object, numFree)
	} else {
		cl.Free = cl.Free[:numFree]
	}
	return cl
}

// Error pool
var errorPool = sync.Pool{
	New: func() interface{} {
		return &Error{}
	},
}

// GetPooledError gets an error from the pool
func GetPooledError(msg string) *Error {
	e := errorPool.Get().(*Error)
	e.Message = msg
	e.Line = 0
	e.Column = 0
	return e
}
