package lock

import (
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/logx"
)

type Pair struct {
	Current string
	Next    string
}

// Keeper manages commit hashes.
type Keeper interface {
	Pair() *Pair
	// Commit writes next hash.
	Commit() error
	// Rollback writes current hash.
	Rollback() error
}

func NewFileKeeper(path filepathx.FilePath) *FileKeeper {
	k := &FileKeeper{
		path: path,
	}
	current, err := path.Read()
	logx.Debug("keeper new",
		logx.S("path", path.String()),
		logx.S("current", current),
		logx.Err(err),
	)
	if err == nil {
		k.pair.Current = current
	}
	return k
}

type FileKeeper struct {
	pair Pair
	path filepathx.FilePath
}

func (f *FileKeeper) Pair() *Pair {
	return &f.pair
}

func (f *FileKeeper) Commit() error {
	logx.Debug("keeper commit",
		logx.S("next", f.pair.Next),
	)
	if f.pair.Next == "" {
		return nil
	}

	if err := f.path.Write(f.pair.Next); err != nil {
		return errorx.Errorf(err, "commit %s into %s", f.pair.Next, f.path)
	}
	return nil
}

func (f *FileKeeper) Rollback() error {
	logx.Debug("keeper rollback",
		logx.S("current", f.pair.Current),
	)
	if f.pair.Current == "" {
		return nil
	}

	if err := f.path.Write(f.pair.Current); err != nil {
		return errorx.Errorf(err, "rollback %s into %s", f.pair.Current, f.path)
	}
	return nil
}
