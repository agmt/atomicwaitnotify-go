//go:build windows

// !!! WaitOnAddress and WakeByAddressSingle/WakeByAddressAll doesn't work on different processes !!!
package atomicwaitnotify

import (
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modsynch                = syscall.NewLazyDLL("API-MS-Win-Core-Synch-l1-2-0.dll")
	procWaitOnAddress       = modsynch.NewProc("WaitOnAddress")
	procWakeByAddressSingle = modsynch.NewProc("WakeByAddressSingle")
	procWakeByAddressAll    = modsynch.NewProc("WakeByAddressAll")
)

func waitOnAddress[T ~uint8 | ~uint16 | ~uint32 | ~uint64](addr *T, old T, timeout time.Duration) error {
	size := unsafe.Sizeof(old)
	var ms uint32
	if timeout == 0 {
		ms = syscall.INFINITE
	} else {
		ms = uint32(timeout.Milliseconds())
	}
	ret, _, err := procWaitOnAddress.Call(
		uintptr(unsafe.Pointer(addr)),
		uintptr(unsafe.Pointer(&old)),
		size,
		uintptr(ms),
	)
	if ret == 0 { // ret bool
		switch err {
		case windows.ERROR_TIMEOUT:
			return syscall.ETIMEDOUT
		default:
			return err
		}
	}
	return nil
}

func WaitUint32(addr *uint32, old uint32) error {
	return waitOnAddress(addr, old, 0)
}

func WaitTimeoutUint32(addr *uint32, old uint32, timeout time.Duration) error {
	return waitOnAddress(addr, old, timeout)
}

func NotifyOneUint32(addr *uint32) error {
	procWakeByAddressSingle.Call(uintptr(unsafe.Pointer(addr)))
	return nil
}

func NotifyAllUint32(addr *uint32) error {
	procWakeByAddressAll.Call(uintptr(unsafe.Pointer(addr)))
	return nil
}

func WaitUint64(addr *uint64, old uint64) error {
	return waitOnAddress(addr, old, 0)
}

func WaitTimeoutUint64(addr *uint64, old uint64, timeout time.Duration) error {
	return waitOnAddress(addr, old, timeout)
}

func NotifyOneUint64(addr *uint64) error {
	procWakeByAddressSingle.Call(uintptr(unsafe.Pointer(addr)))
	return nil
}

func NotifyAllUint64(addr *uint64) error {
	procWakeByAddressAll.Call(uintptr(unsafe.Pointer(addr)))
	return nil
}
