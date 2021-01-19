// Code generated by vfsgen; DO NOT EDIT.

package generated

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	pathpkg "path"
	"time"
)

// CRDs statically implements the virtual filesystem provided to vfsgen.
var CRDs = func() http.FileSystem {
	fs := vfsgen۰FS{
		"/": &vfsgen۰DirInfo{
			name:    "/",
			modTime: time.Date(2021, 1, 20, 3, 34, 59, 969869020, time.UTC),
		},
		"/devops.k8s.io_clustercredentials.yaml": &vfsgen۰CompressedFileInfo{
			name:             "devops.k8s.io_clustercredentials.yaml",
			modTime:          time.Date(2021, 1, 20, 3, 33, 59, 240565065, time.UTC),
			uncompressedSize: 3349,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\xbc\x56\x4d\x6f\x1b\x37\x10\xbd\xef\xaf\x18\xa8\x87\x5c\xaa\x95\x83\x26\xa8\xbb\x37\x57\x4e\x01\xc3\x6d\x60\xc4\x86\x51\xa0\xe8\x81\x4b\x8e\x24\xc6\xdc\x21\x3b\x1c\x0a\x51\x7f\x7d\x41\xee\xca\x5a\xd9\xb2\xa3\xd4\x49\x74\xd2\x92\x9c\xf7\x66\x1e\xe7\x83\xd5\x74\x3a\xad\x54\xb0\xb7\xc8\xd1\x7a\x6a\x40\x05\x8b\x9f\x04\x29\x7f\xc5\xfa\xee\x34\xd6\xd6\xcf\xd6\xaf\xab\x3b\x4b\xa6\x81\x79\x8a\xe2\xbb\x0f\x18\x7d\x62\x8d\xe7\xb8\xb0\x64\xc5\x7a\xaa\x3a\x14\x65\x94\xa8\xa6\x02\x50\x44\x5e\x54\x5e\x8e\xf9\x13\x40\x7b\x12\xf6\xce\x21\x4f\x97\x48\xf5\x5d\x6a\xb1\x4d\xd6\x19\xe4\x02\xbe\xa5\x5e\x9f\xd4\x6f\xea\x93\x62\x31\x51\xc1\x4e\x55\x08\xec\xd7\x68\x8a\x01\x13\x0a\x66\x67\x26\x0d\x4c\x56\x22\x21\x36\xb3\xd9\xd2\xca\x2a\xb5\xb5\xf6\xdd\x6c\x77\x66\xfc\x37\x24\xe7\x66\x3f\x9f\xbe\x79\x7b\x3a\xa9\x00\x34\x63\x71\xeb\xc6\x76\x18\x45\x75\xa1\x01\x4a\xce\x55\x00\xa4\x3a\x6c\x40\xbb\x14\x05\x59\x33\x1a\x24\xb1\xca\xc5\xda\xe0\xda\x87\xad\x0e\x55\x0c\xa8\x73\x48\x4b\xf6\x29\x34\xb0\xbf\xd9\xa3\x0c\x21\x0f\x72\xf5\x80\xf3\x7b\xc0\xb2\xe7\x6c\x94\xcb\xc3\xfb\xbf\xdb\x28\xe5\x4c\x70\x89\x95\x3b\xe4\x52\xd9\x8e\x96\x96\xc9\x29\x3e\x70\xa0\x02\x88\xda\x07\x6c\xe0\x7d\x76\x27\x28\x8d\xa6\x02\x18\x54\x2e\xee\x4d\x87\x78\xd7\xaf\x7b\x30\xbd\xc2\x4e\xf5\x7e\x03\xf8\x80\x74\x76\x75\x71\xfb\xd3\xf5\xde\x32\x80\xc1\xa8\xd9\x06\x29\x77\xf5\xc8\x73\x60\xd4\x9e\x4d\x04\x59\x21\xec\xbc\x01\x4b\x0b\xcf\x5d\x91\x1d\x08\xd1\xa0\x01\xf1\xf7\x98\x00\x4a\x6b\x8c\x83\x55\x8f\x59\xdf\xef\x06\xf6\x01\x59\xec\x56\xd4\xc1\x62\x97\xad\xa3\xd5\x07\xfe\xbd\xca\x21\xf4\xa7\xc0\xe4\x34\xc5\x9e\x63\x90\x01\xcd\x10\x35\xf8\x05\xc8\xca\x46\x60\x0c\x8c\x11\xa9\x4f\xdc\x3d\x60\xc8\x87\x14\x81\x6f\x3f\xa2\x96\x1a\xae\x91\x33\x0c\xc4\x95\x4f\xce\xe4\xec\x5e\x23\x4b\x11\x60\x49\xf6\xdf\x7b\xec\x08\xe2\x0b\xa9\x53\x82\xc3\xbd\xee\x7e\x96\x04\x99\x94\x83\xb5\x72\x09\x7f\x04\x45\x06\x3a\xb5\x01\xc6\xcc\x02\x89\x46\x78\xe5\x48\xac\xe1\x0f\xcf\x58\x14\x6d\x60\x54\x02\xdb\x2a\xd5\xbe\xeb\x12\x59\xd9\xcc\x4a\xc1\xd9\x36\x89\xe7\x38\x33\xb8\x46\x37\x8b\x76\x39\x55\xac\x57\x56\x50\x4b\x62\x9c\xe5\x0a\x2b\xae\x53\xa9\xd4\xba\x33\x3f\xf0\x50\xd7\xf1\xd5\x9e\xaf\xb2\xc9\xd9\x14\x85\x2d\x2d\x47\x1b\xad\xf7\x12\x85\x55\xb8\xf1\x77\xf8\xdc\x5d\xfc\xe6\x19\x72\x4d\x2a\xd3\x41\xee\x17\xe0\x19\x3e\x7a\x4b\xc7\x90\x68\x35\x47\x96\xcf\x80\x6b\x4f\x94\x35\x1b\x25\xd1\x9e\x41\x9f\x81\x0d\xb4\x1b\xc1\xe3\x48\x2f\x71\xd3\xbc\x0c\x22\xe7\xed\xc2\x6a\x25\xf8\x08\xeb\xeb\x89\x83\x2c\xf1\x57\x4b\x8a\x37\xe7\x43\xef\x1d\x95\x89\x31\xa5\x35\x2b\x77\x75\xb0\x8c\x9e\x8d\xea\x49\xca\xed\x46\x5f\x0b\x63\x5f\x9c\x45\x92\x23\x2e\x2b\x07\x3b\x55\xc1\xc6\x52\x45\xf0\xe7\xdb\x93\x5f\x40\x25\x59\xbd\x4c\xee\xc2\x7e\x9c\xd2\xdf\x80\xbc\xa4\x5c\x6e\xb7\xcd\x31\xe7\x51\xb4\x39\xbb\xba\x98\x3f\xa1\xd8\x97\xd2\xef\xc1\xbd\x38\x71\x33\xda\xfc\xec\x88\x7b\xbc\xb9\x7c\x07\x96\x60\xe9\x7c\x5b\xba\x7f\x8a\xf8\x15\x88\x5f\xee\xff\x27\xf9\x7f\xb5\xf0\xa5\x09\x5f\x26\xfc\x33\x03\x28\x4f\x78\xb0\x11\xd4\x80\xd9\x37\xf1\xdd\x9c\xc9\x4b\xb9\x61\x7d\x78\x77\x7d\x03\xdb\xce\x5b\x66\xd1\xc3\xe1\x53\x98\x77\x86\x71\x37\x81\xf2\xbc\xb0\xb4\x40\xee\x67\xd8\x82\x7d\x57\x30\x91\x4c\xf0\x96\xb6\x1d\x31\x27\xc6\x03\xd0\x98\xda\xce\x4a\x1e\x7b\xff\x24\x8c\x92\x47\x55\x0d\xf3\xf2\x72\x83\x16\x21\x05\xa3\x04\x4d\x0d\x17\x04\x73\xd5\xa1\x9b\xab\x88\xdf\x7c\xfe\x64\xa5\xe3\x34\x0b\x7b\xdc\x04\xca\xd5\xfc\x7d\x2e\xbb\x53\x64\x17\x59\xa7\xef\x44\x37\x7a\x4d\x7f\xf6\xb0\x20\x29\x92\x8b\xf3\xa3\x7a\x8f\x1c\x39\xab\x47\x4d\xb2\x98\x3c\xee\x92\x07\x09\x72\x3a\x59\xc6\x51\x61\x4c\xc7\xed\x71\xb4\xba\xf5\xba\x7a\x32\xba\x42\x6f\x1a\x10\x4e\xbd\x61\x14\xcf\x6a\x89\xc3\x4a\x14\x25\xa9\x28\x9d\x9f\x90\x41\xd0\xbc\x7f\xf8\xfc\x9e\x4c\xf6\xde\xd2\xe5\x53\x7b\xea\xef\x2b\x36\xf0\xd7\xdf\x55\x8f\x8a\xe6\x76\xfb\x3c\xce\x8b\xff\x05\x00\x00\xff\xff\x6a\x56\x3f\x73\x15\x0d\x00\x00"),
		},
		"/devops.k8s.io_clusters.yaml": &vfsgen۰CompressedFileInfo{
			name:             "devops.k8s.io_clusters.yaml",
			modTime:          time.Date(2021, 1, 20, 3, 34, 59, 969795569, time.UTC),
			uncompressedSize: 23551,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\xec\x3c\xdb\x72\x1b\x37\x96\xef\xfc\x0a\xac\x76\xa7\x6c\x6d\x44\xca\xd9\xc9\xce\x66\xf9\x92\x52\x24\x25\xa3\x8a\xad\xb0\x44\xd9\xfb\xe0\xc9\x56\x1d\x36\x0e\x49\x0c\xd1\x40\x1b\x17\x4a\x9d\xf5\xfe\xfb\x14\x2e\xdd\x24\xc5\xbe\x92\x92\x9d\x87\xf1\x8b\xc5\x6e\xe0\x00\xe7\x82\x73\x47\x0f\x86\xc3\xe1\x00\x32\xf6\x01\x95\x66\x52\x8c\x09\x64\x0c\x1f\x0d\x0a\xf7\x4b\x8f\x56\xdf\xeb\x11\x93\xe7\xeb\x6f\x07\x2b\x26\xe8\x98\x5c\x5a\x6d\x64\x7a\x87\x5a\x5a\x95\xe0\x15\xce\x99\x60\x86\x49\x31\x48\xd1\x00\x05\x03\xe3\x01\x21\x20\x84\x34\xe0\x1e\x6b\xf7\x93\x90\x44\x0a\xa3\x24\xe7\xa8\x86\x0b\x14\xa3\x95\x9d\xe1\xcc\x32\x4e\x51\x79\xe0\xc5\xd2\xeb\x37\xa3\xef\x46\x6f\xfc\x8c\x13\xc8\xd8\x10\xb2\x4c\xc9\x35\x52\x3f\x41\x09\x34\xe8\x36\x73\x32\x26\x27\x4b\x63\x32\x3d\x3e\x3f\x5f\x30\xb3\xb4\xb3\x51\x22\xd3\xf3\xcd\x98\xed\x3f\x33\xcb\xf9\xf9\x7f\x7d\xff\xdd\x7f\x7e\x7f\x32\x20\x24\x51\xe8\xb7\x75\xcf\x52\xd4\x06\xd2\x6c\x4c\x84\xe5\x7c\x40\x88\x80\x14\xc7\x24\xe1\x56\x1b\x54\x7a\x44\x71\x2d\xb3\x02\xfb\x81\xce\x30\x71\x88\x2c\x94\xb4\xd9\x98\xec\xbe\x0c\x73\x23\xa2\x91\x48\x01\x8c\x7f\xc2\x99\x36\xbf\x6c\x3f\x7d\xcb\xb4\xf1\x6f\x32\x6e\x15\xf0\xcd\xa2\xfe\xa1\x5e\x4a\x65\x6e\x37\x00\x87\x64\x9d\x84\x17\x4c\x2c\x2c\x07\x55\x8e\x1f\x10\xa2\x13\x99\xe1\x98\xf8\xe1\x19\x24\x48\x07\x84\x44\x62\xfa\xe9\x43\x02\x94\x7a\xf6\x00\x9f\x28\x26\x0c\xaa\x4b\xc9\x6d\x2a\x4a\xe0\x14\x75\xa2\x58\x66\x3c\xf9\xef\x97\x58\x00\x27\x54\xe8\x9b\xc9\xc8\x8f\x22\xe4\xef\x5a\x8a\x09\x98\xe5\x98\x8c\xb4\x01\x63\xf5\xc8\xbf\x8e\x6f\x03\xe9\xae\x6e\xa7\xe5\x13\x93\xbb\x6d\x69\xa3\x98\x58\xd4\x2d\x14\xf7\x49\xe4\x9c\x6c\x71\xb7\x76\xc1\x51\x1c\xbf\xb3\xe6\x87\xeb\xbb\xe9\xcd\xaf\xb7\x3d\x56\x4d\xb8\x75\xd8\x65\x4b\xd0\x58\xbf\x98\x7f\xbd\xb3\xd2\xe4\xaf\x17\xd3\xeb\x8e\xeb\xbc\xba\x7c\x2a\x65\x84\x69\x02\xc4\x94\x3f\x15\x66\x0a\x35\x0a\xc3\xc4\x82\x98\x25\x12\x8d\x6a\x8d\xca\x8f\x88\x8b\x10\xf2\xb0\x44\x41\xcc\x92\x69\x22\x67\x7f\xc7\xc4\x90\x07\xd0\x41\x80\x91\x8e\xc8\xab\xfd\xcd\x17\x27\x70\xb4\x27\xe5\x3b\xa8\x5c\xfc\xbc\x8b\x08\x05\x13\x16\x0d\xaf\xd7\xdf\x06\x71\x4b\x96\x98\xc2\x38\x8e\x94\x19\x8a\x8b\xc9\xcd\x87\x3f\x4f\x77\x1e\x93\x5d\xc4\xa3\x80\x3b\x6c\x1d\x52\x61\x2c\x99\x4b\xe5\x7f\x16\x6f\x2f\x26\x37\xe5\xf4\x4c\xc9\x0c\x95\x61\x85\xb4\x87\x7f\x5b\xda\x68\xeb\xe9\x93\xc5\x5e\xb9\xfd\x44\x19\xa2\x4e\x0d\x61\x58\x35\xca\x09\xd2\x88\x82\x13\x30\x4f\xc5\x92\xe8\x9e\x36\x3b\x80\x89\x1b\x04\x22\x12\x7a\x44\xa6\x9e\x1d\xda\x1d\x46\xcb\xa9\xd3\x5e\x6b\x54\x86\x28\x4c\xe4\x42\xb0\xdf\x4b\xd8\x9a\x18\xe9\x17\xe5\x60\x30\x9e\xea\xcd\x3f\x7f\xde\x04\x70\xb2\x06\x6e\xf1\x8c\x80\xa0\x24\x85\x9c\x28\xf4\xec\xb4\x62\x0b\x9e\x1f\xa2\x47\xe4\x9d\x54\x48\x98\x98\xcb\x31\xd9\x52\x71\x85\x16\x4e\x64\x9a\x5a\xc1\x4c\x7e\xee\x15\x2a\x9b\x59\x23\x95\x3e\xa7\xb8\x46\x7e\xae\xd9\x62\x08\x2a\x59\x32\x83\x89\xb1\x0a\xcf\x9d\x06\xf5\x5b\x17\x5e\x13\x8f\x52\xfa\xaf\x2a\xea\x6d\xfd\x6a\x67\xaf\x7b\x12\x1d\xfe\x79\x65\xd6\xc0\x01\xa7\xd6\x82\x68\x87\xa9\x01\x8b\x7d\xe9\xbe\xbb\x9e\xde\x93\x62\x69\xcf\x8c\xa7\xd4\x0f\x02\x5e\x4e\xd4\x1b\x16\x38\x82\x31\x31\x77\x87\xc3\x31\x71\xae\x64\xea\x61\xa2\xa0\x99\x64\xc2\xf8\x1f\x09\x67\x28\x9e\x92\x5f\xdb\x59\xca\x8c\xe3\xfb\x27\x8b\xda\x38\x5e\x8d\xc8\xa5\x37\x4d\x64\x86\xc4\x66\x34\x9c\xa4\x1b\x41\x2e\x21\x45\x7e\xe9\x54\xc2\x4b\x33\xc0\x51\x5a\x0f\x1d\x61\xbb\xb1\x60\xdb\xaa\x3e\x1d\x1c\xa8\xb6\xf5\xa2\x30\x53\x35\xfc\x8a\x07\x70\x9a\x61\xb2\x73\x62\x28\x6a\xa6\x9c\x4c\x1b\x30\xe8\x4e\xc2\xb6\xf9\x6a\x3e\xa9\xf1\xb4\x06\x66\x5d\x3f\x1a\x05\x17\x6a\xb1\x37\x82\xec\x98\xa1\x3a\x38\x0d\x54\x68\xc4\xda\xfb\x17\x61\xc7\x97\x37\x57\x77\xe3\x41\x0f\xa0\x71\xde\xbd\x1b\xd1\x6b\x5e\xe9\xcf\xbc\x03\x01\x8b\xaf\x8b\xbb\x62\xd5\xfb\xdf\x61\xfe\x9d\x15\xce\xba\xb8\x91\x3b\xcc\x57\xe1\xb9\x67\xbb\x14\x06\x98\x40\x35\xea\x43\x0a\xca\x74\xc6\x21\x77\x3e\x48\x2f\x12\x52\xa1\xaf\x64\x0a\x4c\xb4\x6c\xfc\xea\x76\x1a\xc6\x15\x66\x85\x0a\x4d\x68\x78\x62\x35\x52\x32\xcb\xc9\xea\x7b\xed\x4d\x28\x4b\x9c\x0e\xbd\xc2\x39\x58\x6e\x74\x15\x8d\x25\x39\x89\x3c\x1f\x71\x99\x00\x3f\xe9\x85\x2b\x9a\x84\xb6\x6c\xf7\xda\x24\x94\x2c\x25\xa7\xda\x09\xc9\x9c\x2d\xac\xf2\xf6\xc6\x9b\x41\x37\x7f\x7f\xc1\xac\x51\x2c\x9c\x2f\xee\xac\x48\xd5\xbb\xa7\x6b\xc7\xa1\xf1\xe9\x0c\x35\x59\xca\x07\x87\x74\x22\x85\x70\x1a\xd6\x48\x67\xe6\x0a\x90\x95\x10\x03\x96\xa5\x1f\xf8\xd6\x51\xc9\x9b\xae\x12\x3a\x28\x24\xa9\x35\x16\x38\xcf\x09\x3e\xba\x91\x6c\x8d\x95\xc0\x9a\x51\xf3\xd2\x0b\x3f\x31\x8e\x75\x6f\x9f\x6a\xb0\x0b\x37\xd8\x9b\x1c\x41\xa6\xd3\xb7\xe4\xd2\x01\x9f\xb3\xc4\x29\xae\x0b\x6b\x96\x52\x31\x93\x93\xb9\x1b\xe4\x64\xa3\x16\xaa\x97\x04\x8d\x89\x55\x18\xd1\x0d\x8a\x3d\xf1\xbc\x1a\x91\x3b\xfc\x64\xbd\x4e\x64\x73\x62\x9d\xe7\x4d\x80\xdc\xbf\x9d\x16\x74\x74\x63\x6a\x61\x37\x9e\xe3\x88\x34\x2a\xd3\x07\xed\x38\x7c\x0b\xf1\xa4\x44\xdc\xcb\x56\x81\x30\x31\xb2\x01\xe7\xaf\x87\x70\x61\xad\x75\x47\x8c\xaf\x8b\xf1\x4e\x2f\xf9\xfd\xa6\x98\xce\x5c\x60\xb6\xd9\xa9\x3b\x50\x85\x4c\x5e\x57\x1e\xac\xd2\x11\x33\x98\x36\xac\xdc\x09\x83\x62\x10\x28\x05\x79\xcd\x98\x15\xe6\x3d\xb8\xfa\x4b\x18\xbd\xc5\xd4\x15\xe6\x3b\xac\xdc\x66\x58\xc3\xee\xbf\x24\x2b\x55\x04\x5e\x8d\xe3\x30\x1e\xe7\xba\x97\x51\x8e\x6b\x5e\x97\x42\x52\xf3\x3e\x92\x77\x50\xbf\xf1\x4a\xfb\xe8\x83\x70\xa7\xc5\x3a\x68\xd0\xa0\xed\x32\x25\xd7\x8c\xe2\x53\x0d\xbe\x12\x72\xa6\xbd\xd8\x15\xcf\xeb\xc5\xc5\x07\x05\x1e\x98\x97\x5e\x26\xb4\x01\x91\xe0\x8b\xab\x53\xe7\x2b\x5e\x31\xd5\x51\x04\xaf\xc2\xe8\xd2\xb2\x32\x85\x89\x91\x2a\x0f\x9b\x7e\x60\x9c\x93\x8c\x43\x82\x84\xd5\x30\x65\xb3\xe8\xc6\xec\x7a\x23\x7b\xbe\x06\x75\xce\xd9\xec\xdc\x41\x3a\x39\x4e\x77\xd4\xbb\x56\x7d\x5d\xac\x5e\xe7\xfd\xa9\x69\x0d\x9b\xf0\xec\xf2\x5b\x22\xa0\x16\x36\x75\xd1\x4a\x21\x30\x34\x86\x83\x0d\x2b\x7b\xc2\xce\x98\x00\x95\x87\x00\x5f\x59\xe1\xa4\x83\x51\xf4\x61\x14\x18\x96\x90\x4c\xd2\x36\x8a\xd5\x4a\xba\x17\x13\x44\xe5\x6c\xc6\xf4\xe2\xb6\xab\xc2\x9d\x6c\x4d\x21\x1a\x8d\x8e\x38\x4e\x6d\x08\xcd\x2e\xb8\x97\x56\xc3\xd6\x18\xd2\x4d\x0d\x38\x16\x01\xbf\xc7\xd5\xed\x85\x68\xb6\x10\x4e\x11\x39\x05\xf0\xf5\xd5\x74\x48\xb6\xf4\x24\xd0\x74\x67\x52\x0b\x89\x1a\x70\xf0\xc4\xdb\x25\x51\x4c\xfe\xfc\x91\x88\xd4\xa6\xe6\xa3\x9a\xe9\xaf\x8a\x1b\x5e\xce\x11\x5c\xd4\xac\x5b\x1c\xec\x18\x9c\xfe\x14\x46\xfb\x9c\x8c\xa2\x41\x7f\x15\x10\x88\x59\x82\x09\x07\x55\xc0\x8c\x57\xfa\x81\xb3\x3c\x66\x0e\x42\x30\xd0\xd7\x29\xf7\x70\xdf\x81\x8f\xa7\x93\x25\x52\x5b\x67\xf6\x03\xc2\x33\x29\x39\x82\xa8\x18\xe1\xec\x7d\x0d\x3f\x1b\x59\x9d\x75\x50\x74\x54\x9b\xa3\x25\x45\xab\xe4\x48\x18\xcd\xb2\xe4\xa5\x49\x9b\xda\x77\x5a\x25\x83\x03\x15\x61\xb3\x90\x2f\x32\x5b\x1d\x36\xef\x49\xdc\xcf\x93\xf7\x7b\x61\xf3\x22\xb3\x1e\xbe\xf3\x4f\x6b\x65\xa8\x03\x81\x96\x30\x3e\xd0\xd2\xaf\xec\xac\xc1\xd3\xcc\x3a\xd9\xc1\x35\xcb\x9a\x5e\x77\x14\x91\x36\x06\xfb\x22\x06\xcb\x8e\xb1\x68\x66\xc9\x14\x9d\x80\x32\xf9\x1f\x03\x65\x42\xd6\x99\x54\xa6\x19\xd2\x5c\xaa\x14\xcc\x98\x30\x61\xfe\xfc\x1f\x1d\xd6\x64\xc2\xe0\x02\xd5\x8b\xd1\x79\x18\x36\x7d\x38\x1f\x5a\x06\x2c\xa5\x5c\xd5\xd0\xbe\x8f\x83\xd6\xca\x80\x96\x6d\x14\x69\xf7\xb7\x3f\x1e\xa6\x91\x59\xb6\xfe\xcb\x95\x05\x3e\x35\x90\xac\x0e\x06\xa1\x0f\x9b\x99\xd9\x19\x67\xc9\xa1\x5b\xd7\x2b\x96\x5d\x4a\x11\x68\x7d\x88\x55\xe9\x48\xfb\x3a\x9d\x6a\xb3\x85\x02\xda\x45\xa7\xbe\x0f\x23\x8b\x44\x6a\x31\xd3\x1d\xe2\x04\xb5\x1e\x1d\xa8\x14\x53\x49\xbb\x86\xdf\xc5\x0e\xdc\x94\x33\xa7\xdb\x5d\xf0\x12\x2b\x18\x4c\x93\x0b\x6b\xe4\x51\x61\x8b\x36\x0a\x0c\x2e\xf2\x9e\xdb\x29\xa6\xc5\x18\x73\x74\xa4\xba\xa3\x0a\x98\xb8\x95\x14\x7f\xc4\xb9\x54\xf8\xbe\x89\x41\x95\xfb\xfa\x9f\x25\x9a\x25\xaa\x00\x88\x08\x49\x91\xcc\x3c\xa8\x82\x65\x23\x72\xe5\x5e\x35\x6b\x4b\x5f\x5e\xdc\x9f\x4b\x7c\x59\x2e\x91\x69\x8a\x82\x22\x1d\x91\x1f\xad\x21\x42\x1a\x02\x2e\xf2\x94\xb4\x05\xa2\xb2\xc2\x7b\xcc\xe0\x22\xf5\x07\x7d\x46\x80\xcc\xf1\x61\xfb\x71\x86\x46\x8f\xc8\xcd\x9c\xe4\xd2\xaa\x0e\x10\x13\x10\x61\xfd\x24\xc1\xcc\x97\x89\xf0\x31\x43\xce\x91\x86\xca\x53\x62\x95\x42\x61\x3c\x2e\x67\xa1\x22\xe5\x25\xa6\x05\x6c\xac\x23\xce\x90\xcc\x81\x6f\x0a\xcf\x4d\xb2\x55\x7f\xc8\x4b\x59\x87\xc7\xf7\x42\x21\xd0\xbc\x99\x9d\x20\xf2\x5f\xe7\xcd\x43\x86\x1d\x2d\xd0\xf6\xd8\x56\x0b\x49\xf6\xab\xef\x29\x3c\xb2\xd4\xa6\x44\xd8\x74\x86\xca\xf9\x4e\x99\xa4\xd1\x5d\x77\xa4\x9f\xa1\xaf\x8a\x02\xcd\x5b\x08\x4a\x7d\xee\xc5\x7b\x62\xa5\x18\xbe\xf9\x13\x49\x11\x84\x2e\x84\x47\x13\x81\x21\x24\x9f\xb9\x30\xa1\x1d\x28\xcc\x0d\x2a\x82\x6b\x16\xf2\x68\xdf\xbe\x29\x21\xb2\x85\x70\x52\x0b\x22\x0f\x80\xe3\x26\xc9\xc3\x92\x25\xcb\x16\xa8\x29\xe4\x1e\x2f\x8d\x94\x30\x41\xa4\x40\xf2\x20\xd5\x0a\x55\x14\x22\xab\x43\x49\x94\x24\xa0\x70\x6e\x39\xcf\xff\xa5\x0d\xf9\xa7\xba\xea\xcd\x9f\x9a\x25\xea\x71\xb8\xe9\xb0\x18\x32\x61\x86\x52\x0d\x03\xfb\xc6\xc4\xa8\x06\xe9\x3d\xce\x23\x68\x8a\xf5\x98\x00\xce\x7e\x47\xd5\x16\xed\xfd\x54\x0e\x8c\x79\x52\x99\xc1\x27\x8b\xbe\xab\xc6\xc9\x4f\xa8\x9b\x07\x09\x4a\xad\x0e\x47\x36\xcd\x4c\x5e\x5d\xfc\xc9\x50\xa5\x20\x50\x18\x9e\x13\x85\xa9\x5c\x63\x51\x7b\xf6\xc7\x5b\x1b\xa9\x60\x51\x71\x3e\x6b\x8d\x66\xf5\x66\x5d\x90\x5f\x24\xd5\x84\xff\x9b\xa2\x30\x6c\x9e\x87\x6c\x6c\x89\x7d\x94\xe3\x4a\xd2\x16\x05\x18\xce\xe6\x98\xe4\x09\xaf\xd4\x1a\x1d\x6a\x86\x55\xb6\xda\x89\x03\x47\xf3\x15\x8b\x96\x29\x24\x4b\x17\x4c\x8d\x0f\x24\x75\xcc\x02\xbc\x0b\x60\x0a\x5a\xa7\x3e\x20\x2f\x80\x87\x80\xcd\xa3\x41\xe4\xbc\x89\xc8\x55\xa4\x6d\xb3\xaf\xf5\x71\x44\xab\x7e\xe4\x30\x43\x5e\x6b\xb6\xfb\x25\x33\x3b\x28\xe3\xd6\x53\x9c\x81\xd6\x93\xa5\x02\x5d\xeb\x18\x14\xa1\xcc\x2c\x37\x78\x28\xd6\x6e\x95\x07\xa9\xe8\xc1\x64\x6b\x8a\xb8\xba\xc4\x5a\xed\x36\x2e\x53\x6c\x0d\x06\x7f\xc1\xfc\x25\x09\x61\xa0\xa9\x12\xb6\x23\xe6\x37\x73\xdf\xd4\xc1\xe6\x0c\xe9\x59\x50\x27\x92\xe2\x2b\x1d\x61\xd4\xa9\xfd\x96\x24\xe1\x9e\x5d\xf6\xae\x99\x37\x42\xf7\x0e\xae\x57\xb5\xc6\x40\xb2\x0c\xf6\x73\x09\xe1\x78\x9d\xe0\x7c\x8e\x89\x39\x69\x30\x36\x52\x14\x56\x32\xe8\x64\x2a\x51\x7b\x8f\xca\x48\x8e\xce\xb3\xf5\x80\xfc\x2a\xc7\xba\xb6\x61\x33\x3d\x5c\xd9\xa2\x30\x36\xf2\x38\x87\xe9\xa1\x3b\x0c\x03\x3d\xdd\xee\x9d\x7d\x6f\xb1\xbf\x11\xb1\x7d\xb4\x3c\x90\x11\xf9\x00\x9c\xd1\x08\x3f\xd4\x0b\x6e\x65\x91\x24\x3c\x6b\x81\x3d\x51\x38\x47\xb5\x19\xef\x4b\x45\xb7\xf2\xfa\x11\x13\x6b\x3a\x39\x8e\xad\x7e\xd9\x0a\xf3\x83\x89\x16\xc8\xb4\xc2\x3c\xba\x55\x90\x65\x9c\xb5\x95\x9d\xbd\x52\xf3\x32\xf6\x2c\xfb\x37\x2c\xc5\x0b\x4a\x9b\xf2\x20\xfb\x42\x5e\xcc\xd9\xea\x31\x0b\x2c\x63\x29\x12\x30\xc1\x93\x6b\xa9\xdd\x94\x67\xd7\x77\x62\x02\xf5\x11\xcb\x8d\x3f\x2d\x52\xf0\x9c\x3c\x28\x66\x0c\x86\x4e\x8f\x92\x65\x2d\x27\x75\x57\xab\x50\x30\x38\xdc\x69\x02\x3d\x26\x43\xe5\x5c\xa3\x71\x3f\x07\x3d\xa0\x17\x1c\xcb\x44\x2a\x85\x3a\x93\x82\x7a\x57\x5b\x6e\x44\xbc\x85\x46\x2b\xcc\x47\x5f\x26\xab\x18\xce\x58\xc3\x80\x15\xe6\x47\xa5\x1d\x5b\xca\x24\x56\xa3\x12\x95\x8d\x4f\x9d\x10\x6d\x42\x71\x48\x2a\x33\x79\x43\x52\x93\xc2\x1b\x96\x9b\x19\x1c\xe4\xa6\x57\x21\x29\xd0\xb8\x50\xe5\xeb\x78\x88\x71\xf1\x2b\x5c\xb3\xa4\x5f\x67\x59\x9c\xd9\xa1\x29\xee\x76\x33\x72\x27\xbb\x1f\x21\xb4\x67\xf8\x1b\x76\x21\x75\x87\x0d\xfc\x9a\x39\xeb\xc1\xc4\x62\x9a\x6b\x83\xe9\xce\x26\x64\xf1\xce\xed\x40\xfb\xf7\xbd\x36\x90\x81\xd5\x6d\xeb\x4f\xdc\x98\x41\x9f\x0c\x44\x93\x6d\xae\xf2\xd0\xa3\x70\xe4\x3b\x85\x3a\x30\xa1\x77\x36\x74\xda\x3a\xdd\x57\xeb\xa3\x1f\x51\xa4\x4b\xe1\xb1\x68\x79\x0d\xbd\x81\xb7\x36\xad\x3e\xa9\xed\xde\x63\x9b\xef\x98\xc2\xe3\xad\xa4\x38\x91\xf4\x05\x17\x91\x6b\x54\x5a\x72\x7a\xe7\xe8\xf5\xb5\x13\xed\x0d\x2f\x43\x12\x7b\xab\x1e\xbe\x75\x9f\xa5\x93\x9b\x7a\x60\x58\xab\x70\xc1\xb4\x51\x79\x5b\x2b\x6c\x1c\x16\xb2\x8f\xf1\x6f\x8d\xc6\x9d\xb5\x4d\xcf\x4f\x45\xd5\xb8\x45\xdc\x98\x52\x52\x3d\x43\x01\x64\x67\xb3\xef\x3c\x54\x9f\x27\x07\x26\xc2\xf1\x09\x5b\x24\x0a\x39\x98\xb2\xe9\xa4\x44\xa5\xd6\xe0\x85\x0d\x1e\x51\x4f\xee\xd0\xc5\x57\xdb\xc7\x17\xaa\xf1\xc5\x2f\xe7\x22\x41\xb8\x37\x95\x41\x82\x23\x72\x79\x77\xd3\xe8\x32\x64\xdc\x2e\x98\x08\x2d\x49\x8e\x5d\xdb\x57\x00\xb4\x4f\xaa\xcd\x72\xff\x9f\x15\x86\x71\x02\x3e\xc7\xd6\x96\x9f\x94\x21\x73\x30\x97\x56\x14\x41\x41\x71\xab\xa0\xc8\x25\x81\x73\x86\x18\x25\x56\x71\xf2\xc0\xcc\x92\x2c\xa5\x6e\xf6\x7f\xca\x78\x2d\x40\xf4\x17\x45\xf0\xcc\xcf\xf3\xce\x7c\x06\x66\xb9\x7f\x8d\xa1\x11\xe4\xfb\xbb\xb7\x01\xf3\x98\x47\x6c\xf2\xae\x5a\xfb\x43\xba\xfb\xd9\xad\x8d\x34\xad\x1e\x54\x85\x24\x07\x49\x28\x39\xef\x84\x37\xc8\x25\x49\x21\xcb\x9c\xb5\xf3\xb2\xc1\xf9\x66\x8c\x1e\x3d\xab\x76\xd2\x31\xb4\xfa\x9a\x3d\xfb\xb1\x59\xbd\xfa\xbe\xc2\x5e\xdf\x51\x1c\xe9\x64\x75\xd3\x19\x6a\x08\x10\x8d\x19\x28\xaf\x03\xfc\x7b\x47\xb9\xad\x56\xf8\xaa\x58\x93\x99\x57\x7a\xd3\x64\x18\x44\xfa\x5d\x85\x9d\xec\xe5\x69\x18\x14\x20\xcc\xcd\x55\x2f\x2f\xad\xb6\x56\x78\x5c\x9d\xb0\x45\x47\xcb\x5e\xd5\xc9\xde\xb5\xc1\x96\xa3\xd5\x5c\x13\x3c\xbc\x1e\xd8\xa1\x2f\xb4\x67\x1d\xf0\xd0\x1a\xa0\x7f\xdb\xa0\x54\xfa\xd5\xff\xb6\x4a\x79\x8d\xfd\xcf\xdd\x6a\x7f\x4f\xaa\x7b\x8d\xbb\xec\x58\xf7\xeb\x5e\xd9\xeb\x52\xd5\xeb\x52\xd1\x6b\xa9\xe6\x75\xad\xe4\x0d\x0f\xe8\x81\x7d\xb6\x0a\xde\x61\xd5\xbb\x50\x9f\x6b\x4a\x08\xf6\xad\xdc\xc5\xda\x5c\x03\xc8\x3e\x55\xbb\x5d\x3d\xd1\x64\xa2\x9b\x2b\x76\x07\x55\xeb\x0e\x35\x86\xeb\xaa\x7b\xb6\x8d\x9a\xac\x3a\x69\x31\x2c\x8d\x40\xa7\x8b\x8a\xfe\x9a\x75\x87\xab\x8a\x7e\xdc\x76\xce\x6e\x3b\x5e\x84\x99\xb4\xe1\xfe\x67\x80\xb7\x1f\x40\x42\x75\xe8\xd8\x70\x97\x91\x52\x85\x5a\xb7\x06\xb7\x6f\x63\xf9\xb1\x1c\xef\x44\x34\x59\xc2\x8c\x63\xe1\x8e\xd7\x06\xad\xbd\x0a\x5c\x17\x61\x81\x8d\xff\xbf\x4d\x80\xa2\x8b\x38\x2e\xf5\x4a\xd7\x85\x1e\xaa\xa6\xa5\xa6\xcd\x6e\x38\xa7\xb5\x53\x94\xb2\xfd\xe5\x80\x86\xf5\xfe\x20\x65\x1e\x53\xdb\xf7\x59\x73\xa1\x3d\xa2\xe4\x27\x9e\x79\x65\x20\xe7\x64\xe2\xa3\xdc\xb3\xf2\x22\x47\xf9\xf5\x83\x8a\x40\x43\x91\x1b\x51\x8c\x1a\xbd\x44\xb2\xb0\x26\x2c\x69\x48\x17\xba\xe5\x9e\x2f\x55\x78\xf0\x05\x5c\x99\x66\x52\x60\x65\x4c\xd9\xeb\xa0\x5c\x16\x80\x76\x12\x4d\x1b\x33\x95\xc8\x8c\x61\xb8\x4e\x06\x35\x3d\x1b\xe5\x5e\x8a\xa2\x50\x21\xd1\xa1\xac\x7c\xc8\x01\x52\x98\x71\x96\x80\xee\x23\x6d\x25\x26\x77\x71\x72\x35\x46\x83\x7a\x77\x68\x1b\xd3\xcd\xad\x6e\xff\xeb\x00\x1c\xbb\x66\x07\x60\x0d\x8c\x3b\x0d\x38\x1e\xb4\x57\x3a\xda\x7a\x62\xbb\x79\x31\xd1\x2d\xfb\x72\x0b\xc6\xab\xf2\x5f\x6e\xc1\xf8\x95\x82\x2f\xb5\x60\x7b\xaf\x7e\xc9\xe5\xda\x11\x91\x29\xf5\xdd\xfe\x81\x86\xb5\xef\x23\xca\x47\x75\xfd\xbf\x84\x8a\x2d\xce\xf2\x97\xd0\xa6\x0d\x3d\xbc\x3d\xb5\x62\x04\xb4\x71\x20\x28\x1a\x60\x7c\x73\x05\xa9\x08\x6d\xca\x35\x2b\x69\x57\x7c\x44\xe5\x88\x86\x19\x0e\xda\x4c\x94\x9c\xe1\x3d\x4b\xbb\x99\xdf\xb7\xa0\x4d\xa8\xd0\x3e\xf8\x88\x7f\xe6\x02\x82\x90\xff\x0c\x5b\x1d\x0d\x8e\xab\xa7\x76\xe8\xd4\xd1\xe6\x5e\x81\xd0\xac\xf8\x9e\x4e\xcf\x8d\xef\x6c\x97\x98\x12\x54\x11\x58\x3a\x5f\x22\x78\xaf\xf5\x91\xa3\x24\x20\xa4\x0b\xc0\x5f\x1c\xdd\x14\xb5\x86\x45\x37\x1c\xff\x6a\x53\x10\x43\x17\x4d\x79\xaf\x37\x4e\x25\x4c\x50\x7f\x13\x59\x2c\x4a\x49\xf3\x7e\x7a\x2d\x7a\xdc\xd3\xaa\x24\xcc\xc1\x6e\xa3\x42\xd0\x55\x81\x4c\x55\x6e\x45\xb0\x4f\x36\x38\x72\xc3\x07\xa9\xe8\xd9\xe6\x03\x30\x11\xcc\xe6\x74\x14\xbc\x7b\xa5\x5f\x1c\x83\xaa\xa8\xa8\x2e\x2b\x18\x02\x9e\xd8\xf6\x56\x86\x3f\x4f\x4e\x07\xb9\x0c\x11\xf9\xbd\xb2\x0d\x8d\x27\x3f\x01\xd7\x78\x46\xde\x8b\x95\x90\x0f\x87\xef\xbe\xb3\x53\xed\xab\xad\x71\xe7\x45\x81\xb5\xd3\xa9\x3e\x4a\x7b\xd7\x1e\xb2\xe7\xd6\xdd\xfe\xcb\x68\xbd\x7c\x60\x2e\x93\x55\xd5\xb6\x9b\x52\x46\xb5\x27\x75\x87\xd4\x17\x64\xe9\x4e\x29\xe9\x7c\x4a\xc9\xc3\x32\x6f\x2e\xc2\x3a\xce\xb1\xf8\x5d\xb2\x06\x8e\x35\x95\xe7\x25\xf5\xe9\xec\x77\xa0\x57\x53\xf6\x7b\x05\x0e\xcd\x1e\x4d\x93\x1f\xf3\x14\xf6\xcd\x64\xfd\xdd\x0b\xc3\xff\xcb\x73\xc2\xf7\x1f\x9e\xeb\x58\x58\x77\x43\x77\x1a\x06\xfc\xe4\xad\x76\x05\xc7\x1f\x6d\x94\x4d\x8c\xec\xd7\xba\x50\xa7\x4a\x9f\xc8\xd6\x4c\x31\x9c\x6f\xa9\xce\xe7\x15\x2e\x1f\xb7\xf4\xdc\x76\x28\x80\xde\x4c\xbe\x48\xad\x39\x7c\xce\xac\x1b\xb3\x8a\xef\x65\xee\x84\x72\x85\xe7\x55\xba\xd4\xf1\x0b\x71\x3e\xc3\x5a\xa9\xe0\x22\x90\x4f\x56\x1a\x68\xca\x04\xf5\x2d\x8b\x00\xe7\x32\x01\x53\x1f\xbc\xf5\x29\x5f\x37\x66\xa8\xbb\xe5\xa7\x3b\x65\xa7\x33\x30\x06\x95\x18\x93\xff\x7d\xfd\xb7\x6f\x3e\x0f\x4f\x7f\x78\xfd\xfa\xe3\x9b\xe1\x7f\xff\xf6\xcd\xeb\xbf\x8d\xfc\x1f\xff\x7e\xfa\xc3\xe9\xe7\xe2\xc7\x37\xa7\xa7\xaf\x5f\x7f\xfc\xe5\xdd\xcf\xf7\x93\xeb\xdf\xd8\xe9\xe7\x8f\xc2\xa6\xab\xf0\xeb\xf3\xeb\x8f\x78\xfd\x5b\x47\x20\xa7\xa7\x3f\xfc\xdb\xe0\x19\x13\xb7\xbb\x67\x6a\xc3\x87\xa7\x8d\x8a\xe5\x97\xfb\x7c\xda\x71\xeb\x1b\xa0\x95\x96\xb2\xb8\xff\xbe\x11\x2d\x27\x21\xb1\xf2\xc9\xc4\x62\xf7\x6b\x1d\x97\x90\x41\xc2\x4c\x3e\x3a\xe4\xba\x65\x94\x9d\xba\x98\xf1\x9f\x92\xf3\x45\x24\xa7\x50\x30\x3e\x31\x1d\x3e\xc9\x88\x3e\xc7\xf3\xba\xd4\x1a\x02\x52\x3c\x23\x9f\x2c\x08\xc3\x4c\x7e\x5a\x4b\x1b\xa6\xf4\x41\x82\x90\x44\x29\xfa\xa7\x1c\x7c\x45\x39\x28\x8e\xf2\x5e\x9f\xb3\x34\xc0\x6b\x94\xc8\x33\xb7\x57\xa0\x73\x0b\x41\xe5\x97\x07\xa6\x84\x4b\x00\xd3\xa6\x26\x89\x46\x00\x87\xcd\xeb\x5f\x0b\xab\x24\xc3\xde\x43\xff\x7d\x17\xba\xc5\xc1\x78\xcd\x6c\xfb\x89\x9d\x95\xbc\x29\xd6\x8f\xe1\x20\xf9\xbf\xff\x1f\x6c\x22\xc3\x50\xba\x46\x7a\xfb\xf4\x93\xd3\x27\x27\x3b\xdf\x94\xf6\x3f\xb7\xd2\x48\xe4\xe3\x6f\x83\xb0\x30\xd2\x0f\xc5\x17\xa2\xdd\xc3\x7f\x04\x00\x00\xff\xff\xae\x91\x2f\x24\xff\x5b\x00\x00"),
		},
		"/devops.k8s.io_machines.yaml": &vfsgen۰CompressedFileInfo{
			name:             "devops.k8s.io_machines.yaml",
			modTime:          time.Date(2021, 1, 20, 3, 33, 51, 478952657, time.UTC),
			uncompressedSize: 10628,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\xcc\x1a\x5d\x6f\xdb\x38\xf2\xdd\xbf\x62\x90\x7b\xc8\x1d\x10\xcb\xbb\xb7\x5b\x6c\xe1\xb7\x5c\xda\x5e\x83\x36\x5d\x23\x4e\xfb\x72\x38\x14\x63\x71\x6c\x71\x4d\x91\x2a\x3f\x92\x7a\x0f\xf7\xdf\x0f\x24\x25\x59\xb2\x25\xf9\xa3\x09\x70\x7e\xb2\xf8\x31\xdf\x33\x9c\x19\x72\x34\x1e\x8f\x47\x58\xf0\x2f\xa4\x0d\x57\x72\x0a\x58\x70\xfa\x6e\x49\xfa\x2f\x93\xac\x5f\x9b\x84\xab\xc9\xe3\xcf\xa3\x35\x97\x6c\x0a\x37\xce\x58\x95\xdf\x93\x51\x4e\xa7\xf4\x86\x96\x5c\x72\xcb\x95\x1c\xe5\x64\x91\xa1\xc5\xe9\x08\x00\xa5\x54\x16\xfd\xb0\xf1\x9f\x00\xa9\x92\x56\x2b\x21\x48\x8f\x57\x24\x93\xb5\x5b\xd0\xc2\x71\xc1\x48\x07\xe0\x15\xea\xc7\x9f\x92\x5f\x93\x9f\xc2\x8e\x0b\x2c\xf8\x18\x8b\x42\xab\x47\x62\x61\x83\x96\x64\xc9\x13\x73\x31\x85\x8b\xcc\xda\xc2\x4c\x27\x93\x15\xb7\x99\x5b\x24\xa9\xca\x27\xdb\x35\xcd\xbf\x85\x13\x62\xf2\xdb\xeb\x5f\x5f\xbd\xbe\x18\x01\xa4\x9a\x02\x59\x0f\x3c\x27\x63\x31\x2f\xa6\x20\x9d\x10\x23\x00\x89\x39\x4d\x21\xc7\x34\xe3\x92\x4c\xc2\xe8\x51\x15\x15\xf7\x23\x53\x50\xea\x19\x59\x69\xe5\x8a\x29\xb4\x27\xe3\xde\x92\xd1\x28\xa4\xbb\x08\x26\x8c\x08\x6e\xec\x87\xe6\xe8\x47\x6e\x6c\x98\x29\x84\xd3\x28\xb6\x48\xc3\xa0\xc9\x94\xb6\x9f\xb6\x00\xc7\x90\xa7\x71\x82\xcb\x95\x13\xa8\xeb\xf5\x23\x00\x93\xaa\x82\xa6\x10\x96\x17\x98\x12\x1b\x01\x94\xc2\x0c\xdb\xc7\x80\x8c\x05\xf5\xa0\x98\x69\x2e\x2d\xe9\x1b\x25\x5c\x2e\x6b\xe0\x8c\x4c\xaa\x79\x61\x83\xf8\x1f\x32\x82\x54\x38\x4b\x1a\x8a\x0c\x0d\x25\x61\x11\xc0\x1f\x46\xc9\x19\xda\x6c\x0a\x89\xb1\x68\x9d\x49\xc2\x74\x39\x1b\x25\x37\x7b\x7f\x3d\x7f\x5b\x8e\xd8\x8d\xa7\xca\x58\xcd\xe5\xaa\x0b\xcf\xe5\xcd\xae\x1a\x80\x1b\x40\xb0\xf5\xa7\xa6\x42\x93\x21\x69\xb9\x5c\x81\xcd\x08\x0c\xe9\x47\xd2\x61\x45\x89\x04\xe0\x29\x23\x09\x36\xe3\x06\xd4\xe2\x0f\x4a\x2d\x3c\xa1\x89\x1a\x26\x96\xc0\xe5\x3e\xf1\x95\x89\x26\x7b\x66\xd0\x62\xe5\xfa\x9f\x6d\x46\x18\xda\x88\x34\x4e\x3f\xfe\x1c\xf5\x91\x66\x94\xe3\xb4\x5c\xa9\x0a\x92\xd7\xb3\xdb\x2f\xbf\xcc\x5b\xc3\xd0\x66\xbc\xb4\x00\xcf\xad\x67\x2a\xae\x85\xa5\xd2\xe1\xb3\x9a\xbd\x9e\xdd\xd6\xdb\x0b\xad\x0a\xd2\x96\x57\xe6\x10\x7f\x0d\x77\x6d\x8c\xee\x20\xbb\xf4\xf4\xc4\x55\xc0\xbc\x9f\x52\xc4\x5a\x1a\x08\xb1\x92\x05\x50\xcb\x28\xc5\x5a\xe8\x41\x36\x2d\xc0\xe0\x17\xa1\x2c\x05\x9d\xc0\x3c\xa8\xc3\x78\x6b\x75\x82\x79\xf7\x7e\x24\x6d\x41\x53\xaa\x56\x92\xff\x59\xc3\x36\x60\x55\x40\x2a\xd0\x52\x69\xf6\xdb\x5f\x30\x48\x89\x02\x1e\x51\x38\xba\x02\x94\x0c\x72\xdc\x80\xa6\xa0\x4e\x27\x1b\xf0\xc2\x12\x93\xc0\x9d\xd2\x04\x5c\x2e\xd5\x14\x1a\x31\xa0\x0a\x53\xa9\xca\x73\x27\xb9\xdd\x4c\x42\xc4\xe1\x0b\x67\x95\x36\x13\x46\x8f\x24\x26\x86\xaf\xc6\xa8\xd3\x8c\x5b\x4a\xad\xd3\x34\xf1\x21\x26\x90\x2e\x43\xa8\x4a\x72\xf6\x17\x5d\x06\x36\x73\xd9\xa2\x75\xcf\xa2\xe3\x2f\x78\xfb\x80\x06\xbc\xdf\x47\xd3\x8e\x5b\x23\x17\xfb\xd6\x7d\xff\x76\xfe\x00\x15\xea\xa0\x8c\x5d\xe9\x47\x03\xaf\x37\x9a\xad\x0a\xbc\xc0\xb8\x5c\x7a\xe7\xf0\x4a\x5c\x6a\x95\x07\x98\x24\x59\xa1\xb8\xb4\xe1\x23\x15\x9c\xe4\xae\xf8\x8d\x5b\xe4\xdc\x7a\xbd\x7f\x73\x64\xac\xd7\x55\x02\x37\x21\x76\xc3\x82\xc0\x15\x2c\x7a\xd2\xad\x84\x1b\xcc\x49\xdc\xf8\x90\xf0\xd2\x0a\xf0\x92\x36\x63\x2f\xd8\xe3\x54\xd0\x3c\x76\x76\x17\x47\xa9\x35\x26\xaa\x38\xde\xa3\xaf\xd2\x01\xe7\x05\xa5\x51\x6b\x8d\x59\xef\x00\x65\xe0\x4d\x5a\x10\xba\x3d\x34\x1c\x7a\xc2\x19\x4b\xda\x47\xe7\xdd\xa9\x5e\x76\xfc\x6f\x49\xe8\xa5\xb3\xbf\xa7\x1f\x55\xd8\xc6\x45\xf7\x04\x00\xb7\x94\xf7\x4c\x1d\x82\x5a\x8a\xc9\xd8\xfe\xc9\x41\x66\x1a\xc2\xd7\xe9\x0f\xc2\xf0\x86\xca\x35\xb1\x3e\x30\x63\x4f\x67\xef\x9c\xd1\xe9\x68\x08\xf5\x9e\xb5\xec\x2e\x40\xad\x71\xd3\x31\x9f\x29\xb5\xee\x91\x5d\xf3\xf8\x3d\x24\xe5\x83\x02\x38\x40\xa6\x59\xf3\xe2\x46\xc9\x88\xf0\x1c\x43\x38\x92\x80\x6e\x31\x0c\x10\xb7\xe4\x12\x05\xff\x93\x74\x07\xe6\x96\xff\xbd\xab\x17\x06\xf7\x93\xa0\x0a\xfc\xe6\x28\xa4\x50\xde\xff\xe2\x19\x00\x36\x43\x0b\xb9\x33\x21\x4a\x51\x5e\xd8\x2e\xa5\x58\x05\x05\xe9\x1c\x25\x49\x2b\xfc\x91\x92\xab\x47\xaa\xe2\x68\x08\x92\xc6\x2a\x8d\xab\x1d\x6f\x1e\x14\x52\x37\xb1\xde\xbf\xab\x13\x5d\x86\xff\xcc\xc7\xb3\xe5\xc6\x47\x77\xdc\x72\x0f\xcc\xf5\x4a\xb6\x0c\x15\x20\xf8\x92\xd2\x4d\x2a\x3a\xa8\x3a\xa0\x9f\x7e\xdd\x94\x51\xeb\x80\xec\x6f\x22\x05\x3b\x19\x4a\x8e\x81\xac\x12\x44\x4c\x23\x78\x15\x0e\x4b\xa2\x93\x13\xe3\x14\x2f\xa6\xa3\x33\xcc\x4f\xe0\x82\xc4\xff\x81\x9b\x15\x68\xcc\x2c\xd3\x68\xa8\x1b\xc3\x52\xe9\x1c\xed\x14\x16\x1b\x4b\xe7\xf0\xe9\xe1\x3f\x29\xcd\xce\x12\x52\xa1\xb4\x1d\x26\x8b\x4b\xfb\xcb\xdf\x07\x40\xfb\x9c\x6c\x45\xba\x0b\xb6\xe6\x8f\x68\xe9\x03\x6d\x5e\x86\x71\x8b\x5c\xda\x1e\xb5\xb5\x4c\xf5\x76\x19\x0e\x72\xbe\xe4\xc4\xae\xa2\xdb\x29\x46\x97\xa6\x84\x90\x9c\x1e\xf9\xf6\xaa\x20\x0f\x30\xe6\x53\x0f\x1e\x66\x08\x47\xd6\x62\x9a\x11\xf3\x91\x25\xc3\xe8\x1e\x17\xb4\x5c\x52\x6a\x2f\x7a\x8f\x35\x25\x01\xe5\x06\x0a\xc5\x62\xd4\x62\x8a\x0c\xf8\xfc\xca\x2a\x41\x1a\x2d\x05\x30\x01\x47\xf2\x03\xc7\x73\x24\x63\xe8\x74\x6d\x71\x78\x5f\x9e\xa3\x49\xe0\x35\x6e\x8e\x55\x00\x45\x19\x7a\xba\x0b\xc5\x62\xa8\x1d\x80\x0a\xc0\xd4\x3e\x3b\x01\x44\x02\x5f\x50\x70\x56\x42\x37\x80\x9a\xe0\x93\xf2\x15\x0f\x73\x82\xae\x06\x81\xce\x34\x2d\x49\x6f\x57\x87\xc2\xe0\x93\x7a\xfb\x9d\x52\x67\x29\xf9\xd1\x44\x64\xdd\x67\xc1\x07\x45\x15\x85\xb3\xa6\x8d\x37\x82\x05\x01\x16\x85\xe0\xd1\x24\x70\x90\x23\x6f\x4f\x3f\x4c\xb7\x2f\x7e\xaf\x19\xeb\xcf\x7f\xf6\x4d\xb9\xda\xd1\xa8\x1c\xa2\x8a\x78\x4e\x80\x16\x9e\x32\x9e\x66\x7e\x64\x90\xfa\xc8\xb6\xaf\xae\xd1\x03\x4b\xe0\x36\x78\x84\x92\x62\x03\x4f\x9a\x5b\x4b\x32\x14\xb1\xb5\x8a\x06\x3d\xb1\x1d\x2d\x7c\x8d\x31\x6e\x95\xf5\x67\x4a\x27\x24\x07\xc7\x4b\xa6\xd6\x66\x2c\xc9\x52\xa5\x35\x99\xc2\xa7\x4f\xbe\x26\x53\x5b\x43\x1e\x94\xcc\x9a\x36\xc9\x4b\xe7\xb4\xd1\x83\x7a\xa7\xd7\xb4\x79\x99\xb4\xd6\x19\x5f\x9c\xe7\x74\xc6\x41\xd4\xcf\xd4\x18\x78\xd1\x31\xe8\xcf\xad\x8e\xe1\x8a\x84\x53\xb2\xcd\x02\x9d\xe9\xad\xb7\x16\x4a\x09\xc2\xdd\xde\x86\x25\x89\xd2\xde\xbe\x39\xa9\x4a\x0b\x53\xc7\x6f\xe8\x16\xc9\xb8\x59\x24\xee\xcc\x78\x50\x47\x15\xb5\xa1\x25\x77\x44\x59\x1b\xd6\x35\x23\x81\xaf\xe2\xbd\x17\xfa\x7c\x0e\x17\xca\xc5\x5e\x41\x84\x07\x6a\xb9\xc3\x1c\xca\x53\x0b\x60\x64\x4c\x93\x31\x74\x28\xef\xff\x58\xe6\xf7\xf5\x7a\xd0\x84\x69\x86\x0b\x41\x95\x2b\x76\x62\x3e\x3e\x59\x2f\x45\x70\x1d\x11\x84\x76\x34\x72\xd9\x96\x40\xd5\x86\x2b\x51\x5d\x9a\xbe\x54\xd3\x83\x48\x46\xa7\x9f\xd4\xe5\xd6\xa3\x93\x90\x2a\xeb\x1e\x40\x79\x7c\x42\x7b\x0c\xd2\xbb\x36\xc2\xb0\xf1\x0a\x94\x24\xaf\x9c\x99\x5b\x08\x9e\x5e\xc1\xdb\xef\xb1\x69\x77\x3b\xeb\xcf\x7a\x34\xdc\xca\x6a\xd5\x99\x64\x0f\xc5\xc5\x71\x45\x61\xe7\xdc\x9e\xdf\x1c\x11\x0d\xfb\x23\x61\x3a\x50\x51\x9f\x64\x7b\x75\x69\xbe\xb5\x3e\x46\x16\xb9\x30\xb5\xe5\xa5\x4e\x6b\x92\x76\x8b\xb3\x53\x74\x55\xbb\xf6\xae\xcf\x25\x0e\x5b\xa2\x40\x63\x67\x5a\x2d\xc8\x27\x08\x47\x99\xc6\x47\x34\x36\x66\x0d\x4f\xe4\xc1\x2f\x7c\xd6\xe3\x49\xae\x48\xed\x53\xf3\xb1\xe7\xfc\x41\x2b\xf6\x34\x3f\x68\x94\x86\x57\x9d\xfb\x13\x09\x6f\x91\x0b\xb6\x06\x45\x2c\xf6\x03\xbc\x9d\xc7\xd8\xd7\x9f\x81\x29\x40\xa9\x6c\xd6\x55\xf4\x3e\x33\xbb\x39\x19\x83\xab\xe3\x78\x7c\xef\x72\x94\x63\x4d\xc8\x42\xc8\x2c\xb7\x02\x97\x8c\xa7\x18\x9a\xcc\x95\xa5\x85\x28\xdf\xcb\x9e\x08\xb2\xaa\x05\x73\x76\xc0\xd1\x84\x66\xf7\x6a\xa2\x87\xf4\xcf\x92\x7f\x73\x31\xc8\x8c\x7d\xd5\x7b\xb5\x6d\x35\x97\x60\xb6\xde\x51\xe9\xee\xd2\xbc\x38\x07\x5d\x67\x6a\x0f\x07\xe5\xb1\x5a\x36\x4c\xea\xc3\x73\xc7\x3b\xe0\x06\xa5\xaf\x18\x1e\xb4\x1b\x28\x7e\xde\xa1\x30\x74\x05\x9f\xe5\x5a\xaa\x27\xf9\xf2\x01\xff\x61\x53\xd4\xad\x1e\xbf\x69\x9f\xee\x97\x08\xde\xbd\x4e\xf6\xdc\xb1\x5b\xa8\x74\xdd\x45\xc4\x50\x2e\x58\x1e\xba\xb7\x72\xa9\x0e\x64\x2d\x73\x0a\x49\x0b\x67\x66\xe2\x1c\x67\xe1\xaa\xcb\x05\x73\x16\x9b\xba\x07\x58\xb7\x27\x4e\xed\x92\x35\xef\x49\x8e\xe8\x89\xf8\x7c\xe1\xba\xb1\xc5\xa7\x79\x4a\x5b\x62\xb0\xd8\xd2\x70\x4e\x57\x66\xa1\x54\x67\x66\xbc\x47\xc1\x3f\x94\xb2\x70\xfb\xa6\x13\x71\x72\x0e\xe6\xf2\x98\x24\x7d\xef\xa4\x8f\xa4\x9d\x37\x9e\xdd\xbd\xcc\x9d\x9d\x50\x5d\x83\x3e\x1b\x6d\x6b\xd2\x92\xc4\xf1\x14\x7d\x08\xeb\x5f\x80\x0e\xb7\xa0\x99\x56\xdf\x37\x27\x90\x52\x6d\x79\x19\x6a\x04\xd9\xd3\x68\x11\x64\x9f\x9f\x92\xca\x8b\x8f\x31\xdc\xcb\xbb\x6a\x71\x37\x7e\x78\xa7\x74\xe9\xd8\x8d\xa7\x17\x9d\x3d\xc6\xe8\xf4\xe1\xd0\x55\x12\xb8\x2c\xef\x5e\x63\x6f\x3f\x5e\xcf\x72\x12\xe1\x4a\xb8\x08\x3d\xae\xd0\x59\xfa\x48\xa8\x65\x0f\xc8\x5c\x69\x8a\xe9\x49\x8e\xf2\xaf\xaf\xfe\x56\x51\x30\xe6\x2c\xde\xbf\x4e\x27\x93\x1c\xe5\x6f\x89\xd2\xab\x89\xe0\xd2\x7d\xf7\x9f\xe3\x02\x57\x64\xfc\xbf\x57\x93\xed\x86\xe4\x55\x92\xd9\x5c\x5c\x9e\x23\x50\x1f\xaa\x42\x2a\x31\xdf\x18\x4b\xf9\x91\x11\xe9\xf7\x6a\x17\xc4\x6d\xcf\x16\x95\x94\xb9\xcd\x7b\xb3\xa3\x16\x19\xbf\xcf\x21\x2c\x7d\x3e\xdb\x32\x81\x95\xcf\x9f\x8f\x32\xae\x79\xbd\xf8\x99\x8d\x6b\x6b\xb4\x6d\x63\x7a\x68\x59\x59\xd9\x27\xef\xbd\xf8\x54\x70\x4f\x0c\xde\xa3\x85\x4c\x19\x6b\xea\x1b\x7d\x4c\x53\x5f\x71\x6a\x62\x19\xda\xf0\xba\x8a\xa9\xd4\xe5\xd5\xdb\x90\x09\xc9\xf1\xe7\xf9\xe4\x9e\xd8\xd7\xf7\x68\xbf\xce\xdd\xa2\x66\xf9\xeb\x1d\x4a\x5c\x91\x5f\x3a\xf9\x79\xe2\xed\x6d\x72\xff\x7e\x7e\x37\x59\x91\xf5\x86\x30\x8e\xd2\x1b\xfb\x13\x33\x58\xe3\xe9\x1a\x18\x48\x06\x7a\x93\xe6\x96\x4e\xae\x21\xf3\x09\x33\x1c\x9d\x30\xc3\x53\xd6\x79\xc5\xd8\xa8\xd1\xb9\x89\xee\xce\xcd\x50\xf2\x34\xc0\x57\x78\x51\x75\x80\xf0\x52\xe7\x33\xbf\xb4\xf5\xa4\x27\x6c\x6e\xbc\x50\xf0\x34\x18\xab\x5d\x6a\x95\x3e\x85\x88\xbe\xc4\x7d\x47\x7c\x0b\xcd\x69\xd9\x48\xd4\x9f\x57\x7e\x3e\x3d\xa4\x13\x64\xd7\x69\x0f\x7b\x83\xe1\x01\x19\x9b\x82\xd5\x2e\x7a\x58\x79\xfd\xdb\x1c\x71\x8b\xfa\xf9\x4f\x25\x83\xb2\x10\x80\xff\xfc\x77\xb4\xad\x09\xbc\x77\x14\x96\xd8\xa7\xdd\x77\x7f\x17\x17\xad\x87\x7d\xe1\xb3\xd1\x40\x80\x7f\xfd\x7b\x14\x11\x13\xfb\x52\x3d\xd3\xf3\x83\xff\x0b\x00\x00\xff\xff\xce\x1e\x54\x69\x84\x29\x00\x00"),
		},
		"/workload.k8s.io_addons.yaml": &vfsgen۰CompressedFileInfo{
			name:             "workload.k8s.io_addons.yaml",
			modTime:          time.Date(2021, 1, 20, 3, 33, 19, 879922914, time.UTC),
			uncompressedSize: 2364,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\xb4\x55\xcd\x8e\xdb\x36\x10\xbe\xeb\x29\x06\xee\x61\x2f\xb5\x9c\xa0\x29\x1a\xe8\x66\xb8\x41\xbb\x68\x13\x18\xf1\x62\x2f\x45\x0f\x14\x39\x96\x98\xa5\x48\x96\x33\xf4\x76\x5b\xf4\xdd\x0b\x92\x92\x57\x96\xb3\x6d\x81\x22\x3a\x89\x33\x43\xce\xcc\xf7\xcd\x4f\xb5\x5e\xaf\x2b\xe1\xf5\x3d\x06\xd2\xce\x36\x20\xbc\xc6\xdf\x19\x6d\x3a\x51\xfd\xf0\x96\x6a\xed\x36\xa7\xd7\xd5\x83\xb6\xaa\x81\x5d\x24\x76\xc3\x47\x24\x17\x83\xc4\xef\xf1\xa8\xad\x66\xed\x6c\x35\x20\x0b\x25\x58\x34\x15\x80\xb0\xd6\xb1\x48\x62\x4a\x47\x00\xe9\x2c\x07\x67\x0c\x86\x75\x87\xb6\x7e\x88\x2d\xb6\x51\x1b\x85\x21\x3f\x3e\xb9\x3e\xbd\xaa\xdf\xd4\xaf\xf2\x8d\x95\xf0\x7a\x2d\xbc\x0f\xee\x84\x2a\x5f\x08\x16\x19\x53\x30\xab\x06\x56\x3d\xb3\xa7\x66\xb3\xe9\x34\xf7\xb1\xad\xa5\x1b\x36\xcf\x36\xf3\x5f\x1f\x8d\xd9\x7c\xf7\xf6\xcd\xb7\x6f\x57\x15\x80\x0c\x98\xc3\xba\xd3\x03\x12\x8b\xc1\x37\x60\xa3\x31\x15\x80\x15\x03\x36\x20\x94\x4a\x49\x3f\xba\xf0\x60\x9c\x50\x63\xf6\x15\x79\x94\x29\x91\x2e\xb8\xe8\x1b\x58\xaa\xcb\xed\x31\xd5\x02\xd3\x36\x3f\x94\x05\x46\x13\xff\x34\x13\xfe\xac\x89\xb3\xc2\x9b\x18\x84\x99\x9c\x66\x11\x69\xdb\x45\x23\xc2\x4c\x48\xd2\x79\x6c\xe0\x43\xf2\xe0\x85\x44\x55\x01\x8c\x80\x65\x8f\xeb\x64\x9a\x29\x10\x66\x1f\xb4\x65\x0c\x3b\x67\xe2\x30\x41\xbf\x06\x85\x24\x83\xf6\x9c\x21\xbe\xeb\x71\x8c\x03\x7c\x2f\x08\xeb\x6c\x04\xf0\x89\x9c\xdd\x0b\xee\x1b\xa8\x89\x05\x47\xaa\xb3\x7a\xd4\x16\x74\xf6\x3f\x6e\x0f\xef\x46\x09\x3f\xa5\xa8\x88\x83\xb6\xdd\xe7\xfc\xdc\xec\x96\x50\x83\x26\x10\xc0\xe7\x63\x40\x1f\x90\xd0\xb2\xb6\x1d\x70\x8f\x40\x18\x4e\x18\xb2\xc5\xe8\x04\xe0\xb1\x47\x0b\xdc\x6b\x02\xd7\x7e\x42\xc9\xf0\x28\xa8\xb0\x88\xaa\x86\x9b\xeb\xe0\xa7\x32\xac\xaf\xa8\xbe\x48\x65\xfb\xc3\x65\x22\x4a\x70\x71\x5a\xd4\xa7\xd7\x85\x0e\xd9\xe3\x20\x9a\xd1\xd2\x79\xb4\xdb\xfd\xed\xfd\x37\x87\x0b\x31\x5c\x26\xfe\x5e\xc8\x5e\x5b\x4c\xd9\xa6\xa4\x8a\x2d\x1c\x5d\xc8\xc7\x49\xbb\xdd\xdf\x9e\xaf\xfb\xe0\x3c\x06\xd6\x53\x05\x95\x6f\xd6\x92\x33\xe9\xc2\xd9\x4d\x8a\xa7\x58\x81\x4a\xbd\x88\xc5\xeb\x58\x20\xa8\xc6\x14\xc0\x1d\x0b\x8a\x67\xd0\x33\x36\x17\x0f\x43\x32\x12\x76\x04\xba\x86\x43\xa6\x83\x80\x7a\x17\x8d\x4a\x2d\x7c\xc2\xc0\x10\x50\xba\xce\xea\x3f\xce\x6f\x13\xb0\xcb\x4e\x8d\x60\x1c\x4b\xfb\xf9\xcb\x05\x69\x85\x81\x93\x30\x11\xbf\x06\x61\x15\x0c\xe2\x09\x02\x66\x3a\xa3\x9d\xbd\x97\x4d\xa8\x86\xf7\x2e\x20\x68\x7b\x74\x0d\xcc\xfa\x7c\x1a\x45\xd2\x0d\x43\xb4\x9a\x9f\x36\x79\xaa\xe8\x36\xb2\x0b\xb4\x51\x78\x42\xb3\x21\xdd\xad\x45\x90\xbd\x66\x94\x1c\x03\x6e\xd2\x18\xc9\xa1\xdb\x3c\x8e\xea\x41\x7d\x15\xc6\xe1\x45\x37\x17\xb1\x5e\x55\x74\xf9\x72\x3f\xff\x03\x03\xa9\xb5\x4b\x69\x97\xab\x25\x8b\xeb\xea\xfe\xf8\xee\x70\x07\x93\xeb\x4c\xc6\x12\xfd\x52\xe0\xe7\x8b\xf4\x4c\x41\x02\x4c\xdb\x63\x6a\x8e\x44\xe2\x31\xb8\x21\xbf\x89\x56\x79\xa7\x2d\xe7\x83\x34\x1a\xed\x12\x7e\x8a\xed\xa0\x39\xf1\xfe\x5b\x44\xe2\xc4\x55\x0d\xbb\x3c\x9f\xa1\x45\x88\x5e\x95\x4e\xba\xb5\xb0\x13\x03\x9a\x5d\x1a\x09\x5f\x9a\x80\x84\x34\xad\x13\xb0\xff\x8d\x82\xf9\x6a\x59\x1a\x17\xd4\x66\x8a\x69\x56\xbf\xc0\x57\x99\x7d\x07\x8f\xb2\x90\x36\x53\xa6\xfa\x2f\xea\xfa\xe2\xfe\xe7\xfb\x33\x7d\x47\xe7\x96\xa2\x17\x93\x78\x39\xe0\x3c\x6e\xff\x3d\xe4\x6c\x36\x2b\x8f\x4c\x50\x18\x72\x23\x83\x68\x5d\x2c\x65\x50\x9e\x2b\xbd\xbc\x88\xed\x4b\x26\x07\x65\x9d\xfc\x5f\x38\xae\x84\x79\x29\xa8\x06\x38\xc4\x32\xa0\x89\x5d\x10\x1d\xce\x25\xb1\x3d\xb7\xf4\xe4\x7f\x04\x15\xfe\xfc\xab\x7a\xc6\x57\x48\x89\x9e\x51\x7d\x58\x6e\xeb\xd5\xea\x62\x21\xe7\xa3\x74\xb6\xac\x55\x6a\xe0\x97\x5f\xab\xe2\x18\xd5\xfd\xb4\x7a\x93\xf0\xef\x00\x00\x00\xff\xff\x96\x12\xae\x8b\x3c\x09\x00\x00"),
		},
	}
	fs["/"].(*vfsgen۰DirInfo).entries = []os.FileInfo{
		fs["/devops.k8s.io_clustercredentials.yaml"].(os.FileInfo),
		fs["/devops.k8s.io_clusters.yaml"].(os.FileInfo),
		fs["/devops.k8s.io_machines.yaml"].(os.FileInfo),
		fs["/workload.k8s.io_addons.yaml"].(os.FileInfo),
	}

	return fs
}()

