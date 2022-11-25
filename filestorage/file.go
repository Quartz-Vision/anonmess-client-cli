package filestorage

import (
	"io/fs"
	"os"
)

const MAX_OPENED_FILES = 32

type File struct {
	path          string
	flag          int
	perm          fs.FileMode
	file          *os.File
	opened        bool // is the real file opened
	accessCount   uint32
	freeAccess    chan bool
	wakingChannel chan bool
}

var (
	openedFiles       = [MAX_OPENED_FILES]*File{}
	fileWakingChannel = make(chan *File, MAX_OPENED_FILES)
)

func InitFileManager() (err error) {
	go func() {
		for f := range fileWakingChannel {
			for i, file := range openedFiles {
				if file == nil {
					openedFiles[i] = f
					f.opened = true
					break

				}
				if file.suspend() {
					openedFiles[i] = f
					f.opened = true
					break
				}
			}
			f.wakingChannel <- true
		}
	}()

	return nil
}

func OpenFile(path string, flag int, perm fs.FileMode) (file *File) {
	f := &File{
		path:          path,
		flag:          flag,
		perm:          perm,
		file:          nil,
		opened:        false,
		accessCount:   0,
		freeAccess:    make(chan bool, 1),
		wakingChannel: make(chan bool, 1),
	}
	f.freeAccess <- true

	return f
}

func (f *File) suspend() bool {
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

func (f *File) wake() (err error) {
	fileWakingChannel <- f
	<-f.wakingChannel

	f.file, err = os.OpenFile(f.path, f.flag, f.perm)
	if err != nil {
		f.opened = false
	}
	return err
}

func (f *File) ReadAt(b []byte, offset int64) (n int, err error) {
	<-f.freeAccess

	if !f.opened {
		err = f.wake()
		if err != nil {
			f.freeAccess <- true
			return n, err
		}
	}

	n, err = f.file.ReadAt(b, offset)
	f.freeAccess <- true
	return n, err
}

func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	<-f.freeAccess

	if !f.opened {
		err = f.wake()
		if err != nil {
			f.freeAccess <- true
			return ret, err
		}
	}

	ret, err = f.file.Seek(offset, whence)
	f.freeAccess <- true
	return ret, err
}

func (f *File) Write(b []byte) (n int, err error) {
	<-f.freeAccess

	if !f.opened {
		err = f.wake()
		if err != nil {
			f.freeAccess <- true
			return n, err
		}
	}

	n, err = f.file.Write(b)
	f.freeAccess <- true
	return n, err
}

func (f *File) Close() {
	<-f.freeAccess
	if f.opened {
		f.file.Close()
		f.opened = false
	}
	f.freeAccess <- true
}
