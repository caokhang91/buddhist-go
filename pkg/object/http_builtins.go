package object

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func httpRequestBuiltin(args ...Object) Object {
	return httpRequestBuiltinWithName("http_request", args...)
}

func curlBuiltin(args ...Object) Object {
	return httpRequestBuiltinWithName("curl", args...)
}

func httpRequestBuiltinWithName(fnName string, args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	config, ok := args[0].(*Hash)
	if !ok {
		return newError("argument to `%s` must be HASH, got %s", fnName, args[0].Type())
	}

	urlValue, errObj := requiredStringField(config, "url", fnName)
	if errObj != nil {
		return errObj
	}

	method := "GET"
	if methodObj, ok := getHashValue(config, "method"); ok {
		method, errObj = stringFromScalar(methodObj, fnName, "method")
		if errObj != nil {
			return errObj
		}
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		return newError("`%s` method must be non-empty", fnName)
	}

	if paramsObj, ok := getHashValue(config, "params"); ok {
		paramsHash, ok := paramsObj.(*Hash)
		if !ok {
			return newError("`%s` params must be HASH, got %s", fnName, paramsObj.Type())
		}
		updatedURL, err := addQueryParams(urlValue, paramsHash, fnName)
		if err != nil {
			return err
		}
		urlValue = updatedURL
	}

	var bodyReader io.Reader
	if bodyObj, ok := getHashValue(config, "body"); ok {
		bodyReader, errObj = bodyReaderFromObject(bodyObj, fnName)
		if errObj != nil {
			return errObj
		}
	}

	req, err := http.NewRequest(method, urlValue, bodyReader)
	if err != nil {
		return newError("%s failed: %s", fnName, err.Error())
	}

	if headersObj, ok := getHashValue(config, "headers"); ok {
		headersHash, ok := headersObj.(*Hash)
		if !ok {
			return newError("`%s` headers must be HASH, got %s", fnName, headersObj.Type())
		}
		headers, errObj := hashToStringMap(headersHash, fnName, "headers")
		if errObj != nil {
			return errObj
		}
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	client := &http.Client{}
	if timeoutObj, ok := getHashValue(config, "timeout_ms"); ok {
		timeoutValue, errObj := intFromObject(timeoutObj, fnName, "timeout_ms")
		if errObj != nil {
			return errObj
		}
		if timeoutValue < 0 {
			return newError("`%s` timeout_ms must be non-negative", fnName)
		}
		client.Timeout = time.Duration(timeoutValue) * time.Millisecond
	}

	resp, err := client.Do(req)
	if err != nil {
		return newError("%s failed: %s", fnName, err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return newError("%s failed: %s", fnName, err.Error())
	}

	response := newStringHash(map[string]Object{
		"status":  &Integer{Value: int64(resp.StatusCode)},
		"body":    &String{Value: string(body)},
		"headers": headersToHash(resp.Header),
	})

	return response
}

func getHashValue(hash *Hash, key string) (Object, bool) {
	if hash == nil {
		return nil, false
	}
	keyObj := &String{Value: key}
	pair, ok := hash.Pairs[keyObj.HashKey()]
	if !ok {
		return nil, false
	}
	return pair.Value, true
}

func requiredStringField(hash *Hash, key string, fnName string) (string, *Error) {
	value, ok := getHashValue(hash, key)
	if !ok {
		return "", newError("`%s` requires %q", fnName, key)
	}
	return stringFromScalar(value, fnName, key)
}

func stringFromScalar(obj Object, fnName, field string) (string, *Error) {
	switch v := obj.(type) {
	case *String:
		return v.Value, nil
	case *Integer:
		return strconv.FormatInt(v.Value, 10), nil
	case *Float:
		return strconv.FormatFloat(v.Value, 'g', -1, 64), nil
	case *Boolean:
		if v.Value {
			return "true", nil
		}
		return "false", nil
	default:
		return "", newError("`%s` %s must be STRING, got %s", fnName, field, obj.Type())
	}
}

func intFromObject(obj Object, fnName, field string) (int64, *Error) {
	switch v := obj.(type) {
	case *Integer:
		return v.Value, nil
	case *Float:
		return int64(v.Value), nil
	default:
		return 0, newError("`%s` %s must be INTEGER, got %s", fnName, field, obj.Type())
	}
}

func bodyReaderFromObject(obj Object, fnName string) (io.Reader, *Error) {
	switch v := obj.(type) {
	case *Null:
		return nil, nil
	case *String:
		return strings.NewReader(v.Value), nil
	case *Blob:
		return bytes.NewReader(v.Data), nil
	default:
		return nil, newError("`%s` body must be STRING or BLOB, got %s", fnName, obj.Type())
	}
}

func addQueryParams(rawURL string, params *Hash, fnName string) (string, *Error) {
	if params == nil {
		return rawURL, nil
	}
	queryValues, errObj := hashToStringMap(params, fnName, "params")
	if errObj != nil {
		return "", errObj
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", newError("`%s` url invalid: %s", fnName, err.Error())
	}
	values := parsed.Query()
	for key, value := range queryValues {
		values.Set(key, value)
	}
	parsed.RawQuery = values.Encode()
	return parsed.String(), nil
}

func hashToStringMap(hash *Hash, fnName, field string) (map[string]string, *Error) {
	result := make(map[string]string, len(hash.Pairs))
	for _, pair := range hash.Pairs {
		keyObj, ok := pair.Key.(*String)
		if !ok {
			return nil, newError("`%s` %s keys must be STRING, got %s", fnName, field, pair.Key.Type())
		}
		value, errObj := stringFromScalar(pair.Value, fnName, field)
		if errObj != nil {
			return nil, errObj
		}
		result[keyObj.Value] = value
	}
	return result, nil
}

func headersToHash(headers http.Header) *Hash {
	pairs := make(map[string]Object, len(headers))
	for key, values := range headers {
		pairs[key] = &String{Value: strings.Join(values, ",")}
	}
	return newStringHash(pairs)
}

func newStringHash(pairs map[string]Object) *Hash {
	hashPairs := make(map[HashKey]HashPair, len(pairs))
	for key, value := range pairs {
		keyObj := &String{Value: key}
		hashPairs[keyObj.HashKey()] = HashPair{Key: keyObj, Value: value}
	}
	return &Hash{Pairs: hashPairs}
}
