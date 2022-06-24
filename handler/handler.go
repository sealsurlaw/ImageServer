package handler

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/sealsurlaw/gouvre/config"
	"github.com/sealsurlaw/gouvre/errs"
	"github.com/sealsurlaw/gouvre/helper"
	"github.com/sealsurlaw/gouvre/token"
)

type Handler struct {
	BaseUrl                string
	BasePath               string
	tokenizer              *token.Tokenizer
	thumbnailQuality       int
	hashFilename           bool
	pinToIpfs              bool
	whitelistedTokens      []string
	whitelistedIpAddresses []string
	singleUseUploadTokens  map[string]bool
}

func NewHandler(cfg *config.Config) *Handler {
	tokenizer, err := token.NewTokenizer(cfg.EncryptionSecret)
	if err != nil {
		log.Fatal(err)
	}
	return &Handler{
		BaseUrl:                getBaseUrl(cfg),
		BasePath:               getBasePath(cfg),
		tokenizer:              tokenizer,
		thumbnailQuality:       cfg.ThumbnailQuality,
		hashFilename:           cfg.HashFilename,
		pinToIpfs:              cfg.PinToIpfs,
		whitelistedTokens:      cfg.WhitelistedTokens,
		whitelistedIpAddresses: cfg.WhitelistedIpAddresses,
		singleUseUploadTokens:  make(map[string]bool),
	}
}

func getBaseUrl(cfg *config.Config) string {
	baseUrl := cfg.BaseUrl
	if baseUrl == "" {
		baseUrl = fmt.Sprintf("http://localhost:%s", cfg.Port)
	}

	return baseUrl
}

func getBasePath(cfg *config.Config) string {
	basePath := cfg.BasePath
	if basePath == "" {
		bp, err := os.MkdirTemp("/tmp", "imageserver.*")
		if err != nil {
			log.Fatal("Couldn't create tmp directory")
		}
		basePath = bp
	}

	return basePath
}

func (h *Handler) checkFileExists(filename string, encryptionSecret string) (string, error) {
	// open file to make sure it exists
	filename = h.getProperFilename(filename)
	fullFilePath := h.makeFullFilePath(filename)
	fileData, err := helper.OpenFile(fullFilePath)
	if err != nil {
		return "", err
	}

	if h.tryDecryptFile(&fileData, encryptionSecret) != nil {
		return "", errs.ErrBadEncryptionSecret
	}

	return filename, nil
}

func (h *Handler) checkOrCreateThumbnailFile(tp *ThumbnailParameters) (string, error) {
	// open file to make sure it exists
	thumbnailFilename := h.getThumbnailFilename(tp)
	thumbnailfullFilePath := h.makeFullFilePath(thumbnailFilename)
	fileData, err := helper.OpenFile(thumbnailfullFilePath)
	if err != nil {
		// if not found, attempt to make it
		err = h.createThumbnail(tp)
		if err != nil {
			return "", err
		}

		fileData, err = helper.OpenFile(thumbnailfullFilePath)
		if err != nil {
			return "", err
		}
	}

	if h.tryDecryptFile(&fileData, tp.EncryptionSecret) != nil {
		return "", errs.ErrBadEncryptionSecret
	}

	return thumbnailFilename, nil
}

