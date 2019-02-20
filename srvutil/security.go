package srvutil

import (
	"net/http"
)

func defaultHeaders() map[string]string {
	return map[string]string{
		// Prevents clickjacking vulnerabilities by instructing browsers to disallow the site from being rendered in an
		// iFrame. This breaks some sites so ALLOW-FROM is recommended in combination with a Content Security Policy
		// defining permitted FRAME-ANCESTORS
		"X-Frame-Options": "DENY",

		// The X-Content-Type-Options response HTTP header is a marker used by the server to instruct browsers that the
		// MIME types advertised in the Content-Type headers should not be changed and be followed. This prevents
		// unintended vulnerabilities such as XSS by guessing the content-type based on file contents.
		"X-Content-Type-Options": "nosniff",

		// The X-Download-Options is specific to IE 8, and is related to how IE 8 handles downloaded HTML files. If you
		// download an HTML file from a web page and chooses to "Open" it in IE8, it will execute in the context of the
		// web site. That means that any scripts in that file will also execute with the origin of the web site.
		"X-Download-Options": "noopen",

		// A cross-domain policy file is an XML document that grants a web client, such as Adobe Flash Player or Adobe
		// Acrobat (though not necessarily limited to these), permission to handle data across domains. When clients
		// request content hosted on a particular source domain and that content make requests directed towards a domain
		// other than its own, the remote domain needs to host a cross-domain policy file that grants access to the
		// source domain, allowing the client to continue the transaction. The none value ensures no policy files are
		// allowed anywhere on the target server, including this master policy file limiting cross-domain communication.
		// An example of use is when using PDF or Flash.
		"X-Permitted-Cross-Domain-Policies": "none",

		// The HTTP X-XSS-Protection response header is a feature of Internet Explorer, Chrome and Safari that stops
		// pages from loading when they detect reflected cross-site scripting (XSS) attacks.
		"X-Xss-Protection": "1; mode=block",

		// The Referrer-Policy HTTP header governs which referrer information, sent in the Referrer header, should be
		// included with requests made. origin-when-cross-origin sends a full URL when performing a same-origin
		// request, but only send the origin of the document for other cases.
		"Referrer-Policy": "origin-when-cross-origin",

		// Content Security Policy (CSP) is an added layer of security that helps to detect and mitigate certain types
		// of attacks, including Cross Site Scripting (XSS) and data injection attacks. These attacks are used for
		// everything from data theft to site defacement or distribution of malware.
		// block-all-mixed-content; upgrade-insecure-requests; ensures that all requests are HTTPS.
		"Content-Security-Policy": "block-all-mixed-content; upgrade-insecure-requests;",

		// The HTTP Strict-Transport-Security response header (often abbreviated as HSTS or STS) lets a website tell
		// browsers that it should only be accessed using HTTPS, instead of using HTTP. max-age is the time, in seconds,
		// that the browser should remember that a site is only to be accessed using HTTPS. includeSubdomains applies
		// the rule to all of the site's subdomains as well.
		"Strict-Transport-Security": "max-age=631139040; includeSubdomains",
	}
}

type SecurityHeaderOption func(headers map[string]string)

// SecurityHeaderMiddleware set default security headers.
func SecurityHeaderMiddleware(options ...SecurityHeaderOption) func(http.Handler) http.Handler {
	headers := defaultHeaders()
	for _, opt := range options {
		opt(headers)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for key, val := range headers {
				w.Header().Set(key, val)
			}
			next.ServeHTTP(w, r)
		})
	}
}
