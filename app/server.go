package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	HTTP_PROTOCOL         = "HTTP/1.1"
	STATUS_CODE_OK        = "200 OK"
	STATUS_CODE_NOT_FOUND = "404 Not Found"
)

func main() {

	var responseBuilder strings.Builder

	responseBuilder.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	responseBuilder.Write([]byte("Content-Type: text/plain\r\n\r\n"))
	responseBuilder.Write([]byte("Content-Length: \r\n\r\n"))

	fmt.Println(responseBuilder.String())

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {

		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)

		if err != nil {
			fmt.Println("Error reading buffer: ", err.Error())
		}

		request := string(buffer[:n])
		path := strings.Split(request, " ")

		if path[1] == "/" {
			fmt.Println(path)
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
			continue
		}

		if len(path[1]+"echo/") > 1 {
			content, _ := strings.CutPrefix(path[1], "/echo/")
			contentString := fmt.Sprintf("%s\r\n\r\n", content)
			contentLength := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(content))
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
			conn.Write([]byte("Content-Type: text/plain\r\n\r\n"))
			conn.Write([]byte(contentLength))
			conn.Write([]byte(contentString))
			continue
		}

		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		conn.Close()
	}
}
