package playlistfetch

import (
	"context"
	"fmt"
	"muxic/pkg/config"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
)

const youtubeCredentialsFile = "youtube_credentials.json"

// YouTubeService implements Service for YouTube Music.
type YouTubeService struct {
	cfg *config.Config
}

// NewYouTubeService creates a YouTubeService using the loaded config.
func NewYouTubeService(cfg *config.Config) *YouTubeService {
	return &YouTubeService{cfg: cfg}
}

// FetchPlaylists authenticates with YouTube and returns all of the current user's playlists.
func (y *YouTubeService) FetchPlaylists() ([]Playlist, error) {
	ctx := context.Background()
	svc, err := y.getService(ctx)
	if err != nil {
		return nil, fmt.Errorf("youtube auth: %w", err)
	}
	return fetchYouTubePlaylists(ctx, svc)
}

func loadYouTubeOAuthConfig() (*oauth2.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".muxic", youtubeCredentialsFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"YouTube credentials not found.\n"+
				"Download an OAuth 2.0 Desktop credentials JSON from:\n"+
				"  https://console.cloud.google.com/apis/credentials\n"+
				"and save it to: %s", path)
	}
	cfg, err := google.ConfigFromJSON(data, youtube.YoutubeReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("parse YouTube credentials: %w", err)
	}
	cfg.RedirectURL = redirectURL
	return cfg, nil
}

func (y *YouTubeService) getService(ctx context.Context) (*youtube.Service, error) {
	oauthCfg, err := loadYouTubeOAuthConfig()
	if err != nil {
		return nil, err
	}

	var token *oauth2.Token

	if y.cfg.YouTube.Token != nil {
		ts := oauthCfg.TokenSource(ctx, y.cfg.YouTube.Token)
		newToken, err := ts.Token()
		if err == nil {
			token = newToken
			y.cfg.YouTube.Token = newToken
			_ = config.Save(y.cfg)
		} else {
			token = y.cfg.YouTube.Token
		}
	} else {
		authURL := oauthCfg.AuthCodeURL("muxic-youtube", oauth2.AccessTypeOffline)
		code, err := WaitForAuthCode(authURL)
		if err != nil {
			return nil, err
		}
		newToken, err := oauthCfg.Exchange(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("token exchange: %w", err)
		}
		token = newToken
		y.cfg.YouTube.Token = newToken
		if err := config.Save(y.cfg); err != nil {
			return nil, fmt.Errorf("save config: %w", err)
		}
	}

	httpClient := oauthCfg.Client(ctx, token)
	httpClient.Timeout = 30 * time.Second
	svc, err := youtube.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func fetchYouTubePlaylists(ctx context.Context, svc *youtube.Service) ([]Playlist, error) {
	var playlists []Playlist

	call := svc.Playlists.List([]string{"snippet"}).Mine(true).MaxResults(50)
	if err := call.Pages(ctx, func(resp *youtube.PlaylistListResponse) error {
		for _, yt := range resp.Items {
			name := ""
			if yt.Snippet != nil {
				name = yt.Snippet.Title
			}
			fmt.Printf("  Fetching playlist: %s\n", name)
			tracks, err := fetchYouTubeTracks(ctx, svc, yt.Id)
			if err != nil {
				return fmt.Errorf("playlist %q: %w", name, err)
			}
			playlists = append(playlists, Playlist{
				Name:   name,
				Tracks: tracks,
			})
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list playlists: %w", err)
	}

	return playlists, nil
}

func fetchYouTubeTracks(ctx context.Context, svc *youtube.Service, playlistID string) ([]Track, error) {
	var tracks []Track

	call := svc.PlaylistItems.List([]string{"snippet"}).PlaylistId(playlistID).MaxResults(50)
	if err := call.Pages(ctx, func(resp *youtube.PlaylistItemListResponse) error {
		for _, item := range resp.Items {
			if item.Snippet == nil {
				continue
			}
			artist := strings.TrimSuffix(item.Snippet.VideoOwnerChannelTitle, " - Topic")
			title := item.Snippet.Title
			if artist == "Music Library Uploads" {
				// User-uploaded track: no artist metadata available.
				// Attempt to parse "Artist - Title" from the video title.
				if idx := strings.Index(title, " - "); idx != -1 {
					artist = title[:idx]
					title = title[idx+3:]
				} else {
					artist = ""
				}
			}
			tracks = append(tracks, Track{
				Artist: artist,
				Title:  title,
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return tracks, nil
}
