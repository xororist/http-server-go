package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	HTTP_PROTOCOL         = "HTTP/1.1"
	STATUS_CODE_OK        = "200 OK"
	STATUS_CODE_NOT_FOUND = "404 Not Found"
	DEFAULT_PORT          = "4221"
)

var dir string

func main() {
	flagDirectory := flag.String("directory", "", "Directory to use")
	flag.Parse()
	if flagDirectory != nil && *flagDirectory != "" {
		dir = *flagDirectory
	}

	l, err := net.Listen("tcp", "0.0.0.0:"+DEFAULT_PORT)
	if err != nil {
		fmt.Println("Failed to bind to port", DEFAULT_PORT)
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func sendResponse(conn net.Conn, status string, body string, length int) {
	response := fmt.Sprintf("%s %s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
		HTTP_PROTOCOL, status, length, body)

	conn.Write([]byte(response))
}

func writeAndClose(conn net.Conn, statusCode, body string) {
	conn.Write([]byte(statusCode + "\r\n"))
	conn.Write([]byte("Content-Type: text/plain\r\n\r\n"))
	conn.Write([]byte(body + "\r\n"))
	conn.Close()
}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading buffer: ", err.Error())
	}
	request := string(buffer[:n])

	path := strings.Fields(request)

	switch {
	case len(path) == 1 || strings.TrimSpace(path[1]) == "/":
		writeAndClose(conn, "HTTP/1.1 200 OK", "Hello, World!")
	case strings.TrimSpace(path[1]) == "/user-agent":
		userAgent := getHeaderValue(request, "User-Agent")
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
		conn.Write([]byte(response))
		conn.Close()
	case strings.TrimSpace(path[0]) == "POST" && strings.HasPrefix(path[1], "/files/"):
		handleFilePostRequest(conn, path, request)
	case strings.HasPrefix(path[1], "/echo/"):
		handleEchoRequest(conn, path)
	case strings.HasPrefix(path[1], "/files/"):
		handleFileGetRequest(conn, path)
	default:
		writeAndClose(conn, "HTTP/1.1 404 Not Found", "")
	}
}

func getHeaderValue(request string, header string) string {
	lines := strings.Split(request, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, header+":") {
			return strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		}
	}
	return ""
}

func handleFilePostRequest(conn net.Conn, path []string, request string) {
	filename := strings.TrimPrefix(path[1], "/files/")
	absFilepath := filepath.Join(dir, filename)
	data := strings.TrimSpace(strings.SplitN(request, "\r\n\r\n", 2)[1])

	err := os.WriteFile(absFilepath, []byte(data), 0644)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		conn.Close()
		return
	}

	conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
	conn.Close()
}

func handleEchoRequest(conn net.Conn, path []string) {
	content := strings.TrimPrefix(path[1], "/echo/")
	if content == "" {
		sendResponse(conn, STATUS_CODE_NOT_FOUND, "", 0)
	} else {
		length := len(content)
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", length, content)
		conn.Write([]byte(response))
		conn.Close()
	}
}

func handleFileGetRequest(conn net.Conn, path []string) {
	filename := strings.TrimPrefix(path[1], "/files/")
	absFilepath := filepath.Join(dir, filename)
	fileInfo, err := os.Stat(absFilepath)
	if err != nil || fileInfo.IsDir() {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		conn.Close()
		return
	}
	bs, err := os.ReadFile(absFilepath)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		conn.Close()
		return
	}
	conn.Write(buildBytesStreamResp(bs))
	conn.Close()
}

func buildBytesStreamResp(bs []byte) []byte {
	b := bytes.Buffer{}
	b.WriteString("HTTP/1.1 200 OK\r\n")
	b.WriteString("Content-Type: application/octet-stream\r\n")
	b.WriteString("Content-Length: " + strconv.Itoa(len(bs)) + "\r\n")
	b.WriteString("\r\n")
	b.Write(bs)
	return b.Bytes()
}
