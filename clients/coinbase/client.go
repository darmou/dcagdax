package coinbase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/sberserker/dcagdax/clients/coinbasev3"
)

type Client struct {
	BaseURL string
	Secret  string
	Key     string
}

func NewClient(secret, key, passphrase string) *Client {
	client := Client{
		BaseURL: "https://api.coinbase.com/v2",
		Secret:  secret,
		Key:     key,
	}

	return &client
}

func (c *Client) Request(method string, url string,
	params, result interface{}) (res *http.Response, err error) {
	var data []byte
	body := bytes.NewReader(make([]byte, 0))

	if params != nil {
		data, err = json.Marshal(params)
		if err != nil {
			return res, err
		}

		body = bytes.NewReader(data)
	}

	fullURL := fmt.Sprintf("%s%s", c.BaseURL, url)
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return res, err
	}

	uri := fmt.Sprintf("%s %s/v2%s", method, req.Host, url)
	jwt, err := coinbasev3.BuildJWT(uri, c.Key, c.Secret)
	if err != nil {
		return res, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))

	client := http.Client{}
	res, err = client.Do(req)
	if err != nil {
		return res, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 && res.StatusCode != 201 {
		defer res.Body.Close()
		// coinbaseError := Error{}
		// decoder := json.NewDecoder(res.Body)
		// if err := decoder.Decode(&coinbaseError); err != nil {
		// 	return res, err
		// }
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return res, err
		}

		mgs := fmt.Sprintf("status: %d, body: %s", res.StatusCode, string(bodyBytes))
		return res, errors.New(mgs)
	}

	if result != nil {
		decoder := json.NewDecoder(res.Body)
		if err = decoder.Decode(result); err != nil {
			return res, err
		}
	}

	return res, nil
}
