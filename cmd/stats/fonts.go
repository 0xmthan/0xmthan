package main

import (
	_ "embed"
	"sync"
)

//go:embed fonts/jbmono-400.woff2
var jbmono400 []byte

//go:embed fonts/jbmono-600.woff2
var jbmono600 []byte

//go:embed fonts/jbmono-head.woff2
var jbmonoHead []byte

var fontText = sync.OnceValue(func() string {
	return face(jbmono400, 400) + face(jbmono600, 600)
})

var fontHead = sync.OnceValue(func() string {
	return face(jbmonoHead, 600)
})
