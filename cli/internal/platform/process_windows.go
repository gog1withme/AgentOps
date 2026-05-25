//go:build windows

package platform

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func PauseProcess(pid int) error {
	if pid <= 0 {
		return nil
	}
	handle, err := windows.OpenProcess(windows.PROCESS_SUSPEND_RESUME, false, uint32(pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	ntdll := windows.NewLazySystemDLL("ntdll.dll")
	ntSuspendProcess := ntdll.NewProc("NtSuspendProcess")
	r, _, e := ntSuspendProcess.Call(uintptr(handle))
	if r != 0 {
		if e != syscall.Errno(0) {
			return e
		}
	}
	return nil
}

func KillProcess(pid int) error {
	if pid <= 0 {
		return nil
	}
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	return windows.TerminateProcess(handle, 1)
}

func ShellName() string {
	if comspec := os.Getenv("COMSPEC"); comspec != "" {
		return comspec
	}
	return "powershell.exe"
}

func DetectShellType() string {
	if os.Getenv("PSModulePath") != "" {
		return "powershell"
	}
	return "cmd"
}

func _unused() {
	_ = unsafe.Pointer(nil)
}
