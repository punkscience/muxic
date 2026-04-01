package cmd

import (
	"log"
	"muxic/pkg/config"
	"muxic/pkg/playlistfetch"
	"os"

	"github.com/spf13/cobra"
)

var playlistFetchCmd = &cobra.Command{
	Use:   "playlist-fetch",
	Short: "Fetches playlists from a streaming service and writes them to text files.",
	Long: `Authenticates with Spotify or YouTube Music and fetches all of the
user's playlists, writing each one to a text file in the output directory.
Each line is formatted as: artist - album - trackNumber - title.

Spotify requires client credentials in ~/.muxic/config.json (see the error
message on first run for setup instructions).

YouTube Music requires a credentials file at ~/.muxic/youtube_credentials.json.
Download an OAuth 2.0 Desktop app JSON from:
  https://console.cloud.google.com/apis/credentials
On first run a browser window will open for authorization.

OAuth tokens are stored in ~/.muxic/config.json and reused on subsequent runs.`,
	Run: func(cmd *cobra.Command, args []string) {
		service, _ := cmd.Flags().GetString("service")
		outputDir, _ := cmd.Flags().GetString("output")

		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		var svc playlistfetch.Service
		switch service {
		case "spotify":
			svc = playlistfetch.NewSpotifyService(cfg)
		case "youtube-music":
			svc = playlistfetch.NewYouTubeService(cfg)
		default:
			log.Fatalf("Unknown service %q. Use --service spotify or --service youtube-music.", service)
		}

		log.Printf("Fetching playlists from %s...", service)
		playlists, err := svc.FetchPlaylists()
		if err != nil {
			log.Fatalf("Failed to fetch playlists: %v", err)
		}

		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create output directory %q: %v", outputDir, err)
		}

		totalTracks := 0
		for _, pl := range playlists {
			if err := playlistfetch.WritePlaylist(pl, outputDir); err != nil {
				log.Printf("Error writing playlist %q: %v", pl.Name, err)
				continue
			}
			totalTracks += len(pl.Tracks)
		}

		log.Printf("Done. Wrote %d playlist(s) with %d total track(s) to %q.", len(playlists), totalTracks, outputDir)
	},
}

func init() {
	rootCmd.AddCommand(playlistFetchCmd)

	playlistFetchCmd.Flags().String("service", "", "Streaming service to fetch from: spotify or youtube-music (required)")
	playlistFetchCmd.Flags().String("output", "./playlists", "Directory to write playlist text files")
	_ = playlistFetchCmd.MarkFlagRequired("service")
}
