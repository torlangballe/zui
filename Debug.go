package zgo

import (
	"fmt"
	"runtime/debug"
	"sync"
)

var mutex sync.Mutex
var storePrintLines = 0
var storedLines []string
var lastStampTime Time
var printHooks []func(s string)
var logAllOutput bool

func DebugIsRelease() bool {
	return false
}

func IsMinIOS11() bool {
	return false
}

func DebugPrint(items ...interface{}) {
	var str = ""
	if lastStampTime.Since() > 3.0 {
		lastStampTime = TimeNow()
		str = lastStampTime.GetString("============= yy-MM-dd' 'HH:mm:ssZZ =============\n", "", nil)
	}
	for i, item := range items {
		if i != 0 {
			str += " "
		}
		if item == nil {
			str += "<nil>"
		} else {
			str += fmt.Sprintf("%v", item)
		}
	}
	mutex.Lock()
	if !logAllOutput {
		if storePrintLines != 0 {
			if len(storedLines) > storePrintLines {
				storedLines = StrRemovedFirst(storedLines)
			}
			storedLines = append(storedLines, str)
		}
	}
	for _, h := range printHooks {
		h(str)
	}
	mutex.Unlock()
	fmt.Println(str)
}

func ErrorOnRelease() {
	if DebugIsRelease() {
		var n = 100
		for n > 0 {
			DebugPrint("Should not run on ")
			n--
		}
	}
}

func DebugLoadSavedLog(prefix string) {
	file := FoldersGetFileInFolderType(FoldersTemporary, prefix+"/zdebuglog.txt")
	str, _ := file.LoadString()
	storedLines = StrSplit(str, "\n")
}

func DebugAppendToFileAndClearLog(prefix string) {
	file := FoldersGetFileInFolderType(FoldersTemporary, prefix+"/zdebuglog.txt")
	if file.GetDataSizeInBytes() > 5*1024*1024 {
		file.Remove()
		storedLines = append([]string{"--- ZDebug Cleared early part of large stored log."}, storedLines...)
	}
	storedLines = storedLines[:0]
}

func DebugAssert(success bool) {
	if !success {
		panic("assert failed:\n" + string(debug.Stack()))
	}
}
