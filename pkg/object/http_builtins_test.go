package object

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPRequestBuiltinGETWithParams(t *testing.T) {
	errCh := make(chan error, 1)
	recordErr := func(err error) {
		select {
		case errCh <- err:
		default:
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			recordErr(errors.New("unexpected method"))
		}
		if r.URL.Query().Get("q") != "buddhist" {
			recordErr(errors.New("missing query parameter"))
		}
		if r.URL.Query().Get("page") != "2" {
			recordErr(errors.New("unexpected page parameter"))
		}
		if r.Header.Get("X-Test") != "yes" {
			recordErr(errors.New("missing header"))
		}
		w.Header().Set("X-Reply", "ok")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("pong"))
	}))
	defer server.Close()

	config := newTestHash(map[string]Object{
		"method":  &String{Value: "GET"},
		"url":     &String{Value: server.URL},
		"params":  newTestHash(map[string]Object{"q": &String{Value: "buddhist"}, "page": &Integer{Value: 2}}),
		"headers": newTestHash(map[string]Object{"X-Test": &String{Value: "yes"}}),
	})

	builtin := GetBuiltinByName("http_request")
	if builtin == nil {
		t.Fatalf("http_request builtin not found")
	}
	result := builtin.Fn(config)
	if errObj, ok := result.(*Error); ok {
		t.Fatalf("http_request returned error: %s", errObj.Inspect())
	}

	hash, ok := result.(*Hash)
	if !ok {
		t.Fatalf("expected hash response, got %T", result)
	}

	status, ok := getTestHashInt(hash, "status")
	if !ok || status != http.StatusCreated {
		t.Fatalf("unexpected status: %d", status)
	}
	body, ok := getTestHashString(hash, "body")
	if !ok || body != "pong" {
		t.Fatalf("unexpected body: %q", body)
	}

	headers, ok := getTestHashHash(hash, "headers")
	if !ok {
		t.Fatalf("expected headers hash")
	}
	reply, ok := getTestHashString(headers, "X-Reply")
	if !ok || reply != "ok" {
		t.Fatalf("unexpected header value: %q", reply)
	}

	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}

func TestCurlBuiltinPOSTBody(t *testing.T) {
	errCh := make(chan error, 1)
	recordErr := func(err error) {
		select {
		case errCh <- err:
		default:
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			recordErr(errors.New("unexpected method"))
		}
		if r.Header.Get("Content-Type") != "text/plain" {
			recordErr(errors.New("missing content type"))
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			recordErr(err)
		}
		if string(body) != "hello" {
			recordErr(errors.New("unexpected body"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	config := newTestHash(map[string]Object{
		"method":  &String{Value: "POST"},
		"url":     &String{Value: server.URL},
		"headers": newTestHash(map[string]Object{"Content-Type": &String{Value: "text/plain"}}),
		"body":    &String{Value: "hello"},
	})

	builtin := GetBuiltinByName("curl")
	if builtin == nil {
		t.Fatalf("curl builtin not found")
	}
	result := builtin.Fn(config)
	if errObj, ok := result.(*Error); ok {
		t.Fatalf("curl returned error: %s", errObj.Inspect())
	}

	hash, ok := result.(*Hash)
	if !ok {
		t.Fatalf("expected hash response, got %T", result)
	}
	status, ok := getTestHashInt(hash, "status")
	if !ok || status != http.StatusOK {
		t.Fatalf("unexpected status: %d", status)
	}

	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}

func newTestHash(pairs map[string]Object) *Hash {
	hashPairs := make(map[HashKey]HashPair, len(pairs))
	for key, value := range pairs {
		keyObj := &String{Value: key}
		hashPairs[keyObj.HashKey()] = HashPair{Key: keyObj, Value: value}
	}
	return &Hash{Pairs: hashPairs}
}

func getTestHashValue(hash *Hash, key string) (Object, bool) {
	keyObj := &String{Value: key}
	pair, ok := hash.Pairs[keyObj.HashKey()]
	if !ok {
		return nil, false
	}
	return pair.Value, true
}

func getTestHashString(hash *Hash, key string) (string, bool) {
	value, ok := getTestHashValue(hash, key)
	if !ok {
		return "", false
	}
	str, ok := value.(*String)
	if !ok {
		return "", false
	}
	return str.Value, true
}

func getTestHashInt(hash *Hash, key string) (int64, bool) {
	value, ok := getTestHashValue(hash, key)
	if !ok {
		return 0, false
	}
	integer, ok := value.(*Integer)
	if !ok {
		return 0, false
	}
	return integer.Value, true
}

func getTestHashHash(hash *Hash, key string) (*Hash, bool) {
	value, ok := getTestHashValue(hash, key)
	if !ok {
		return nil, false
	}
	child, ok := value.(*Hash)
	if !ok {
		return nil, false
	}
	return child, true
}
