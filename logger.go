//By Izan Beltrán Ferreiro - izanbf.es

package main

import (
	"fmt"
    "time"
    "os"
    "strings"
)

type Logger struct{
	path string
}

func (logger *Logger) Write(data []byte) {  // Write to file
	var f *os.File
	var err error

	f, err = os.OpenFile(logger.path, os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		f, _ = os.Create(logger.path)
	}

	defer f.Close()

	_, err = f.Write(data)

	if err != nil {
		fmt.Printf("\n\t!!!!ERROR: CAN'T WRITE FILE\n")
		os.Exit(1)
	}
}

func (logger Logger) Log(_type string, format string, args ...interface{}) {  //Log to file and Print
	err_log := fmt.Sprintf("\n%s - %v: %v\n", _type, time.Now(), fmt.Sprintf(format, args...))
	fmt.Print(err_log)
	logger.Write([]byte(err_log))
}

func (logger Logger) LogFatal(format string, args ...interface{}) {   //Log and exit
	fmt.Printf("->%v\n", args)
	logger.Log("!FATAL ERROR", format, args...)
	os.Exit(1)
}

func (logger Logger) Print(_type string, message string, args ...interface{}) {  //Print without Log to File
	fmt.Printf("\n%s - %v: %v\n", _type, time.Now(), fmt.Sprintf(message, args))
}

func (logger Logger) Init() {   //Init variables
	sTime := fmt.Sprintf("%v", time.Now())
	divisorLen := len(sTime)
	if divisorLen % 2 != 0 {
		divisorLen++
	}

	d1 := strings.Repeat("-", divisorLen*2)
	d2 := strings.Repeat("-", divisorLen/2)

	divisor := fmt.Sprintf("%s\n\n%s%v%s\n", d1, d2, sTime, d2)

	logger.Write([]byte(divisor))
}