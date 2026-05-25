package platform

import "runtime"

func GOOS() string {
	return runtime.GOOS
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsUnix() bool {
	return runtime.GOOS == "darwin" || runtime.GOOS == "linux"
}
