package resreader

import (
	"compress/gzip"
	"errors"
	"github.com/andybalholm/brotli"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
)

func ReadCloserFor(res *http.Response) (io.ReadCloser, error) {
	if res == nil {
		return nil, errors.New("http response was nil")
	}

	if res.Body == nil {
		return nil, errors.New("http response body was nil")
	}

	switch res.Header.Get("content-encoding") {
	case "gzip":
		return gzip.NewReader(res.Body)
	case "br":
		return ioutil.NopCloser(brotli.NewReader(res.Body)), nil
	default:
		return res.Body, nil
	}
}

func ReadBody(res *http.Response) ([]byte, error) {
	defer Close(res)

	rdr, err := ReadCloserFor(res)
	if err != nil {
		return nil, err
	}
	defer rdr.Close()

	return ioutil.ReadAll(rdr)
}

func Parse[T interface{}](res *http.Response, parser func(io.Reader) *T) (*T, error) {
	defer Close(res)

	rdr, err := ReadCloserFor(res)
	if err != nil {
		return nil, err
	}
	defer rdr.Close()

	return parser(rdr), nil
}

func ParseErr[T interface{}](res *http.Response, parser func(io.Reader) (*T, error)) (*T, error) {
	defer Close(res)

	rdr, err := ReadCloserFor(res)
	if err != nil {
		return nil, err
	}
	defer rdr.Close()

	return parser(rdr)
}

func ParseDoc(res *http.Response) (*html.Node, error) {
	return ParseErr[html.Node](res, htmlquery.Parse)
}

type Decoder interface {
	Decode(obj interface{}) error
}

func Decode[D Decoder, O interface{}](newDecoder func(io.Reader) D, res *http.Response, obj *O) (*O, error) {
	defer Close(res)

	rdr, err := ReadCloserFor(res)
	if err != nil {
		return nil, err
	}
	defer rdr.Close()

	dec := newDecoder(rdr)
	if err := dec.Decode(obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func Decode_[D Decoder, O interface{}](newDecoder func(io.Reader) D, res *http.Response, obj *O) error {
	_, err := Decode(newDecoder, res, obj)
	return err
}

func Close(res *http.Response) error {
	if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
		res.Body.Close()
		return err
	}

	if err := res.Body.Close(); err != nil {
		return err
	}

	return nil
}
