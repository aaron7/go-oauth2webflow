// Package oauth2webflow allows you to authorize with an OAuth2 Authorization
// Code Flow without copy and pasting codes. It opens the AuthCodeURL with the
// system browser and listens for a http://localhost:5000 callback. If the flow
// completes, it returns an oauth2.Token, automatically closing the browser window.
package oauth2webflow

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"

	"golang.org/x/oauth2"
)

// BrowserAuthCodeFlow attempts the OAuth2 Authorization Code Flow by opening
// the AuthCodeURL given in oauth2.Config with the system browser and the
// RedirectURL set as http://localhost:5000. It then listens for the callback
// and if the flow completes, it returns an oauth2.Token, automatically closing
// the browser window.
func BrowserAuthCodeFlow(ctx context.Context, conf *oauth2.Config) (*oauth2.Token, error) {
	var token *oauth2.Token
	secretState := randomString(10)

	conf.RedirectURL = "http://localhost:5000"
	url := conf.AuthCodeURL(secretState, oauth2.AccessTypeOffline)

	// Open the authorize url in the system web browser
	log.Printf("If a web browser window did not open, please visit: %v", url)
	err := openURLBrowser(url)
	if err != nil {
		return token, err
	}

	// Make a channel for the AccessToken to return (with buffer of 1 so we don't block)
	c := make(chan *oauth2.Token, 1)

	// Create a listener which we can close
	l, err := net.Listen("tcp", ":5000")
	if err != nil {
		return token, err
	}

	// Start the callback http server
	err = http.Serve(l, callbackHandler(ctx, conf, l, c, secretState))
	if err != nil {
		// Ignore "use of closed network connection" error caused by l.close(). Note: would like to fix this.
		if err.Error() != "accept tcp [::]:5000: use of closed network connection" {
			return token, err
		}
	}

	// Return the token from the channel
	token = <-c
	return token, nil
}

func callbackHandler(ctx context.Context, conf *oauth2.Config, l net.Listener, c chan *oauth2.Token, secretState string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ignore other requests such as favicon
		if r.URL.Path != "/" {
			return
		}

		// Get code and state
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		// Close the browser window using JavaScript
		fmt.Fprint(w, `<script type="text/javascript">window.close()</script>`)

		// Check if state is valid
		if state != secretState {
			log.Fatal("callbackHandler: state invalid")
		}

		// Wrap the context to skip invalid certificates
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		sslcli := &http.Client{Transport: tr}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, sslcli)

		// Exchange the code for a token
		token, err := conf.Exchange(ctx, code)
		if err != nil {
			log.Fatal(err)
		}

		// Send the Token through the channel
		c <- token

		// Stop the HTTP server
		defer l.Close()
	})
}
