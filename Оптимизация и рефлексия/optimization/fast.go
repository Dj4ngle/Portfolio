package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

type User struct {
	Browsers []string
	Email    string
	Name     string
}

// easyjson9f2eff5fDecode12hwData release unmarshalling
func easyjson9f2eff5fDecode12hwData(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					v1 := in.String()
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "email":
			out.Email = in.String()
		case "name":
			out.Name = in.String()
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9f2eff5fDecode12hwData(&r, v)
	return r.Error()
}

// FastSearch implements a quick file search
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("There is no file: ", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	seenBrowsers := make(map[string]bool, 120)

	androidRE := regexp.MustCompile("Android")
	msieRE := regexp.MustCompile("MSIE")

	fmt.Fprintln(out, "found users:")

	var isAndroid bool
	var isMSIE bool

	userIndex := 0
	user := User{}
	for scanner.Scan() {
		line := scanner.Bytes()

		err = user.UnmarshalJSON(line)
		if err != nil {
			log.Println("JSON Unmarshal error:", err)
			continue
		}

		isAndroid = false
		isMSIE = false

		browsers := user.Browsers

		for _, browser := range browsers {
			if androidRE.MatchString(browser) {
				isAndroid = true
				if !seenBrowsers[browser] {
					seenBrowsers[browser] = true
				}
			}
			if msieRE.MatchString(browser) {
				isMSIE = true
				if !seenBrowsers[browser] {
					seenBrowsers[browser] = true
				}
			}
		}

		if isAndroid && isMSIE {
			email := strings.ReplaceAll(user.Email, "@", " [at] ")
			fmt.Fprintf(out, "[%d] %s <%s>\n", userIndex, user.Name, email)
		}
		userIndex++
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}
