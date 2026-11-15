package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	rec := httptest.NewRecorder()
	writeProblem(rec, http.StatusConflict, "already_approved", "承認済みの申し込みは編集できません")
	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Content-Type"))
	fmt.Print(rec.Body.String())
}