type vfsgen۰FS map[string]interface{}

func (fs vfsgen۰FS) Open(path string) (http.File, error) {
	path = pathpkg.Clean("/" + path)
	f, ok := fs[path]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}

	switch f := f.(type) {
	case *vfsgen۰CompressedFileInfo:
		gr, err := gzip.NewReader(bytes.NewReader(f.compressedContent))
		if err != nil {
			// This should never happen because we generate the gzip bytes such that they are always valid.
			panic("unexpected error reading own gzip compressed bytes: " + err.Error())
		}
		return &vfsgen۰CompressedFile{
			vfsgen۰CompressedFileInfo: f,
			gr:                        gr,
		}, nil
	case *vfsgen۰DirInfo:
		return &vfsgen۰Dir{
			vfsgen۰DirInfo: f,
		}, nil
	default:
		// This should never happen because we generate only the above types.
		panic(fmt.Sprintf("unexpected type %T", f))
	}
}

// vfsgen۰CompressedFileInfo is a static definition of a gzip compressed file.
type vfsgen۰CompressedFileInfo struct {
	name              string
	modTime           time.Time
	compressedContent []byte
	uncompressedSize  int64
}

func (f *vfsgen۰CompressedFileInfo) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("cannot Readdir from file %s", f.name)
}
func (f *vfsgen۰CompressedFileInfo) Stat() (os.FileInfo, error) { return f, nil }

