package input

import (
	"github.com/elastic/libbeat/logp"
	"os"
	"reflect"
)

type FileStateOS struct {
	IdxHi uint64 `json:"idxhi,omitempty"`
	IdxLo uint64 `json:"idxlo,omitempty"`
	Vol   uint64 `json:"vol,omitempty"`
}

// GetOSFileState returns the platform specific FileStateOS
func GetOSFileState(info *os.FileInfo) *FileStateOS {

	// Gathering fileStat (which is fileInfo) through reflection as otherwise not accessible
	// See https://github.com/golang/go/blob/90c668d1afcb9a17ab9810bce9578eebade4db56/src/os/stat_windows.go#L33
	fileStat := reflect.ValueOf(info).Elem()

	// Get the three fields required to uniquely identify file und windows
	// More details can be found here: https://msdn.microsoft.com/en-us/library/aa363788(v=vs.85).aspx
	// Uint should already return uint64, but making sure this is the case
	// The required fiels can be found here: https://github.com/golang/go/blob/master/src/os/types_windows.go#L78
	fileState := &FileStateOS{
		IdxHi: uint64(fileStat.FieldByName("idxhi").Uint()),
		IdxLo: uint64(fileStat.FieldByName("idxlo").Uint()),
		Vol:   uint64(fileStat.FieldByName("vol").Uint()),
	}

	return fileState
}

// IsSame file checks if the files are identical
func (fs *FileStateOS) IsSame(state *FileStateOS) bool {
	return fs.IdxHi == state.IdxHi && fs.IdxLo == state.IdxLo && fs.Vol == state.Vol
}

// SafeFileRotate safely rotates an existing file under path and replaces it with the tempfile
func SafeFileRotate(path, tempfile string) error {
	old := path + ".old"
	var e error

	// In Windows, one cannot rename a file if the destination already exists, at least
	// not with using the os.Rename function that Golang offers.
	// This tries to move the existing file into an old file first and only do the
	// move after that.
	if e = os.Remove(old); e != nil {
		logp.Debug("filecompare", "delete old: %v", e)
		// ignore error in case old doesn't exit yet
	}
	if e = os.Rename(path, old); e != nil {
		logp.Debug("filecompare", "rotate to old: %v", e)
		// ignore error in case path doesn't exist
	}

	if e = os.Rename(tempfile, path); e != nil {
		logp.Err("rotate: %v", e)
		return e
	}
	return nil
}
