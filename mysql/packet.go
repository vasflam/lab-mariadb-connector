package mysql

import (
    "encoding/binary"
    "bytes"
    "math"
    "crypto/sha1"
    "github.com/vasflam/lab-mysql-connector/mysql/capabilities"
    "fmt"
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
    fmt.Printf("stage1: %v\ndigest: %v\nstage2: %v\n", stage1, digest, stage2)
    for i := 0; i < len(digest); i++ {
        fmt.Printf("%v : %v = %v\n", stage1[i], digest[i], stage1[i] ^ digest[i])
        digest[i] = stage1[i] ^ digest[i]
    }
    return digest
}

/**
 *
 */
type Packet struct {
    pos    int
    payload []byte
    hasHeader bool
    markedPos int
    hasMarkedPos bool
}

func (p Packet) bytes() []byte {
    return p.payload
}

func (p Packet) length() int {
    return len(p.payload)
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
    if (p.pos + n >= len(p.payload)) {
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
    buf := bytes.NewBuffer(p.payload[p.pos:p.pos+4])
    var value uint32
    binary.Read(buf, binary.LittleEndian, &value)
    p.pos += 4
    return value
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

func (p *Packet) writeLengthCoded(length uint64) {
    if length < 0xfb {
        p.writeUInt8(uint8(length))
        return
    }

    if length < 65536 {
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

/**
 * See: https://mariadb.com/kb/en/connection/#initial-handshake-packet
 */
type HandshakeRequestPacket struct {
    status       uint16
    scramble     []byte
    collation    uint8
    connection   uint32
    pluginName   string
    capabilities int64

    serverVersion    string
    protocolVersion  uint8
    pluginDataLength uint8

    rawPacket *Packet
}

func CreateHandshakeRequestPacket(packet *Packet) *HandshakeRequestPacket {
    hsr := &HandshakeRequestPacket{rawPacket:packet}

    if (packet.hasHeader) {
        packet.skip(4)
    }

    hsr.protocolVersion = packet.readUInt8()
    hsr.serverVersion = packet.readStringNullEnded()
    hsr.connection = packet.readUInt32()
    hsr.scramble = packet.readBytes(8) // scramble 1st part (authentication seed)
    packet.skip(1)
    hsr.capabilities = int64(packet.readUInt16())
    hsr.collation = packet.readUInt8()
    hsr.status = packet.readUInt16()
    hsr.capabilities += int64(packet.readUInt16()) << 16
    fmt.Printf("caps1: %d\n", hsr.capabilities)

    if hsr.capabilities & capabilities.PLUGIN_AUTH != 0 {
        hsr.pluginDataLength = packet.readUInt8()
    } else {
        packet.skip(1)
    }

    packet.skip(6)
    if hsr.capabilities & capabilities.MYSQL  != 0 {
        packet.skip(4)
    } else {
        hsr.capabilities += int64(packet.readUInt32()) << 32
    }

    fmt.Printf("srcamble1: %v\n", []byte(hsr.scramble))
    if hsr.capabilities & capabilities.SECURE_CONNECTION != 0 {
        scramble2 := packet.readBytes(int(math.Max(12, float64(hsr.pluginDataLength)-9)))
        hsr.scramble = append(hsr.scramble, scramble2...)
        packet.skip(1)
    }

    if hsr.capabilities & capabilities.PLUGIN_AUTH != 0 {
        hsr.pluginName = packet.readStringNullEnded()
    }

    fmt.Printf("pversion=%d\n", hsr.protocolVersion)
    fmt.Printf("sversion=%s\n", hsr.serverVersion)
    fmt.Printf("connection=%d\n", hsr.connection)
    fmt.Printf("scrable=%v\n", hsr.scramble)
    fmt.Printf("caps=%d\n", hsr.capabilities)
    fmt.Printf("collation=%d\n", hsr.collation)
    fmt.Printf("status=%d\n", hsr.status)
    fmt.Printf("pluginname=%s\n", hsr.pluginName)

    return hsr
}

/**
 *
 */
func CreateHandshakeResponsePacket(
     hsreq *HandshakeRequestPacket, 
     config *Config,
 ) *Packet {
     clientCapabilities := capabilities.DEFAULT
     if hsreq.capabilities & capabilities.PLUGIN_AUTH != 0 {
         clientCapabilities |= capabilities.PLUGIN_AUTH
     }

     /*
     if hsreq.capabilities & capabilities.MYSQL == 0 {
         clientCapabilities |= capabilities.MARIADB_CLIENT_EXTENDED_TYPE_INFO
     }
     */

     /*
     if hsreq.capabilities & capabilities.MARIADB_CLIENT_STMT_BULK_OPERATIONS != 0 {
         clientCapabilities |= capabilities.MARIADB_CLIENT_STMT_BULK_OPERATIONS
     }
     */

     if config.Database != "" && (hsreq.capabilities & capabilities.CONNECT_WITH_DB != 0) {
         clientCapabilities |= capabilities.CONNECT_WITH_DB
     }

     var authToken []byte
     var authPlugin string
     switch pluginName := hsreq.pluginName; pluginName {
     case "mysql_clear_password":
         authToken = []byte(config.Password)
         authPlugin = pluginName
     case "mysql_native_password":
         authToken = hashPassword(config.Password, hsreq.scramble)
         authPlugin = pluginName
     default:
         panic(`Only 'mysql_native_password' and 'mysql_clear_password' authentication is supported`)
    }

    packet := &Packet{}
    packet.writeEmptyHeader()
    fmt.Printf("op1: %d (%d)\n", uint32(clientCapabilities & 0xffffffff), clientCapabilities)
    packet.writeUInt32(uint32(clientCapabilities & 0xffffffff))
    //packet.writeUInt32(uint32(28222218))
    packet.writeUInt32(1073741824) // 1MB
    packet.writeUInt8(hsreq.collation)
    for i := 0; i < 19; i++ {
        packet.writeUInt8(0)
    }
    fmt.Printf("op2: %d\n", uint32(clientCapabilities >> 32))
    //packet.writeUInt32(28)
    packet.writeUInt32(uint32(clientCapabilities >> 32))
    packet.writeBytes([]byte(config.Username))
    packet.writeUInt8(0)

    if hsreq.capabilities & capabilities.PLUGIN_AUTH_LENENC_CLIENT_DATA != 0 {
        fmt.Printf("atoken: %v\n", authToken)
        packet.writeLengthCoded(uint64(len(authToken)))
        packet.writeBytes(authToken)
    } else if hsreq.capabilities & capabilities.SECURE_CONNECTION != 0 {
        packet.writeUInt8(hsreq.pluginDataLength)
        packet.writeBytes(authToken)
    } else {
        packet.writeBytes(authToken)
        packet.writeUInt8(0)
    }

    if clientCapabilities  & capabilities.CONNECT_WITH_DB != 0 {
        packet.writeBytes([]byte(config.Database))
        packet.writeUInt8(0)
    }

    if hsreq.capabilities & capabilities.PLUGIN_AUTH != 0 {
        packet.writeBytes([]byte(authPlugin))
        packet.writeUInt8(0)
    }

    if hsreq.capabilities & capabilities.CONNECT_ATTRS != 0 {
        packet.writeUInt8(0)
    }

    packet.updateHeader()
    packet.setSequence(1)
    return packet
}

/**
 * See https://mariadb.com/kb/en/err_packet/
 */
type ErrorPacket struct {
    *Packet
}

func CreateErrorPacket(packet *Packet) *ErrorPacket {
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

/**
 * See https://mariadb.com/kb/en/ok_packet/
 */
type OkPacket struct {
    *Packet
}

func CreateOkPacket(packet *Packet) *OkPacket {
    return &OkPacket{Packet: packet}
}

