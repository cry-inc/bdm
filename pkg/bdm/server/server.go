package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

// StartServer sets up a HTTP server and starts it.
func StartServer(port uint16, handler http.Handler) {
	binding := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(binding, handler))
}

// StartServerTLS sets up a HTTPS server with an existing certifacte and key and starts it.
func StartServerTLS(port uint16, certPath, certKeyPath string, handler http.Handler) {
	binding := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServeTLS(binding, certPath, certKeyPath, handler))
}

// StartServerLetsEncrypt sets up a LE HTTPS server for the specified domain.
// If certCacheFolder is empty, there will be no certificate caching.
func StartServerLetsEncrypt(port uint16, letsEncryptDomain, certCacheFolder string, handler http.Handler) {
	var manager autocert.Manager
	if len(certCacheFolder) > 0 {
		manager = autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(letsEncryptDomain),
			Cache:      autocert.DirCache(certCacheFolder),
		}
	} else {
		manager = autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(letsEncryptDomain),
		}
	}

	binding := fmt.Sprintf(":%d", port)
	tlsConfig := tls.Config{
		GetCertificate: manager.GetCertificate,
	}

	server := &http.Server{
		Addr:      binding,
		TLSConfig: &tlsConfig,
		Handler:   handler,
	}

	// Port 80 is required for LE certificate acquisition and forwarding
	go func() {
		httpHandler := manager.HTTPHandler(createFallbackHandler(port))
		log.Fatal(http.ListenAndServe(":80", httpHandler))
	}()

	// Key and cert are coming from LE
	log.Fatal(server.ListenAndServeTLS("", ""))
}

// Creates a handler that forwards all traffic to HTTPS at the specified port
func createFallbackHandler(port uint16) http.HandlerFunc {
	fallback := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "Use HTTPS", http.StatusBadRequest)
			return
		}
		target := "https://" + replacePort(r.Host, port) + r.URL.RequestURI()
		http.Redirect(w, r, target, http.StatusFound)
	}
	return http.HandlerFunc(fallback)
}

// Replaces the port of a hostname:port combination with the supplied port
func replacePort(host string, newPort uint16) string {
	hostOnly, _, err := net.SplitHostPort(host)
	if err != nil {
		return net.JoinHostPort(host, fmt.Sprint(newPort))
	}
	return net.JoinHostPort(hostOnly, fmt.Sprint(newPort))
}
