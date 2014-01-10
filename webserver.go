package main

import (
    "encoding/json"
    "fmt"
    "github.com/stamp/go-json-rest"
    "io/ioutil"
    "net/http"
)

// Webserver that serves static files
var chttp = http.NewServeMux()
var handler = rest.ResourceHandler{}

// The webserver
type Response map[string]interface{} /*{{{*/
func (r Response) String() (s string) {
    b, err := json.Marshal(r)
    if err != nil {
        s = ""
        return
    }
    s = string(b)
    return
}                 /*}}}*/
func webStart() { /*{{{*/
    chttp.Handle("/", http.FileServer(http.Dir("./site/")))
    http.HandleFunc("/", webHandler)

    handler.SetRoutes(
        rest.Route{"GET", "/api/users/:id", GetUser},
    )

    go http.ListenAndServe(":8080", nil)
    fmt.Println()
}                                                         /*}}}*/
func webHandler(w http.ResponseWriter, r *http.Request) { /*{{{*/
    // Select what to serve
    if len(r.URL.Path) > 3 && r.URL.Path[1:4] == "api" {
        handler.ServeHTTP(w, r)
        return
    }

    switch r.URL.Path[1:] {
    case "": // The main page (index.htm)
        // Handle simulate tag scans
        if r.FormValue("tag") != "" {
            //
        }

        // Serve the index.htm
        body, _ := ioutil.ReadFile("./site/index.htm")
        fmt.Fprintf(w, string(body))
    default: // Serve all files in the web folder
        chttp.ServeHTTP(w, r)
    }
}   /*}}}*/

func GetUser(w *rest.ResponseWriter, req *rest.Request) {

}
