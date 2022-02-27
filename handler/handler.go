package handler

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sealsurlaw/ImageServer/config"
	"github.com/sealsurlaw/ImageServer/errs"
	"github.com/sealsurlaw/ImageServer/helper"
	"github.com/sealsurlaw/ImageServer/linkstore"
)

type Handler struct {
	LinkStore              linkstore.LinkStore
	BaseUrl                string
	BasePath               string
	thumbnailQuality       int
	hashFilename           bool
	whitelistedTokens      []string
	whitelistedIpAddresses []string
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		LinkStore:              getLinkStore(cfg),
		BaseUrl:                getBaseUrl(cfg),
		BasePath:               getBasePath(cfg),
		thumbnailQuality:       cfg.ThumbnailQuality,
		hashFilename:           cfg.HashFilename,
		whitelistedTokens:      cfg.WhitelistedTokens,
		whitelistedIpAddresses: cfg.WhitelistedIpAddresses,
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

func getLinkStore(cfg *config.Config) linkstore.LinkStore {
	var linkStore linkstore.LinkStore
	linkStore = linkstore.NewMemoryLinkStore()
	if cfg.PostgresqlConfig.Enabled {
		ls, err := linkstore.NewPostgresqlLinkStore(cfg)
		if err != nil {
			fmt.Println("PostgreSQL connection failed. Falling back to memory link store.")
		} else {
			linkStore = ls
			fmt.Println("Connected to PostgreSQL link store.")
		}
	} else {
		fmt.Println("Connected to Memory link store.")
	}

	return linkStore
}

func (h *Handler) checkFileExists(filename string) (string, error) {
	// open file to make sure it exists
	filename = h.getProperFilename(filename)
	fullFilename := h.makeFullFilename(filename)
	file, err := os.Open(fullFilename)
	if err != nil {
		return "", err
	}
	err = file.Close()
	if err != nil {
		return "", err
	}

	return fullFilename, nil
}

func (h *Handler) checkOrCreateThumbnailFile(tp *ThumbnailParameters) (string, error) {
	// open file to make sure it exists
	thumbnailFilename := h.getThumbnailFilename(tp)
	thumbnailFullFilename := h.makeFullFilename(thumbnailFilename)
	file, err := os.Open(thumbnailFullFilename)
	if err != nil {
		// if not found, attempt to make it
		err = h.createThumbnail(tp)
		if err != nil {
			return "", err
		}
		file, err = os.Open(thumbnailFullFilename)
		if err != nil {
			return "", err
		}
	}
	err = file.Close()
	if err != nil {
		return "", err
	}

	return thumbnailFullFilename, nil
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
	fullFilename := h.makeFullFilename(fn)
	file, err := os.Open(fullFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	// create thumbnail
	thumbFile, err := helper.CreateThumbnail(
		file,
		tp.Resolution,
		tp.Cropped,
		h.thumbnailQuality,
	)
	if err != nil {
		return err
	}

	thumbnailFilename := h.getThumbnailFilename(tp)
	thumbnailFullFilename := h.makeFullFilename(thumbnailFilename)

	// update the deps file
	err = h.updateDepsFile(fullFilename, thumbnailFullFilename)
	if err != nil {
		return err
	}

	// write file
	fileData, err := ioutil.ReadAll(thumbFile)
	if err != nil {
		return err
	}

	err = h.createDirectories(thumbnailFilename)
	if err != nil {
		return err
	}
	err = os.WriteFile(thumbnailFullFilename, fileData, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) deleteDepFiles(fullFilename string) error {
	depFullFilename := fmt.Sprintf("%s_deps", fullFilename)
	depFile, err := os.Open(depFullFilename)
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

	fmt.Println(depFullFilename)
	err = os.Remove(depFullFilename)
	if err != nil {
		return nil
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

func (h *Handler) makeFullFilename(filename string) string {
	return fmt.Sprintf("%s/%s", h.BasePath, filename)
}

func (h *Handler) makeTokenUrl(token int64) string {
	return fmt.Sprintf("%s/link/%d", h.BaseUrl, token)
}

func (h *Handler) tryToAddLink(
	fullFileName string,
	expiresAt *time.Time,
) (int64, error) {
	maxRetries := 10
	var token int64
	tries := 0
	for true {
		tries++

		token = rand.Int63()
		err := h.LinkStore.AddLink(token, &linkstore.Link{
			FullFilename: fullFileName,
			ExpiresAt:    expiresAt,
		})

		if err == nil {
			break
		}
		if err != errs.ErrTokenAlreadyExists {
			return 0, err
		}

		// should never happen, but prevents a forever loop
		if tries == maxRetries {
			return 0, errs.ErrTooManyAttempts
		}
	}

	return token, nil
}

func (h *Handler) updateDepsFile(fullFilename, thumbnailFullFilename string) error {
	depsFilename := fmt.Sprintf("%s_deps", fullFilename)
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

	_, err = depsFile.WriteString(thumbnailFullFilename + "\n")
	if err != nil {
		return err
	}

	return nil
}
