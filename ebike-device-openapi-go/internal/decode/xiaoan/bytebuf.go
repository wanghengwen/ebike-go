package xiaoan

import (
	"encoding/binary"
	"fmt"
	"math"
)

// ByteBuf simulates Java's Netty ByteBuf for sequential reading.
// On buffer underrun, the first error is stored and subsequent reads are no-ops.
type ByteBuf struct {
	data   []byte
	offset int
	err    error
}

func NewByteBuf(data []byte) *ByteBuf {
	return &ByteBuf{data: data, offset: 0}
}

func (b *ByteBuf) Err() error {
	return b.err
}

func (b *ByteBuf) markOOB(method string, need int) {
	if b.err != nil {
		return
	}
	remain := len(b.data) - b.offset
	b.err = fmt.Errorf("bytebuf %s: need %d bytes at offset %d, remain %d", method, need, b.offset, remain)
}

func (b *ByteBuf) ReadUnsignedByte() uint8 {
	if b.err != nil {
		return 0
	}
	if b.offset >= len(b.data) {
		b.markOOB("ReadUnsignedByte", 1)
		return 0
	}
	val := b.data[b.offset]
	b.offset++
	return val
}

func (b *ByteBuf) ReadShort() int16 {
	if b.err != nil {
		return 0
	}
	if b.offset+2 > len(b.data) {
		b.markOOB("ReadShort", 2)
		return 0
	}
	val := int16(binary.BigEndian.Uint16(b.data[b.offset:]))
	b.offset += 2
	return val
}

func (b *ByteBuf) ReadUnsignedShort() uint16 {
	if b.err != nil {
		return 0
	}
	if b.offset+2 > len(b.data) {
		b.markOOB("ReadUnsignedShort", 2)
		return 0
	}
	val := binary.BigEndian.Uint16(b.data[b.offset:])
	b.offset += 2
	return val
}

func (b *ByteBuf) ReadInt() int32 {
	if b.err != nil {
		return 0
	}
	if b.offset+4 > len(b.data) {
		b.markOOB("ReadInt", 4)
		return 0
	}
	val := int32(binary.BigEndian.Uint32(b.data[b.offset:]))
	b.offset += 4
	return val
}

func (b *ByteBuf) ReadUnsignedInt() uint32 {
	if b.err != nil {
		return 0
	}
	if b.offset+4 > len(b.data) {
		b.markOOB("ReadUnsignedInt", 4)
		return 0
	}
	val := binary.BigEndian.Uint32(b.data[b.offset:])
	b.offset += 4
	return val
}

// ReadFloatLE reads a float32 in Little Endian format as per Java Bin41.
func (b *ByteBuf) ReadFloatLE() float32 {
	if b.err != nil {
		return 0
	}
	if b.offset+4 > len(b.data) {
		b.markOOB("ReadFloatLE", 4)
		return 0
	}
	bits := binary.LittleEndian.Uint32(b.data[b.offset:])
	b.offset += 4
	return math.Float32frombits(bits)
}

func (b *ByteBuf) ReadCharSequence(length int) string {
	if b.err != nil {
		return ""
	}
	if b.offset+length > len(b.data) {
		b.markOOB("ReadCharSequence", length)
		return ""
	}
	val := string(b.data[b.offset : b.offset+length])
	b.offset += length
	return val
}

func (b *ByteBuf) ReadBytes(length int) []byte {
	if b.err != nil {
		return nil
	}
	if b.offset+length > len(b.data) {
		b.markOOB("ReadBytes", length)
		return nil
	}
	val := b.data[b.offset : b.offset+length]
	b.offset += length
	return val
}

func (b *ByteBuf) ReadableBytes() int {
	return len(b.data) - b.offset
}
