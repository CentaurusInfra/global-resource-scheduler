/*
Copyright 2020 Authors of Arktos.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

//This functon creates latency log csv file
//usage example: go run main.go -readfile=klog.txt -writefile=latency.csv
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	Latency = "LatencyLog"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	readfilename := flag.String("readfile", "", "latency data file name")
	writefilename := flag.String("writefile", "./latency.txt", "latency data file name")

	flag.Parse()
	rf, err := os.Open(*readfilename)
	check(err)
	wf, err := os.Create(*writefilename)
	check(err)

	defer rf.Close()
	defer wf.Close()

	// Start reading from the file with a reader.
	fmt.Printf("readfile : %s, writefile= %s\n", *readfilename, *writefilename)
	reader := bufio.NewReader(rf)
	var line string
	for {
		//read a line
		line, err = reader.ReadString('\n')
		fmt.Printf("line : %v\n", line)
		if err != nil || err == io.EOF {
			break
		}
		//process a line
		ndx := strings.Index(line, Latency)
		if ndx >= 0 {
			_, err = wf.WriteString(line)
			if err != nil {
				check(err)
			}
		}
	}
	if err != io.EOF {
		fmt.Printf("Failed with error: %v\n", err)
	}
}
