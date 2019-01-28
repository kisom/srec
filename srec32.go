package srec

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
)

type record32 struct {
	Raw       []byte
	ByteCount uint8
	Address   uint32
	Data      []byte
	Checksum  byte
}

func newRecord32(address uint32, data []byte) *record32 {
	r := &record32{
		Address: address,
		Data:    data,
	}
	r.ByteCount = uint8(len(r.Data)) + 5

	// Compute the checksum
	checksum := uint32(r.ByteCount)
	var addr [4]byte
	binary.BigEndian.PutUint32(addr[:], r.Address)
	checksum += uint32(addr[0]) + uint32(addr[1])
	checksum += uint32(addr[2]) + uint32(addr[3])
	for i := range r.Data {
		checksum += uint32(r.Data[i])
	}

	r.Checksum = byte(checksum & 0xFF)
	r.Checksum = ^r.Checksum

	buf := &bytes.Buffer{}
	buf.WriteByte(r.ByteCount)
	buf.Write(addr[:])
	buf.Write(r.Data)
	buf.WriteByte(r.Checksum)
	r.Raw = buf.Bytes()

	return r
}

func data32(data []byte) string {
	s := ""
	var addr uint32
	var count uint32

	for {
		l := min(len(data), 32)
		if l == 0 {
			break
		}

		r := newRecord32(addr, data[:l])
		s += encodeRecord(3, r.Raw)
		count++
		addr += uint32(l)
		data = data[l:]
	}

	return s
}

func terminator32(exec uint32) string {
	r := newRecord32(exec, []byte{})
	return encodeRecord(7, r.Raw)
}

// Dump32 returns an S37 32-bit record. Memory devices (such as EEPROMs)
// should use an exec address of 0.
func Dump32(header, data []byte, exec uint32) string {
	s := header16(header)
	s += data32(data)
	s += terminator32(exec)
	return strings.ToUpper(s)
}

// Copy32 copies the data from the reader to an S37 32-bit record in the
// writer. Memory devices (such as EEPROMs) should use an exec address of 0.
func Copy32(header []byte, exec uint32, r io.Reader, w io.Writer) error {
	_, err := w.Write([]byte(header16(header)))
	if err != nil {
		return err
	}

	address := uint32(0)
	data := make([]byte, 32)
	for {
		n, err := r.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		rec := newRecord32(address, data)
		_, err = w.Write([]byte(encodeRecord(3, rec.Raw)))
		if err != nil {
			return err
		}
		address += uint32(n)
	}

	_, err = w.Write([]byte(terminator32(exec)))
	if err != nil {
		return err
	}

	return nil
}
