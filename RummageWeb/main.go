package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/igpkb/navsearch/Rummage"
	/*"strings"*/)

func main() {
	Rummage.Shell, Rummage.Client = Rummage.ConnectServer("https://mainnet.infura.io/v3/5c04e573d61b4e5a8fc0f3312becfdbc", "k51qzi5uqu5dknzcklexyqi8kfhfy8oh8ai7tinrlaa0m6sm6hk1mwangfnafn") //, "127.0.0.1",3333)
	RunCrons("all", "corpus", "crawlCIDs.txt")
	r := newRouter()
	err := http.ListenAndServe(":80", r)
	if err != nil {
		panic(err.Error())
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/hello", handler).Methods("GET")
	staticFileDirectory := http.Dir("./web/")
	staticFileHandler := http.StripPrefix("/web/", http.FileServer(staticFileDirectory))
	r.PathPrefix("/web/").Handler(staticFileHandler).Methods("GET")
	r.HandleFunc("/search", getSearchResultHandler).Methods("GET")
	r.HandleFunc("/getcrawlTargets", getCrawlTargetHandler).Methods("GET")
	r.HandleFunc("/crawlTarget", createCrawlTargetHandler).Methods("GET")
	return r
}
