package file

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	gorepo "github.com/kgoins/go-repo"
	"github.com/kgoins/go-repo/codecs"
)

type Options struct {
	// Directory must be an absolute filepath
	Directory string

	// FilenameExtension is optional
	FilenameExtension string
}

func NewOptions(dir, ext string) (Options, error) {
	fullDir, err := filepath.Abs(dir)
	if err != nil {
		return Options{}, err
	}

	return Options{
		fullDir, ext,
	}, nil
}

// File system based Repo implementation
// Note: operates on absolute paths only
type FileRepo[T gorepo.Identifiable] struct {
	// For locking the locks map
	locksLock *sync.Mutex
	// For locking file access.
	fileLocks map[string]*sync.RWMutex

	opts  Options
	codec codecs.Codec[T]
}

// Ensure FileRepo implements Repo
var _ gorepo.Repo[gorepo.Identifiable] = &FileRepo[gorepo.Identifiable]{}

func NewRepo[T gorepo.Identifiable](
	options Options,
	c ...codecs.Codec[T],
) (*FileRepo[T], error) {

	if options.Directory == "" {
		return nil, errors.New("directory cannot be empty")
	}

	codec := codecs.NewDefaultCodec[T]()
	if len(c) > 0 {
		codec = c[0]
	}

	err := os.MkdirAll(options.Directory, 0700)
	if err != nil {
		return nil, err
	}

	return &FileRepo[T]{
		locksLock: new(sync.Mutex),
		fileLocks: make(map[string]*sync.RWMutex),

		opts:  options,
		codec: codec,
	}, nil
}

func (s FileRepo[T]) keyToFilepath(key string) string {
	filename := url.PathEscape(key)
	if s.opts.FilenameExtension != "" {
		filename = filename + "." + s.opts.FilenameExtension
	}

	return filepath.Clean(s.opts.Directory + "/" + filename)
}

func (s FileRepo[T]) stripExtension(key string) string {
	if s.opts.FilenameExtension == "" {
		return key
	}

	return strings.TrimRight(
		key,
		"."+s.opts.FilenameExtension,
	)
}

func (s FileRepo[T]) listKeys() ([]string, error) {
	keys := []string{}

	err := filepath.WalkDir(
		s.opts.Directory,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if path == s.opts.Directory {
				return nil
			}

			if d.IsDir() {
				return filepath.SkipDir
			}

			key := s.stripExtension(d.Name())
			keys = append(keys, key)

			return nil
		},
	)

	return keys, err
}

func (s FileRepo[T]) Count() (int64, error) {
	keys, err := s.listKeys()
	if err != nil {
		return 0, err
	}

	return int64(len(keys)), nil
}

func (s FileRepo[T]) GetAll() ([]T, error) {
	keys, err := s.listKeys()
	if err != nil {
		return nil, err
	}

	all := make([]T, 0, len(keys))
	for _, key := range keys {
		entry, found, err := s.Get(key)
		if !found {
			continue
		}

		if err != nil {
			return nil, err
		}

		all = append(all, entry)
	}

	return all, nil
}

func (s FileRepo[T]) Add(entry T) error {
	if len(entry.GetID()) == 0 {
		return errors.New("empty id not allowed")
	}

	data, err := s.codec.Marshal(entry)
	if err != nil {
		return err
	}

	filePath := s.keyToFilepath(entry.GetID())

	lock := s.aquireLock(entry.GetID())
	lock.Lock()
	defer lock.Unlock()

	return ioutil.WriteFile(filePath, data, 0600)
}

func (s FileRepo[T]) Get(id string) (entry T, found bool, err error) {
	lock := s.aquireLock(id)
	filePath := s.keyToFilepath(id)

	lock.RLock()
	data, err := ioutil.ReadFile(filePath)
	lock.RUnlock()

	if err != nil {
		if os.IsNotExist(err) {
			found = false
			err = nil
			return
		}
		return
	}

	found = true
	entry, err = s.codec.Unmarshal(data)

	return
}

func (s FileRepo[T]) Remove(id string) error {
	lock := s.aquireLock(id)
	filePath := s.keyToFilepath(id)

	lock.Lock()
	defer lock.Unlock()

	err := os.Remove(filePath)
	if os.IsNotExist(err) {
		return nil
	}

	return err
}

func (s FileRepo[T]) Close() error {
	s.fileLocks = nil
	return nil
}

// aquireLock returns an existing file lock or creates a new one
// no two goroutines may create a lock for a filename that doesn't have a lock yet
func (s FileRepo[T]) aquireLock(key string) *sync.RWMutex {
	s.locksLock.Lock()

	lock, found := s.fileLocks[key]
	if !found {
		lock = new(sync.RWMutex)
		s.fileLocks[key] = lock
	}

	s.locksLock.Unlock()
	return lock
}
