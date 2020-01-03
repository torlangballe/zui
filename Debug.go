package zui

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/ztime"
)

var mutex sync.Mutex
var storePrintLines = 0
var storedLines []string
var lastStampTime time.Time
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
	if ztime.Since(lastStampTime) > 3.0 {
		lastStampTime = time.Now()
		str = lastStampTime.Format("============= 06-01-02' '15:04:05 =============\n")
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
				zslice.Behead(&storedLines)
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
	str, _ := zfile.ReadFromFile(file)
	storedLines = strings.Split(str, "\n")
}

func DebugAppendToFileAndClearLog(prefix string) {
	file := FoldersGetFileInFolderType(FoldersTemporary, prefix+"/zdebuglog.txt")
	if zfile.GetSize(file) > 5*1024*1024 {
		os.Remove(file)
		storedLines = append([]string{"--- ZDebug Cleared early part of large stored log."}, storedLines...)
	}
	storedLines = storedLines[:0]
}

func DebugAssert(success bool) {
	if !success {
		panic("assert failed:\n" + string(debug.Stack()))
	}
}
