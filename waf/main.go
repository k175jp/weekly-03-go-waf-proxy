package main

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var wafRegexSignatures = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:on[a-z]+\s*=|javascript\s*:)`),
	regexp.MustCompile(`(?i)<\/?script\b[^>]*>`),
	regexp.MustCompile(`(?i)<\/?.+(?:href|src|style)\s*=`),
	regexp.MustCompile(`(?i)'?\s*or\s+[^=]+=\s*[^=\s]+|'\s*--(.*)$|'\s*#|'\s*\/\*`),
	regexp.MustCompile(`(?i)(?:\.\.\/|\.\.\\\w+)|\/etc\/(?:passwd|hosts)`),
	regexp.MustCompile(`(?i)union(?:\s|%0a|%0d|\/\*[\s\S]*?\*\/)+select`),
}

func fullyURLDecode(input string) string {
	current := input
	for {
		decoded, err := url.QueryUnescape(current)
		if err != nil || decoded == current {
			break
		}
		current = decoded
	}
	return current
}

func fullyHTMLDecode(input string) string {
	current := input
	for {
		decoded := html.UnescapeString(current)
		if decoded == current {
			break
		}
		current = decoded
	}
	return current
}

var escapeRegex = regexp.MustCompile(`(?i)\\u[0-9a-f]{4}|\\x[0-9a-f]{2}`)

func decodeEscapes(input string) string {
	return escapeRegex.ReplaceAllStringFunc(input, func(match string) string {
		quoted := fmt.Sprintf(`"%s"`, match)
		unquoted, err := strconv.Unquote(quoted)
		if err != nil {
			return match
		}
		return unquoted
	})
}

func normalize(input string) string {
	current := input
	for {
		previous := current

		current = fullyURLDecode(current)
		current = fullyHTMLDecode(current)
		current = decodeEscapes(current)

		if current == previous {
			break
		}
	}
	return current
}

func isMalicious(input string) bool {
	normalized := normalize(input)
	lowerInput := strings.ToLower(normalized)
	for _, re := range wafRegexSignatures {
		if re.MatchString(lowerInput) {
			log.Printf("[WAF DETECT] Matched Pattern: %s", re.String())
			return true
		}
	}
	return false
}

func main() {
	targetURL, err := url.Parse("http://backend-app:8080")
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[WAF LOG] Incoming request: %s %s", r.Method, r.URL.Path)

		// todo
		for _, values := range r.URL.Query() {
			for _, val := range values {
				if isMalicious(val) {
					log.Printf("[🚨 WAF BLOCK] Malicious Query Param detected: %s", val)
					http.Error(w, "403 Forbidden: Blocked by WAF", http.StatusForbidden)
					return
				}
			}
		}

		if r.Body != nil && r.Body != http.NoBody {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				bodyStr := string(bodyBytes)
				if isMalicious(bodyStr) {
					log.Printf("[🚨 WAF BLOCK] Malicious Body Payload detected")
					http.Error(w, "403 Forbidden: Blocked by WAF", http.StatusForbidden)
					return
				}
			}
		}

		proxy.ServeHTTP(w, r)
	})

	log.Println("🚀 WAF Reverse Proxy running on http://localhost:9000")
	log.Printf("👉 Forwarding requests to: %s\n\n", targetURL)

	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}