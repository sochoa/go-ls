package perm

import "testing"

func TestNew(t *testing.T) {
	const (
		Read    = 4 // Binary: 100
		Write   = 2 // Binary: 010
		Execute = 1 // Binary: 001
	)
	tests := []struct {
		name     string
		mode     uint8
		expected SymbolicPermission
	}{
		{"No permissions", 0, SymbolicPermission{Read: false, Write: false, Execute: false}},
		{"Execute only", Execute, SymbolicPermission{Read: false, Write: false, Execute: true}},
		{"Write only", Write, SymbolicPermission{Read: false, Write: true, Execute: false}},
		{"Write and execute", Write + Execute, SymbolicPermission{Read: false, Write: true, Execute: true}},
		{"Read only", Read, SymbolicPermission{Read: true, Write: false, Execute: false}},
		{"Read and execute", Read + Execute, SymbolicPermission{Read: true, Write: false, Execute: true}},
		{"Read and write", Read + Write, SymbolicPermission{Read: true, Write: true, Execute: false}},
		{"Full permissions", Read + Write + Execute, SymbolicPermission{Read: true, Write: true, Execute: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := New(tt.mode)
			if result != tt.expected {
				t.Errorf("New(%d): got %v, want %v", tt.mode, result, tt.expected)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		input    SymbolicPermission
		expected string
	}{
		{"No permissions", SymbolicPermission{Read: false, Write: false, Execute: false}, "---"},
		{"Execute only", SymbolicPermission{Read: false, Write: false, Execute: true}, "--x"},
		{"Write only", SymbolicPermission{Read: false, Write: true, Execute: false}, "-w-"},
		{"Write and execute", SymbolicPermission{Read: false, Write: true, Execute: true}, "-wx"},
		{"Read only", SymbolicPermission{Read: true, Write: false, Execute: false}, "r--"},
		{"Read and execute", SymbolicPermission{Read: true, Write: false, Execute: true}, "r-x"},
		{"Read and write", SymbolicPermission{Read: true, Write: true, Execute: false}, "rw-"},
		{"Full permissions", SymbolicPermission{Read: true, Write: true, Execute: true}, "rwx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("String(): got %q, want %q", result, tt.expected)
			}
		})
	}
}
