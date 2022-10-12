package mariadb

import (
    "encoding/binary"
    "bytes"
    "crypto/sha1"
    "fmt"
)

const (
    PacketTypeOK = 0x00
    PacketTypeLOCALINFILE = 0xfb
    PacketTypeEOF = 0xfe
    PacketTypeERR = 0xff
)

type PacketDirection int
const (
    IncomingPacket PacketDirection = iota
    OutgoingPacket
)

/**
 * Hash password for mysql_native_password auth
 */
func hashPassword(password string, salt []byte) []byte {
    var hash [20]byte
    stage1 := make([]byte, 20)
    stage2 := make([]byte, 20)
    hash = sha1.Sum([]byte(password)) 
    copy(stage1, hash[:])
    hash = sha1.Sum(stage1)
    copy(stage2, hash[:])
    h := sha1.New()
    h.Write(salt)
    h.Write(stage2)
    digest := h.Sum(nil)
    for i := 0; i < len(digest); i++ {
        digest[i] = stage1[i] ^ digest[i]
    }
    return digest
}

/**
 *
 */
type Packet struct {
    payload []byte
    pos    int
    hasHeader bool
    direction PacketDirection
    storedPos int
}

func (p Packet) bytes() []byte {
    return p.payload
}

func (p Packet) length() int {
    return len(p.payload)
}

func (p Packet) payloadLength() int {
    length := p.length()
    if p.hasHeader {
        length -= 4
    }
    return length
}

func (p *Packet) resetPos() {
    p.pos = 0
}

func (p *Packet) skip(n int) {
    p.pos += n
}

func (p *Packet) peek() byte {
    return p.payload[p.pos]
}

func (p *Packet) peekAt(pos int) byte {
    if pos <= len(p.payload) {
        return p.payload[pos]
    }
    return 0
}

func (p *Packet) readStringNullEnded() string {
    s, err := bytes.NewBuffer(p.payload[p.pos:]).ReadString(0x00)
    if err != nil {
        return ""
    }
    p.pos += len(s)
    return s[:len(s)-1]
}

func (p *Packet) readBytes(n int) []byte {
    if (p.pos + n > len(p.payload)) {
        return []byte{}
    }
    buf := p.payload[p.pos:p.pos+n]
    p.pos += n
    return buf
}

func (p *Packet) readBytesRest() []byte {
    return p.payload[p.pos:]
}

func (p *Packet) readInt8() int8 {
    return int8(p.readUInt8())
}

func (p *Packet) readUInt8() uint8 {
    v := p.payload[p.pos]
    p.pos++
    return v
}

func (p *Packet) readInt16() int16 {
    return int16(p.readUInt16())
}

func (p *Packet) readUInt16() uint16 {
    v := binary.LittleEndian.Uint16(p.payload[p.pos:p.pos+2])
    p.pos += 2
    return v

    buf := bytes.NewBuffer(p.payload[p.pos:p.pos+2])
    var value uint16
    binary.Read(buf, binary.LittleEndian, &value)
    p.pos += 2
    return value
}

func (p *Packet) readInt24() int32 {
    return int32(p.readUInt32())
}

func (p *Packet) readUInt24() uint32 {
    normalized := append(p.payload[p.pos:p.pos+3], 0)
    v := binary.LittleEndian.Uint32(normalized)
    return v
}

func (p *Packet) readInt32() int32 {
    return int32(p.readUInt32())
}

func (p *Packet) readUInt32() uint32 {
    v := binary.LittleEndian.Uint32(p.payload[p.pos:p.pos+4])
    p.pos += 4
    return v
}

func (p *Packet) readInt64() int64 {
    return int64(p.readUInt64())
}

func (p *Packet) readUInt64() uint64 {
    buf := []byte{0,0,0,0,0,0,0,0}
    if p.pos + 8 > len(p.payload) {
        buf = p.payload[p.pos:]
        p.pos = len(p.payload)
    } else {
        buf = p.payload[p.pos:p.pos+8]
        p.pos += 8
    }
    bufLen := len(buf)
    if len(buf) < 8 {
        for i := 0; i < 8 - bufLen; i++ {
            buf = append(buf, 0)
        }
    }
    v := binary.LittleEndian.Uint64(buf)
    return v
}

func (p *Packet) readBytesEncodedLength() []byte {
    b := int(p.readUInt8())
    return p.readBytes(b)
}

func (p *Packet) readUIntEncodedLength() int {
    length := int(p.readUInt8())
    if length < 0xfb {
        return length
    } else if length < 65536 {
        return int(p.readUInt16())
    } else if length < 16777216 {
        return int(p.readUInt24())
    } else {
        return int(p.readUInt64())
    }
}

func (p *Packet) readUIntPrefixLength() int {
    length := int(p.readUInt8())
    buf := make([]byte, length)
    copy(buf, p.readBytes(length))
    packet := &Packet{payload:buf}
    v := int(packet.readUInt64())
    fmt.Printf("buf=%v = %d\n", buf, v)
    return int(packet.readUInt64())
}

func (p *Packet) readStringLengthEncoded() string {
    length := int(p.readUInt8())
    return string(p.readBytes(length))
}

func (p *Packet) readStringLengthEncodedNULLABLE() (string, bool) {
    if p.peek() == 0xfb {
        return "", true
    }
    length := int(p.readUInt8())
    return string(p.readBytes(length)), false
}

