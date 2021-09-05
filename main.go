package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

var staticTypes = []string{"ico", "png", "svg", "jpg", "jpeg", "css", "js", "json", "webmanifest"}

// Check url request file
func checkStatic(lastPathName string) bool {
	for _, staticType := range staticTypes {
		if staticType == lastPathName {
			return true
		}
	}

	return false
}

func handler(w http.ResponseWriter, r *http.Request) {
	// The second part of path is remote URL
	// https://localhost:8080/example.com -> example.com
	remoteUrl := r.URL.Path[1:]
	query := r.URL.RawQuery
	// fmt.Println(r.URL.RawQuery)

	// Check URL is "https://localhost:8080/"
	if remoteUrl == "" && query == "" {
		fmt.Fprintf(w, "Hello, world!")
	} else {

		// Get the first part of remote URL
		// https://localhost:8080/example.com/about -> [example, com]
		// https://localhost:8080/about -> [about]
		// https://localhost:8080/favicon.ico -> [favicon, ico]
		remoteUrlComponents := strings.Split(remoteUrl, "/")
		remoteHostnameUrl := strings.Split(remoteUrlComponents[0], ".")

		// Check URL relative
		// https://localhost:8080/about -> [about] -> example.com/about
		// https://localhost:8080/favicon.ico -> [favicon, ico] -> example.com/favicon.ico
		remoteHostnameUrlLen := len(remoteHostnameUrl)
		if remoteHostnameUrlLen == 1 || checkStatic(remoteHostnameUrl[remoteHostnameUrlLen-1]) {
			// https://localhost:8080/about -> https://localhost:8080/example.com
			referer := r.Header.Get("Referer")
			refererComponents := strings.Split(referer, "/")
			// fmt.Println("Referer " + referer + "\n")

			if len(refererComponents) > 2 {
				remoteUrl = "/" + refererComponents[3] + "/" + remoteUrl
			}

			// fmt.Println(query)
			if query != "" {
				remoteUrl = remoteUrl + "?" + query
			}

			// fmt.Println("Redirect " + remoteUrl + "\n")
			// Redirect relative path /example.com/about
			http.Redirect(w, r, remoteUrl, http.StatusFound)
			return
		}

		// fmt.Println(query)
		if query != "" {
			remoteUrl = remoteUrl + "?" + query
		}
		// example.com -> https://example.com
		remoteUrl = "https://" + remoteUrl

		// Debug log
		// fmt.Println("Get " + remoteUrl + "\n")

		// GET https://example.com
		client := &http.Client{}
		request, _ := http.NewRequest("GET", remoteUrl, nil)

		response, err := client.Do(request)
		if err != nil {
			fmt.Println(remoteUrl)
		}
		defer response.Body.Close()

		// Header https://example.com -> Header https://localhost:8080/example.com
		remoteHeader := response.Header
		for headerType, headerValue := range remoteHeader {
			w.Header().Set(headerType, strings.Join(headerValue, ";"))
		}

		// Body https://example.com -> Body https://localhost:8080/example.com
		_, err = io.Copy(w, response.Body)
		if err != nil {
			fmt.Println(err)
			fmt.Println(remoteUrl)
		}
	}
}

func main() {
	// https://localhost:8080/
	// https://localhost:8080/example.com
	// https://localhost:8080/example.com/about
	// https://localhost:8080/about
	http.HandleFunc("/", handler)

	// Listen at localhost:8080
	http.ListenAndServe(":8080", nil)
}
