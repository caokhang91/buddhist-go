package object

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/caokhang91/buddhist-go/pkg/tracing"
)

// ClosureCaller is a function type that can call a closure with arguments
// This is set by the VM to avoid import cycles
type ClosureCaller func(closure *Closure, args ...Object) (Object, error)

var (
	closureCallerMu sync.RWMutex
	closureCaller   ClosureCaller
)

// SetClosureCaller sets the function to call closures (called by VM)
func SetClosureCaller(caller ClosureCaller) {
	closureCallerMu.Lock()
	defer closureCallerMu.Unlock()
	closureCaller = caller
}

// ClearClosureCaller clears the closure caller
func ClearClosureCaller() {
	closureCallerMu.Lock()
	defer closureCallerMu.Unlock()
	closureCaller = nil
}

// callProgressCallback calls a progress callback if provided
func callProgressCallback(callback *Closure, data map[string]Object) {
	if callback == nil {
		return
	}

	closureCallerMu.RLock()
	caller := closureCaller
	closureCallerMu.RUnlock()

	if caller == nil {
		// No caller available, skip callback
		return
	}

	// Create hash from data
	hash := newStringHash(data)

	// Call the closure using the registered caller
	_, err := caller(callback, hash)
	if err != nil {
		// Silently ignore callback errors to not break the main operation
		// Could log to tracing if needed
		if tracing.IsEnabled() {
			tracing.Trace("Progress callback error: %v", err)
		}
	}
}

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

	// Check for progress callback
	var progressCallback *Closure
	if progressObj, ok := getHashValue(config, "progress"); ok {
		if closure, ok := progressObj.(*Closure); ok {
			progressCallback = closure
		} else {
			return newError("`%s` progress must be FUNCTION, got %s", fnName, progressObj.Type())
		}
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

	// Call progress callback: "connecting"
	callProgressCallback(progressCallback, map[string]Object{
		"stage":  &String{Value: "connecting"},
		"url":    &String{Value: urlValue},
		"method": &String{Value: method},
	})

	// Network I/O tracing
	tracing.TraceNetwork("Sending HTTP %s request to: %s", method, urlValue)
	if headersObj, ok := getHashValue(config, "headers"); ok {
		if headersHash, ok := headersObj.(*Hash); ok {
			headers, _ := hashToStringMap(headersHash, fnName, "headers")
			if len(headers) > 0 {
				tracing.TraceNetwork("Request headers: %v", headers)
			}
		}
	}
	if bodyReader != nil {
		tracing.TraceNetwork("Request body present")
	}

	// Call progress callback: "sending"
	callProgressCallback(progressCallback, map[string]Object{
		"stage":  &String{Value: "sending"},
		"url":    &String{Value: urlValue},
		"method": &String{Value: method},
	})

	networkStart := time.Now()
	resp, err := client.Do(req)
	networkDuration := time.Since(networkStart)
	
	if err != nil {
		// Call progress callback: "error"
		callProgressCallback(progressCallback, map[string]Object{
			"stage": &String{Value: "error"},
			"error": &String{Value: err.Error()},
		})
		tracing.TraceNetwork("Request failed: %s", err.Error())
		return newError("%s failed: %s", fnName, err.Error())
	}
	defer resp.Body.Close()

	// Call progress callback: "received"
	callProgressCallback(progressCallback, map[string]Object{
		"stage":  &String{Value: "received"},
		"status": &Integer{Value: int64(resp.StatusCode)},
		"url":    &String{Value: urlValue},
	})

	tracing.TraceNetwork("Received response: status=%d, duration=%v", resp.StatusCode, networkDuration)

	// Call progress callback: "reading"
	callProgressCallback(progressCallback, map[string]Object{
		"stage": &String{Value: "reading"},
	})

	readStart := time.Now()
	body, err := io.ReadAll(resp.Body)
	readDuration := time.Since(readStart)
	if err != nil {
		// Call progress callback: "error"
		callProgressCallback(progressCallback, map[string]Object{
			"stage": &String{Value: "error"},
			"error": &String{Value: err.Error()},
		})
		tracing.TraceNetwork("Failed to read response body: %s", err.Error())
		return newError("%s failed: %s", fnName, err.Error())
	}

	// Call progress callback: "complete"
	callProgressCallback(progressCallback, map[string]Object{
		"stage":    &String{Value: "complete"},
		"status":   &Integer{Value: int64(resp.StatusCode)},
		"bodySize": &Integer{Value: int64(len(body))},
		"duration": &Integer{Value: int64((networkDuration + readDuration).Milliseconds())},
	})

	tracing.TraceNetwork("Read response body: %d bytes, duration=%v", len(body), readDuration)
	tracing.TraceNetwork("Total network I/O time: %v", networkDuration+readDuration)

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
