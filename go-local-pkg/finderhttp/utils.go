package finderhttp

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"unsafe"
)

type SyncPool[T any] struct {
	p sync.Pool
}

func (p *SyncPool[T]) Get() *T {
	data, ok := p.p.Get().(*T)
	if ok {
		return data
	}
	return new(T)
}

func (p *SyncPool[T]) Put(t *T) {
	p.p.Put(t)
}

type SyncMap[K any, V any] struct {
	m sync.Map
}

func (s *SyncMap[K, V]) Store(k K, v V) {
	s.m.Store(k, v)
}

func (s *SyncMap[K, V]) Load(k K) (v V, ok bool) {
	vv, ok := s.m.Load(k)
	if ok {
		return vv.(V), true
	}
	return v, false
}

func (s *SyncMap[K, V]) Delete(k K) {
	s.m.Delete(k)
}

func (s *SyncMap[K, V]) Range(f func(k K, v V) bool) {
	s.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

type Set[K any] struct {
	m SyncMap[K, bool]
}

func (s *Set[K]) Add(k K) {
	s.m.Store(k, true)
}

func (s *Set[K]) Delete(k K) {
	s.m.Delete(k)
}

func (s *Set[K]) Exists(k K) bool {
	_, ok := s.m.Load(k)
	return ok
}

func (s *Set[K]) ToSlice() (res []K) {
	s.m.Range(func(k K, v bool) bool {
		res = append(res, k)
		return true
	})
	return
}

type comparableKey interface {
	string | int | int32 | int64 | uint | uint32 | uint64
}

type SyncMapList[K comparableKey, V any] struct {
	m     sync.Map
	vlist atomic.Pointer[[]V]

	lock sync.RWMutex
}

type mapkv[K comparableKey, V any] struct {
	k K
	v V
}

func (s *SyncMapList[K, V]) Store(k K, v V) {
	s.lock.Lock()

	s.m.Store(k, v)

	s.valAsList()
	s.lock.Unlock()
}

func (s *SyncMapList[K, V]) Load(k K) (v V, ok bool) {
	vv, ok := s.m.Load(k)
	if ok {
		return vv.(V), true
	}
	return v, false
}

func (s *SyncMapList[K, V]) Delete(k K) {
	s.lock.Lock()

	s.m.Delete(k)

	s.valAsList()

	s.lock.Unlock()
}

func (s *SyncMapList[K, V]) Range(f func(k K, v V) bool) {
	s.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

func (s *SyncMapList[K, V]) valAsList() (res []V) {
	kvs := []mapkv[K, V]{}
	s.Range(func(k K, v V) bool {
		// res = append(res, v)
		kvs = append(kvs, mapkv[K, V]{k, v})
		return true
	})

	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].k < kvs[j].k
	})

	for _, kv := range kvs {
		res = append(res, kv.v)
	}
	s.vlist.Store(&res)
	return res
}

func (s *SyncMapList[K, V]) ValList() []V {
	p := s.vlist.Load()
	if p == nil {
		return s.valAsList()
	}
	return *p
}

func String(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return StringOf(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func StringOf(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func BytesOf(s string) []byte {
	p := (*[2]uintptr)(unsafe.Pointer(&s))
	bptr := [3]uintptr{p[0], p[1], p[1]}
	return *(*[]byte)(unsafe.Pointer(&bptr))
}
