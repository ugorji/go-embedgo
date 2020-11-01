// Copyright (c) 2012-2020 Ugorji Nwoke. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

// embedgo will take a set of files which may be directories, zip-based archives (.zip, .jar, etc) or other files.
// If a zip-based archive, it will treat it as a directory tree.
//
// For any files in there whose (relative) path matches a regex (and doesn't match a regex if defined),
// it's contents will be written out into a go file as a string, and accessible via a MemFS.
//
// Usage:
//   embedgo [-match regex] [-notmatch regex] -n file...
//   embedgo [-match regex] [-notmatch regex] [-out output.go] [-prefix xyz] [-pkg pkg] [-tags tags] file...
package main

