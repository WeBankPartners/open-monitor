package funcs

import (
	"net/http"
	"fmt"
)

func InitHttpServer(port int)  {
	http.Handle("/write", http.HandlerFunc(write))
	http.Handle("/read", http.HandlerFunc(read))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}