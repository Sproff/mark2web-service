package web

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/thealamu/mark2web-service/internal/pkg/db"
	"github.com/thealamu/mark2web-service/internal/pkg/log"
	"github.com/thealamu/mark2web-service/internal/pkg/mark2web"
)

func TestE2E(t *testing.T) {
	service := &mark2web.Service{
		DB: &db.FSDatabase{
			BaseDir: os.TempDir(),
		},
	}
	s := &m2wserver{
		service: service,
		logger:  log.New("DEBUG"),
		Server:  httpServer(),
	}
	s.setupRoutes()
	handler := s.Handler

	// Submit markdown
	testMarkdownBytes := []byte("# Markdown Data")

	mPartReader, contentType, err := createMultipart(testMarkdownBytes)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", mPartReader)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected code '%d', got '%d'", http.StatusCreated, rr.Code)
	}

	gotURL, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Get HTML
	id := getLastPath(string(gotURL))

	expectedHTMLBytes := []byte("<h1>Markdown Data</h1>")
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", id), nil)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected code '%d', got '%d'", http.StatusOK, rr.Code)
	}

	gotHTMLBytes, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(gotHTMLBytes, expectedHTMLBytes) {
		t.Fatalf("expected HTML content to be '%s', got '%s'", expectedHTMLBytes, gotHTMLBytes)
	}
}

func createMultipart(filedata []byte) (io.Reader, string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// TODO(thealamu): Look into CreateFormFile. Why do I have to provide a filename?
	fw, err := w.CreateFormFile("file", "file")
	if err != nil {
		return nil, "", err
	}

	fw.Write(filedata)
	w.Close()

	return &b, w.FormDataContentType(), nil
}
