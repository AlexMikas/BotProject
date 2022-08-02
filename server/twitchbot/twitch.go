// Twitch API

/*
To identify your application to the API, every request must include your application’s client ID,
either explicitly or implicitly by including an OAuth token.
If you use an OAuth token in your request, the API figures out the client ID for you.

Requests can include both a client ID and an OAuth token.
Requests without either one fail with an HTTP 400 error.
*/
package twitchbot

import (
	"encoding/json"
	"io"

	//"context"
	//"errors"
	"fmt"
	"net/http"
	/*
		"github.com/hortbot/hortbot/internal/pkg/oauth2x"
		"golang.org/x/oauth2"
		"golang.org/x/oauth2/clientcredentials"
		"golang.org/x/oauth2/twitch" */)

/*
type httpClient struct {
	cli     *http.Client
	ts      oauth2.TokenSource
	headers http.Header
}

var twitchEndpoint = oauth2.Endpoint{
	AuthURL:   twitch.Endpoint.AuthURL,
	TokenURL:  twitch.Endpoint.TokenURL,
	AuthStyle: oauth2.AuthStyleInParams,
}
*/
const (
	krakenRoot = "https://api.twitch.tv/kraken"
	helixRoot  = "https://api.twitch.tv/helix"
)

type TwitchUser struct {
	ID          int64  `json:"_id"`
	Bio         string `json:"bio"`
	CreatedAt   string `json:"created_at"`
	DisplayName string `json:"display_name,omitempty"`
	Logo        string `json:"logo"`
	Name        string `json:"name"`
	UserType    string `json:"type"`
	UpdatedAt   string `json:"updated_at"`
}
type Userr struct {
	ID          int64  `json:"_id"`
	Name        string `json:"login"`
	DisplayName string `json:"display_name,omitempty"`
}

func DecodeSingle(r io.Reader, v interface{}) error {
	d := json.NewDecoder(r)
	if err := d.Decode(v); err != nil {
		return err
	}

	if _, err := d.Token(); err != io.EOF {
		return err
	}

	return nil
}

func (bb *BasicBot) GetUserByName(userName string) {
	//client := &http.Client{}
	req, _ := http.NewRequest("GET", krakenRoot+"/users?login="+userName, nil)
	req.Header = bb.headers

	response, err := bb.client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	//defer response.Body.Close()
	/*
		body := &struct {
			Data []*Userr `json:"data"`
		}{}*/
	//tu := &TwitchUser{}
	// marshal unmarshal
	//DecodeSingle(response.Body, body)
	//json.NewDecoder(response.Body).Decode(body)
	//json.NewDecoder(strings.NewReader(strings.NewReader(string(response.Body))).Decode(tu)

	var result map[string]interface{}

	json.NewDecoder(response.Body).Decode(&result)

	//users := body.Data
	// TODO:: Return json
	// log.Println(result["users"])
	//return result
}