func (f *vfsgen۰CompressedFileInfo) GzipBytes() []byte {
	return f.compressedContent
}

func (f *vfsgen۰CompressedFileInfo) Name() string       { return f.name }
func (f *vfsgen۰CompressedFileInfo) Size() int64        { return f.uncompressedSize }
func (f *vfsgen۰CompressedFileInfo) Mode() os.FileMode  { return 0444 }
func (f *vfsgen۰CompressedFileInfo) ModTime() time.Time { return f.modTime }
func (f *vfsgen۰CompressedFileInfo) IsDir() bool        { return false }
func (f *vfsgen۰CompressedFileInfo) Sys() interface{}   { return nil }

// vfsgen۰CompressedFile is an opened compressedFile instance.
type vfsgen۰CompressedFile struct {
	*vfsgen۰CompressedFileInfo
	gr      *gzip.Reader
	grPos   int64 // Actual gr uncompressed position.
	seekPos int64 // Seek uncompressed position.
}

func (f *vfsgen۰CompressedFile) Read(p []byte) (n int, err error) {
	if f.grPos > f.seekPos {
		// Rewind to beginning.
		err = f.gr.Reset(bytes.NewReader(f.compressedContent))
		if err != nil {
			return 0, err
		}
		f.grPos = 0
	}
	if f.grPos < f.seekPos {
		// Fast-forward.
		_, err = io.CopyN(ioutil.Discard, f.gr, f.seekPos-f.grPos)
		if err != nil {
			return 0, err
		}
		f.grPos = f.seekPos
	}
	n, err = f.gr.Read(p)
	f.grPos += int64(n)
	f.seekPos = f.grPos
	return n, err
}
func (f *vfsgen۰CompressedFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.seekPos = 0 + offset
	case io.SeekCurrent:
		f.seekPos += offset
	case io.SeekEnd:
		f.seekPos = f.uncompressedSize + offset
	default:
		panic(fmt.Errorf("invalid whence value: %v", whence))
	}
	return f.seekPos, nil
}
func (f *vfsgen۰CompressedFile) Close() error {
	return f.gr.Close()
}

