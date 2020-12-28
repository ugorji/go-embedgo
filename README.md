# embedgo

This repository contains the `embedgo` command.

To install:

```
go get github.com/ugorji/go-embedgo/cmd/embedgo
```

# Package Documentation

archives (.zip, .jar, etc) or other files. If a zip-based archive, it will
treat it as a directory tree.

For any files in there whose (relative) path matches a regex (and doesn't
match a regex if defined), it's contents will be written out into a go file
as a string, and accessible via a MemFS.

Usage:

```
    embedgo [-match regex] [-notmatch regex] -n file...
    embedgo [-match regex] [-notmatch regex] [-out output.go] [-prefix xyz] [-pkg pkg] [-tags tags] file...
