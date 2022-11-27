package filestorage

import (
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"sync"
)

const (
	MAX_OPENED_FILES      = 32 // MAX_OPENED_FILES % WAKING_WORKERS_NUMBER = 0
	WAKING_WORKERS_NUMBER = 4
)

// A file that persists its state between closes.
// Allows to open more files than available in the system for a process
type ManagedFile struct {
	Path          string
	flag          int
	perm          fs.FileMode
	pos           int64 // current seek position in the file
	size          int64
	file          *os.File
	opened        bool // is the real file opened
	wakingChannel chan bool
	rwmutex       sync.RWMutex
	wakingMutex   sync.Mutex
}

var (
	openedFiles         = [MAX_OPENED_FILES]*ManagedFile{}
	globalWakingChannel = make(chan *ManagedFile, MAX_OPENED_FILES)
)

// manages some set of file slots for opened files and tries to wake up suspended files if possible
func startWakingManager(fileSlots []*ManagedFile) {
	slotsLen := len(fileSlots)
root:
	for f := range globalWakingChannel {
		for i, tfile := range fileSlots {
			if tfile == nil {
				fileSlots[i] = f
				f.wakingChannel <- true
				continue root
			}
		}
		for _, i := range rand.Perm(slotsLen) {
			if fileSlots[i].trySuspend() {
				fileSlots[i] = f
				f.wakingChannel <- true
				continue root
			}
		}
		f.wakingChannel <- false
	}
}

func InitFileManager() (err error) {
	slotsPerWorker := MAX_OPENED_FILES / WAKING_WORKERS_NUMBER
	for i := 0; i < WAKING_WORKERS_NUMBER; i++ {
		go startWakingManager(openedFiles[i*slotsPerWorker : (i+1)*slotsPerWorker])
	}

	return nil
}

func NewFile(filePath string, flag int, perm fs.FileMode) (file *ManagedFile, err error) {
	f := &ManagedFile{
		Path:          filePath,
		flag:          flag,
		perm:          perm,
		pos:           0,
		size:          -1,
		file:          nil,
		opened:        false,
		rwmutex:       sync.RWMutex{},
		wakingMutex:   sync.Mutex{},
		wakingChannel: make(chan bool, 1),
	}

	return f, os.MkdirAll(path.Dir(filePath), os.ModePerm)
}

// Tries to close the file. Returns true if the file is closed eventually
func (f *ManagedFile) trySuspend() bool {
	if !f.rwmutex.TryLock() {
		return false
	}

	if f.opened {
		f.file.Close()
	}
	f.opened = false
	f.rwmutex.Unlock()
	return true
}

// Waits for a free place in openedFiles and then opens the fail
func (f *ManagedFile) wake() (err error) {
	f.wakingMutex.Lock()

	if f.opened {
		f.wakingMutex.Unlock()
		return nil
	}

	for !f.opened {
		globalWakingChannel <- f
		f.opened = <-f.wakingChannel
	}

	f.file, err = os.OpenFile(f.Path, f.flag, f.perm)
	if err != nil {
		f.opened = false
	} else {
		if f.pos != 0 {
			f.file.Seek(f.pos, io.SeekStart)
		}
		if f.size == -1 {
			f.size, _ = f.file.Seek(0, io.SeekEnd)
		}
	}

	f.wakingMutex.Unlock()
	return err
}

func (f *ManagedFile) ReadAt(b []byte, offset int64) (nRead int64, err error) {
	f.rwmutex.RLock()

	if !f.opened { // don't do the expesive call with mutex
		if err = f.wake(); err != nil {
			f.rwmutex.RUnlock()
			return 0, err
		}
	}

	if offset < 0 {
		offset += f.size + 1
	}

	n, err := f.file.ReadAt(b, offset)

	f.rwmutex.RUnlock()
	return int64(n), err
}

func (f *ManagedFile) Seek(offset int64, whence int) (ret int64, err error) {
	f.rwmutex.Lock()

	if (whence == io.SeekStart && offset == f.pos) || (whence == io.SeekEnd && (f.size-offset) == f.pos) {
		pos := f.pos
		f.rwmutex.Unlock()
		return pos, nil
	}

	if !f.opened {
		if err = f.wake(); err != nil {
			f.rwmutex.Unlock()
			return ret, err
		}
	}

	ret, err = f.file.Seek(offset, whence)
	f.rwmutex.Unlock()
	return ret, err
}

func (f *ManagedFile) Write(b []byte) (nWritten int64, err error) {
	f.rwmutex.Lock()

	if !f.opened {
		if err = f.wake(); err != nil {
			f.rwmutex.Unlock()
			return 0, err
		}
	}

	n, err := f.file.Write(b)
	newSize := f.pos + int64(n)
	if newSize > f.size {
		f.size = newSize
	}
	f.rwmutex.Unlock()
	return int64(n), err
}

func (f *ManagedFile) WriteAt(b []byte, offset int64) (nWritten int64, err error) {
	f.rwmutex.Lock()

	if !f.opened {
		if err = f.wake(); err != nil {
			f.rwmutex.Unlock()
			return 0, err
		}
	}

	if offset < 0 {
		offset += f.size + 1
	}

	n, err := f.file.WriteAt(b, offset)
	newSize := offset + int64(n)
	if err != nil {
		f.rwmutex.Unlock()
		return 0, err
	}
	if newSize > f.size {
		f.size = newSize
	}

	f.rwmutex.Unlock()
	return int64(n), err
}

func (f *ManagedFile) Append(data []byte) (pos int64, err error) {
	_, err = f.WriteAt(data, f.size)
	return f.size - int64(len(data)), err
}

func (f *ManagedFile) Size() (length int64, err error) {
	if f.size == -1 {
		err = f.wake()
	}
	return f.size, err
}

func (f *ManagedFile) Close() {
	f.rwmutex.Lock()
	if f.opened {
		f.file.Close()
		f.opened = false
	}
	f.rwmutex.Unlock()
}
