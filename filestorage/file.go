package filestorage

import (
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path"
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
	freeAccess    chan bool
	wakingChannel chan bool
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
			if fileSlots[i].suspend() {
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
	fmt.Println(openedFiles)

	f := &ManagedFile{
		Path:          filePath,
		flag:          flag,
		perm:          perm,
		pos:           0,
		size:          -1,
		file:          nil,
		opened:        false,
		freeAccess:    make(chan bool, 1),
		wakingChannel: make(chan bool, 1),
	}
	f.freeAccess <- true

	return f, os.MkdirAll(path.Dir(filePath), os.ModePerm)
}

// Tries to close the file. Returns true if the file is closed eventually
func (f *ManagedFile) suspend() bool {
	select {
	case <-f.freeAccess:
		if !f.opened {
			f.freeAccess <- true
			return true
		}
		f.file.Close()
		f.opened = false
		f.freeAccess <- true
		return true
	default:
		return false
	}
}

// Waits for a free place in openedFiles and then opens the fail
func (f *ManagedFile) wake() (err error) {
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
	return err
}

func (f *ManagedFile) ReadAt(b []byte, offset int64) (nRead int64, err error) {
	<-f.freeAccess

	if !f.opened {
		if err = f.wake(); err != nil {
			f.freeAccess <- true
			return 0, err
		}
	}

	if offset < 0 {
		offset += f.size + 1
	}

	n, err := f.file.ReadAt(b, offset)
	f.freeAccess <- true
	return int64(n), err
}

func (f *ManagedFile) Seek(offset int64, whence int) (ret int64, err error) {
	<-f.freeAccess

	if (whence == io.SeekStart && offset == f.pos) || (whence == io.SeekEnd && (f.size-offset) == f.pos) {
		pos := f.pos
		f.freeAccess <- true
		return pos, nil
	}

	if !f.opened {
		if err = f.wake(); err != nil {
			f.freeAccess <- true
			return ret, err
		}
	}

	ret, err = f.file.Seek(offset, whence)
	f.freeAccess <- true
	return ret, err
}

func (f *ManagedFile) Write(b []byte) (nWritten int64, err error) {
	<-f.freeAccess

	if !f.opened {
		if err = f.wake(); err != nil {
			f.freeAccess <- true
			return 0, err
		}
	}

	n, err := f.file.Write(b)
	newSize := f.pos + int64(n)
	if newSize > f.size {
		f.size = newSize
	}
	f.freeAccess <- true
	return int64(n), err
}

func (f *ManagedFile) WriteAt(b []byte, offset int64) (nWritten int64, err error) {
	<-f.freeAccess

	if !f.opened {
		if err = f.wake(); err != nil {
			f.freeAccess <- true
			return 0, err
		}
	}

	if offset < 0 {
		offset += f.size + 1
	}

	n, err := f.file.WriteAt(b, offset)
	if err != nil {
		f.freeAccess <- true
		return 0, err
	}
	newSize := offset + int64(n)
	if newSize > f.size {
		f.size = newSize
	}

	f.freeAccess <- true
	return int64(n), err
}

func (f *ManagedFile) Append(data []byte) (pos int64, err error) {
	_, err = f.WriteAt(data, f.size)
	return f.size - int64(len(data)), err
}

func (f *ManagedFile) Size() (length int64, err error) {
	if f.size == -1 {
		<-f.freeAccess
		err = f.wake()
		f.freeAccess <- true
	}
	return f.size, err
}

func (f *ManagedFile) Close() {
	<-f.freeAccess
	if f.opened {
		f.file.Close()
		f.opened = false
	}
	f.freeAccess <- true
}
