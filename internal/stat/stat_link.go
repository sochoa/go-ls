package stat

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"syscall"
)

type StatLink struct {
	Stat
	Targets []string `json:"targets"`
}

var _ CommonStat = (*StatLink)(nil)

func (s StatLink) GetType() string {
	return s.Type
}

func (s StatLink) Json(pretty bool) (string, error) {
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
		return string(statBytes), fmt.Errorf("failed to marshal StatLink type struct to json: %w", err)
	}
	return string(statBytes), nil
}

func NewLink(j CommonStat) (*StatLink, error) {
	var s = j.(Stat)
	if s.Type != SymbolicLinkFileType {
		return nil, nil
	}
	sl := &StatLink{
		Stat:    s,
		Targets: []string{},
	}
	return NewLinkWithDeps(sl, os.Readlink, syscall.Lstat, filepath.IsAbs, filepath.Join, filepath.Dir)
}

func NewLinkWithDeps(
	j CommonStat,
	readlink func(string) (string, error),
	lstat func(string, *syscall.Stat_t) error,
	isAbs func(string) bool,
	filepathJoin func(...string) string,
	filepathDir func(string) string,
) (*StatLink, error) {
	var s = j.(Stat)
	if s.Type != SymbolicLinkFileType {
		return nil, nil
	}
	sl := &StatLink{
		Stat:    s,
		Targets: []string{},
	}
	currentPath := s.AbsolutePath
	for i := 0; i < 10; i++ {
		var linkStat syscall.Stat_t
		err := lstat(currentPath, &linkStat)
		if err != nil {
			return nil, fmt.Errorf("error lstat-ing %s: %w", currentPath, err)
		}
		if slices.Contains(sl.Targets, currentPath) {
			return nil, fmt.Errorf("symlink loop detected at %s", currentPath)
		}
		sl.Targets = append(sl.Targets, currentPath)
		if linkStat.Mode&syscall.S_IFMT != syscall.S_IFLNK {
			break
		}
		target, err := readlink(currentPath)
		if err != nil {
			return nil, fmt.Errorf("error reading symlink %s: %w", currentPath, err)
		}
		if !isAbs(target) {
			target = filepathJoin(filepathDir(currentPath), target)
		}
		currentPath = target
	}
	return sl, nil
}
