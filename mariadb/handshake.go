package mariadb

import (
    "math"
    "github.com/vasflam/lab-mysql-connector/mariadb/capabilities"
)

/**
 * See: https://mariadb.com/kb/en/connection/#initial-handshake-packet
 */
type HandshakeRequest struct {
    status       uint16
    scramble     []byte
    collation    uint8
    connection   uint32
    pluginName   string
    capabilities uint64

    serverVersion    string
    protocolVersion  uint8
    pluginDataLength uint8
}

func parseHandshakeRequest(packet *Packet) *HandshakeRequest{
    hsr := &HandshakeRequest{}

    if (packet.hasHeader) {
        packet.skip(4)
    }

    hsr.protocolVersion = packet.readUInt8()
    hsr.serverVersion = packet.readStringNullEnded()
    hsr.connection = packet.readUInt32()
    hsr.scramble = packet.readBytes(8) // scramble 1st part (authentication seed)
    packet.skip(1)
    hsr.capabilities = uint64(packet.readUInt16())
    hsr.collation = packet.readUInt8()
    hsr.status = packet.readUInt16()
    hsr.capabilities += uint64(packet.readUInt16()) << 16

    if hsr.capabilities & capabilities.PLUGIN_AUTH != 0 {
        hsr.pluginDataLength = packet.readUInt8()
    } else {
        packet.skip(1)
    }

    packet.skip(6)
    if hsr.capabilities & capabilities.MYSQL  != 0 {
        packet.skip(4)
    } else {
        hsr.capabilities += uint64(packet.readUInt32()) << 32
    }

    if hsr.capabilities & capabilities.SECURE_CONNECTION != 0 {
        scramble2 := packet.readBytes(int(math.Max(12, float64(hsr.pluginDataLength)-9)))
        hsr.scramble = append(hsr.scramble, scramble2...)
        packet.skip(1)
    }

    if hsr.capabilities & capabilities.PLUGIN_AUTH != 0 {
        hsr.pluginName = packet.readStringNullEnded()
    }

    return hsr
}

/**
 *
 */
func createHandshakeResponsePacket(
     hsreq *HandshakeRequest, 
     config *Config,
     info *ConnectionInfo,
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
    packet.writeUInt32(uint32(clientCapabilities & 0xffffffff))
    packet.writeUInt32(1073741824) // 1MB
    packet.writeUInt8(hsreq.collation)
    for i := 0; i < 19; i++ {
        packet.writeUInt8(0)
    }
    packet.writeUInt32(uint32(clientCapabilities >> 32))
    packet.writeBytes([]byte(config.Username))
    packet.writeUInt8(0)

    if hsreq.capabilities & capabilities.PLUGIN_AUTH_LENENC_CLIENT_DATA != 0 {
        packet.writeLengthEncoded(uint64(len(authToken)))
        packet.writeBytes(authToken)
    } else if hsreq.capabilities & capabilities.SECURE_CONNECTION != 0 {
        packet.writeUInt8(hsreq.pluginDataLength)
        packet.writeBytes(authToken)
    } else {
        packet.writeBytes(authToken)
        packet.writeUInt8(0)
    }

    if clientCapabilities & capabilities.CONNECT_WITH_DB != 0 {
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
    info.clientCapabilities = clientCapabilities
    return packet
}

