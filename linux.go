//go:build linux

package atomicwaitnotify

import (
	"math"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	FUTEX_WAIT            = 0
	FUTEX_WAKE            = 1
	FUTEX_FD              = 2
	FUTEX_REQUEUE         = 3
	FUTEX_CMP_REQUEUE     = 4
	FUTEX_WAKE_OP         = 5
	FUTEX_LOCK_PI         = 6
	FUTEX_UNLOCK_PI       = 7
	FUTEX_TRYLOCK_PI      = 8
	FUTEX_WAIT_BITSET     = 9
	FUTEX_WAKE_BITSET     = 10
	FUTEX_WAIT_REQUEUE_PI = 11
	FUTEX_CMP_REQUEUE_PI  = 12
	FUTEX_LOCK_PI2        = 13

	FUTEX_PRIVATE_FLAG   = 128
	FUTEX_CLOCK_REALTIME = 256
	FUTEX_CMD_MASK       = ^(FUTEX_PRIVATE_FLAG | FUTEX_CLOCK_REALTIME)
)

// futexWait blocks until the futex value at addr is changed or timeout occurs.
func futexWait(addr *uint32, val uint32, timeout *unix.Timespec) error {
	_, _, errno := unix.Syscall6(
		unix.SYS_FUTEX,
		uintptr(unsafe.Pointer(addr)),
		uintptr(FUTEX_WAIT),
		uintptr(val),
		uintptr(unsafe.Pointer(timeout)),
		uintptr(unsafe.Pointer(nil)),
		uintptr(0),
	)
	switch errno {
	case 0:
		return nil
	case unix.EAGAIN:
		return nil
	default:
		return errno
	}
}

// futexWake wakes up n threads waiting on the futex at addr.
func futexWake(addr *uint32, n int32) error {
	_, _, errno := unix.RawSyscall6(
		unix.SYS_FUTEX,
		uintptr(unsafe.Pointer(addr)),
		uintptr(FUTEX_WAKE),
		uintptr(n),
		uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(nil)),
		uintptr(0),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func WaitUint32(addr *uint32, old uint32) error {
	return futexWait(addr, old, nil)
}

func WaitTimeoutUint32(addr *uint32, old uint32, timeout time.Duration) error {
	tm := unix.NsecToTimespec(timeout.Nanoseconds())
	return futexWait(addr, old, &tm)
}

func NotifyOneUint32(addr *uint32) error {
	return futexWake(addr, 1)
}

func NotifyAllUint32(addr *uint32) error {
	return futexWake(addr, math.MaxInt32)
}

func WaitUint64(addr *uint64, old uint64) error {
	// futex is 32-bit, so to avoid deadlock due to race-condition and alteration of high half, use timeout=2s
	// https://github.com/llvm/llvm-project/blob/main/libcxx/src/atomic.cpp#L58
	tm := unix.Timespec{Sec: 2}
	return futexWait((*uint32)(unsafe.Pointer(addr)), uint32(old), &tm)
}

func WaitTimeoutUint64(addr *uint64, old uint64, timeout time.Duration) error {
	tm := unix.NsecToTimespec(timeout.Nanoseconds())
	return futexWait((*uint32)(unsafe.Pointer(addr)), uint32(old), &tm)
}

func NotifyOneUint64(addr *uint64) error {
	return futexWake((*uint32)(unsafe.Pointer(addr)), 1)
}

func NotifyAllUint64(addr *uint64) error {
	return futexWake((*uint32)(unsafe.Pointer(addr)), math.MaxInt32)
}
