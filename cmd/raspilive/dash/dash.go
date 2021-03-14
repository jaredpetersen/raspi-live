package dash

import (
	"errors"
	"os"
	"os/signal"
	"time"

	"github.com/jaredpetersen/raspilive/internal/ffmpeg/dash"
	"github.com/jaredpetersen/raspilive/internal/raspivid"
	"github.com/jaredpetersen/raspilive/internal/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const serverShutdownDeadline = 10 * time.Second

// Cfg represents the configuration for DASH.
type Cfg struct {
	Width          int
	Height         int
	Fps            int
	HorizontalFlip bool
	VerticalFlip   bool
	Port           int
	Directory      string
	TLSCert        string
	TLSKey         string
	SegmentTime    int // Segment length target duration in seconds
	PlaylistSize   int // Maximum number of playlist entries
	StorageSize    int // Maximum number of unreferenced segments to keep on disk before removal
}

// Cmd is a DASH command for Cobra.
var Cmd = &cobra.Command{
	Use:   "dash",
	Short: "Stream video using DASH",
	Long:  "Stream video using DASH",
}

func init() {
	cfg := Cfg{}

	Cmd.Flags().IntVar(&cfg.Width, "width", 1920, "video width")

	Cmd.Flags().IntVar(&cfg.Height, "height", 1080, "video height")

	Cmd.Flags().IntVar(&cfg.Fps, "fps", 30, "video framerate")

	Cmd.Flags().BoolVar(&cfg.HorizontalFlip, "horizontal-flip", false, "horizontally flip video")

	Cmd.Flags().BoolVar(&cfg.VerticalFlip, "vertical-flip", false, "vertically flip video")

	Cmd.Flags().IntVar(&cfg.Port, "port", 0, "static file server port")

	Cmd.Flags().StringVar(&cfg.Directory, "directory", "", "static file server directory")

	Cmd.Flags().StringVar(&cfg.TLSCert, "tls-cert", "", "static file server TLS certificate")

	Cmd.Flags().StringVar(&cfg.TLSKey, "tls-key", "", "static file server TLS key")

	Cmd.Flags().IntVar(&cfg.SegmentTime, "segment-time", 0, "target segment duration in seconds")

	Cmd.Flags().IntVar(&cfg.PlaylistSize, "playlist-size", 0, "maximum number of playlist entries")

	Cmd.Flags().IntVar(&cfg.StorageSize, "storage-size", 0, "maximum number of unreferenced segments to keep on disk before removal")

	Cmd.Flags().SortFlags = false

	Cmd.Run = func(cmd *cobra.Command, args []string) {
		streamDash(cfg)
	}
}

func streamDash(cfg Cfg) {
	raspiStream := newRaspiStream(cfg)
	muxer := newMuxer(cfg)
	srv := newServer(cfg)

	// Set up a channel for exiting
	stop := make(chan struct{})
	osStopper(stop)

	// Serve files generated by the video stream
	go func() {
		err := srv.ListenAndServe()
		if err != nil && errors.Is(err, server.ErrInvalidDirectory) {
			log.Fatal().Msg("Directory does not exist")
		}
		if err != nil {
			log.Fatal().Msg("Encountered an error serving video")
		}
		stop <- struct{}{}
	}()

	// Stream video
	go func() {
		if err := mux(raspiStream, muxer); err != nil {
			log.Fatal().Msg("Encountered an error muxing video")
		}
		stop <- struct{}{}
	}()

	// Wait for a stop signal
	<-stop

	log.Info().Msg("Shutting down")

	raspiStream.Video.Close()
	srv.Shutdown(serverShutdownDeadline)
}

func newRaspiStream(cfg Cfg) *raspivid.Stream {
	raspiOptions := raspivid.Options{
		Width:          cfg.Width,
		Height:         cfg.Height,
		Fps:            cfg.Fps,
		HorizontalFlip: cfg.HorizontalFlip,
		VerticalFlip:   cfg.VerticalFlip,
	}

	raspiStream, err := raspivid.NewStream(raspiOptions)
	if err != nil {
		log.Fatal().Msg("Encountered an error streaming video from the Raspberry Pi Camera Module")
	}

	return raspiStream
}

func newMuxer(cfg Cfg) *dash.Muxer {
	return &dash.Muxer{
		Directory: cfg.Directory,
		Options: dash.Options{
			Fps:          cfg.Fps,
			SegmentTime:  cfg.SegmentTime,
			PlaylistSize: cfg.PlaylistSize,
			StorageSize:  cfg.StorageSize,
		},
	}
}

func newServer(cfg Cfg) *server.Static {
	return &server.Static{
		Port:      cfg.Port,
		Directory: cfg.Directory,
		Cert:      cfg.TLSCert,
		Key:       cfg.TLSKey,
	}
}

func osStopper(stop chan struct{}) {
	// Set up a channel for OS signals so that we can quit gracefully if the user terminates the program
	// Once we get this signal, sent a message to the stop channel
	osStop := make(chan os.Signal, 1)
	signal.Notify(osStop, os.Interrupt, os.Kill)

	go func() {
		<-osStop
		stop <- struct{}{}
	}()
}

func mux(raspiStream *raspivid.Stream, muxer *dash.Muxer) error {
	if err := muxer.Mux(raspiStream.Video); err != nil {
		return err
	}
	if err := raspiStream.Start(); err != nil {
		return err
	}
	if err := muxer.Wait(); err != nil {
		return err
	}
	if err := raspiStream.Wait(); err != nil {
		return err
	}
	return nil
}
