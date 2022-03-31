package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	jsonContentType = "application/json"
)

var errTransport = errors.New("failed to deliver payload")

type HTTPClient interface {
	MakeJSONRequest(request, response interface{}, method, reqURL string) error
}

type httpClient struct {
	client http.Client
}

func (client *httpClient) MakeJSONRequest(request, response interface{}, method, reqURL string) error {
	var bodyReader io.Reader
	if request != nil {
		body, err := json.Marshal(request)
		if err != nil {
			return errors.WithStack(err)
		}
		bodyReader = bytes.NewReader(body)
	}
	resBody, err := client.makeRequest(bodyReader, jsonContentType, method, reqURL)
	if err != nil {
		return err
	}
	defer resBody.Close()

	if response != nil {
		return unmarshalJSON(resBody, response)
	}
	return nil
}

func (client *httpClient) makeRequest(bodyReader io.Reader, contentType, method, reqURL string) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.Header.Set("Content-Type", contentType)
	res, err := client.client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if res.StatusCode >= http.StatusBadRequest {
		_ = res.Body.Close()
		return nil, errTransport
	}
	if res.StatusCode != http.StatusOK {
		_ = res.Body.Close()
		return nil, errors.New("invalid status code: " + res.Status)
	}
	return res.Body, nil
}

func NewHTTPClient(client http.Client) HTTPClient {
	return &httpClient{
		client: client,
	}
}

func unmarshalJSON(body io.Reader, response interface{}) error {
	content, err := io.ReadAll(body)
	if err != nil {
		return errors.WithStack(err)
	}
	err = json.Unmarshal(content, response)
	return errors.WithStack(err)
}
