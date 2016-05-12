package icarus

import (
	"fmt"
	"net/http"
)



func handleBase(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi, there, you're at %s", r.URL.Path[1:])
}


func Serve(loc string) {
	http.HandleFunc("/", handleBase)
	http.ListenAndServe(loc, nil)
}
