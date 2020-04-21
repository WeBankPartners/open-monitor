package funcs

import "net/http"

func InitHttpServer()  {
	http.Handle("/write", http.HandlerFunc(write))
	http.Handle("/read", http.HandlerFunc(read))
	http.ListenAndServe(":9202", nil)
}