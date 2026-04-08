package contenttype

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/oaswrap/gswag/test/util"
)

type UploadRequest struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

type UploadResponse struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

type Report struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Rows  int    `json:"rows"`
}

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /upload", func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(ct)
		if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
			util.WriteErrorJSON(w, http.StatusBadRequest, "expected multipart/form-data")
			return
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		var filename string
		var size int64
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				util.WriteErrorJSON(w, http.StatusBadRequest, "bad multipart")
				return
			}
			b, _ := io.ReadAll(p)
			if p.FormName() == "file" {
				filename = p.FileName()
				if filename == "" {
					filename = "upload"
				}
				size = int64(len(b))
			}
		}
		util.WriteJSON(w, http.StatusOK, UploadResponse{
			ID:       fmt.Sprintf("file-%d", size),
			Filename: filename,
			Size:     size,
		})
	})

	mux.HandleFunc("GET /report", func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "text/csv") {
			w.Header().Set("Content-Type", "text/csv")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "id,title,rows\n1,Monthly Report,42\n")
			return
		}
		util.WriteJSON(w, http.StatusOK, Report{ID: 1, Title: "Monthly Report", Rows: 42})
	})

	return mux
}
