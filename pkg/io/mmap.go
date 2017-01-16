package io

import (
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

// Mmap maps file and supports many useful methods
type Mmap struct {
	MmapBytes   []byte
	FileName    string
	FileLen     int64
	FilePointer int64
	MapType     int64
	FileHandler *os.File
}

const preAllocatedSpace int64 = 1024 * 1024

const (
	// ModeAppend is append mode for mmap
	ModeAppend = iota
	// ModeCreate is create mode for mmap
	ModeCreate
)

// NewMmap initializes mmap struct
func NewMmap(filename string, mode int) (*Mmap, error) {
	m := &Mmap{
		MmapBytes: make([]byte, 0),
		FileName:  filename,
	}

	fileMode := os.O_RDWR
	fileCreateMode := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	if mode == ModeCreate {
		fileMode = fileCreateMode
	}
	file, err := os.OpenFile(filename, fileMode, 0664)
	if err != nil {
		return nil, err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	m.FileLen = fileInfo.Size()
	if mode == ModeCreate || m.FileLen == 0 {
		syscall.Ftruncate(int(file.Fd()), m.FileLen+preAllocatedSpace)
		m.FileLen = m.FileLen + preAllocatedSpace
	}
	m.MmapBytes, err = syscall.Mmap(int(file.Fd()), 0, int(m.FileLen), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, errors.Wrap(err, "mapping file error")
	}
	m.FileHandler = file
	return m, nil
}

// Unmap cancels mmap
func (m *Mmap) Unmap() error {
	err := syscall.Munmap(m.MmapBytes)
	if err != nil {
		return errors.Wrap(err, "unmap error")
	}
	if err = m.FileHandler.Close(); err != nil {
		return errors.Wrap(err, "close file error")
	}
	return nil
}

// Sync flushes data into disk
func (m *Mmap) Sync() error {
	h := m.header()
	_, _, err := syscall.Syscall(syscall.SYS_MSYNC, h.Data, uintptr(h.Len), syscall.MS_SYNC)
	if err != 0 {
		return errors.Wrap(err, "sync error")
	}
	return nil
}

func (m *Mmap) header() *reflect.SliceHeader {
	return (*reflect.SliceHeader)(unsafe.Pointer(&m.MmapBytes))
}

// GetPointer returns the current pointer location which indicates file content so far
func (m *Mmap) GetPointer() int64 {
	return m.FilePointer
}

// SetFileEnd sets end location of file
func (m *Mmap) SetFileEnd(fileLen int64) {
	m.FilePointer = fileLen
}

// AppendBytes appends bytes and return file pointer location
func (m *Mmap) AppendBytes(val []byte) uint64 {
	offset := uint64(m.FilePointer)
	valLen := int64(len(val))
	if err := m.checkFilePointerOutOfRange(valLen); err != nil {
		return 0
	}

	dst := m.MmapBytes[m.FilePointer : m.FilePointer+valLen]
	copy(dst, val)
	m.FilePointer += valLen
	return offset
}

// AppendStringWithLen appends string into file
func (m *Mmap) AppendStringWithLen(val string) error {
	if err := m.AppendInt64(int64(len(val))); err != nil {
		return err
	}
	if err := m.appendString(val); err != nil {
		return err
	}
	return nil
}

// ReadInt64 reads int64 number
func (m *Mmap) ReadInt64(start int64) int64 {
	return int64(binary.LittleEndian.Uint64(m.MmapBytes[start : start+8]))
}

// WriteInt64 writes int64 number into file
func (m *Mmap) WriteInt64(start, val int64) error {
	binary.LittleEndian.PutUint64(m.MmapBytes[start:start+8], uint64(val))
	return nil
}

// AppendInt64 appends int64 number into file and updates file pointer location
func (m *Mmap) AppendInt64(val int64) error {
	if err := m.checkFilePointerOutOfRange(8); err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(m.MmapBytes[m.FilePointer:m.FilePointer+8], uint64(val))
	m.FilePointer += 8
	return nil
}

// ReadUint64 reads uint64 number
func (m *Mmap) ReadUint64(start uint64) uint64 {
	return binary.LittleEndian.Uint64(m.MmapBytes[start : start+8])
}

// WriteUint64 writes uint64 number into file
func (m *Mmap) WriteUint64(start int64, val uint64) error {
	binary.LittleEndian.PutUint64(m.MmapBytes[start:start+8], val)
	return nil
}

// AppendUint64 appends uint64 number into file and updates file pointer location
func (m *Mmap) AppendUint64(val uint64) error {
	if err := m.checkFilePointerOutOfRange(8); err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(m.MmapBytes[m.FilePointer:m.FilePointer+8], val)
	m.FilePointer += 8
	return nil
}

// ReadStringWithLen returns string
func (m *Mmap) ReadStringWithLen(start uint64) string {
	valLen := m.ReadInt64(int64(start))
	return m.readString(int64(start+8), valLen)
}

func (m *Mmap) readString(start, valLen int64) string {
	return string(m.MmapBytes[start : start+valLen])
}

func (m *Mmap) appendString(val string) error {
	valLen := int64(len(val))
	if err := m.checkFilePointerOutOfRange(valLen); err != nil {
		return err
	}
	dst := m.MmapBytes[m.FilePointer : m.FilePointer+valLen]
	copy(dst, []byte(val))
	m.FilePointer += valLen
	return nil
}

// checkFilePointerOutOfRange checks if new data to be appended exceeds file length
func (m *Mmap) checkFilePointerOutOfRange(valLen int64) error {
	if m.FilePointer+valLen >= m.FileLen {
		err := syscall.Ftruncate(int(m.FileHandler.Fd()), m.FileLen+preAllocatedSpace)
		if err != nil {
			return errors.Wrap(err, "truncate file error")
		}
		m.FileLen += preAllocatedSpace
		syscall.Munmap(m.MmapBytes)
		m.MmapBytes, err = syscall.Mmap(int(m.FileHandler.Fd()), 0, int(m.FileLen), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
		if err != nil {
			return errors.Wrap(err, "mmaping file error")
		}
	}
	return nil
}

func (m *Mmap) String() string {
	return fmt.Sprintf("mmap - name: %s, len: %d, ptr: %d, type: %d", m.FileName, m.FileLen, m.FilePointer, m.MapType)
}
