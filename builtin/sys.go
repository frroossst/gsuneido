// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package builtin

import (
	"errors"
	"io/fs"
	"log"
	"net"
	"os"
	"runtime"
	"strings"

	. "github.com/apmckinlay/gsuneido/core"
)

var sysMem = systemMemory() // cache

var _ = builtin(SystemMemory, "()")

func SystemMemory() Value {
	return Int64Val(int64(sysMem))
}

var _ = builtin(MemoryArena, "()")

func MemoryArena() Value {
	return Int64Val(int64(HeapSys()))
}

func HeapSys() uint64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.HeapSys
}

var _ = builtin(GetCurrentDirectory, "()")

func GetCurrentDirectory() Value {
	dir, err := os.Getwd()
	if err != nil {
		panic("GetCurrentDirectory: " + err.Error())
	}
	return SuStr(dir)
}

// NOTE: temp file is NOT deleted automatically on exit
// (same as cSuneido, but different from jSuneido)
var _ = builtin(GetTempFileName, "(path, prefix)")

func GetTempFileName(path, prefix Value) Value {
	f, err := os.CreateTemp(ToStr(path), ToStr(prefix)+"*.tmp")
	if err != nil {
		panic("GetTempFileName: " + err.Error())
	}
	filename := f.Name()
	f.Close()
	filename = strings.Replace(filename, `\`, `/`, -1)
	return SuStr(filename)
}

var _ = builtin(CreateDir, "(dirname)")

func CreateDir(th *Thread, args []Value) Value {
	path := ToStr(args[0])
	err := os.Mkdir(path, 0755)
	if errors.Is(err, os.ErrExist) {
		if info, err2 := os.Stat(path); err2 == nil && info.Mode().IsDir() {
			// not return-throw
			return SuStr("CreateDir " + path + ": already exists")
		}
		err = errors.New(path + ": exists but is not a directory")
	}
	if err != nil {
		th.ReturnThrow = true
		return SuStr("CreateDir: " + err.Error())
	}
	return True
}

func init() { // TEMP for transition
	Global.Builtin("CreateDirectory",
		builtinVal("CreateDirectory", CreateDir, "(dirname)"))
}

var _ = builtin(DeleteFileApi, "(filename)")

func DeleteFileApi(th *Thread, args []Value) Value {
	path := ToStr(args[0])
	err := deleteFile(path) // see sys_unix.go and sys_windows.go
	if errors.Is(err, os.ErrNotExist) {
		// not return-throw
		return SuStr("DeleteFileApi " + path + ": does not exist")
	}
	if err != nil {
		th.ReturnThrow = true
		return SuStr("DeleteFileApi " + path + ": " + err.Error())
	}
	return True
}

var _ = builtin(FileExistsQ, "(filename)")

func FileExistsQ(arg Value) Value {
	path := ToStr(arg)
	info, err := os.Stat(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Println("INFO: FileExists?", path, err)
	}
	return SuBool(err == nil && info.Mode().IsRegular())
}

var _ = builtin(DirExistsQ, "(filename)")

func DirExistsQ(arg Value) Value {
	path := ToStr(arg)
	info, err := os.Stat(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Println("INFO: DirExists?", path, err)
	}
	return SuBool(err == nil && info.Mode().IsDir())
}

var _ = builtin(MoveFile, "(from, to)")

func MoveFile(th *Thread, args []Value) Value {
	from := ToStr(args[0])
	to := ToStr(args[1])
	err := os.Rename(from, to)
	if err == nil {
		return True
	}
	th.ReturnThrow = true
	return SuStr("MoveFile: " + err.Error())
}

var _ = builtin(DeleteDir, "(dir)")

func DeleteDir(th *Thread, args []Value) Value {
	path := ToStr(args[0])
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		// not return-throw
		return SuStr("DeleteDir " + path + ": does not exist")
	}
	th.ReturnThrow = true
	if err != nil {
		return SuStr("DeleteDir: " + err.Error())
	}
	if !info.Mode().IsDir() {
		return SuStr("DeleteDir " + path + ": not a directory")
	}
	err = os.RemoveAll(path)
	if err != nil {
		return SuStr("DeleteDir: " + err.Error())
	}
	return True
}

var _ = builtin(GetMacAddresses, "()")

func GetMacAddresses() Value {
	ob := &SuObject{}
	if intfcs, err := net.Interfaces(); err == nil {
		for _, intfc := range intfcs {
			if s := string(intfc.HardwareAddr); s != "" {
				ob.Add(SuStr(s))
			}
		}
	}
	return ob
}

var _ = builtin(GetTempPath, "()")

func GetTempPath() Value {
	s := os.TempDir()
	s = strings.ReplaceAll(s, `\`, `/`)
	if !strings.HasSuffix(s, "/") {
		s += "/"
	}
	return SuStr(s)
}

var _ = builtin(OSName, "()")

func OSName() Value {
	os := runtime.GOOS
	os = strings.Replace(os, "darwin", "macos", 1)
	return SuStr(os)
}
