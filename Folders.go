package zgo

import (
	"os"

	"github.com/torlangballe/zutil/zfile"
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

func FoldersGetFileInFolderType(ftype FolderType, addPath string) FilePath {
	dir := false
	if ftype == FoldersTemporaryUniqueFolder {
		f := zfile.CreateTempFilePath(addPath)
		fp := FilePathMake(f, dir)
		return fp
	}

	switch ftype {
	case FoldersAppSupport:
		return FilePathMake("~/"+addPath, dir)

	case FoldersCaches:
		udir, _ := os.UserCacheDir()
		return FilePathMake(udir+addPath, dir)

	case FoldersTemporary:
		return FilePathMake(os.TempDir()+addPath, dir)

	case FoldersPreferences:
		return FilePathMake("~/"+addPath, dir)
	}
	return FilePath{}
}

func ZGetResourceFilePath(subPath string) FilePath {
	return FilePath{}
}
