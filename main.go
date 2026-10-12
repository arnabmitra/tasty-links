package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight OPTIONS request immediately
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

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

	modifyResponse := func(resp *http.Response) error {
		resp.Header.Set("Access-Control-Allow-Origin", "*")
		resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		resp.Header.Set("Access-Control-Allow-Headers", "*")
		return nil
	}
	prodProxy.ModifyResponse = modifyResponse
	certProxy.ModifyResponse = modifyResponse

	// Route API requests through the proxy with CORS & preflight handling
	http.Handle("/proxy/prod/", corsMiddleware(http.StripPrefix("/proxy/prod", prodProxy)))
	http.Handle("/proxy/cert/", corsMiddleware(http.StripPrefix("/proxy/cert", certProxy)))

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
