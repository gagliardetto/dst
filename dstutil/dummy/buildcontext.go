package dummy

import (
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-billy.v4/memfs"
)

func BuildContext(src Item) (bc *build.Context) {

	fs := memfs.New()
	src.create("/gopath/", "src", fs)

	// This is from build.hasSubdir - which reports if dir is within root by performing
	// lexical analysis only.
	hasSubDir := func(root, dir string) (rel string, ok bool) {
		const sep = string(filepath.Separator)
		root = filepath.Clean(root)
		if !strings.HasSuffix(root, sep) {
			root += sep
		}
		dir = filepath.Clean(dir)
		if !strings.HasPrefix(dir, root) {
			return "", false
		}
		return filepath.ToSlash(dir[len(root):]), true
	}

	goroot := "/goroot/"
	gopath := "/gopath/"

	bc = &build.Context{
		GOARCH:        build.Default.GOARCH,
		GOOS:          build.Default.GOOS,
		GOROOT:        goroot,
		GOPATH:        gopath,
		CgoEnabled:    build.Default.CgoEnabled,
		UseAllFiles:   build.Default.UseAllFiles,
		Compiler:      build.Default.Compiler,
		BuildTags:     build.Default.BuildTags,
		ReleaseTags:   build.Default.ReleaseTags,
		InstallSuffix: build.Default.InstallSuffix,

		// By default, Import uses the operating system's file system calls
		// to read directories and files. To read from other sources,
		// callers can set the following functions. They all have default
		// behaviors that use the local file system, so clients need only set
		// the functions whose behaviors they wish to change.

		// JoinPath joins the sequence of path fragments into a single path.
		// If JoinPath is nil, Import uses filepath.Join.
		JoinPath: filepath.Join,

		// SplitPathList splits the path list into a slice of individual paths.
		// If SplitPathList is nil, Import uses filepath.SplitList.
		SplitPathList: filepath.SplitList,

		// IsAbsPath reports whether path is an absolute path.
		// If IsAbsPath is nil, Import uses filepath.IsAbs.
		IsAbsPath: filepath.IsAbs,

		// IsDir reports whether the path names a directory.
		// If IsDir is nil, Import calls os.Stat and uses the result's IsDir method.
		IsDir: func(name string) bool {
			if _, ok := hasSubDir(goroot, name); ok {
				fi, err := os.Stat(name)
				return err == nil && fi.IsDir()
			}
			info, err := fs.Lstat(name)
			if err != nil {
				return false
			}
			return info.IsDir()
		},

		// HasSubdir reports whether dir is lexically a subdirectory of
		// root, perhaps multiple levels below. It does not try to check
		// whether dir exists.
		// If so, HasSubdir sets rel to a slash-separated path that
		// can be joined to root to produce a path equivalent to dir.
		// If HasSubdir is nil, Import uses an implementation built on
		// filepath.EvalSymlinks.
		HasSubdir: hasSubDir,

		// ReadDir returns a slice of os.FileInfo, sorted by Name,
		// describing the content of the named directory.
		// If ReadDir is nil, Import uses ioutil.ReadDir.
		ReadDir: func(name string) (fi []os.FileInfo, err error) {
			if _, ok := hasSubDir(goroot, name); ok {
				return ioutil.ReadDir(name)
			}
			return fs.ReadDir(name)
		},

		// OpenFile opens a file (not a directory) for reading.
		// If OpenFile is nil, Import uses os.Open.
		OpenFile: func(path string) (io.ReadCloser, error) {
			if _, ok := hasSubDir(goroot, path); ok {
				return os.Open(path)
			}
			return fs.Open(path)
		},
	}
	return bc
}
