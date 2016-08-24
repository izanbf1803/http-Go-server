//By Izan Beltrán Ferreiro - izanbf.es

package main

import (
	"os"
	"io"
	"io/ioutil"
    "net"
    "net/url"
    "strings"
    "bytes"
    "strconv"
    "bufio"
)

var FILE_BLACKLIST map[string]bool

const (
	AUTHOR = "izanbf1803"
	BUF_SIZE = 16384
	BUF_SIZE_FILE = 4096
	SEND_BUF_SIZE = 65535
	SERVER_PATH = "server/"
	IS_FILE = 0
	IS_DIR = 1
	IS_404 = 2
)

var BASE_PATH string
var PORT string
var HTTPnewline string = "\r\n"
var logger *Logger
var configMap map[string]string
var configKeys []string
var SEPARATOR []string

type Request struct {
	conn net.Conn
	typeof, path, real_path, httpv, txt, contentType, dir_path string
	lines []string
	data []byte
	file *os.File
	fileMode os.FileMode
	fileSize int64
	fileType int
	exists bool
	template map[string]string
}

type Status struct {
	code int
	text string
}

func (req *Request) SendHeader(header string) {
	req.conn.Write([]byte( header+HTTPnewline ))
}

func main() {
	FILE_BLACKLIST := map[string]bool {
	    "server/log": true,
	    "server/config": true,
	}
	_ = FILE_BLACKLIST //Bypass "not used" error

	loadConfig()
	MIMEinit()
	logger = &Logger{path: SERVER_PATH+"log"}
	logger.Init()

	ln, err := net.Listen("tcp", ":"+PORT)

	if err != nil {
		logger.LogFatal("Can't listen on port %v", PORT)
	}

	logger.Log("/* Support: izanbf.es *\\	SERVER INFO [v1.0.1]", "-> SERVER STARTED!")

	for {
		conn, err := ln.Accept()

		if err != nil {
			logger.Log("ERROR", "Can't accept socket on %v", conn.RemoteAddr())
		} else {
			go handleConnection(conn)
		}
	}
}

func loadConfig() {  //load configuration from 'config' file
	SEPARATOR = make([]string, 2, 2)
	configMap = make(map[string]string)
	configKeys = make([]string, 0)

	var err error

	f, err := os.Open(SERVER_PATH+"config")
	if err != nil {
		logger.LogFatal("Can't open config file")
	}
	defer f.Close()

	var txt string
	buf := make([]byte, BUF_SIZE_FILE)
	var currentByte int64

	for {
		_, err = f.ReadAt(buf, currentByte)
		txt += string(bytes.Trim(buf, "\x00"))
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.LogFatal("Can't read config file")
		}
		currentByte += BUF_SIZE_FILE
	}

	configKeys = strings.Split(txt, "\r\n")

	for _, v := range configKeys {
		keys := strings.Split(v, ":")
		configMap[keys[0]] = strings.Trim(keys[1], " ")
	}

	BASE_PATH = configMap["DIR"][1:]
	PORT = configMap["PORT"]
	SEPARATOR = strings.Split(configMap["VAR_SYMBOL"], ",")
}

func handleConnection(conn net.Conn) {  //handle connection
	defer conn.Close()

	req := &Request{}

	buf := make([]byte, BUF_SIZE)
	size, err := conn.Read(buf)
	req.data = make([]byte, size, size)

	for i := 0; buf[i] != byte(00); i++ {
		req.data[i] = buf[i]
	}

	reqSetup(conn, req)
	defer req.file.Close()

	if err != nil{
		logger.Log("ERROR", "Can't read on socket %v, %v", conn, err)
		return
	}

	reqServe(req)
}

func reqSetup(conn net.Conn, req *Request) {  //setup Request struct variables

	defer func() {
		_ = recover()
	}()

	var err error

	req.conn = conn
	req.txt = string(req.data)
	req.lines = make([]string, 1)
	req.lines = strings.Split(req.txt, HTTPnewline)

	reqHead := strings.Split(req.lines[0], " ")

	req.typeof = reqHead[0]
	req.real_path, _ = url.QueryUnescape(reqHead[1])
	req.path = BASE_PATH

	if req.real_path != "/" {
		req.path = req.path + req.real_path
		req.real_path = req.real_path[1:]
		logger.Log("->", req.path)
		logger.Log("->", req.real_path)
	}

	req.httpv = reqHead[2]
	checkIndex(req)

	req.fileType = IS_FILE
	req.file, err = os.Open(req.path)
	if err != nil {
        req.fileType = IS_404
    }

    if req.fileType != IS_404 {
	    fi, err := req.file.Stat()
	    if err != nil {
	        logger.Log("ERROR", "Can't get info from file %v", req.path)
	    }
	    req.fileMode = fi.Mode()
	    req.fileSize = fi.Size()
	    req.contentType = getMime(&req.path)
	} else {
		req.contentType = "text/html"
	}

    if req.fileMode.IsDir() {	
    	req.fileType = IS_DIR
    	req.contentType = "text/html"
    	req.dir_path = req.real_path
    }
}

