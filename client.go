package main

import (
	"fmt"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

// https://www.youtube.com/watch?v=74c4z28izds
// https://github.com/pion/webrtc
// https://medium.com/@icelain/a-guide-to-building-a-realtime-http-audio-streaming-server-in-go-24e78cf1aa2c
func main() {
	// conn, err := net.Dial("tcp", "184.146.91.48:8993") // Replace "localhost:8080" with the address of your server
	// if err != nil {
	// 	fmt.Println("Error connecting to server:", err)
	// 	return
	// }
	// defer conn.Close()

	// fmt.Println("Connected to server")
	f, err := os.Open("Dream.mp3")
	if err != nil {
		fmt.Println(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		fmt.Println(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done

}
