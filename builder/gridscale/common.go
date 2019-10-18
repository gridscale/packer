package gridscale

import (
	"github.com/gridscale/gsclient-go/v3"
	"math/rand"
	"strings"
)

const CharSetAlphaNum = "abcdefghijklmnopqrstuvwxyz012346789"

// randString generates a random alphanumeric string of the length specified
func randString(strlen int) string {
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = CharSetAlphaNum[rand.Intn(len(CharSetAlphaNum))]
	}
	return string(result)
}

// getHeaderMapFromStr converts string (format: "key1:val1,key2:val2")
// to a HTTP header map
func convertStrToHeaderMap(str string) map[string]string {
	result := make(map[string]string)
	// split string into comma separated headers
	headers := strings.Split(str, ",")
	for _, header := range headers {
		if header != "" {
			// split each header into a key and a value
			kv := strings.Split(header, ":")
			if len(kv) == 2 {
				result[kv[0]] = kv[1]
			}
		}
	}
	return result
}

// removeErrorContainsHTTPCodes returns nil, if the error of HTTP error
//has status code that is in the given list of http status codes
func removeErrorContainsHTTPCodes(err error, errorCodes ...int) error {
	if requestError, ok := err.(gsclient.RequestError); ok {
		if containsInt(errorCodes, requestError.StatusCode) {
			err = nil
		}
	}
	return err
}

// containsInt check if an int array contains a specific int.
func containsInt(arr []int, target int) bool {
	for _, a := range arr {
		if a == target {
			return true
		}
	}
	return false
}
