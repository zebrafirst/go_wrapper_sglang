package finderhttp

import (
	"context"
	"hash"
	"hash/crc32"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"unsafe"
)

// 轮询
func LBRoundRobin() Lb {
	return &roundRobin{}
}

// 普通hash
func LBHash(hashKeys []string) Lb {
	return &hhash{HashKeys: hashKeys}
}

// 一致性hash
// vitrualNodes 每个节点的虚拟节点数量,建议10 以上,默认100
// hashKeys hash 字段，值需要能够从 context 中获取到
func LBConsistencyHash(hashKeys []string, vitrualNodes int) Lb {
	if vitrualNodes <= 0 {
		vitrualNodes = 100
	}
	return &consistencyHash{hashKeys: hashKeys, virtualNodes: vitrualNodes}
}

type roundRobin struct {
	seed atomic.Uint64
}

func (w *roundRobin) Get(ctx context.Context, addr []*Target) (string, error) {
	if len(addr) == 0 {
		return "", ErrNoPeerAddrFound
	}
	idx := w.seed.Add(1)

	return addr[idx%uint64(len(addr))].Address, nil
}

type hhash struct {
	// 使用的hash key 集合, val 需要写到context 中
	// 例如：
	// ctx.WithValue(parentCtx,"name",name)
	// hashlb.Get(ctx,addr)
	HashKeys []string
}

var (
	hashPools sync.Pool
)

func (w *hhash) Get(ctx context.Context, addr []*Target) (string, error) {
	if len(addr) == 0 {
		return "", ErrNoPeerAddrFound
	}
	idx := sum32Context(w.HashKeys, ctx)
	return addr[idx%uint32(len(addr))].Address, nil
}

func sum32Context(keys []string, ctx context.Context) uint32 {
	h, ok := hashPools.Get().(hash.Hash32)
	if !ok {
		h = crc32.NewIEEE()
	} else {
		h.Reset()
	}
	for _, key := range keys {
		val := ctx.Value(key)
		h.Write(BytesOf(String(val)))
	}
	v := h.Sum32()
	hashPools.Put(h)
	return v
}

type consistencyHash struct {
	hashKeys     []string
	currentPtr   atomic.Uintptr
	targets      atomic.Pointer[[]*Target]
	lock         sync.Mutex
	virtualNodes int
}

func (w *consistencyHash) updateTargets(addr []*Target) {
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&addr))
	// addr 地址发生了变化，说明地址有更新，需要重新计算hash环
	if w.currentPtr.Load() != sh.Data {
		w.lock.Lock()
		// double check 防止重复初始化
		if w.currentPtr.Load() == sh.Data {
			w.lock.Unlock()
			return
		}
		tgs := make([]*Target, 0, len(addr)*w.virtualNodes)
		for _, t := range addr {
			// 生成虚拟节点，让负载更加均衡
			tgs = append(tgs, genVirtualNode(t, w.virtualNodes)...)
		}
		// 根据hash 排序
		sort.Slice(tgs, func(i, j int) bool {
			return tgs[i].AddressSum32 < tgs[j].AddressSum32
		})

		w.targets.Store(&tgs)
		w.currentPtr.Store(sh.Data)
		w.lock.Unlock()
	}
}

func (w *consistencyHash) Get(ctx context.Context, addr []*Target) (string, error) {
	if len(addr) == 0 {
		return "", ErrNoPeerAddrFound
	}

	w.updateTargets(addr)

	tgs := *w.targets.Load()

	hash32 := sum32Context(w.hashKeys, ctx)

	i := sort.Search(len(tgs), func(i int) bool {
		return tgs[i].AddressSum32 >= hash32
	})

	return tgs[i%len(tgs)].Address, nil
}

// 生成虚拟节点
func genVirtualNode(t *Target, n int) (res []*Target) {
	// 虚拟节点的hash 重新生成时不能变化，因此每次需要使用固定的种子来生成虚拟节点的hash 值
	// 这里使用节点hash 作为随机数种子来生成虚拟节点的hash，
	rands := rand.New(rand.NewSource(int64(t.AddressSum32)))
	res = make([]*Target, 0, n)
	for i := 0; i < n; i++ {
		res = append(res, &Target{
			Address:      t.Address,
			AddressSum32: rands.Uint32(),
		})
	}
	return res
}
