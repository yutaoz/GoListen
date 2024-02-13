package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
)

const (
	PACKET_REQUEST_SONG = iota
	PACKET_MUSIC_STREAM_SIZE
	PACKET_MUSIC_STREAM
	PACKET_END_CONNECTION
)

type Packet struct {
	m_packetType uint32
	m_dataSize   uint32
	m_data       [256 - (4 * 2)]byte
}

func ReadPacket(conn net.Conn, packet *Packet) bool {
	buf := make([]byte, 256)
	packetRead := false

	for {
		packetLen, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			break
		}

		if packetLen == 256 {
			packetRead = true
			break
		}
	}

	if packetRead {
		packet.m_packetType = binary.LittleEndian.Uint32(buf[0:4])
		packet.m_dataSize = binary.LittleEndian.Uint32(buf[4:8])
		if packet.m_dataSize != 0 {
			copy(packet.m_data[0:256-(4*2)], buf[8:256])
		}
	}

	return packetRead
}

func SendPacket(conn net.Conn, packet *Packet) {
	buf := make([]byte, 256)
	binary.LittleEndian.PutUint32(buf[0:], packet.m_packetType)
	binary.LittleEndian.PutUint32(buf[4:], packet.m_dataSize)
	copy(buf[8:256], packet.m_data[0:packet.m_dataSize])

	conn.Write(buf)
}

func SendMusic(conn net.Conn, song string) {
	musicFileName := song
	musicContent, err := ioutil.ReadFile(musicFileName)
	if err != nil {
		fmt.Println("couldnt read music file")
	}

	var musicFileLen uint32 = uint32(len(musicContent))
	var musicSizePack Packet
	musicSizePack.m_packetType = PACKET_MUSIC_STREAM_SIZE
	musicSizePack.m_dataSize = 4
	binary.LittleEndian.PutUint32(musicSizePack.m_data[0:], musicFileLen)
	SendPacket(conn, &musicSizePack)

	fmt.Println("Sending music")

	var dataLoadMaxSize uint32 = 256 - (4 * 2)
	var currentMusicPos uint32 = 0

	for currentMusicPos < musicFileLen {
		var dataSize uint32 = dataLoadMaxSize

		if currentMusicPos+dataLoadMaxSize > musicFileLen {
			dataSize = musicFileLen - currentMusicPos
		}

		var musicPack Packet
		musicPack.m_packetType = PACKET_MUSIC_STREAM
		musicPack.m_dataSize = dataSize
		copy(musicPack.m_data[0:musicPack.m_dataSize], musicContent[currentMusicPos:currentMusicPos+dataSize])
		SendPacket(conn, &musicPack)
		fmt.Print("*")
		currentMusicPos += dataSize

	}
	fmt.Println("")
	fmt.Println("All music has been sent!")

	var endPacket Packet
	endPacket.m_packetType = PACKET_END_CONNECTION
	endPacket.m_dataSize = 4
	binary.LittleEndian.PutUint32(endPacket.m_data[0:], 4)
	SendPacket(conn, &endPacket)

}

func main() {
	fmt.Println("Waiting for connection")
	l, err := net.Listen("tcp", ":4998")
	if err != nil {
		fmt.Println("Network error")
	}
	defer l.Close()

	c, err := l.Accept()
	if err != nil {
		fmt.Println("Network error")
	}

	fmt.Println("Client connected: ", c.RemoteAddr().String())

	clientDisconnect := false
	for {

		// on client disconnect
		if clientDisconnect {
			break
		}

		// client packet handler
		var clientPacket Packet
		if ReadPacket(c, &clientPacket) {
			switch clientPacket.m_packetType {

			case PACKET_REQUEST_SONG:
				songRequested := string(clientPacket.m_data[0 : clientPacket.m_dataSize-1])
				fmt.Println("Song req: ", songRequested)

				SendMusic(c, songRequested)
			case PACKET_END_CONNECTION:
				fmt.Println("Disconnect packet recevied")
				clientDisconnect = true

			}
		}
	}
}
