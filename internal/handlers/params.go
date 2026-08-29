package handlers

import (
	"net/http"
	"strconv"
)

func pathInt32(r *http.Request, key string) (int32, bool) {
	v, err := strconv.ParseInt(r.PathValue(key), 10, 32)
	if err != nil {
		return 0, false
	}

	return int32(v), true
}

func pathInt64(r *http.Request, key string) (int64, bool) {
	v, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil {
		return 0, false
	}

	return v, true
}
