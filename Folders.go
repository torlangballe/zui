package zui

import (
	"os"

	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /30/10/15.

type FolderType int

const (
	FoldersPreferences           FolderType = 1
	FoldersCaches                           = 4
	FoldersTemporary                        = 8
	FoldersAppSupport                       = 16
	FoldersTemporaryUniqueFolder            = 32
)

func FoldersGetFileInFolderType(ftype FolderType, addPath string) string {
	if ftype == FoldersTemporaryUniqueFolder {
		f := zfile.CreateTempFilePath(addPath)
		return f
	}

	switch ftype {
	case FoldersAppSupport:
		return zfile.ExpandTildeInFilepath("~/" + addPath)

	case FoldersCaches:
		udir, _ := os.UserCacheDir()
		return udir + addPath

	case FoldersTemporary:
		return os.TempDir() + addPath

	case FoldersPreferences:
		return zfile.ExpandTildeInFilepath("~/" + addPath)
	}
	zlog.Fatal(nil, "wrong type:", ftype)
	return ""
}

func ZGetResourceFilePath(subPath string) string {
	return ""
}
