package coinbase

import "fmt"

type FiatCurrencies struct {
	Data []struct {
		Id      string `json:"id"`
		Name    string `json:"name"`
		MinSize string `json:"min_size"`
	} `json:"data"`
}

func (c *Client) Currencies(id string) (FiatCurrencies, error) {
	var response FiatCurrencies

	_, err := c.Request("GET", fmt.Sprintf("/currencies"), nil, &response)

	return response, err
}
