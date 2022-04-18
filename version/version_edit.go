package version

import "SimpleKV/sstable"

type VersionEdit struct {
	//logNumber      uint64
	//logNumber      uint64
	//prevFileNumber uint64
	//nextFileNumber uint64

	deletes []*TableMeta
	adds    []*TableMeta
}

type TableMeta struct {
	f     *FileMetaData
	level int
}

func NewVersionEdit() *VersionEdit {
	return &VersionEdit{
		deletes: make([]*TableMeta, 0),
		adds:    make([]*TableMeta, 0),
	}
}

func (ve *VersionEdit) AddFile(level int, t *sstable.Table) {
	fm := &FileMetaData{
		id:       t.Fid(),
		largest:  t.MaxKey,
		smallest: t.MinKey,
	}
	ve.adds = append(ve.adds, &TableMeta{f: fm, level: level})
}