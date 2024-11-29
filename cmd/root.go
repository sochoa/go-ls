/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sochoa/go-ls/internal/stat"
	"github.com/spf13/cobra"
)

const (
	outputTypeJson = "json"
	outputTypeText = "text"
)

var (
	listLong   bool
	jsonPretty bool
	outputType string
	rootCmd    = &cobra.Command{
		Use: "ls",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				args = []string{os.Getenv("PWD")}
			}

			count := 0
			argCount := len(args)
			if argCount > 1 && outputType == outputTypeJson {
				fmt.Printf("[")
			}

			for idx := 0; idx < len(args); idx++ {
				arg := args[idx]
				arg = strings.TrimSpace(arg)
				if arg == "" {
					continue
				}

				if count > 0 && outputType == outputTypeJson {
					fmt.Printf(",")
				}

				matches, err := filepath.Glob(arg)
				if err != nil || len(matches) == 0 {
					fmt.Fprintf(os.Stderr, "No matches found for %s\n", arg)
					continue
				} else if len(matches) > 1 {
					left := args[:idx]
					right := args[idx+1:]
					args = append(append(left, matches...), right...)
				}

				for _, match := range matches {
					var s syscall.Stat_t
					err := syscall.Stat(match, &s)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error statting %s: %v\n", match, err)
						continue
					}

					// Get stats for the current item
					var m stat.CommonStat = stat.New(match, &s)
					if m.GetType() == stat.SymbolicLinkFileType {
						var l *stat.StatLink
						l, err = stat.NewLink(m)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error reading symlink %s: %v\n", match, err)
							continue
						} else if l != nil {
							m = *l
						} else {
							fmt.Fprintf(os.Stderr, "Error reading symlink %s: %v\n", match, err)
							continue
						}
					}

					if listLong {
						jsonStr, err := m.Json(jsonPretty)
						if err != nil {
							fmt.Fprintf(os.Stderr, "error: %v\n", err)
						} else {
							fmt.Printf(jsonStr)
						}
					} else {
						fmt.Println(m.GetAbsolutePath())
					}

					// If the current item is a directory, get its immediate children
					if m.GetType() == stat.DirectoryFileType {
						children, err := os.ReadDir(match)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", match, err)
							continue
						}

						for _, child := range children {
							childPath := filepath.Join(match, child.Name())
							var childStat syscall.Stat_t
							err := syscall.Stat(childPath, &childStat)
							if err != nil {
								fmt.Fprintf(os.Stderr, "Error statting %s: %v\n", childPath, err)
								continue
							}

							childMeta := stat.New(childPath, &childStat)
							if listLong {
								jsonStr, err := childMeta.Json(jsonPretty)
								if err != nil {
									fmt.Fprintf(os.Stderr, "error: %v\n", err)
								} else {
									fmt.Printf("  %s\n", jsonStr) // Indent for clarity
								}
							} else {
								fmt.Printf("%s\n", childMeta.AbsolutePath)
							}
						}
					}

					count++
				}
			}

			if argCount > 1 && outputType == outputTypeJson {
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
	rootCmd.Flags().StringVar(&outputType, "output", "text", "output type (text or json)")
}
