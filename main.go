package main

import (
	"net/http"
	"time"

	"strings"

	"fmt"

	"io/ioutil"

	"hash/crc32"

	"strconv"

	"os"

	"github.com/husobee/vestigo"
)

func main() {
	// New request router
	mux := vestigo.NewRouter()

	mux.HandleFunc("/", Handle404)
	//mux.HandleFunc("/favicon.ico", Handle404)
	// Only one handler
	mux.HandleFunc("/:path", HandleGetImage)

	http.ListenAndServe(":7004", mux)
}

// HandleGetImage receives incoming requests and decides what to do with them
func HandleGetImage(w http.ResponseWriter, r *http.Request) {
	var file []byte
	site := vestigo.Param(r, "path")
	site = strings.TrimSuffix(site, ".ico")
	info, err := os.Stat("./cache/" + site + ".ico")
	if !os.IsNotExist(err) && (time.Now().Unix()-info.ModTime().Unix()) < 604800 {
		file, err = ioutil.ReadFile("./cache/" + site + ".ico")
		if err != nil {
			w.Write([]byte("Could not open file"))
		}
	} else {
		favresp, err := http.Get("http://" + site + "/favicon.ico")
		if favresp != nil && favresp.StatusCode == http.StatusOK {
			if err != nil {
				fmt.Println(err)
			}
			defer favresp.Body.Close()
			file, err = ioutil.ReadAll(favresp.Body)
			if err != nil {
				fmt.Println(err)
			}
			ioutil.WriteFile("./cache/"+site+".ico", file, 0777)
			os.Chtimes("./cache/"+site+".ico", time.Now(), time.Now())
		} else {
			file, err = ioutil.ReadFile("./fallback.ico")
		}

	}
	fmt.Println(site)

	if r.Header.Get("If-None-Match") == "" {
		w.Header().Set("Cache-Control", "max-age=604800, public")
		w.Header().Set("Content-Type", "image/vnd.microsoft.icon")
		w.Header().Set("Etag", strconv.FormatUint(uint64(crc32.ChecksumIEEE(file)), 32))
		w.Write(file)
	} else if r.Header.Get("If-None-Match") == strconv.FormatUint(uint64(crc32.ChecksumIEEE(file)), 32) {
		w.Header().Set("Cache-Control", "max-age=604800, public")
		w.WriteHeader(http.StatusNotModified)
	}

}

// Handle404 Handles a user requesting a random invalid path
func Handle404(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store, no-cache, must-revalidate, max-age=0")
	if strings.Contains(r.Header.Get("Accept"), "html") {
		w.Write([]byte("404 not found"))
	} else {
		w.Write([]byte("send fallback image"))
	}
}
