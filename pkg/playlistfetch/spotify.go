package playlistfetch

import (
	"context"
	"fmt"
	"muxic/pkg/config"

	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const spotifySetupInstructions = `Spotify credentials not configured.
To use playlist-fetch with Spotify:
  1. Go to https://developer.spotify.com/dashboard and create an app.
  2. Set the redirect URI to http://localhost:8080/callback in the app settings.
  3. Add your credentials to ~/.muxic/config.json:
     {
       "spotify": {
         "client_id": "YOUR_CLIENT_ID",
         "client_secret": "YOUR_CLIENT_SECRET"
       }
     }`

// SpotifyService implements Service for Spotify.
type SpotifyService struct {
	cfg *config.Config
}

// NewSpotifyService creates a SpotifyService using the loaded config.
func NewSpotifyService(cfg *config.Config) *SpotifyService {
	return &SpotifyService{cfg: cfg}
}

// FetchPlaylists authenticates with Spotify and returns all of the current user's playlists.
func (s *SpotifyService) FetchPlaylists() ([]Playlist, error) {
	if s.cfg.Spotify.ClientID == "" || s.cfg.Spotify.ClientSecret == "" {
		return nil, fmt.Errorf(spotifySetupInstructions)
	}

	ctx := context.Background()
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("spotify auth: %w", err)
	}

	return fetchSpotifyPlaylists(ctx, client)
}

func (s *SpotifyService) getClient(ctx context.Context) (*spotify.Client, error) {
	auth := spotifyauth.New(
		spotifyauth.WithClientID(s.cfg.Spotify.ClientID),
		spotifyauth.WithClientSecret(s.cfg.Spotify.ClientSecret),
		spotifyauth.WithRedirectURL(redirectURL),
		spotifyauth.WithScopes(
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopePlaylistReadCollaborative,
		),
	)

	// Reuse existing token if present and valid.
	if s.cfg.Spotify.Token != nil {
		httpClient := auth.Client(ctx, s.cfg.Spotify.Token)
		client := spotify.New(httpClient)
		// Refresh if expired.
		newToken, err := auth.RefreshToken(ctx, s.cfg.Spotify.Token)
		if err == nil {
			s.cfg.Spotify.Token = newToken
			_ = config.Save(s.cfg)
			httpClient = auth.Client(ctx, newToken)
			client = spotify.New(httpClient)
		}
		return client, nil
	}

	// Run OAuth flow.
	state := "muxic-spotify"
	authURL := auth.AuthURL(state)
	code, err := WaitForAuthCode(authURL)
	if err != nil {
		return nil, err
	}

	token, err := auth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}

	s.cfg.Spotify.Token = token
	if err := config.Save(s.cfg); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	return spotify.New(httpClient), nil
}

func fetchSpotifyPlaylists(ctx context.Context, client *spotify.Client) ([]Playlist, error) {
	var playlists []Playlist

	page, err := client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return nil, fmt.Errorf("list playlists: %w", err)
	}

	for {
		for _, sp := range page.Playlists {
			tracks, err := fetchSpotifyTracks(ctx, client, sp.ID)
			if err != nil {
				return nil, fmt.Errorf("playlist %q: %w", sp.Name, err)
			}
			playlists = append(playlists, Playlist{
				Name:   sp.Name,
				Tracks: tracks,
			})
		}
		if page.Next == "" {
			break
		}
		if err := client.NextPage(ctx, page); err != nil {
			return nil, fmt.Errorf("next playlist page: %w", err)
		}
	}

	return playlists, nil
}

func fetchSpotifyTracks(ctx context.Context, client *spotify.Client, playlistID spotify.ID) ([]Track, error) {
	var tracks []Track

	page, err := client.GetPlaylistItems(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	for {
		for _, item := range page.Items {
			ft := item.Track.Track
			if ft == nil {
				continue // skip episodes or unavailable items
			}
			artist := ""
			if len(ft.Artists) > 0 {
				artist = ft.Artists[0].Name
			}
			tracks = append(tracks, Track{
				Artist:      artist,
				Album:       ft.Album.Name,
				TrackNumber: int(ft.TrackNumber),
				Title:       ft.Name,
			})
		}
		if page.Next == "" {
			break
		}
		if err := client.NextPage(ctx, page); err != nil {
			return nil, fmt.Errorf("next track page: %w", err)
		}
	}

	return tracks, nil
}
