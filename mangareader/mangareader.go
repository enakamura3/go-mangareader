package mangareader

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

var basePath = "/tmp"
var urlraw = "https://www.mangareader.net"
var lastPage = 999

/*Exec download on mangareader*/
func Exec(serie string, initChapter int, endChapter int) {
	for c := initChapter; c <= endChapter; c++ {
		var pages [][]byte
		lastPage := getLastChapterPage(urlraw, serie, c)
		if lastPage == 0 {
			break
		}
		for p := 1; p <= lastPage; p++ {
			page := getImageFile(getImageURL(urlraw, serie, c, p))
			pages = append(pages, page)
		}
		saveImageFile(pages, serie, c, basePath)
		savePdfFile(pages, serie, c, basePath)
	}
}

func getLastChapterPage(url string, serie string, chapter int) int {

	pageURL := fmt.Sprintf("%s/%s/%d/%d", url, serie, chapter, lastPage)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Get(pageURL)
	if err != nil {
		log.Printf("error getting last page url from serie '%v' chapter '%v'", serie, pageURL)

	}
	defer res.Body.Close()
	location := res.Header.Get("location")
	locations := strings.Split(location, "/")
	lastPage, err := strconv.Atoi(locations[len(locations)-1])
	if err != nil {
		lastPage = 0
	}
	return lastPage
}

func getImageURL(url string, serie string, chapter int, page int) string {
	pageURL := fmt.Sprintf("%s/%s/%d/%d", url, serie, chapter, page)
	fmt.Printf("page url: %v\n", pageURL)
	res, err := http.Get(pageURL)
	if err != nil {
		log.Printf("error getting image url from page: %v", pageURL)
	}
	defer res.Body.Close()

	/*retry*/
	if res.StatusCode == 524 {
		return getImageURL(url, serie, chapter, page)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("error reading body content from: %v", pageURL)
	}

	bodys := string(body)

	re := regexp.MustCompile("https+[^,]+\\.jpg")
	imageURL := re.FindString(bodys)
	fmt.Printf("image url: %v\n", imageURL)
	return imageURL
}

func getImageFile(url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		log.Printf("error getting image file from: %v", url)
	}
	defer res.Body.Close()

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("error reading body content from: %v", url)
	}
	return image
}

func saveImageFile(pages [][]byte, serie string, chapter int, basePath string) {
	path := filepath.Join(basePath, serie, fmt.Sprint(chapter))
	err := os.MkdirAll(path, 0777)
	if err != nil {
		log.Printf("error creating directory: %v", path)
	}
	for k, v := range pages {
		filePath := fmt.Sprintf(filepath.Join(path, fmt.Sprintf("%d.jpg", k+1)))
		err = ioutil.WriteFile(filePath, v, 0666)
		if err != nil {
			log.Printf("error writing file: %v", filePath)
		}
		log.Printf("image download at: %s\n", filePath)
	}
}

func savePdfFile(pages [][]byte, serie string, chapter int, basePath string) {
	path := filepath.Join(basePath, serie)
	err := os.MkdirAll(path, 0777)
	if err != nil {
		log.Printf("error creating directory: %v", path)
	}
	pdf := gofpdf.New("P", "mm", "A4", "")
	for k, v := range pages {
		rdr := bytes.NewReader(v)
		pdf.AddPage()
		imageName := fmt.Sprint(k + 1)
		opt := gofpdf.ImageOptions{ImageType: "jpg", ReadDpi: true}
		_ = pdf.RegisterImageOptionsReader(imageName, opt, rdr)
		err := pdf.Error()
		if err != nil {
			pdf.ClearError()
			log.Printf("error parsing image '%v'", imageName)
			pdf.SetFont("Arial", "B", 16)
			pdf.Cell(40, 10, fmt.Sprintf("error parsing image '%v'", imageName))
		} else {
			pdf.ImageOptions(imageName, 0, 0, 0, 0, false, opt, 0, "")
		}
	}
	filePath := filepath.Join(path, fmt.Sprintf("%v-%v.pdf", serie, chapter))
	err = pdf.OutputFileAndClose(filePath)
	if err != nil {
		log.Printf("error creating pdf file for serie '%v' chapter '%v'", serie, chapter)

	}
}
