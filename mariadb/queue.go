package mariadb

// used for transporting packet between channels
type queuePacket struct {
    c chan queuePacket
    packet *Packet
    error error
}

func createQueuePacket(packet *Packet) queuePacket {
    return queuePacket{
         make(chan queuePacket, 10),
         packet,
         nil,
    }
}

func createQueuePacketError(err error) queuePacket {
    return queuePacket{error:err}
}
