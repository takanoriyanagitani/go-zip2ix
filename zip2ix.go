package zip2ix

import (
	"archive/zip"
	"bufio"
	"encoding/asn1"
	"errors"
	"io"
	"io/fs"
	"os"
)

type Offset int64

type CompressedSize int64

type CompressionMethod = asn1.Enumerated

const (
	CompressionMethodUnspecified CompressionMethod = 0
	CompressionMethodStore       CompressionMethod = 100
	CompressionMethodDeflate     CompressionMethod = 108
)

type LeastZipIndexInfo struct {
	Name string `asn1:"utf8"`
	Offset
	CompressedSize
	CompressionMethod
}

type ZipFile struct{ *zip.File }

func (z ZipFile) Header() *zip.FileHeader { return &z.File.FileHeader }

func (z ZipFile) Name() string      { return z.Header().Name }
func (z ZipFile) RawMethod() uint16 { return z.Header().Method }

func (z ZipFile) Size() CompressedSize {
	return CompressedSize(z.Header().CompressedSize64)
}

func (z ZipFile) Method() CompressionMethod {
	switch z.RawMethod() {
	case zip.Store:
		return CompressionMethodStore
	case zip.Deflate:
		return CompressionMethodDeflate
	default:
		return CompressionMethodUnspecified
	}
}

func (z ZipFile) DataOffset() (Offset, error) {
	i, e := z.File.DataOffset()
	return Offset(i), e
}

func (z ZipFile) ToLeastIndex() (LeastZipIndexInfo, error) {
	o, e := z.DataOffset()
	return LeastZipIndexInfo{
		Name:              z.Name(),
		Offset:            o,
		CompressedSize:    z.Size(),
		CompressionMethod: z.Method(),
	}, e
}

type ZipReader struct{ *zip.Reader }

func (r ZipReader) ToIndexDerBytes() ([]byte, error) {
	var files []*zip.File = r.Reader.File
	var fcnt int = len(files)
	var vix []LeastZipIndexInfo = make([]LeastZipIndexInfo, 0, fcnt)

	for _, file := range files {
		zfile := ZipFile{File: file}
		lix, e := zfile.ToLeastIndex()
		if nil != e {
			return nil, e
		}
		vix = append(vix, lix)
	}

	return asn1.Marshal(vix)
}

func (r ZipReader) IndexToWriter(wtr io.Writer) error {
	der, e := r.ToIndexDerBytes()
	if nil != e {
		return e
	}
	_, e = wtr.Write(der)
	return e
}

type ZipFileLike struct {
	io.ReaderAt
	Size int64
}

func (l ZipFileLike) ToReader() (ZipReader, error) {
	rdr, e := zip.NewReader(l.ReaderAt, l.Size)
	return ZipReader{Reader: rdr}, e
}

func (l ZipFileLike) IndexToWriter(wtr io.Writer) error {
	rdr, e := l.ToReader()
	if nil != e {
		return e
	}
	return rdr.IndexToWriter(wtr)
}

type OsFile struct{ *os.File }

func (f OsFile) Close() error { return f.File.Close() }

func (f OsFile) ToStat() (fs.FileInfo, error) { return f.File.Stat() }
func (f OsFile) Size() (int64, error) {
	i, e := f.ToStat()
	if nil != e {
		return 0, e
	}
	return i.Size(), nil
}

func (f OsFile) AsReaderAt() io.ReaderAt { return f.File }
func (f OsFile) ToZipFileLike() (ZipFileLike, error) {
	sz, e := f.Size()
	return ZipFileLike{
		ReaderAt: f.AsReaderAt(),
		Size:     sz,
	}, e
}

func (f OsFile) IndexToWriter(wtr io.Writer) error {
	zf, e := f.ToZipFileLike()
	if nil != e {
		return e
	}
	return zf.IndexToWriter(wtr)
}

type OsFilename string

func (n OsFilename) ToFile() (OsFile, error) {
	f, e := os.Open(string(n))
	return OsFile{File: f}, e
}

func (n OsFilename) IndexToWriter(wtr io.Writer) error {
	f, e := n.ToFile()
	if nil != e {
		return e
	}
	defer f.Close()
	return f.IndexToWriter(wtr)
}

func (n OsFilename) IndexToStdout() error {
	var bw *bufio.Writer = bufio.NewWriter(os.Stdout)
	e := n.IndexToWriter(bw)
	return errors.Join(e, bw.Flush())
}

func ZipfilenameToStdout(filename string) error {
	of := OsFilename(filename)
	return of.IndexToStdout()
}
