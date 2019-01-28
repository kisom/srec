package srec

import (
	"bytes"
	"encoding/binary"
	"io"
)

type record16 struct {
	Raw       []byte
	ByteCount uint8
	Address   uint16
	Data      []byte
	Checksum  byte
}

func newRecord16(address uint16, data []byte) *record16 {
	r := &record16{
		Address: address,
		Data:    data,
	}
	r.ByteCount = uint8(len(r.Data)) + 3

	// Compute the checksum
	checksum := uint32(r.ByteCount)
	var addr [2]byte
	binary.BigEndian.PutUint16(addr[:], r.Address)
	checksum += uint32(addr[0]) + uint32(addr[1])
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

func header16(h []byte) string {
	r := newRecord16(0, h)
	return encodeRecord(0, r.Raw)
}

func data16(data []byte) string {
	s := ""
	var addr uint16
	var count uint16

	for {
		l := min(len(data), 32)
		if l == 0 {
			break
		}

		r := newRecord16(addr, data[:l])
		s += encodeRecord(1, r.Raw)
		count++
		addr += uint16(l)
		data = data[l:]
	}

	r := newRecord16(uint16(count), []byte{})
	s += encodeRecord(5, r.Raw)

	return s
}

func terminator16(exec uint16) string {
	r := newRecord16(exec, []byte{})
	return encodeRecord(9, r.Raw)
}

// Dump16 returns an S19 16-bit record. Memory devices (such as EEPROMs)
// should use an exec address of 0.
func Dump16(header, data []byte, exec uint16) string {
	s := header16(header)
	s += data16(data)
	s += terminator16(exec)
	return s
}

// Dump16 copies the data from the reader to an S19 16-bit record in the
// writer. Memory devices (such as EEPROMs) should use an exec address of 0.
func Copy16(header []byte, exec uint16, r io.Reader, w io.Writer) error {
	_, err := w.Write([]byte(header16(header)))
	if err != nil {
		return err
	}

	address := uint16(0)
	count := uint16(0)
	data := make([]byte, 32)
	for {
		n, err := r.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		rec := newRecord16(address, data[:n])
		_, err = w.Write([]byte(encodeRecord(1, rec.Raw)))
		if err != nil {
			return err
		}
		address += uint16(n)
		count++
	}

	rec := newRecord16(uint16(count), []byte{})
	_, err = w.Write([]byte(encodeRecord(5, rec.Raw)))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(terminator16(exec)))
	if err != nil {
		return err
	}

	return nil
}
