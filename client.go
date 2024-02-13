package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
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
			fmt.Println("packet read err")
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

func CreateSongRequestPacket(packet *Packet, song string) {
	packet.m_packetType = PACKET_REQUEST_SONG
	packet.m_dataSize = uint32(len(song) + 1)
	copy(packet.m_data[0:packet.m_dataSize], song)
}

func main() {
	conn, err := net.Dial("tcp", "localhost:3017")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server")
	var SongReqPack Packet
	CreateSongRequestPacket(&SongReqPack, "Dream.mp3")
	SendPacket(conn, &SongReqPack)

	var musicBuffer []byte

	var endConnection bool

	for !endConnection {
		var serverPacket Packet
		if ReadPacket(conn, &serverPacket) {
			switch serverPacket.m_packetType {

			case PACKET_MUSIC_STREAM_SIZE:
				fmt.Println("streaming")

			case PACKET_MUSIC_STREAM:
				musicBuffer = append(musicBuffer, serverPacket.m_data[:serverPacket.m_dataSize]...)

				if err != nil {
					fmt.Println("Error decoding music:", err)
					return
				}

				if err != nil {
					fmt.Println("Error playing music:", err)
					return
				}

			case PACKET_END_CONNECTION:
				fmt.Println("End")
				endConnection = true
				break
			}
		}
	}

	fmt.Println("start")
	streamer, format, err := mp3.Decode(ioutil.NopCloser(bytes.NewReader(musicBuffer)))
	defer streamer.Close()
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(streamer)
	// wait until the music has finished playing
	select {}

	// f, err := os.Open("Dream.mp3")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// streamer, format, err := mp3.Decode(f)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// defer streamer.Close()

	// speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	// done := make(chan bool)
	// speaker.Play(beep.Seq(streamer, beep.Callback(func() {
	// 	done <- true
	// })))

	// <-done

}
