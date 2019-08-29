package zgo

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/torlangballe/zutil/ustr"
	"github.com/torlangballe/zutil/zfile"
)

//  Created by Tor Langballe on /31/10/15.

type OutputStream int

type FileInfo struct {
	DataSize int64
	Created  Time
	Modified Time
	Accessed Time
	IsLocked bool
	IsHidden bool
	IsFolder bool
	IsLink   bool
}

type FilePathWalkOptions int

const (
	FilePathWalkNone         FilePathWalkOptions = 0
	FilePathWalkSubFolders                       = 1
	FilePathWalkGetInfo                          = 2
	FilePathWalkGetInvisible                     = 4
)

type FilePath struct {
	fpath string
}

//type FileUrl FilePath

func FilePathMake(spath string, isDir, dirUnknow bool) FilePath {
	return FilePath{spath}
}

func (f FilePath) IsEmpty() bool {
	return f.fpath == ""
}

func (f FilePath) ParentPath() string {
	return filepath.Dir(f.fpath)
}

func (f FilePath) OpenOutput(append bool) (*OutputStream, *Error) {
	return nil, ErrorNew("couldn't make stream", 0, "")
}

func (f FilePath) String() string {
	return f.fpath
}

func (f FilePath) IsFolder() bool {
	return StrTail(f.fpath, 1) == "/"
}

func (f FilePath) IsPhysicallyAFolder() bool {
	return zfile.IsFolder(f.fpath)
}

func (f FilePath) Exists() bool {
	return zfile.DoesFileExist(f.fpath)
}

func (f FilePath) CreateFolder(withIntermediates bool) (err *Error) {
	if withIntermediates {
		return ErrorFromErr(os.MkdirAll(f.fpath, 0775|os.ModeDir))
	}
	return ErrorFromErr(os.Mkdir(f.fpath, 0775|os.ModeDir))
}

func (f FilePath) GetDisplayName() string {
	_, name := filepath.Split(f.fpath)
	return name
}

func FilePathGetLegalFilename(filename string) string {
	var str = StrUrlEncode(filename)
	if len(str) > 200 {
		str = ustr.HashTo32Hex(filename) + "_" + StrTail(str, 200)
	}
	return str
}

func FilePathGetPathParts(path string) (string, string, string, string) { // place/a.txt = "place" "a.txt" "a" ".txt"
	dir, name := filepath.Split(path)
	ext := filepath.Ext(name)
	stub := StrBody(name, 0, len(name)-len(ext))
	return dir, name, stub, ext
}

func (f *FilePath) SetModified(t Time) *Error {
	err := zfile.SetModified(f.fpath, t.Time)
	return ErrorFromErr(err)
}

func (f FilePath) GetFiles(options FilePathWalkOptions, wildcard string) map[FilePath]FileInfo {
	m := map[FilePath]FileInfo{}
	return m
}

func (f FilePath) Walk(options FilePathWalkOptions, wildcard string, foreach func(FilePath, *FileInfo) bool) *Error {
	err := filepath.Walk(f.fpath, func(fpath string, info os.FileInfo, err error) error {
		if options&FilePathWalkGetInvisible == 0 && strings.HasPrefix(fpath, ".") {
			return nil
		}
		if options&FilePathWalkSubFolders == 0 && info.IsDir() {
			return nil
		}
		var finfo FileInfo
		f := FilePathMake(fpath, info.IsDir(), false)
		getInfoFromStat(&finfo, info)
		foreach(f, &finfo)
		return nil
	})
	if err != nil {
		return ErrorFromErr(err)
	}

	return nil
}

func (f FilePath) CopyTo(to FilePath) *Error {
	return nil
}

func (f FilePath) MoveTo(to FilePath) *Error {
	return nil
}

func (f FilePath) LinkTo(to FilePath) *Error {
	return nil
}

func (f FilePath) ResolveSimlinkOrSelf() FilePath {
	return FilePath{}
}

func (f FilePath) Remove() *Error {
	return nil
}

func (f FilePath) RemoveContents() (err *Error, errors []*Error) {
	return
}

func (f FilePath) AppendedPath(path string, isDir bool) FilePath {
	str := filepath.Join(f.fpath, path)
	fp := FilePathMake(str, false, true)
	return fp
}

func getInfoFromStat(info *FileInfo, stat os.FileInfo) {
	info.DataSize = stat.Size()
	info.Modified = Time{stat.ModTime()}
	//	info.IsLocked = stat.IsLocked
	info.IsFolder = stat.IsDir()
	fstat := stat.Sys().(*syscall.Stat_t)
	info.Created = getCreatedTimeFromStatT(fstat)
}

func (f FilePath) GetInfo() (info FileInfo, err *Error) {
	stat, e := os.Stat(f.fpath)
	if e != nil {
		err = ErrorFromErr(e)
		return
	}
	getInfoFromStat(&info, stat)
	return info, nil
}

func (f FilePath) GetModified() Time {
	info, err := f.GetInfo()
	if err != nil {
		return Time{}
	}
	return info.Modified
}

func (f FilePath) GetCreated() Time {
	info, err := f.GetInfo()
	if err != nil {
		return Time{}
	}
	return info.Created
}

func (f FilePath) GetDataSizeInBytes() int64 {
	info, err := f.GetInfo()
	if err != nil {
		return -1
	}
	return info.DataSize
}
