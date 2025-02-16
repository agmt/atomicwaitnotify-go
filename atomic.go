package atomicwaitnotify

import (
	"sync/atomic"
	"unsafe"
)

type Uint32 struct {
	atomic.Uint32
}

func WrapAtomicUint32(val *uint32) *Uint32 {
	if uintptr(unsafe.Pointer(val))%unsafe.Alignof(*val) != 0 {
		panic("unaligned pointer")
	}
	return (*Uint32)(unsafe.Pointer(val))
}

func (u *Uint32) Wait(old uint32) {
	WaitUint32((*uint32)(unsafe.Pointer(&u.Uint32)), old)
}

func (u *Uint32) NotifyOne() {
	NotifyOneUint32((*uint32)(unsafe.Pointer(&u.Uint32)))
}

func (u *Uint32) NotifyAll() {
	NotifyAllUint32((*uint32)(unsafe.Pointer(&u.Uint32)))
}

type Uint64 struct {
	atomic.Uint64
}

func WrapAtomicUint64(val *uint64) *Uint64 {
	if uintptr(unsafe.Pointer(val))%unsafe.Alignof(*val) != 0 {
		panic("unaligned pointer")
	}
	return (*Uint64)(unsafe.Pointer(val))
}

func (u *Uint64) Wait(old uint64) {
	WaitUint64((*uint64)(unsafe.Pointer(&u.Uint64)), old)
}

func (u *Uint64) NotifyOne() {
	NotifyOneUint64((*uint64)(unsafe.Pointer(&u.Uint64)))
}

func (u *Uint64) NotifyAll() {
	NotifyAllUint64((*uint64)(unsafe.Pointer(&u.Uint64)))
}
