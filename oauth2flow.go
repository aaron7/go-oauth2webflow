package oauth2flow

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
)

// OAuth2Config respresents configuration for an OAuth2 provider
type OAuth2Config struct {
	AuthorizeURL string
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scope        string
}

// AuthCodeFlow attempts the OAuth2 Authorization Code Flow
func AuthCodeFlow(settings OAuth2Config) AccessToken {
	responseType := "code"
	redirectURI := "http://localhost:5000"
	secretState := randomString(10)

	// Open the authorize url in the system web browser
	url := createAuthorizationURL(settings, responseType, redirectURI, secretState)
	log.Printf("If a web browser window did not open, please visit: %v", url)
	openURLBrowser(url)

	// Make a channel for the AccessToken to return (with buffer of 1 so we don't block)
	c := make(chan AccessToken, 1)

	// Start the callback http server
	l, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatal(err)
	}
	http.Serve(l, callbackHandler(l, settings, redirectURI, secretState, c))

	// Return the token from the channel
	return <-c
}

// AccessToken represents a response from the Token url
type AccessToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func callbackHandler(l net.Listener, settings OAuth2Config, redirectURI string, secretState string, c chan AccessToken) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		// Check state is valid
		if state != secretState {
			log.Fatal("callbackHandler: state invalid")
			return
		}

		// Create a form
		form := url.Values{
			"grant_type":   {"authorization_code"},
			"code":         {code},
			"redirect_uri": {redirectURI},
		}

		// Create http client skipping SSL verify
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		// Make the POST request to the token url
		req, _ := http.NewRequest("POST", settings.TokenURL, bytes.NewBufferString(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		// Add the client id and secret in the Authorization header
		clientIDSecret := settings.ClientID + ":" + settings.ClientSecret
		encodedClientIDSecret := base64.StdEncoding.EncodeToString([]byte(clientIDSecret))
		req.Header.Add("Authorization", "Basic "+encodedClientIDSecret)
		resp, _ := client.Do(req)

		// Decode the AccessToken response
		var token AccessToken
		defer resp.Body.Close()
		_ = json.NewDecoder(resp.Body).Decode(&token)

		// Close window using JavaScript
		fmt.Fprint(w, `<script type="text/javascript">window.close()</script>`)

		// Send back AccessToken through channel
		c <- token

		// Stop HTTP server
		l.Close()
	})
}

func createAuthorizationURL(settings OAuth2Config, responseType string, redirectURL string, state string) string {
	u, err := url.Parse(settings.AuthorizeURL)
	if err != nil {
		log.Fatal(err)
	}

	q := u.Query()
	q.Set("client_id", settings.ClientID)
	q.Set("response_type", responseType)
	q.Set("redirect_uri", redirectURL)
	q.Set("state", state)
	q.Set("scope", settings.Scope)

	u.RawQuery = q.Encode()
	return u.String()
}