func (h *Handler) createDirectories(filename string) error {
	// check if directories need to be created
	if strings.Contains(filename, "/") {
		filenameSplit := strings.Split(filename, "/")
		path := h.BasePath
		for _, dir := range filenameSplit[:len(filenameSplit)-1] {
			path += "/" + dir
			_, err := os.ReadDir(path)
			if err == nil {
				continue
			}
			err = os.Mkdir(path, 0700)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *Handler) createThumbnail(tp *ThumbnailParameters) error {
	// open file
	fn := h.getProperFilename(tp.Filename)
	fullFilePath := h.makeFullFilePath(fn)
	fileData, err := helper.OpenFile(fullFilePath)
	if err != nil {
		return err
	}

	if h.tryDecryptFile(&fileData, tp.EncryptionSecret) != nil {
		return errs.ErrBadEncryptionSecret
	}

	contentType := http.DetectContentType(fileData)
	if contentType == "image/gif" {
		return errs.ErrGif
	}

	// create thumbnail
	thumbData, err := helper.CreateThumbnail(
		fileData,
		tp.Resolution,
		tp.Cropped,
		h.thumbnailQuality,
	)
	if err != nil {
		return err
	}

	thumbnailFilename := h.getThumbnailFilename(tp)
	thumbnailfullFilePath := h.makeFullFilePath(thumbnailFilename)

	// update the deps file
	err = h.updateDepsFile(fullFilePath, thumbnailfullFilePath)
	if err != nil {
		return err
	}

	err = h.createDirectories(thumbnailFilename)
	if err != nil {
		return err
	}

	if h.tryEncryptFile(&thumbData, tp.EncryptionSecret) != nil {
		return errs.ErrBadEncryptionSecret
	}

	err = os.WriteFile(thumbnailfullFilePath, thumbData, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) deleteDepFiles(fullFilePath string) error {
	depfullFilePath := fmt.Sprintf("%s_deps", fullFilePath)
	depFile, err := os.Open(depfullFilePath)
	if err != nil {
		return nil
	}

	scanner := bufio.NewScanner(depFile)
	for scanner.Scan() {
		err = os.Remove(scanner.Text())
		if err != nil {
			return nil
		}
	}
	err = depFile.Close()
	if err != nil {
		return err
	}

	err = os.Remove(depfullFilePath)
	if err != nil {
		return nil
	}

	return nil
}

func (h *Handler) tryDecryptFile(fileData *[]byte, encryptionSecret string) error {
	if encryptionSecret == "" {
		return nil
	}

	var err error
	fileDataCopy := []byte{}
	if encryptionSecret != "" {
		fileDataCopy, err = helper.Decrypt(*fileData, encryptionSecret)
		if err != nil {
			return errs.ErrBadEncryptionSecret
		}
	}

	*fileData = fileDataCopy

	return nil
}

func (h *Handler) tryEncryptFile(fileData *[]byte, encryptionSecret string) error {
	if encryptionSecret == "" {
		return nil
	}

	var err error
	if encryptionSecret != "" {
		*fileData, err = helper.Encrypt(*fileData, encryptionSecret)
		if err != nil {
			return errs.ErrBadEncryptionSecret
		}
	}
	return nil
}

func (h *Handler) getProperFilename(filename string) string {
	if h.hashFilename {
		filename = helper.CalculateHash(filename)
		filename = fmt.Sprintf("%s/%s/%s", string(filename[0]), string(filename[1]), filename)
	}

	return filename
}

func (h *Handler) getThumbnailFilename(tp *ThumbnailParameters) string {
	filename := tp.Filename
	var thumbnailFilename string
	if h.hashFilename {
		filename += strconv.Itoa(tp.Resolution)
		if tp.Cropped {
			filename += "crop"
		}
		thumbnailFilename = h.getProperFilename(filename)
	} else {
		thumbnailFilename = fmt.Sprintf("%s_%d", filename, tp.Resolution)
		if tp.Cropped {
			thumbnailFilename += "_crop"
		}
	}

	return thumbnailFilename
}

func (h *Handler) hasWhitelistedIpAddress(r *http.Request) bool {
	ip := helper.GetIpAddress(r)
	for _, ipAddr := range h.whitelistedIpAddresses {
		if ipAddr == "*" {
			return true
		}

		ip1 := net.ParseIP(ip)
		ip2 := net.ParseIP(ipAddr)
		if ip1.Equal(ip2) {
			return true
		}
	}

	return false
}

func (h *Handler) hasWhitelistedToken(r *http.Request) bool {
	// If empty, block
	if len(h.whitelistedTokens) == 0 {
		return false
	}

	// allow all for '*'
	if len(h.whitelistedTokens) == 1 && h.whitelistedTokens[0] == "*" {
		return true
	}

	authentication := r.Header.Get("Authorization")
	if authentication == "" {
		return false
	}

	authSplit := strings.Split(authentication, " ")
	if len(authSplit) != 2 {
		return false
	}
	if authSplit[0] != "Bearer" {
		return false
	}

	auth := false
	for _, token := range h.whitelistedTokens {
		if token == authSplit[1] {
			auth = true
			break
		}
	}

	return auth
}

func (h *Handler) makeFullFilePath(filename string) string {
	return fmt.Sprintf("%s/%s", h.BasePath, filename)
}

func (h *Handler) makeTokenUrl(token string) string {
	return fmt.Sprintf("%s/links/%s", h.BaseUrl, token)
}

func (h *Handler) makeUploadTokenUrl(token string) string {
	return fmt.Sprintf("%s/uploads/%s", h.BaseUrl, token)
}

func (h *Handler) updateDepsFile(fullFilePath, thumbnailfullFilePath string) error {
	depsFilename := fmt.Sprintf("%s_deps", fullFilePath)
	depsFile, err := os.OpenFile(depsFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		err = os.WriteFile(depsFilename, []byte{}, 0600)
		if err != nil {
			return err
		}
		depsFile, err = os.OpenFile(depsFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
	}
	defer depsFile.Close()

	_, err = depsFile.WriteString(thumbnailfullFilePath + "\n")
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) writeFile(fileData []byte, filename string, encryptionSecret string) error {
	filename = h.getProperFilename(filename)
	err := h.createDirectories(filename)
	if err != nil {
		return err
	}

	contentType := http.DetectContentType(fileData)
	if !helper.IsSupportedContentType(contentType) {
		msg := fmt.Sprintf("Content type %s not supported.", contentType)
		return fmt.Errorf(msg)
	}

	fullFilePath := h.makeFullFilePath(filename)

	// delete files from dep file including itself
	err = h.deleteDepFiles(fullFilePath)
	if err != nil {
		return err
	}

	if h.tryEncryptFile(&fileData, encryptionSecret) != nil {
		return err
	}

	// write file
	err = os.WriteFile(fullFilePath, fileData, 0600)
	if err != nil {
		return err
	}

	return nil
}
