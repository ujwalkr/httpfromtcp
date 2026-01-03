package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func toStr(b []byte) string {
	out := ""
	for _, b := range b {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

func response200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func response400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func response500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		statusCode := response.StatusOK
		body := response200()
		target := req.RequestLine.RequestTarget
		if target == "/yourproblem" {
			statusCode = response.StatusBadRequest
			body = response400()
			h.Replace("content-length", fmt.Sprintf("%d", len(body)))
			h.Replace("content-type", "text/html")
		} else if target == "/myproblem" {
			statusCode = response.StatusInternalServerError
			body = response500()
			h.Replace("content-length", fmt.Sprintf("%d", len(body)))
			h.Replace("content-type", "text/html")
		} else if target == "/video" {
			f, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				statusCode = response.StatusBadRequest
				body = response400()
			}
			h.Replace("content-type", "video/mp4")
			h.Replace("content-length", fmt.Sprintf("%d", len(f)))
			body = f

		} else if strings.HasPrefix(target, "/httpbin") {
			r, err := http.Get("https://httpbin.org" + target[len("/httpbin"):])
			if err != nil {
				statusCode = response.StatusInternalServerError
				body = response500()
			} else {
				w.WriteStatusLine(response.StatusOK)
				h.Delete("content-length")
				h.Set("Transfer-Encoding", "chunked")
				h.Replace("content-type", "text/plain")
				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")
				w.WriteHeaders(*h)
				fullBody := []byte{}
				for {
					data := make([]byte, 32)
					n, err := r.Body.Read(data)
					if err != nil {
						break
					}
					fullBody = append(fullBody, data...)
					w.WriteBody(fmt.Appendf(nil, "%x\r\n", n))
					w.WriteBody(data)
					w.WriteBody([]byte("\r\n"))
				}
				sha := sha256.Sum256(fullBody)
				trailers := headers.NewHeaders()
				trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", toStr(sha[:])))
				trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
				w.WriteBody([]byte("0\r\n"))
				w.WriteHeaders(*trailers)
				w.WriteBody([]byte("\r\n"))
				return
			}

		}
		w.WriteStatusLine(statusCode)
		w.WriteHeaders(*h)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
