package perm

type SymbolicPermission struct {
    Read    bool
    Write   bool
    Execute bool
}

func (p SymbolicPermission) String() string {
    var str string
    ternary := func(b bool, t, f string) string {
        if b {
            return t
        }
        return f
    }
    str += ternary(p.Read, "r", "-")
    str += ternary(p.Write, "w", "-")
    str += ternary(p.Execute, "x", "-")
    return str
}

func New(mode uint8) SymbolicPermission {
    var (
        read    bool
        write   bool
        execute bool
    )
    const (
        readOffset    = 4
        writeOffset   = 2
        executeOffset = 1
    )

    // bitIsSet determines whether a specific bit is set in an unsigned int.
    // Example:
    //   mode := uint8(5) // Binary: 101
    //   bitIsSet(mode, 4) // Returns true (read bit is set)
    //      0b101 & 0b100 = 0b100 (comparison: 0b100 == 0b100 → true)
    //   bitIsSet(mode, 2) // Returns false (write bit is not set)
    //      0b101 & 0b010 = 0b000 (comparison: 0b000 == 0b010 → false)
    //   bitIsSet(mode, 1) // Returns true (execute bit is set)
    //      0b101 & 0b001 = 0b001 (comparison: 0b001 == 0b001 → true)
    bitIsSet := func(mode, offset uint8) bool {
        return mode&offset == offset
    }
    read = bitIsSet(mode, readOffset)
    write = bitIsSet(mode, writeOffset)
    execute = bitIsSet(mode, executeOffset)

    return SymbolicPermission{
        Read:    read,
        Write:   write,
        Execute: execute,
    }
}
