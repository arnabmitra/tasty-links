package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	prodURL, _ := url.Parse("https://api.tastyworks.com")
	certURL, _ := url.Parse("https://api.cert.tastyworks.com")

	// Create reverse proxies
	prodProxy := httputil.NewSingleHostReverseProxy(prodURL)
	certProxy := httputil.NewSingleHostReverseProxy(certURL)

	// Tastytrade requires the Host header to match their domain
	directorProd := prodProxy.Director
	prodProxy.Director = func(req *http.Request) {
		directorProd(req)
		req.Host = prodURL.Host
	}

	directorCert := certProxy.Director
	certProxy.Director = func(req *http.Request) {
		directorCert(req)
		req.Host = certURL.Host
	}

	// Route API requests through the proxy to bypass browser CORS
	http.Handle("/proxy/prod/", http.StripPrefix("/proxy/prod", prodProxy))
	http.Handle("/proxy/cert/", http.StripPrefix("/proxy/cert", certProxy))

	// Serve static files and route root "/" to the referral links page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/templates/referral.html", http.StatusFound)
			return
		}
		http.FileServer(http.Dir(".")).ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Go Server listening on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
