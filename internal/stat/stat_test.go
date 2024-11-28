package stat

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"os/user"
	"path/filepath"
	"syscall"
	"testing"
)

func mockUserLookup(uid string) (*user.User, error) {
	if uid == "1000" {
		return &user.User{Username: "testuser"}, nil
	}
	return nil, errors.New("user not found")
}

func mockPathBasename(path string) string {
	return filepath.Base(path)
}

func mockPathAbs(path string) (string, error) {
	return filepath.Abs(path)
}

func TestNewWithDepsPositive(t *testing.T) {
	stat := &syscall.Stat_t{
		Mode:          syscall.S_IFREG | 0644,
		Size:          12345,
		Uid:           1000,
		Gid:           1000,
		Atimespec:     syscall.Timespec{Sec: 1609459200, Nsec: 0}, // 2021-01-01 00:00:00 UTC
		Mtimespec:     syscall.Timespec{Sec: 1609459300, Nsec: 0}, // 2021-01-01 00:01:40 UTC
		Ctimespec:     syscall.Timespec{Sec: 1609459400, Nsec: 0}, // 2021-01-01 00:03:20 UTC
		Birthtimespec: syscall.Timespec{Sec: 1609459500, Nsec: 0}, // 2021-01-01 00:05:00 UTC
		Blksize:       4096,
		Blocks:        12,
		Nlink:         2,
	}

	statResult := NewWithDeps("testfile.txt", stat, mockUserLookup, mockPathBasename, mockPathAbs)

	assert.Equal(t, "testfile.txt", statResult.BaseName)
	assert.Equal(t, "file", statResult.Type)
	assert.Equal(t, "testuser", statResult.Owner)
	assert.Equal(t, "testuser", statResult.UserName)
	assert.Equal(t, uint32(4096), statResult.BlockSize)
	assert.Equal(t, uint64(12), statResult.NumBlocks)
	assert.Equal(t, uint16(2), statResult.HardLinkReferenceCount)
	assert.Equal(t, "644", statResult.Permissions.Octal)
}

func TestNewWithDepsUnknownUser(t *testing.T) {
	stat := &syscall.Stat_t{
		Mode:          syscall.S_IFREG | 0755,
		Size:          54321,
		Uid:           9999,
		Gid:           9999,
		Atimespec:     syscall.Timespec{Sec: 1609459200, Nsec: 0},
		Mtimespec:     syscall.Timespec{Sec: 1609459300, Nsec: 0},
		Ctimespec:     syscall.Timespec{Sec: 1609459400, Nsec: 0},
		Birthtimespec: syscall.Timespec{Sec: 1609459500, Nsec: 0},
		Blksize:       4096,
		Blocks:        12,
		Nlink:         2,
	}

	statResult := NewWithDeps("testfile.txt", stat, mockUserLookup, mockPathBasename, mockPathAbs)

	assert.Equal(t, "testfile.txt", statResult.BaseName)
	assert.Equal(t, "file", statResult.Type)
	assert.Equal(t, "unknown", statResult.Owner)
	assert.Equal(t, "unknown", statResult.UserName)
	assert.Equal(t, "unknown", statResult.GroupName)
}

func TestNewWithDepsInvalidPath(t *testing.T) {
	stat := &syscall.Stat_t{
		Mode: syscall.S_IFDIR | 0700,
		Size: 1024,
		Uid:  1000,
		Gid:  1000,
	}

	mockPathAbsErr := func(path string) (string, error) {
		return "", errors.New("invalid path")
	}

	statResult := NewWithDeps("invalidpath", stat, mockUserLookup, mockPathBasename, mockPathAbsErr)

	assert.Equal(t, "directory", statResult.Type)
	assert.Equal(t, "unknown", statResult.AbsolutePath) // AbsolutePath should fallback to "unknown"
}

func TestNewWithDepsSpecialFileTypes(t *testing.T) {
	tests := []struct {
		mode     uint32
		fileType string
	}{
		{syscall.S_IFDIR, "directory"},
		{syscall.S_IFREG, "file"},
		{syscall.S_IFLNK, "symlink"},
		{syscall.S_IFIFO, "fifo"},
		{syscall.S_IFSOCK, "socket"},
		{syscall.S_IFCHR, "character_device"},
		{syscall.S_IFBLK, "block_device"},
	}
	for _, tt := range tests {
		t.Run(tt.fileType, func(t *testing.T) {
			stat := &syscall.Stat_t{
				Mode: uint16(tt.mode),
				Uid:  1000,
				Gid:  1000,
			}
			statResult := NewWithDeps("testfile", stat, mockUserLookup, mockPathBasename, mockPathAbs)
			assert.Equal(t, tt.fileType, statResult.Type)
		})
	}
}
