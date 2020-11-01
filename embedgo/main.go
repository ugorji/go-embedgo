// Copyright (c) 2012-2020 Ugorji Nwoke. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ugorji/go-common/flagutil"
	"github.com/ugorji/go-common/vfs"
)

const vfsPkgName = "github.com/ugorji/go-common/vfs"

func onErr(err error) {
	if err != nil {
		// zz.Debugf("error: %v", err)
		// log.Fatalf("error: %v", err)
		panic(err)
	}
}

func fileInfoToString(path string, fi vfs.FileInfo) string {
	var pat = "%s"
	if fi.IsDir() {
		pat = "*%s/ (dir)"
	}
	return fmt.Sprintf(pat+" - %d bytes - on %v", path, fi.Size(), fi.ModTime())
}

var one = []byte{0}

type writer struct {
	*os.File
}

func (x writer) b(v []byte) writer {
TOP:
	n, err := x.Write(v)
	onErr(err)
	if n == len(v) {
		return x
	}
	v = v[n:]
	goto TOP
}

func (x writer) s(v string) writer {
	return x.b([]byte(v))
}

func (x writer) n(b byte) writer {
	one[0] = b
	return x.b(one)
}

func (x writer) nl() writer {
	return x.n('\n')
}

func (x writer) f(v string, args ...interface{}) writer {
	return x.s(fmt.Sprintf(v, args...))
}

func main() {
	// create a Vfs will all the inputs
	// grab all matches for the regex
	// consolidate all in a list
	// read them all and write to the file one by one (as private global constants)
	// - use a hashEncoding that takes a clean path and outputs it
	// write out a method that will return a *vfs.MemFS based on the constant strings

	var match, notMatch flagutil.RegexpFlagValue
	var prefix, pkg, vfsPkg, tags, outfile string
	var doListing bool

	flag.StringVar(&prefix, "prefix", "embed", "`prefix` for all constants created")
	flag.StringVar(&outfile, "out", "", "`output file`")
	flag.StringVar(&pkg, "pkg", "main", "`package` for generated file")
	flag.StringVar(&vfsPkg, "vfspkg", vfsPkgName, "`vfs package` name")
	flag.StringVar(&tags, "t", "", "`build tag list` to put into generated file")

	flag.Var(&match, "match", "`regex` for files names to match")
	flag.Var(&notMatch, "notmatch", "`regex` for files names to not match")
	flag.BoolVar(&doListing, "n", false, "`list files` that will be embedded but do not create the output file")

	flag.Parse()

	var fs vfs.Vfs
	fs.Adds(false, flag.Args()...)

	paths := fs.Matches(match.Regexp(), notMatch.Regexp(), false)
	var mpaths = make(map[string]vfs.FileInfo)

	var f vfs.File
	var fi vfs.FileInfo
	var err error
	var n int

	var w writer
	var bb, bs []byte

	var out *os.File = os.Stdout

	if outfile != "" {
		out, err = os.Create(outfile)
		onErr(err)
	}

	const hextable = "0123456789abcdef"

	if !doListing {
		bs = make([]byte, 512)
		bb = []byte{'\\', 'x', 0, 0}

		w = writer{out}

		if tags != "" {
			w.f("// +build %s", tags).nl().nl()
		}

		w.f("package %s", pkg).nl().nl()

		w.s(`import "time"`).nl()
		w.f(`import "%s"`, vfsPkg).nl().nl()

		w.s("const (").nl()
	}

	for _, s := range paths {
		// zz.Debug2f("path: %s", s)
		f, err = fs.Find(s)
		onErr(err)
		fi, err = f.Stat()
		onErr(err)
		mpaths[s] = fi
		if doListing {
			fmt.Printf("\t%s\n", fileInfoToString(s, fi))
			continue
		}
		w.f(`	// %s`, s).nl()
		w.f(`	%s%x = "`, prefix, s)
		n, err = f.Read(bs)
		for err != io.EOF {
			onErr(err)
			for i := 0; i < n; i++ {
				bb[2] = hextable[bs[i]>>4]
				bb[3] = hextable[bs[i]&0x0f]
				w.b(bb)
			}
			n, err = f.Read(bs)
		}
		w.n('"').nl().nl()
	}

	if doListing {
		return
	}

	w.n(')').nl().nl()

	mpathsAdded := make(map[string]struct{}, len(paths)*2)
	w.f(`func %sFS() (fs *vfs.MemFS) {`, prefix).nl()
	w.s(`	fs = new(vfs.MemFS)`).nl()
	w.s(`	var zeroTime time.Time`).nl()
	w.s(`	_ = zeroTime`).nl()
	for _, s := range paths {
		p1 := s
		for n = strings.LastIndexByte(p1, '/'); n > 0; n = strings.LastIndexByte(p1, '/') {
			p1 = s[:n]
			// zz.Debug("p1: %s", p1)
			if _, ok := mpathsAdded[p1]; !ok {
				mpathsAdded[p1] = struct{}{}
				w.f("\tfs.AddFile(nil, `%s`, 0, zeroTime, ``)", p1).nl()
			}
		}
		fi := mpaths[s]
		n = strings.LastIndexByte(s, '/')
		if n > 0 {
			p1 = s[:n]
			w.f("\tfs.AddFile(fs.GetFile(`%s`), `%s`, %d, time.Unix(%d, %d), %s%x)", p1, s, fi.Size(),
				fi.ModTime().Unix(), fi.ModTime().Nanosecond(), prefix, s).nl()
		}
	}
	w.s(`	fs.Seal()`).nl()
	w.s(`	return fs`).nl()
	w.n('}').nl()
	if out != os.Stdout {
		w.Close()
	}
	// zz.Debugf("all done")
}
