package playlistfetch

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
)

const callbackPort = "8080"
const callbackPath = "/callback"
const redirectURL = "http://localhost:" + callbackPort + callbackPath

// WaitForAuthCode starts a local HTTP server on localhost:8080/callback,
// opens authURL in the browser, blocks until the OAuth redirect arrives,
// and returns the authorization code.
func WaitForAuthCode(authURL string) (string, error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Addr: ":" + callbackPort, Handler: mux}

	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		errParam := r.URL.Query().Get("error")
		if errParam != "" {
			fmt.Fprintf(w, "Authorization failed: %s. You may close this tab.", errParam)
			errCh <- fmt.Errorf("oauth error: %s", errParam)
			return
		}
		fmt.Fprintf(w, "Authorization successful! You may close this tab.")
		codeCh <- code
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser automatically. Please visit:\n%s\n", authURL)
	}

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		_ = srv.Shutdown(context.Background())
		return "", err
	}

	_ = srv.Shutdown(context.Background())
	return code, nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("unsupported platform")
	}
	return exec.Command(cmd, args...).Start()
}
