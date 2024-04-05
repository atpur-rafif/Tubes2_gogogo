package main

import (
	"fmt"
	"net/http"
	"net/url"
)

func main() {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		parsed, _ := url.ParseQuery(r.URL.RawQuery)

		starts, exist := parsed["start"]
		if !exist {
			fmt.Fprintf(w, "Empty start parameter")
			return
		}

		ends, exist := parsed["end"]
		if !exist {
			fmt.Fprintf(w, "Empty end parameter")
			return
		}

		start := starts[0]
		end := ends[0]

		fmt.Println(start, end)
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
