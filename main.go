package main

import (
	"comicInfo/cbz"
	"syscall"
	"unsafe"
)

/**
 * 2024/2/5
 * add by stardust
**/

func IntPtr(n int) uintptr {
	return uintptr(n)
}
func StrPtr(s string) uintptr {
	return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))
}

// ShowMessage windows下的另一种DLL方法调用
func ShowMessage(tittle, text string) {
	user32dll, _ := syscall.LoadLibrary("user32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")
	MessageBoxW := user32.NewProc("MessageBoxW")
	_, _, _ = MessageBoxW.Call(IntPtr(0), StrPtr(text), StrPtr(tittle), IntPtr(0))
	defer func(handle syscall.Handle) {
		_ = syscall.FreeLibrary(handle)
	}(user32dll)
}

func main() {
	if cbz.GetInfo() {
		ShowMessage("处理结果", "生成成功")
	} else {
		ShowMessage("处理结果", "生成失败,可以查看日志确定错误")
	}

}
