package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func SendJson(w http.ResponseWriter, obj interface{}) {
	j, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func SendImage(w http.ResponseWriter, file *os.File) {
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	contentType := http.DetectContentType(fileData)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(fileData)))
	w.Write(fileData)
}
