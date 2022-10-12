package mariadb

/**
 * Used for transporting packet between channels
 */
type QueuePacket struct {
    c chan QueuePacket
    packet *Packet
    error error
}

func createQueuePacket(packet *Packet) QueuePacket {
    return QueuePacket{
         make(chan QueuePacket, 10),
         packet,
         nil,
    }
}

func createQueuePacketError(err error) QueuePacket {
    return QueuePacket{error:err}
}
