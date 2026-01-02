package main

import (
	"bytes"
	"fmt"
	request "httpfromtcp/internal/request"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 8)

	go func() {
		defer f.Close()
		defer close(out)

		str := ""
		for {
			data := make([]byte, 8)

			n, err := f.Read(data)
			if err != nil {
				// fmt.Println(err)
				break
			}
			data = data[:n]

			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				data = data[i+1:]
				out <- str
				str = ""
			}
			str += string(data)
		}
		if len(str) > 0 {
			out <- str
		}

	}()

	return out
}

func main() {
	// f, error := os.Open("messages.txt")
	listner, error := net.Listen("tcp", ":42069")
	if error != nil {
		fmt.Println(error)
	}

	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Fatal("Error", "error", err)
		}
		// lines := getLinesChannel(conn)
		// for line := range lines {
		// 	fmt.Println(line)
		// }

		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("Error", "error", err)
		}
		requestLine := request.RequestLine
		headers := request.Headers

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", requestLine.Method)
		fmt.Printf("- Target: %s\n", requestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", requestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		headers.ForEach(func(k, v string) {
			fmt.Printf("- %s: %s\n", k, v)
		})
		fmt.Printf("Body:\n")
		fmt.Printf("%s\n", request.Body)
	}

}
