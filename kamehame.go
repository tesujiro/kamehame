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
	sema = make(chan struct{}, concurrency) // semaphore for limiting concurrency
	rep := regexp.MustCompile(`[\s\t\r]+`)

	input := bufio.NewScanner(buf)
	for i := 0; input.Scan(); i++ {
		wg.Add(1)

		line := input.Text()
		col := rep.Split(line, -1)

		if len(col) < 2 || len(col) > 3 {
			fmt.Fprintf(os.Stderr, "File format error(%s)\n", line)
			os.Exit(1)
		}
		var method, url string = col[0], col[1]
		var tmpl string

		if len(col) == 3 {
			tmpl = col[2]
		}

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

func getRequest(method, url, tmpl string) *http.Request {
	var buf *bytes.Buffer
	if tmpl != "" {
		buf = new(bytes.Buffer)
		getTemplate(tmpl).Execute(buf, nil)
		//getTemplate(tmpl).Execute(os.Stderr, nil)
	}
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "http.NewRequest: %v\n", err)
		os.Exit(1)
	}
	switch method {
	case "POST":
		//req.Header.Set("X-Custom-Header", "myvalue")
		// TODO: JSON以外も対応すること
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func fetch(method, url, tmpl string) {
	sema <- struct{}{}        // acquire token
	defer func() { <-sema }() // release token
	defer wg.Done()
	start := time.Now()

	req := getRequest(method, url, tmpl)

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "http.Client.Do(): %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("%d %s %6.6f\n", resp.StatusCode, url, time.Since(start).Seconds())
}

//!-
