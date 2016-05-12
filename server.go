package icarus

import (
	"fmt"
	"net/http"
)



func handleBase(w http.ResponseWriter, r *http.Request) {
	page, err := PageFromRedis(r.URL.Path[1:])
	if err != nil {
		fmt.Fprintf(w, "Error opening slug '%s'\n%v\n\n%v", r.URL.Path[1:], err, page)
	} else {
		fmt.Fprintf(w, "Hi, there, you're at slug '%s'\n\n%v", r.URL.Path[1:], page)
	}

	
}


func Serve(loc string) {
	http.HandleFunc("/", handleBase)
	http.ListenAndServe(loc, nil)
}
