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
	// 加载user32.dll
	user32dll, _ := syscall.LoadLibrary("user32.dll")
	// 创建一个新的lazyDLL
	user32 := syscall.NewLazyDLL("user32.dll")
	// 获取MessageBoxW方法
	MessageBoxW := user32.NewProc("MessageBoxW")
	// 调用MessageBoxW方法
	_, _, _ = MessageBoxW.Call(IntPtr(0), StrPtr(text), StrPtr(tittle), IntPtr(0))
	// 释放资源
	defer func(handle syscall.Handle) {
		_ = syscall.FreeLibrary(handle)
	}(user32dll)
}

func main() {
	if err := cbz.GetInfo(); err == nil {
		ShowMessage("生成成功", "文件生成成功!!!")
	} else {
		ShowMessage("生成失败", err.Error())
	}

}
