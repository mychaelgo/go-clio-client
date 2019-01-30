package utils

import (
	"fmt"
	"strings"
)

type CookieMap map[string]string

func CookieStringToMap(cookieString string) CookieMap {
	mapCookie := map[string]string{}

	s := strings.Split(cookieString, ";")
	for _, v := range s {
		c := strings.Split(v, "=")
		if len(c) > 1 {
			mapCookie[c[0]] = c[1]
		}
	}

	return mapCookie
}

func CookieMapToString(cookieMap CookieMap) string {
	stringCookie := ""

	for k, v := range cookieMap {
		stringCookie += fmt.Sprintf("%s=%s;", k, v)
	}

	return stringCookie
}

func MergeCookieMap(maps ...CookieMap) CookieMap {
	result := make(CookieMap)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
