go-oauth2webflow
=============

[![GoDoc](https://godoc.org/github.com/aaron7/go-oauth2webflow?status.png)](https://godoc.org/github.com/aaron7/go-oauth2webflow)

go-oauth2webflow allows you to authorize with an OAuth2 Authorization Code Flow
endpoint without having to copy and paste codes from the callback url. It uses
[golang.org/x/oauth2](https://golang.org/x/oauth2).

The package opens the authorize url with the system browser, setting
`redirect_uri` set to `http://localhost:5000`, and listens for the callback
using a http server. If the flow completes, it returns an `oauth2.Token` and
automatically closes the browser window.

Please ensure `http://localhost:5000` is set as an authorized redirect URI.

Note: created this project as part of learning go

Todo: tests

## Example

```go
package main

import (
	"context"
	"log"

	"github.com/aaron7/go-oauth2webflow"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     "a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0",
		ClientSecret: "b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1",
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}

	token, err := oauth2webflow.BrowserAuthCodeFlow(ctx, conf)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Token: %+v", token)
}
```