func (p *Packet) writeUInt8(i uint8) {
    p.payload = append(p.payload, i)
}

func (p *Packet) writeUInt16(i uint16) {
    buf := make([]byte, 2)
    binary.LittleEndian.PutUint16(buf, i)
    p.payload = append(p.payload, buf...)
}

func (p *Packet) writeUInt24(i uint32) {
    if i < 16777216 {
        buf := make([]byte, 4)
        binary.LittleEndian.PutUint32(buf, i)
        buf = buf[0:3]
        p.payload = append(p.payload, buf...)
    }
}

func (p *Packet) writeUInt32(i uint32) {
    buf := make([]byte, 4)
    binary.LittleEndian.PutUint32(buf, i)
    p.payload = append(p.payload, buf...)
}

func (p *Packet) writeUInt64(i uint64) {
    buf := make([]byte, 8)
    binary.LittleEndian.PutUint64(buf, i)
    p.payload = append(p.payload, buf...)
}

func (p *Packet) writeBytes(b []byte) {
    p.payload = append(p.payload, b...)
}

func (p *Packet) writeLengthEncoded(length uint64) {
    if length < 0xfb {
        p.writeUInt8(uint8(length))
    } else if length < 65536 {
        p.writeUInt8(0xfc)
        p.writeUInt16(uint16(length))
    } else if length < 16777216 {
        p.writeUInt8(0xfd)
        p.writeUInt24(uint32(length))
    } else {
        p.writeUInt8(0xfe)
        p.writeUInt64(length)
    }
}

func (p *Packet) writeHeader(b []byte) {
    if !p.hasHeader {
        header := b[0:4]
        p.payload = append(header, p.payload...)
        p.hasHeader = true
    }
}

func (p *Packet) writeEmptyHeader() {
    p.writeHeader([]byte{0, 0, 0, 0})
}

func (p *Packet) updateHeader() {
    if !p.hasHeader {
        p.writeEmptyHeader()
    }

    length := len(p.payload) - 4
    temp := Packet{}
    temp.writeUInt24(uint32(length))
    header := temp.bytes()

    if p.hasHeader {
        for i := 0; i < 3; i++ {
            p.payload[i] = header[i]
        }
    }
}

func (p *Packet) setSequence(i uint8) {
    if p.hasHeader {
        p.payload[3] = i
    }
}

func (p Packet) getSequence() uint8 {
    if p.hasHeader {
        return p.payload[3]
    }
    return 0
}

func (p Packet) haveStoredPos() bool {
    return p.storedPos > -1
}

func (p Packet) storePos(withReset bool) {
    p.storedPos = p.pos
    if withReset {
        p.pos = 0
    }
}

func (p Packet) restorePos() {
    if p.storedPos > -1 {
        p.pos = p.storedPos
        p.storedPos = -1
    }
}

func (p Packet) isEOF() bool {
    return p.direction == IncomingPacket &&
        p.peekAt(4) == PacketTypeEOF &&
        p.payloadLength() < 9
}

func (p Packet) isERR() bool {
    return p.direction == IncomingPacket && p.peekAt(4) == PacketTypeERR
}

func (p Packet) isOK() bool {
    return p.direction == IncomingPacket &&
        p.peekAt(4) == PacketTypeOK &&
        p.payloadLength() < 0xffffff
}

func (p Packet) isLOCALINFILE() bool {
    return p.direction == IncomingPacket && p.peekAt(4) == PacketTypeLOCALINFILE
}

/**
 * See https://mariadb.com/kb/en/err_packet/
 */
type ErrorPacket struct {
    *Packet
}

func createErrorPacket(packet *Packet) *ErrorPacket {
    e := &ErrorPacket{Packet:packet}
    return e
}

func (p *ErrorPacket) code() int {
    p.resetPos()
    p.skip(5)
    v := p.Packet.readUInt16()
    p.resetPos()
    return int(v)
}

// TODO mark and restore position in packet
func (p *ErrorPacket) error() string {
    p.resetPos()
    message := ""
    if p.code() != 0xff {
        p.skip(7)
        if string(p.peek()) == "#" {
            p.skip(1)
            message = fmt.Sprintf("#[%s] %s", p.readBytes(5), p.readBytesRest())
        } else {
            message = string(p.readBytesRest())
        }
    } else {
        message = "progress error reporting is not supported"
    }
    p.resetPos()
    return message
}

func createQuitPacket() *Packet {
    packet := &Packet{}
    packet.writeUInt8(COM_QUIT)
    packet.updateHeader()
    packet.direction = OutgoingPacket
    return packet
}

func createInitDbPacket(dbname string) *Packet {
    packet := &Packet{}
    packet.writeUInt8(COM_INIT_DB)
    packet.writeBytes([]byte(dbname))
    packet.writeUInt8(0)
    packet.updateHeader()
    packet.direction = OutgoingPacket
    return packet
}

func createQueryPacket(query string) *Packet {
    packet := &Packet{}
    packet.writeUInt8(COM_QUERY)
    packet.writeBytes([]byte(query))
    packet.writeUInt8(0)
    packet.updateHeader()
    packet.direction = OutgoingPacket
    return packet
}

func createPingPacket() *Packet {
    packet := &Packet{}
    packet.writeUInt8(COM_PING)
    packet.updateHeader()
    packet.direction = OutgoingPacket
    return packet
}
