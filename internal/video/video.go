package video

import (
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/jaredpetersen/raspilive/internal/raspivid"
)

// Muxer is a video transformation device for modifying raw video into a format suitable for the user.
type Muxer interface {
	Mux(video io.ReadCloser) error
	Wait() error
}

// MuxAndServe muxes video from the Raspberry Pi Camera and serves the static file content.
//
// Blocks until either the video stream process ends or the video muxing process ends.
func MuxAndServe(raspiStream raspivid.Stream, muxer Muxer, server *http.Server) {
	// If any member of the wait group ends early, quit
	var wg sync.WaitGroup
	wg.Add(1)

	// Serve files generated by the video stream
	go func() {
		log.Println("Server started", server.Addr)
		log.Fatal(server.ListenAndServe())
		wg.Done()
	}()

	// Stream video
	go func() {
		log.Fatal(mux(raspiStream, muxer))
		wg.Done()
	}()

	wg.Wait()
}

func mux(raspiStream raspivid.Stream, muxer Muxer) error {
	err := muxer.Mux(raspiStream.Video)
	if err != nil {
		return err
	}

	err = raspiStream.Start()
	if err != nil {
		return err
	}

	// If any member of the wait group ends early, quit
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		err = raspiStream.Wait()
		wg.Done()
	}()

	go func() {
		err = muxer.Wait()
		wg.Done()
	}()

	wg.Wait()

	return nil
}
