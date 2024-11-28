/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
    "fmt"
    "os"
    "strings"
    "syscall"

    "github.com/sochoa/go-ls/internal/stat"
    "github.com/spf13/cobra"
)

var (
    listLong   bool
    jsonPretty bool
    rootCmd    = &cobra.Command{
        Use: "ls",
        RunE: func(cmd *cobra.Command, args []string) error {
            if len(args) == 0 {
                args = []string{os.Getenv("PWD")}
            }
            count := 0
            argCount := len(args)
            if argCount > 1 {
                fmt.Printf("[")
            }
            for _, arg := range args {
                arg = strings.TrimSpace(arg)
                if arg == "" {
                    continue
                }
                if count > 0 {
                    fmt.Printf(",")
                }
                var s syscall.Stat_t
                err := syscall.Stat(arg, &s)
                if err != nil {
                    continue // does not exist
                }
                m := stat.New(arg, &s)
                if listLong {
                    jsonStr, err := m.Json(jsonPretty)
                    if err != nil {
                        fmt.Fprintf(os.Stderr, "error: %v\n", err)
                    } else {
                        fmt.Printf(jsonStr)
                    }
                } else {
                    fmt.Println(m.AbsolutePath)
                }
                count++
            }
            if argCount > 1 {
                fmt.Printf("]")
            }
            fmt.Println()
            return nil
        },
    }
)

func Execute() {
    err := rootCmd.Execute()
    if err != nil {
        os.Exit(1)
    }
}

func init() {
    rootCmd.Flags().BoolVarP(&listLong, "long", "l", false,
        "use a long listing format")
    rootCmd.Flags().BoolVarP(&jsonPretty, "json", "j", false, "use json output")
}
