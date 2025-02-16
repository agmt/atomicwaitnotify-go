//go:build darwin

package atomicwaitnotify

import (
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	// https://github.com/apple-oss-distributions/xnu/blob/8d741a5de7ff4191bf97d57b9f54c2f6d4a15585/bsd/kern/syscalls.master#L856
	SYS_ULOCK_WAIT2 = 544 // Available since macOS 11

	// https://github.com/apple-oss-distributions/xnu/blob/main/bsd/sys/ulock.h
	UL_COMPARE_AND_WAIT          = 1
	UL_UNFAIR_LOCK               = 2
	UL_COMPARE_AND_WAIT_SHARED   = 3
	UL_UNFAIR_LOCK64_SHARED      = 4
	UL_COMPARE_AND_WAIT64        = 5
	UL_COMPARE_AND_WAIT64_SHARED = 6

	ULF_WAKE_ALL                   = 0x00000100
	ULF_WAKE_THREAD                = 0x00000200
	ULF_WAKE_ALLOW_NON_OWNER       = 0x00000400
	ULF_WAIT_WORKQ_DATA_CONTENTION = 0x00010000
	ULF_WAIT_CANCEL_POINT          = 0x00020000
	ULF_WAIT_ADAPTIVE_SPIN         = 0x00040000
	ULF_NO_ERRNO                   = 0x01000000

	UL_OPCODE_MASK   = 0x000000FF
	UL_FLAGS_MASK    = 0xFFFFFF00
	ULF_GENERIC_MASK = 0xFFFF0000

	ULF_WAIT_MASK = (ULF_NO_ERRNO |
		ULF_WAIT_WORKQ_DATA_CONTENTION |
		ULF_WAIT_CANCEL_POINT | ULF_WAIT_ADAPTIVE_SPIN)
	ULF_WAKE_MASK = (ULF_NO_ERRNO |
		ULF_WAKE_ALL |
		ULF_WAKE_THREAD |
		ULF_WAKE_ALLOW_NON_OWNER)
)

// Go suggests to use libSystem, but `__ulock_wait` is private and not available in Go's runtime
// int ulock_wait(uint32_t operation, void *addr, uint64_t value, uint32_t timeout)

// https://github.com/apple-oss-distributions/xnu/blob/8d741a5de7ff4191bf97d57b9f54c2f6d4a15585/bsd/kern/syscalls.master#L856
// int ulock_wait2(uint32_t operation, void *addr, uint64_t value, uint64_t timeout, uint64_t value2)
func ulock_wait2(operation uint32, addr unsafe.Pointer, value uint64, timeout uint64, value2 uint64) error {
	_, _, errno := unix.Syscall6(
		SYS_ULOCK_WAIT2,
		uintptr(operation),
		uintptr(addr),
		uintptr(value),
		uintptr(timeout),
		uintptr(value2),
		0)
	// https://github.com/jart/cosmopolitan/blob/master/libc/intrin/ulock.c#L34
	switch errno {
	case syscall.Errno(0):
		return nil
	case syscall.EINTR:
		return nil
	default:
		return errno
	}
}

// https://github.com/apple-oss-distributions/xnu/blob/8d741a5de7ff4191bf97d57b9f54c2f6d4a15585/bsd/kern/syscalls.master#L819
// int ulock_wake(uint32_t operation, void *addr, uint64_t wake_value)
func ulock_wake(operation uint32, addr unsafe.Pointer, wake_value uint64) error {
	_, _, errno := unix.RawSyscall6(
		unix.SYS_ULOCK_WAKE,
		uintptr(operation),
		uintptr(addr),
		uintptr(wake_value),
		0, 0, 0)
	// https://github.com/jart/cosmopolitan/blob/master/libc/intrin/ulock.c#L74
	switch errno {
	case syscall.Errno(0):
		return nil
	case syscall.ENOENT:
		return nil
	default:
		return errno
	}
}

func WaitUint32(addr *uint32, old uint32) error {
	return ulock_wait2(UL_COMPARE_AND_WAIT_SHARED, unsafe.Pointer(addr), uint64(old), 0, 0)
}

func WaitTimeoutUint32(addr *uint32, old uint32, timeout time.Duration) error {
	return ulock_wait2(UL_COMPARE_AND_WAIT_SHARED, unsafe.Pointer(addr), uint64(old), uint64(timeout.Nanoseconds()), 0)
}

func NotifyOneUint32(addr *uint32) error {
	return ulock_wake(UL_COMPARE_AND_WAIT_SHARED, unsafe.Pointer(addr), 0)
}

func NotifyAllUint32(addr *uint32) error {
	return ulock_wake(UL_COMPARE_AND_WAIT_SHARED|ULF_WAKE_ALL, unsafe.Pointer(addr), 0)
}

func WaitUint64(addr *uint64, old uint64) error {
	return ulock_wait2(UL_COMPARE_AND_WAIT64_SHARED, unsafe.Pointer(addr), uint64(old), 0, 0)
}

func WaitTimeoutUint64(addr *uint64, old uint64, timeout time.Duration) error {
	return ulock_wait2(UL_COMPARE_AND_WAIT64_SHARED, unsafe.Pointer(addr), uint64(old), uint64(timeout.Nanoseconds()), 0)
}

func NotifyOneUint64(addr *uint64) error {
	return ulock_wake(UL_COMPARE_AND_WAIT64_SHARED, unsafe.Pointer(addr), 0)
}

func NotifyAllUint64(addr *uint64) error {
	return ulock_wake(UL_COMPARE_AND_WAIT64_SHARED|ULF_WAKE_ALL, unsafe.Pointer(addr), 0)
}
