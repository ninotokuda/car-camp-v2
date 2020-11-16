package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

var _schema_graphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x54\x4d\x6b\xdc\x30\x10\xbd\xeb\x57\xcc\xb2\x17\x07\xf6\xd2\x1e\xcd\x92\x4b\x02\xc1\x94\x86\x16\x77\x0f\xa5\x94\x20\xec\x69\x2a\x62\xcb\xaa\x24\x17\x96\x92\xff\x5e\x46\x23\x7b\xe5\x8f\x6c\x76\x2f\xbb\x9e\xd1\xf8\xbd\x99\x79\xcf\x72\xd5\x6f\x6c\x25\xfc\x13\x00\x7f\x7a\xb4\xc7\x1c\xbe\xd2\x9f\x00\x68\x7b\x2f\xbd\xea\x74\x0e\x9f\xe3\x93\x78\x15\xc2\x1f\x0d\x72\x49\x78\xa7\x51\xce\xdf\x75\xad\x91\x5a\xa1\xcb\x6e\x72\xf8\xc1\xd1\xf1\xe7\x46\x00\xc8\xa6\xf9\x26\x1b\xd4\x9e\x8f\xf8\x39\x9c\xf8\xf0\x98\x99\x97\x1c\x4a\x6f\x95\x7e\xde\xec\xc0\x9d\x82\x9b\x1c\xb8\x58\x00\x54\x0c\x38\x00\x99\x49\x55\x8a\xd9\x19\xd4\x1c\x96\xa6\xf3\x2e\x63\x8e\xa2\x4e\x29\xbc\xb4\xfe\x5e\x7a\x1c\x72\x3b\x40\x5d\xa7\x09\xc2\xa4\xd7\x03\x62\xa4\x66\xb8\x18\xcc\xf1\x7c\xef\x4e\x60\x57\xe2\xf7\x0e\x2d\x83\xd3\xd3\x95\x9d\xce\xd9\x27\xc8\xce\x74\x67\xb7\x4b\x85\xb1\x81\x6c\x7a\x72\x70\x68\x49\x68\x56\x7a\x90\x3e\x88\x0d\x50\x59\x94\x1e\x79\xc9\x99\x96\x2d\x26\x04\x35\xba\xca\x2a\xc3\x96\x19\xb3\xc6\x76\xbf\x54\x83\x45\x2b\x9f\xd3\x1d\x30\xc4\x26\x80\xf6\xa6\x3e\x81\xbe\xd5\xf3\x0e\x2e\xa0\xbb\x84\x8d\x47\xa0\xf9\xd7\x0c\xb2\x26\x72\x9d\x2c\x9d\xc2\xde\x86\x95\x94\x58\x75\xba\x76\x39\x14\xda\x87\x41\x55\x85\xdf\x51\x0f\x31\x83\x3f\x26\x4d\x4f\x47\x58\x9d\x20\x2a\x93\xee\xa5\x3c\xa7\xe4\xc5\xdd\x5c\xc7\x6c\xd1\xa1\xfd\xfb\x1e\xf5\xc2\xb3\x94\x48\x07\x1e\x31\x47\x3b\x91\xb9\x82\x95\xbe\x7c\x3a\x89\xb6\x0d\x2f\x0a\x80\x72\x91\x7c\xda\x87\x5f\x55\xdf\x0a\x80\x47\x55\xbd\xa4\x73\x8c\xd7\x51\xbc\x72\x56\x80\xa3\x9c\x0b\xec\x7d\x3c\x88\xc8\x77\x0b\xd5\x89\x2e\xf5\x9b\x00\xb8\x5f\x2e\x6d\xec\x80\x0d\x16\x3f\x92\x69\x0b\xec\x83\xed\x8c\x71\xd1\x10\x97\x8d\xa7\x73\xf2\x75\x7a\xca\x3f\x94\xc5\x87\x21\x31\xd2\x71\x1f\xc9\xc7\x70\xb0\xcd\xa2\x69\xd2\x66\x65\x67\x74\x75\xcc\xdb\x9d\x35\x4b\xbe\x7c\xf2\xaa\xc5\xdb\xed\x9e\xca\x87\x2a\xd3\xf9\xe9\x0e\xcb\xc9\x0d\x15\xb6\x98\x7e\x4c\x14\xaf\xf9\x97\x7a\x9a\x18\x78\x26\x87\x80\xb8\xf1\x59\xf2\x30\x33\xe0\x1b\xaa\x2d\x96\xc6\x23\x4f\x24\x78\x28\x8b\x8f\x8b\x8a\xd1\x8c\xaf\xe2\x7f\x00\x00\x00\xff\xff\xda\x29\xa8\x41\x38\x07\x00\x00")

func schema_graphql() ([]byte, error) {
	return bindata_read(
		_schema_graphql,
		"schema.graphql",
	)
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() ([]byte, error){
	"schema.graphql": schema_graphql,
}
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() ([]byte, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"schema.graphql": &_bintree_t{schema_graphql, map[string]*_bintree_t{
	}},
}}