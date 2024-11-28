package stat

import (
	"encoding/json"
	"fmt"
	"github.com/sochoa/go-ls/internal/perm"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"syscall"
	"time"
)

type Stat struct {
	SizeBytes              int64     `json:"size_bytes"`
	Mode                   uint16    `json:"mode"`
	UserID                 uint32    `json:"user_id"`
	UserName               string    `json:"user_name"`
	GroupID                uint32    `json:"group_id"`
	GroupName              string    `json:"group_name"`
	Owner                  string    `json:"owner"`
	LastAccessedTime       time.Time `json:"last_accessed_time"`
	LastModifiedTime       time.Time `json:"last_modified_time"`
	CreateTime             time.Time `json:"create_time"`
	BirthTime              time.Time `json:"birth_time"`
	BlockSize              uint32    `json:"block_size"`
	NumBlocks              uint64    `json:"num_blocks"`
	HardLinkReferenceCount uint16    `json:"hard_link_reference_count"`
	Permissions            struct {
		Octal    string `json:"octal"`
		Symbolic struct {
			Owner perm.SymbolicPermission `json:"owner"`
			Group perm.SymbolicPermission `json:"group"`
			Other perm.SymbolicPermission `json:"other"`
		} `json:"symbolic"`
	} `json:"permissions"`
	BaseName     string `json:"basename"`
	AbsolutePath string `json:"absolute_path"`
	Type         string `json:"type"`
}

func New(n string, stat *syscall.Stat_t) Stat {
	var m Stat
	m.BaseName = path.Base(n)
	m.AbsolutePath, _ = filepath.Abs(n)
	if stat.Mode&syscall.S_IFDIR == syscall.S_IFDIR {
		m.Type = "directory"
	} else if stat.Mode&syscall.S_IFREG == syscall.S_IFREG {
		m.Type = "file"
	} else if stat.Mode&syscall.S_IFLNK == syscall.S_IFLNK {
		m.Type = "symlink"
	} else if stat.Mode&syscall.S_IFIFO == syscall.S_IFIFO {
		m.Type = "fifo"
	} else if stat.Mode&syscall.S_IFSOCK == syscall.S_IFSOCK {
		m.Type = "socket"
	} else if stat.Mode&syscall.S_IFCHR == syscall.S_IFCHR {
		m.Type = "character_device"
	} else if stat.Mode&syscall.S_IFBLK == syscall.S_IFBLK {
		m.Type = "block_device"
	} else {
		m.Type = "unknown"
	}
	m.SizeBytes = stat.Size
	m.Mode = stat.Mode
	m.UserID = stat.Uid
	m.GroupID = stat.Gid
	m.LastAccessedTime = time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec)
	m.LastModifiedTime = time.Unix(stat.Mtimespec.Sec, stat.Mtimespec.Nsec)
	m.CreateTime = time.Unix(stat.Ctimespec.Sec, stat.Ctimespec.Nsec)
	m.BirthTime = time.Unix(stat.Birthtimespec.Sec, stat.Birthtimespec.Nsec)
	m.BlockSize = uint32(stat.Blksize)
	m.NumBlocks = uint64(stat.Blocks)

	var err error
	u, err := user.LookupId(fmt.Sprintf("%d", stat.Uid))
	if err == nil {
		m.Owner = u.Username
		m.UserName = u.Username
	} else {
		m.Owner = "unknown"
		m.UserName = "unknown"
	}

	u, err = user.LookupId(fmt.Sprintf("%d", stat.Gid))
	if err == nil {
		m.GroupName = u.Username
	} else {
		m.GroupName = "unknown"
	}

	m.HardLinkReferenceCount = uint16(stat.Nlink)

	octalPerm := os.FileMode(stat.Mode) & os.ModePerm
	m.Permissions.Octal = fmt.Sprintf("%o", octalPerm)

	const (
		ownerStatTOffset = 6
		groupStatTOffset = 3
		otherStatTOffset = 1
	)

	// https://man7.org/linux/man-pages/man7/inode.7.html
	// who-has-what-perms section starts at offset 6 and ends at offset 10.

	// The >> operator in Go shifts the bits of a number to the right.
	// Think of the number as a row of lights (1 = on, 0 = off).
	// - >> 1 moves all the lights 1 step to the right, filling empty spaces on the left:
	//   - Positive numbers fill with 0 (e.g., 8 >> 1: 00001000 -> 00000100 = 4).
	//   - Negative numbers fill with 1 to keep the number negative.
	// - Bits that fall off the right edge disappear.
	var (
		ownerPerms = uint8(octalPerm >> ownerStatTOffset)
		groupPerms = uint8(octalPerm >> groupStatTOffset)
		otherPerms = uint8(octalPerm >> otherStatTOffset)
	)
	m.Permissions.Symbolic.Owner = perm.New(ownerPerms)
	m.Permissions.Symbolic.Group = perm.New(groupPerms)
	m.Permissions.Symbolic.Other = perm.New(otherPerms)
	return m
}

func (s Stat) Json(pretty bool) (string, error) {
	var (
		statBytes []byte
		err       error
	)
	if pretty {
		statBytes, _ = json.MarshalIndent(s, "", "  ")
	} else {
		statBytes, _ = json.Marshal(s)
	}
	if err != nil {
		return string(statBytes), fmt.Errorf("failed to marshal Stat type struct to json: %w", err)
	}
	return string(statBytes), nil
}