func reqServe(req *Request) {
	switch req.typeof {
		case "GET":
			GEThandle(req)
	}
}

func GEThandle(req *Request) {  //handle GET request
	if req.contentType == "ERROR" {
		return
	}
	sendHeaders(req)
	switch req.fileType {
		case IS_FILE:
			sendFile(req)
		case IS_DIR:
			listDir(req)
		case IS_404:
			send404(req)
	}
}

func checkIndex(req *Request) {  //Check if PATH is a file
	if configMap["USE_INDEX_FILES"] == "1" {
		_indexPath := (req.path)
		if req.path[len(req.path)-1:len(req.path)] != "/" {
			_indexPath += "/"
		}
		_indexPath += "index.html"
		req.exists = Exists(&_indexPath)
		if req.exists {
			req.path = _indexPath
		} else {
			req.exists = Exists(&req.path)
		}	
	}
}

func sendHeaders(req *Request) {
	status := &Status{code: 200, text: "OK"}
	
	if !req.exists {
		status = &Status{code: 404, text: "Not Found"}
	}

	req.SendHeader(req.httpv+" "+strconv.Itoa(status.code))
	req.SendHeader("Content-Type: "+req.contentType)
	req.SendHeader("Accept-Ranges: bytes")
}

func send404 (req *Request) {
	req.path = SERVER_PATH+"404.html"
	req.file, _ = os.Open(req.path) 
	fi, _ := req.file.Stat()
	req.fileSize = fi.Size()

    sendFile(req)
}

func sendFile(req *Request) {
    req.SendHeader("Content-Length: "+strconv.FormatInt(req.fileSize, 10))
    req.SendHeader("") //END OF HEADERS

	buf := make([]byte, SEND_BUF_SIZE)
	var currentByte int64 = 0

	for{
		var err error
		_, err = req.file.ReadAt(buf, currentByte)

		req.conn.Write(buf)

		if err == io.EOF{
            break
		}

        currentByte += SEND_BUF_SIZE
	}
}

func sendTxt(req *Request, res *string) {
    req.SendHeader("Content-Length: "+strconv.Itoa(len(*res)))
    req.SendHeader("") //END OF HEADERS

	req.conn.Write([]byte(*res))
}

func listDir(req *Request) {
	req.template = make(map[string]string)

	lines := ""

	req.template["title"] = "/"
	if req.dir_path != "/" {
		req.template["title"] += req.dir_path
	}
	req.template["files"] = ""
	req.template["dirs"] = ""

	files, dirs := getFilesInDir(req)
	for _, f := range files {
		req.template["files"] += "<div class='__temp_info_' type='file' data='"+f+"'></div>"
	}
	for _, d := range dirs {
		req.template["dirs"] += "<div class='__temp_info_' type='dir' data='"+d+"'></div>"
	}

	file, err := os.Open(SERVER_PATH+"list.html")
	if err != nil {
		logger.Log("ERROR", "Can't open list file")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines += scanner.Text()
	}

	for k, _ := range req.template {
		lines = strings.Replace(lines, SEPARATOR[0]+k+SEPARATOR[1], req.template[k], -1)
	}

	sendTxt(req, &lines) 
}

func getFilesInDir(req *Request) ([]string, []string) {
	files := make([]string, 0)
	dirs := make([]string, 0)
	
	fList, _ := ioutil.ReadDir("./"+BASE_PATH+"/"+req.dir_path)
    for _, f := range fList {
        fName := f.Name()
        fName = strings.Replace(fName, "/", "", -1)
        
        if !FILE_BLACKLIST[fName] {
        	if f.Mode().IsRegular() {
        		files = append(files, fName)
        	} else {
        		dirs = append(dirs, fName)
        	}
        }
    }

    return files, dirs
}

func Exists(path *string) (bool) {
    _, err := os.Stat(*path)
    if err == nil {
    	return true
    } else {
    	return false
    }
}