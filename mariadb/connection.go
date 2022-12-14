package mariadb

import (
    "fmt"
    "net"
    "context"
    "time"
    _ "log"
    "strconv"
    "github.com/vasflam/lab-mysql-connector/mariadb/capabilities"
)

const COM_QUIT = 0x01
const COM_INIT_DB = 0x02
const COM_QUERY = 0x03
const COM_PING = 0x0e
const COM_RESET_CONN = 0x1f

//  Connection configuration. 
//  Uri in format 'host:port'
type Config struct {
    Uri string
    Username string
    Password string
    Database string
    Timeout time.Duration
}

type connectionInfo struct {
    serverVersion   string
    protocolVersion uint8
    serverCapabilities uint64
    clientCapabilities uint64
}

// Describe database connection
type Connection struct {
    ctx    context.Context
    cancel context.CancelFunc
    ready  bool
    config Config
    socket net.Conn

    info connectionInfo
    packetQueue chan queuePacket
    sequence uint8
    lastInsertId int
    affectedRows int
}

// Establish connection with database
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
        info: connectionInfo{},
        packetQueue: make(chan queuePacket),
    }

    err = connection.init()
    if err != nil {
        return nil, err
    }

    go connection.drainQueue()
    return connection, nil
}

// Gracefuly close conenction
func (c *Connection) Close() {
    <- c.communicate(createQuitPacket())
}

func (c *Connection) LastInsertId() int {
    return c.lastInsertId
}

func (c *Connection) AffectedRows() int {
    return c.affectedRows
}

func (c *Connection) recv() (*Packet, error) {
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

    buf := make([]byte, size)
    n, err = c.socket.Read(buf)
    if err != nil {
        return nil, fmt.Errorf("Failed to read packet payload")
    }
    packet.writeBytes(buf)
    packet.direction = incomingPacket

    if packet.isERR() {
        er := createErrorPacket(packet)
        return nil, fmt.Errorf("mysql error [%d]: %s", er.code(), er.error())
    }
    return packet, nil
}

func (c *Connection) send(packet *Packet) error {
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

// Sends packet to command queue
func (c *Connection) communicate(packet *Packet) chan queuePacket {
    q := createQueuePacket(packet)
    go func() {
        c.packetQueue <- q
    }()
    return q.c
}

/**
 * Do hanshake with server
 * See https://mariadb.com/kb/en/connection/
 */

// See https://mariadb.com/kb/en/connection/
func (c *Connection) init() error {
    packet, err := c.recv()
    if err != nil {
        return err
    }

    request := parseHandshakeRequest(packet)
    c.info = connectionInfo{
        protocolVersion: request.protocolVersion,
        serverVersion: request.serverVersion,
        serverCapabilities: request.capabilities,
    }

    response := createHandshakeResponsePacket(request, &c.config, &c.info)
    err = c.send(response)
    if err != nil {
        return err
    }

    packet, err = c.recv()
    if err != nil {
        return err
    }

    packet.skip(4)
    if packet.peek() == 0xfe {
        // Authentication Swith Request
        // must send packet with hashed password without headers. etc.
        packet.skip(1)
        //pluginName := packet.readStringNullEnded()
        return fmt.Errorf("handshake: Authentication Switch Request is unsupported yet")
    }
    
    c.ready = true
    return nil
}

func (c *Connection) drainQueue() {
    ticker := time.NewTicker(10 * time.Second)
    recvPackets := func (c *Connection, q *queuePacket, initiator *Packet) {
        for {
            packet, err := c.recv()
            if err != nil {
                q.c <- queuePacket{error:err}
                break
            }

            if packet.isLOCALINFILE() {
                err := fmt.Errorf("Unsupported packet LOCAL_INFILE")
                q.c <- createQueuePacketError(err)
                break
            }

            q.c <- createQueuePacket(packet)

            if packet.isOK() || packet.isEOF() {
                break
            }
        }
        close(q.c)
    }

    for {
        select {
        case q := <- c.packetQueue:
            err := c.send(q.packet)
            if err != nil {
                q.c <- queuePacket{error: err}
            } else {
                recvPackets(c, &q, q.packet)
            }
        case <-ticker.C:
            q := c.communicate(createPingPacket())
            for _ = range q {}
        case <-c.ctx.Done():
            return
        }
    }
}

// run mysql commands
func (c *Connection) Query(query string) (QueryResultRows, error){
    q := c.communicate(createQueryPacket(query))
    // read first packet to get columns packets
    response := <- q
    if response.error != nil {
        return nil, response.error
    }
    packet := response.packet
    if packet.isOK() {
        packet.skip(4)
        packet.skip(1)
        c.affectedRows = packet.readUIntLengthEncoded() // affected rows
        c.lastInsertId = int(packet.readUIntLengthEncoded())
        return nil, nil
    }

    columnCount := int(packet.peekAt(4))
    rows := QueryResultRows{}
    columns := []tableColumn{}
    for i := 0; i < columnCount; i++ {
        response := <- q
        if response.error != nil {
            return nil, response.error
        }
        packet := response.packet
        packet.skip(4)
        _ = packet.readStringLengthEncoded() // catalog
        _ = packet.readStringLengthEncoded() // schema
        _ = packet.readStringLengthEncoded() // tableAlias
        _ = packet.readStringLengthEncoded() // table
        columnAlias := packet.readStringLengthEncoded()
        _ = packet.readStringLengthEncoded() // column
        if c.info.clientCapabilities & capabilities.MARIADB_CLIENT_EXTENDED_TYPE_INFO != 0 {
            count := packet.readUIntLengthEncoded()
            for i = 0; i < count; i++ {
                _ = packet.readUInt8() // type
                _ = packet.readStringLengthEncoded() // value
            }
        }

        fixedFields := packet.readUIntLengthEncoded()
        charset := packet.readUInt16()
        maxColSize := packet.readUInt32()
        fieldType := packet.readUInt8()
        fieldDetailFlag := packet.readUInt16()
        decimals := packet.readUInt8()
        unused := packet.readUInt16()
        column := tableColumn{
            columnAlias,
            fixedFields,
            charset,
            maxColSize,
            fieldType,
            fieldDetailFlag,
            decimals,
            unused,
        }
        columns = append(columns, column)
    }

    i := 0
    for {
        response := <- q
        if response.error != nil {
            return nil, response.error
        }
        packet := response.packet

        if packet.isEOF() || packet.isOK() {
            break
        }
        packet.skip(4)
        row := QueryResultRow{}
        for i := 0; i < len(columns); i++ {
            column := columns[i]
            kind := column.kind

            var value interface{}
            if kind == MYSQL_TYPE_TINY ||
                kind == MYSQL_TYPE_SHORT ||
                kind == MYSQL_TYPE_LONG {
                    str, isNULL := packet.readStringLengthEncodedNULLABLE()
                    if isNULL {
                        value = nil
                    } else {
                        var err error
                        value, err = strconv.Atoi(str)
                        if err != nil {
                            return nil, err
                        }
                    }
            }
            row[column.name] = value
        }
        rows = append(rows, row)
        i += 1
    }

    return rows, nil
}

