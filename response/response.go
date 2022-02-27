package response

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sealsurlaw/ImageServer/errs"
)

type GetLinkResponse struct {
	Url       string     `json:"url"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

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

func SendError(w http.ResponseWriter, code int, msg string, er ...error) {
	res := errs.NewErrorResponse(code, msg, er...)
	SendJson(w, res)
}

func SendBadRequest(w http.ResponseWriter, missingField string) {
	msg := fmt.Sprintf("No %s provided.", missingField)
	SendError(w, http.StatusBadRequest, msg)
}

func SendMethodNotFound(w http.ResponseWriter) {
	SendError(w, http.StatusBadRequest, "Method not found.")
}

func SendInvalidAuthToken(w http.ResponseWriter) {
	SendError(w, http.StatusNotFound, "Invalid auth token.", errs.ErrNotAuthorized)
}

func SendCouldntFindImage(w http.ResponseWriter, err error) {
	SendError(w, http.StatusNotFound, "Couldn't find image.", err)
}
