package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
)

func check(path string) {
	spec, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	doc, err := libopenapi.NewDocument(spec)
	if err != nil {
		fmt.Println(path, "読めない:", err)
		return
	}
	v, errs := validator.NewValidator(doc)
	if len(errs) > 0 {
		fmt.Println(path, "validator:", errs)
		return
	}
	ok, verrs := v.ValidateDocument()
	fmt.Printf("%s: 定義として正しい=%v\n", path, ok)
	for _, e := range verrs {
		fmt.Println("  -", e.Message, e.Reason)
		for _, se := range e.SchemaValidationErrors {
			fmt.Println("    ", se.Reason)
		}
	}
	if path != "openapi.yaml" {
		return
	}

	// 実際のレスポンスが定義に合っているかも確かめる
	cases := []struct {
		status int
		ctype  string
		body   string
	}{
		{200, "application/json", `{"id":42,"title":"写真展","status":"approved"}`},
		{404, "application/problem+json", `{"type":"about:blank","title":"Not Found","status":404,"code":"not_found"}`},
		{200, "application/json", `{"id":42,"title":"写真展","status":"done"}`},
		{404, "application/problem+json", `{"error":"not found"}`},
	}
	for _, c := range cases {
		req, _ := http.NewRequest("GET", "https://example.com/applications/42", nil)
		res := &http.Response{
			StatusCode: c.status,
			Header:     http.Header{"Content-Type": {c.ctype}},
			Body:       io.NopCloser(bytes.NewBufferString(c.body)),
			Request:    req,
		}
		ok, errs := v.ValidateHttpResponse(req, res)
		fmt.Printf("  レスポンス %d %s: %v\n", c.status, c.body, ok)
		for _, e := range errs {
			for _, se := range e.SchemaValidationErrors {
				fmt.Println("     ", se.Reason)
			}
			if len(e.SchemaValidationErrors) == 0 {
				fmt.Println("     ", e.Message)
			}
		}
	}
}

func main() {
	for _, p := range os.Args[1:] {
		check(p)
	}
}
