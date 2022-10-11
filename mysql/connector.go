package mysql

import (
    "fmt"
    "net"
    "context"
    "time"
    "log"
)


type Config struct {
    Uri string
    Username string
    Password string
    Database string
}

type Connection struct {
    ctx    context.Context
    cancel context.CancelFunc
    ready  bool
    config Config
    socket net.Conn

    serverVersion   string
    protocolVersion uint8

    commandQueue chan Command
    sequence uint8
}

func Connect(config Config, parentCtx context.Context) (*Connection, error) {
    socket, err := net.Dial("tcp", config.Uri)
    if err != nil {
        return nil, err
    }
    ctx, cancel := context.WithCancel(parentCtx)
    connection := &Connection{
        ctx: ctx,
        cancel: cancel,
        config: config,
        socket: socket,
        ready: false,
        commandQueue: make(chan Command),
    }

    err = connection.init()
    if err != nil {
        return nil, err
    }

    go connection.drainQueue()
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
    c.sequence = packet.getSequence()
    log.Printf("Read packet sequence = %d; header=%v\n", c.sequence, header)

    buf := make([]byte, size)
    n, err = c.socket.Read(buf)
    if err != nil {
        return nil, fmt.Errorf("Failed to read packet payload")
    }
    packet.writeBytes(buf)

    if packet.peekAt(4) == 0xff {
        er := CreateErrorPacket(packet)
        return nil, fmt.Errorf("mysql error [%d]: %s", er.code(), er.error())
    }

    if packet.peekAt(4) == 0xFB {
        return nil, fmt.Errorf("Unsupported packet LOCAL_INFILE")
    }

    return packet, nil
}

func (c *Connection) sendPacket(packet *Packet) error {
    //packet.setSequence(c.sequence + 1)
    buf := packet.bytes()
    log.Printf("Send packet: %v\n", buf)
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
        return err
    }

    packet, err = c.readPacket()
    if err != nil {
        return err
    }

    packet.skip(4)
    if packet.peek() == 0xfe {
        // must send packet with hashed password without headers. etc.
        packet.skip(1)
        pluginName := packet.readStringNullEnded()
        log.Printf("init: Authentication Switch Request. Plugin=%s\n", pluginName)
        log.Printf("init packet rest: %v\n", packet.readBytesRest())
    }
    
    log.Printf("init packet: %v\n", packet.payload)
    c.ready = true
    return nil
}

func (c *Connection) drainQueue() {
    ticker := time.NewTicker(10 * time.Second)
    recvPackets := func (c *Connection, command *Command) {
        eofPackets := 0
        for {
            packet, err := c.readPacket()
            if err != nil {
                command.response <- Command{error:err}
                break
            }
            command.response <- Command{packet:packet}
            if packet.length() < 9 && packet.peekAt(4) == 0xfe {
                // EOF packet
                eofPackets += 1
            }
            if eofPackets > 1 {
                break
            }
        }
        close(command.response)
    }

    for {
        select {
        case command := <- c.commandQueue:
            err := c.sendPacket(command.packet)
            if err != nil {
                command.response <- Command{error: err}
            } else {
                recvPackets(c, &command)
            }
        case <-ticker.C:
            fmt.Printf("ticker")
        case <-c.ctx.Done():
            return
        }
    }
}

func (c *Connection) Query(query string) (string, error){
    packet := &Packet{}
    packet.writeEmptyHeader()
    packet.writeUInt8(0x03)
    packet.writeBytes([]byte(query))
    packet.updateHeader()
    log.Printf("current sequence: %d\n", c.sequence)
    fmt.Printf("Query packet = %v\n", packet)
    command := Command{
        response: make(chan Command, 2),
        packet: packet,
    }
    c.commandQueue <- command
    response := <- command.response
    if response.error != nil {
        return "", response.error
    }
    fmt.Printf("%#v\n", response.error)
    return "", nil
}

/**
 *
 */

type Command struct {
    response chan Command
    packet *Packet
    error error
}

