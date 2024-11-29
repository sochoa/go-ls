package stat

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"syscall"
	"testing"
)

func TestNewLinkWithDepsSingleSymlink(t *testing.T) {
	j := Stat{
		Type:         SymbolicLinkFileType,
		AbsolutePath: "/path/to/symlink",
	}
	mockLstat := func(path string, stat *syscall.Stat_t) error {
		switch path {
		case "/path/to/symlink":
			stat.Mode = syscall.S_IFLNK
			return nil
		case "/path/to/target":
			stat.Mode = syscall.S_IFREG // Regular file to indicate the symlink resolves to a file
			return nil
		default:
			return fmt.Errorf("unexpected lstat call")
		}
	}
	mockReadlink := func(path string) (string, error) {
		if path == "/path/to/symlink" {
			return "/path/to/target", nil
		}
		return "", fmt.Errorf("unexpected readlink call")
	}
	mockIsAbs := func(path string) bool { return true }
	mockFilepathJoin := filepath.Join
	mockFilepathDir := filepath.Dir

	result, err := NewLinkWithDeps(j, mockReadlink, mockLstat, mockIsAbs, mockFilepathJoin, mockFilepathDir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, []string{"/path/to/symlink", "/path/to/target"}, result.Targets)
}

func TestNewLinkWithDepsChainSymlink(t *testing.T) {
	j := Stat{
		Type:         SymbolicLinkFileType,
		AbsolutePath: "/path/to/symlink1",
	}
	mockLstat := func(path string, stat *syscall.Stat_t) error {
		switch path {
		case "/path/to/symlink1", "/path/to/symlink2":
			stat.Mode = syscall.S_IFLNK
			return nil
		case "/path/to/target":
			stat.Mode = syscall.S_IFREG
			return nil
		default:
			return fmt.Errorf("unexpected lstat call")
		}
	}
	mockReadlink := func(path string) (string, error) {
		switch path {
		case "/path/to/symlink1":
			return "/path/to/symlink2", nil
		case "/path/to/symlink2":
			return "/path/to/target", nil
		default:
			return "", fmt.Errorf("unexpected readlink call")
		}
	}
	mockIsAbs := func(path string) bool { return true }
	mockFilepathJoin := filepath.Join
	mockFilepathDir := filepath.Dir

	result, err := NewLinkWithDeps(j, mockReadlink, mockLstat, mockIsAbs, mockFilepathJoin, mockFilepathDir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, []string{"/path/to/symlink1", "/path/to/symlink2", "/path/to/target"}, result.Targets)
}

func TestNewLinkWithDepsNonSymlink(t *testing.T) {
	j := Stat{
		Type:         "regular",
		AbsolutePath: "/path/to/file",
	}
	mockLstat := func(path string, stat *syscall.Stat_t) error {
		return nil
	}
	mockReadlink := func(path string) (string, error) { return "", nil }
	mockIsAbs := func(path string) bool { return true }
	mockFilepathJoin := filepath.Join
	mockFilepathDir := filepath.Dir

	result, err := NewLinkWithDeps(j, mockReadlink, mockLstat, mockIsAbs, mockFilepathJoin, mockFilepathDir)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestNewLinkWithDepsBrokenSymlink(t *testing.T) {
	j := Stat{
		Type:         SymbolicLinkFileType,
		AbsolutePath: "/path/to/symlink",
	}
	mockLstat := func(path string, stat *syscall.Stat_t) error {
		if path == "/path/to/symlink" {
			stat.Mode = syscall.S_IFLNK
			return nil
		}
		return fmt.Errorf("file does not exist: %s", path) // Simulate broken link for "/path/to/missing"
	}
	mockReadlink := func(path string) (string, error) {
		if path == "/path/to/symlink" {
			return "/path/to/missing", nil
		}
		return "", fmt.Errorf("unexpected readlink call")
	}
	mockIsAbs := func(path string) bool { return true }
	mockFilepathJoin := filepath.Join
	mockFilepathDir := filepath.Dir

	result, err := NewLinkWithDeps(j, mockReadlink, mockLstat, mockIsAbs, mockFilepathJoin, mockFilepathDir)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "file does not exist")
}

func TestNewLinkWithDepsCyclicSymlink(t *testing.T) {
	j := Stat{
		Type:         SymbolicLinkFileType,
		AbsolutePath: "/path/to/symlink1",
	}
	mockLstat := func(path string, stat *syscall.Stat_t) error {
		stat.Mode = syscall.S_IFLNK
		return nil
	}
	mockReadlink := func(path string) (string, error) {
		if path == "/path/to/symlink1" {
			return "/path/to/symlink2", nil
		}
		if path == "/path/to/symlink2" {
			return "/path/to/symlink1", nil
		}
		return "", fmt.Errorf("unexpected readlink call")
	}
	mockIsAbs := func(path string) bool { return true }
	mockFilepathJoin := filepath.Join
	mockFilepathDir := filepath.Dir

	result, err := NewLinkWithDeps(j, mockReadlink, mockLstat, mockIsAbs, mockFilepathJoin, mockFilepathDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "symlink loop detected")
	require.Nil(t, result)
}
