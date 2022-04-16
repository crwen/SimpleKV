package lsm

import (
	"SimpleKV/file"
	"SimpleKV/utils"
	"SimpleKV/utils/codec"
	"SimpleKV/utils/convert"
	"SimpleKV/utils/errs"
	"bufio"
	"bytes"
	"io"
	"os"
	"sync"
)

const walFileExt string = ".wal"

// Wal
type WalFile struct {
	f    *file.MmapFile
	lock *sync.RWMutex

	opt     *file.Options
	buf     *bytes.Buffer
	writeAt uint32
	size    uint32
}

type WalHeader struct {
	checksum uint32
	keyLen   uint16
	ValueLen uint16
	types    uint8
}

// OpenWalFile _
func OpenWalFile(opt *file.Options) *WalFile {
	omf, err := file.OpenMmapFile(opt.FileName, os.O_CREATE|os.O_RDWR, opt.MaxSz)
	wf := &WalFile{f: omf, lock: &sync.RWMutex{}, opt: opt}
	wf.buf = &bytes.Buffer{}
	wf.size = uint32(len(wf.f.Data))
	errs.Err(err)
	return wf
}

// Write
// +---------------------------------------------------+
// | checksum | key len | value len | type | key:value |
// +---------------------------------------------------+
func (wal *WalFile) Write(entry *utils.Entry) error {
	wal.lock.Lock()
	defer wal.lock.Unlock()

	h := WalHeader{}
	h.keyLen = uint16(len(entry.Key))
	h.ValueLen = uint16(len(entry.Key))
	// checksum + key len + value len + type = 4 + 2 + 2 + 1 = 9
	total := h.keyLen + h.ValueLen + 9

	buf := make([]byte, total)
	// write key len , write len and type
	copy(buf[4:6], convert.U16ToBytes(h.keyLen))
	copy(buf[6:8], convert.U16ToBytes(h.ValueLen))
	buf[8] = 1
	wal.buf.Bytes()
	// write key value
	copy(buf[9:9+len(entry.Key)], entry.Key)
	pos := 9 + len(entry.Key)
	copy(buf[pos:], entry.Value) // write value

	h.checksum = codec.CalculateU32Checksum(buf[4:])
	copy(buf[:4], convert.U32ToBytes(h.checksum)) // write checksum

	dst, err := wal.f.Bytes(int(wal.writeAt), int(total))
	if err != nil {
		return err
	}
	copy(dst, buf)
	wal.writeAt += uint32(total)
	return nil
}

func (wal *WalFile) Iterate(fn func(e *utils.Entry) error) (uint32, error) {
	wal.lock.Lock()
	defer wal.lock.Unlock()
	reader := bufio.NewReader(wal.f.NewReader(int(0)))

	//data := wal.f.Data
	//data, err := io.ReadAll(reader)
	//if err != nil {
	//	errs.Panic(err)
	//}
	for {
		buf := make([]byte, 9)
		if _, err := io.ReadFull(reader, buf); err != nil {
			break
		}
		h := WalHeader{}
		h.checksum = convert.BytesToU32(buf[0:4])
		h.keyLen = convert.BytesToU16(buf[4:6])
		h.ValueLen = convert.BytesToU16(buf[6:8])
		h.types = buf[8]
		b := make([]byte, h.keyLen+h.ValueLen)

		//io.ReadFull(reader, buf)
		if _, err := io.ReadFull(reader, b); err != nil {
			break
		}
		//total := 9 + h.keyLen + h.ValueLen
		//key := data[9 : 9+h.keyLen]
		key := b[:h.keyLen]
		//value := data[9+h.keyLen : total]
		value := b[h.keyLen : h.keyLen+h.ValueLen]
		//data = data[total:]
		err := fn(&utils.Entry{Key: key, Value: value})
		buf = append(buf, b...)
		if err := codec.VerifyU32Checksum(buf[4:], h.checksum); err != nil {
			break
		}
		if err != nil {
			break
		}
	}
	return 1, nil
}

func (wal *WalFile) Fid() uint64 {
	return wal.opt.FID
}

func (wal *WalFile) Close() error {
	filename := wal.f.Fd.Name()
	if err := wal.f.Close(); err != nil {
		return err
	}
	return os.Remove(filename)
}

func (wal *WalFile) Name() string {
	return wal.f.Fd.Name()
}

func (wal *WalFile) Size() uint32 {
	return wal.writeAt
}
