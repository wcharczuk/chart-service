package util

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	exception "github.com/blendlabs/go-exception"
)

// Compress gzip compresses the bytes.
func Compress(contents []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(contents)
	err := w.Flush()
	if err != nil {
		return nil, exception.Wrap(err)
	}
	err = w.Close()
	if err != nil {
		return nil, exception.Wrap(err)
	}

	return b.Bytes(), nil
}

// Decompress gzip decompresses the bytes.
func Decompress(contents []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(contents))
	if err != nil {
		return nil, exception.Wrap(err)
	}
	defer r.Close()
	decompressed, err := ioutil.ReadAll(r)
	return decompressed, exception.Wrap(err)
}
