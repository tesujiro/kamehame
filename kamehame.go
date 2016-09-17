//!+

// Kamehame sends http requests and prints results at each specified URL.
package kamehame

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"sync"
	"text/template"
	"time"
)

// sema is a counting semaphore for limiting concurrency in http requesting.
var (
	sema chan struct{}
	wg   sync.WaitGroup
)

func Wave(concurrency int, tps int, buf io.Reader) {
	var start = time.Now()
	sema = make(chan struct{}, concurrency)
	input := bufio.NewScanner(buf)
	rep := regexp.MustCompile(`[\s\t\r]+`)

	for i := 0; input.Scan(); i++ {
		wg.Add(1)

		line := input.Text()
		col := rep.Split(line, -1)

		var method = col[0]
		var url = col[1]
		var tmpl = col[2]

		/* for i := 0; i < len(col); i++ {
			fmt.Printf("col[%d]=%s\n", i, col[i])
		} */

		go fetch(method, url, tmpl)

		// sleep difference expected time from elapsed time
		var wait = time.Duration(float64((i+1)/tps))*time.Second - time.Since(start)
		if wait > 0 {
			time.Sleep(wait)
		}
	}
	//fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
	wg.Wait()
}

var (
	templates = make(map[string]*template.Template)
	count     = 0
	funcMap   = template.FuncMap{
		"increment": func() int {
			count++
			return count
		},
	}
)

func getTemplate(tmpl string) *template.Template {
	t, ok := templates[tmpl]
	if !ok {
		//fmt.Printf("NEW==%s==\n", tmpl)
		//t = template.Must(template.New(path.Base(tmpl)).Funcs(funcMap).ParseFiles(tmpl))
		var err error
		t, err = template.New(path.Base(tmpl)).Funcs(funcMap).ParseFiles(tmpl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "template.ParseFIles(%s): %v\n", tmpl, err)
			os.Exit(1)
		}
		templates[tmpl] = t
	}
	return t
}

func getPostRequest(url, tmpl string) *http.Request {
	// TODO: JSON以外も対応すること
	jsonBuf := new(bytes.Buffer)
	getTemplate(tmpl).Execute(jsonBuf, nil)
	//getTemplate(tmpl).Execute(os.Stderr, nil)
	req, _ := http.NewRequest("POST", url, jsonBuf)
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")
	return req
}

func fetch(method, url, tmpl string) {
	sema <- struct{}{}        // acquire token
	defer func() { <-sema }() // release token
	defer wg.Done()
	start := time.Now()

	var req *http.Request
	switch method {
	case "POST":
		req = getPostRequest(url, tmpl)

	case "GET":
		req, _ = http.NewRequest(method, url, nil)
	}

	client := new(http.Client)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "http.Client.Do(): %v\n", err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	fmt.Printf("%d %s %6.6f\n", resp.StatusCode, url, time.Since(start).Seconds())
}

//!-
