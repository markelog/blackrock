package disk

import (
	"encoding/binary"
	"io/ioutil"
	"os"
	"path"

	"github.com/jackdoe/blackrock/depths"
	log "github.com/sirupsen/logrus"
)

type InvertedWriter struct {
	descriptors        map[string]*os.File
	maxOpenDescriptors int
	root               string
}

func NewInvertedWriter(root string, maxOpenDescriptors int) (*InvertedWriter, error) {
	return &InvertedWriter{
		maxOpenDescriptors: maxOpenDescriptors,
		descriptors:        map[string]*os.File{},
		root:               root,
	}, nil
}
func (fw *InvertedWriter) Close() {
	for k, v := range fw.descriptors {
		v.Close()
		delete(fw.descriptors, k)
	}
}
func (fw *InvertedWriter) Read(maxDocuments int64, tagKey uint64, tagValue string) []int64 {
	dir, filename := depths.PathForTag(fw.root, tagKey, tagValue)
	fn := path.Join(dir, filename)
	file, err := os.OpenFile(fn, os.O_RDONLY, 0600)
	if os.IsNotExist(err) {
		log.Infof("missing file %s, returning empty", fn)
		return []int64{}
	}
	fi, err := file.Stat()
	if err != nil {
		log.Warnf("failed to read file stats: %s, error: %s", fn, err.Error())
		return []int64{}
	}
	// read to the closest multiple of 8
	log.Infof("reading %s, size: %d", fn, fi.Size())
	total := fi.Size() / int64(8)
	seek := int64(0)
	if maxDocuments > 0 && total > maxDocuments {
		seek = (total - maxDocuments) * 8
	}
	file.Seek(seek, 0)
	log.Infof("seek %d, total %d max requested: %d", seek, total, maxDocuments)

	postings, err := ioutil.ReadAll(file)
	if err != nil {
		log.Warnf("failed to read file: %s, error: %s", fn, err.Error())
		return []int64{}
	}
	n := len(postings) / 8
	longed := make([]int64, n)
	j := 0
	for i := 0; i < n*8; i += 8 {
		longed[j] = int64(binary.LittleEndian.Uint64(postings[i:]))
		j++
	}
	return longed
}

func (fw *InvertedWriter) Append(docId int64, tagKey uint64, tagValue string) error {
	dir, fn := depths.PathForTag(fw.root, tagKey, tagValue)
	filename := path.Join(dir, fn)
	f, ok := fw.descriptors[filename]
	if !ok {
		if len(fw.descriptors) > fw.maxOpenDescriptors {
			log.Warnf("clearing descriptor cache len: %d", len(fw.descriptors))
			for dk, fd := range fw.descriptors {
				fd.Close()
				delete(fw.descriptors, dk)
			}
		}

		log.Infof("openning %s", filename)
		os.MkdirAll(dir, 0700)
		fd, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		f = fd
		fw.descriptors[filename] = fd
	}
	log.Infof("writing document id %d at %s", docId, filename)
	data := make([]byte, 8)

	binary.LittleEndian.PutUint64(data, uint64(docId))
	_, err := f.Write(data)
	if err != nil {
		return err
	}
	return nil
}