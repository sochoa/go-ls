## Introduction

Hey all.  I was recently asked this question:  

> describe what happens when `ls -l *` is run on the Linux command line

My description was very wrong.  But, now that I've got some time and room to learn a bit more, I've taken a few minutes to write up the actual answer.  

## Listing in Bash on Linux

### The _Actual_ Output

Here's an example of the output:  

```
➜  pwd $ ls -l *
-rw-r--r--@ 1 user  group    0 Month 01 14:05 LICENSE
-rw-r--r--@ 1 user  group  202 Month 01 14:05 go.mod
-rw-r--r--@ 1 user  group  896 Month 01 14:05 go.sum
-rw-r--r--@ 1 user  group  138 Month 01 14:05 main.go

cmd:
total 8
-rw-r--r--@ 1 user  group  1444 Month 01 14:05 root.go

internal:
total 16
-rw-r--r--@ 1 user  group   37 Month 01 16:05 main.go
-rw-r--r--@ 1 user  group  225 Month 01 14:10 main_test.go
```

Here are some things we see so far:  

1. It shows the current working directory and all its non-hidden contents.  
2. It shows permissions of the files in symbolic notation (as opposed to file mode notation like 0755).  These permissions include owner, group, and other permissions (in that order) along with `@` indicating extended attributes like for SELinux.  
3. It shows the total number of child elements below each directory child below the current working directory (pwd).  
4. This listing shows the user and group indicated on the child (file or dir)
5. The last mod date/time for the file.  

### How does `ls` get this information?

* `stat` - for getting file and directory info for each item listed, even with symbolic links and not using the `-l` (long-listing) option.  
* `lstat` - for getting metadata information for when using `-l` (long-listing) on symbolic link file types.

## Code Sample

In an attempt to implement `ls` with `-l` support in go, [here](https://github.com/sochoa/go-ls/blob/main) is a code sample.

### What were the difficult parts?

* **Mode Parsing** - Here's the `stat_t` [mode parsing](https://github.com/sochoa/go-ls/blob/main/internal/stat/stat.go#L82-L98) section.  The hardest part here was to find the right order of parsing.  Some of these file types are also other file types.    

* **User Lookup** - It turns out that some user lookups fail.  That was annoying, so I had to fail over to "unknown".  

* **Permissions** - Parsing permissions so that it was understandable in the application layer was an interesting challenge.  Mostly it was difficult to explain through well-named variable names the concept of  single integer values being used to represent a series of values.  This is interesting as a single integer can be converted to binary and be used to indicate enum values within range of the original value's binary equivilent.  There were two operations that I ended up using to read the binary values: 

1. bit shifting to the right (`>>`) - This removed the parts of the binary value of the original int so that the resulting shifted value can be `&`-ed with the offset of the part of the original value that we are looking for.  
2. bit is set via bitmask (`&`) - When the value is shifted to the right so that the resulting value when `v & offset == offset`.  

* **Symbolic Link De-referencing** - The interesting part here was checking for link loops and then having a test confirming successful detection.  

## Conclusion

In all, this experiment was very informative.  Learning how to parse integer values with bit-wise offsets and then to apply bit-mask in order to determine if a value is set is an interesting and efficient technique.  
