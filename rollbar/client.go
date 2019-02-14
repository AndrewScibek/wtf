package rollbar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/wtfutil/wtf/wtf"
)

var ROLLBAR_HOST = map[bool]string{
	false: "travis-ci.org",
	true:  "travis-ci.com",
}

func ItemsFor() (*ActiveItems, error) {
	items := &ActiveItems{}

	access_token := wtf.Config.UString("wtf.mods.rollbar.access_token", "")
	rollbarAPIURL.Host = "api.rollbar.com/api/1/items"

	resp, err := rollbarItemRequest(access_token)
	if err != nil {
		return items, err
	}

	parseJson(&items, resp.Body)

	return items, nil
}

/* -------------------- Unexported Functions -------------------- */

var (
	rollbarAPIURL = &url.URL{Scheme: "https", Path: "/"}
)

func rollbarItemRequest(access_token string) (*http.Response, error) {
	params := url.Values{}
	params.Add("access_token", access_token)
	params.Add("status", "active")

	requestUrl := rollbarAPIURL.ResolveReference(&url.URL{RawQuery: params.Encode()})

	req, err := http.NewRequest("GET", requestUrl.String(), nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf(resp.Status)
	}

	return resp, nil
}

func parseJson(obj interface{}, text io.Reader) {
	jsonStream, err := ioutil.ReadAll(text)
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonStream))

	for {
		if err := decoder.Decode(obj); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
}