// vfsgen۰DirInfo is a static definition of a directory.
type vfsgen۰DirInfo struct {
	name    string
	modTime time.Time
	entries []os.FileInfo
}

func (d *vfsgen۰DirInfo) Read([]byte) (int, error) {
	return 0, fmt.Errorf("cannot Read from directory %s", d.name)
}
func (d *vfsgen۰DirInfo) Close() error               { return nil }
func (d *vfsgen۰DirInfo) Stat() (os.FileInfo, error) { return d, nil }

func (d *vfsgen۰DirInfo) Name() string       { return d.name }
func (d *vfsgen۰DirInfo) Size() int64        { return 0 }
func (d *vfsgen۰DirInfo) Mode() os.FileMode  { return 0755 | os.ModeDir }
func (d *vfsgen۰DirInfo) ModTime() time.Time { return d.modTime }
func (d *vfsgen۰DirInfo) IsDir() bool        { return true }
func (d *vfsgen۰DirInfo) Sys() interface{}   { return nil }

// vfsgen۰Dir is an opened dir instance.
type vfsgen۰Dir struct {
	*vfsgen۰DirInfo
	pos int // Position within entries for Seek and Readdir.
}

func (d *vfsgen۰Dir) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == io.SeekStart {
		d.pos = 0
		return 0, nil
	}
	return 0, fmt.Errorf("unsupported Seek in directory %s", d.name)
}

func (d *vfsgen۰Dir) Readdir(count int) ([]os.FileInfo, error) {
	if d.pos >= len(d.entries) && count > 0 {
		return nil, io.EOF
	}
	if count <= 0 || count > len(d.entries)-d.pos {
		count = len(d.entries) - d.pos
	}
	e := d.entries[d.pos : d.pos+count]
	d.pos += count
	return e, nil
}
