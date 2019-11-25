package inputvalidationtest

import "strings"

var (
	String37Long = strings.Repeat("a", 37)
)

const (
	NonPrintableString   = "\u0000"
	PrintableString      = "汉ʋǟʟɨɖ ɨռքʊȶ!لْحُرُوف ٱلْعَرَبِيَّر"
	WhitespaceCharString = "\u0009"
	ValidName            = "thi5-1npu7.15-valid"
)
