package web

import "net/http"
import "io/ioutil"
import "fmt"
import "os"


func Get(data chan []byte, url string){
	response, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load list: %s\n", err)
		return
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	data <- contents
}
/*
func wget(data chan byte[], url string, args ...interface{}){
	response, err := http.Get(fmt.Sprintf(url,args))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load list: %s\n", err)
		return
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("%s", err)
		   os.Exit(1)
	}
	data <- contents
}*/