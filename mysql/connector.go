package mysql

import "fmt"
import "net"
//import "encoding/binary"
//import "github.com/vasflam/lab-mysql-connector/mysql/capabilities"


type Config struct {
    Uri string
    Username string
    Password string
    Database string
}

type Connection struct {
    ready  bool
    config Config
    socket net.Conn

    serverVersion   string
    protocolVersion uint8
    sequence int
}

func Connect(config Config) (*Connection, error) {
    socket, err := net.Dial("tcp", config.Uri)
    if err != nil {
        return nil, err
    }
    connection := &Connection{
        config: config,
        socket: socket,
        ready: false,
    }

    err = connection.init()
    if err != nil {
        return nil, err
    }

    return connection, nil
}

func (c Connection) ServerVersion() string {
    return c.serverVersion
}

func (c Connection) ProtocolVersion() int {
    return int(c.protocolVersion)
}

func (c *Connection) readPacket() (*Packet, error) {
    header := make([]byte, 4)
    n, err := c.socket.Read(header)
    if err != nil {
        return nil, err
    }

    if n != 4 {
        return nil, fmt.Errorf("Read less than requried")
    }
    packet := &Packet{}
    packet.writeHeader(header)
    size := packet.readUInt24()

    buf := make([]byte, size)
    n, err = c.socket.Read(buf)
    if err != nil {
        return nil, fmt.Errorf("Failed to read packet payload")
    }
    packet.writeBytes(buf)

    if packet.peekAt(4) == 0xff {
        er := CreateErrorPacket(packet)
        er.error()
        return nil, fmt.Errorf("Got mysql error, code=%d", er.code())
    }

    return packet, nil
}

func (c *Connection) sendPacket(packet *Packet) error {
    packet.updateHeader()
    packet.setSequence(0)
    buf := packet.bytes()
    n, err := c.socket.Write(buf)
    if err != nil {
        return err
    }

    if n != len(buf) {
        return fmt.Errorf("Sent less bytes than required")
    }

    return nil
}

func (c *Connection) init() error {
    packet, err := c.readPacket()
    if err != nil {
        return err
    }

    hsreq := CreateHandshakeRequestPacket(packet)
    hsres := CreateHandshakeResponsePacket(hsreq, &c.config)
    err = c.sendPacket(hsres)
    if err != nil {
        panic(err)
    }

    packet, err = c.readPacket()
    if err != nil {
        panic(err)
    }
    packet.skip(4)

    if packet.peek() == 0xff {
        errorPacket := CreateErrorPacket(packet)
        fmt.Printf("%d\n", errorPacket.code)
    }
    fmt.Printf("%v\n", packet)
    
    // Prepare response
    //hsrep := CreateHandshakeResponsePacket(hsreq, c.config, capabilities.DEFAULT)
    return nil
}

