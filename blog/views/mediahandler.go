package views

import (
	"blog/blog/config"
	"blog/blog/models"
	"blog/blog/requestfilter"
	"blog/blog/session"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"crypto/sha256"

	_ "blog/blog/filters" //import pongo  plugins

	"github.com/aws/aws-sdk-go/aws"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"

	
)

type jsonErrorMessage struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Media media
func Media(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s\n", err)
	}

	// Get List
	media, err := models.AllMediaSortedByDate()
	if err != nil {
		fmt.Println(err)
	}

	template, err := SiteTemplate("/admin/media.html")
	//template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Index",
		"media":     media,
		"user":      sess.User,
		"bodyclass": "",
		"hidetitle": true,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// ViewMedia View the media
func ViewMedia(w http.ResponseWriter, r *http.Request) {

	var media models.MediaModel

	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		fmt.Printf("Error getting url variable, id: %s", val)
	}

	// Load Media
	err = media.GetMedia(vars["id"])
	if err != nil {
		fmt.Println(err)
		return
	}

	template, err := SiteTemplate("/admin/mediaview.html")
	//template := "templates/admin/mediaview.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":           "View Media",
		"media":           media,
		"user":            sess.User,
		"bodyclass":       "",
		"fluid":           true,
		"hidetitle":       true,
		"exposureprogram": media.GetExposureProgramTranslated(),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// MediaAdd add media
func MediaAdd(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	template, err := SiteTemplate("/admin/mediaadd.html")
	//template := "templates/admin/mediaadd.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// PutMedia Upload file to server
func PutMedia(w http.ResponseWriter, r *http.Request) {

	errorMessage := "{\"status\":\"error\", \"message\": \"error: %s in %s\",\"file\":\"error\"}\n"

	vars := mux.Vars(r)
	var media models.MediaModel

	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "getting session")
		return
	}

	//err = r.ParseForm()
	err = r.ParseMultipartForm(128 << 20)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "parsing multipart form")

		fmt.Println(err)
		return
	}

	keywords := r.FormValue("keywords")
	description := r.FormValue("description")
	title := r.FormValue("title")
	category := r.FormValue("category")
	location := r.FormValue("location")

	file, handler, err := r.FormFile("file") // Retrieve the file from form data
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "opening file")
		return
	}

	defer file.Close() // Close the file when we finish

	// This is path which we want to store the file
	f, err := os.OpenFile("temp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "error storing file")
		return
	}

	defer f.Close()

	// Copy the file to the destination path
	_, err = io.Copy(f, file)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "copying file to destination path")
		return
	}

	rf, err := os.OpenFile("temp/"+handler.Filename, os.O_RDONLY, 0666)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "opening file")
		return
	}
	defer rf.Close()

	// Get exif
	err = media.ExifExtractor(rf)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "extracting exif")
		return
	}

	h := sha256.New()
	if _, err := io.Copy(h, rf); err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "creating sha265")
		return
	}
	sha256 := hex.EncodeToString(h.Sum(nil))

	media.Keywords = keywords
	media.Checksum = string(sha256)
	media.Description = description
	media.Category = category
	media.FileName = handler.Filename
	media.Title = title
	media.Location = location
	media.S3Uploaded = "false"

	err = media.InsertMedia()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "Inserting media into database")
		return
	}

	// Get s3 key
	s3KeyGenerator(&media)

	go addFileToS3("temp/"+handler.Filename, media)
	//fmt.Println("File uploaded")

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"file %s uploaded\",\"file\":\"%s\"}\n", vars["id"], handler.Filename)
	return
}

// MediaEdit Delete media from the database and s3
func MediaEdit(w http.ResponseWriter, r *http.Request) {

	// Media Object
	var media models.MediaModel

	// Form Management Variables
	formTitle := ""
	formTitleError := false
	formDescription := ""
	formDescriptionError := false
	formKeywords := ""
	formKeywordsError := false
	formCategory := ""
	formCategoryError := false
	formLocation := ""
	formLocationError := false

	//http Session
	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		fmt.Printf("Error getting url variable, id: %s", val)
	}

	// Load Media
	err = media.GetMedia(vars["id"])
	if err != nil {
		fmt.Println(err)
		return
	}

	// Test if we are a POST to capture form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		// Loading form
		media.Title = r.FormValue("title")
		media.Keywords = r.FormValue("keywords")
		media.Description = r.FormValue("description")
		media.Category = r.FormValue("category")
		media.Location = r.FormValue("location")

		// Do validation here
		validate := true
		if media.Title == "" {
			validate = false
			formTitle = "Please provide a title"
			formTitleError = true
		}

		if media.Keywords == "" {
			validate = false
			formKeywords = "Please provide keywords"
			formKeywordsError = true
		}

		if media.Description == "" {
			validate = false
			formDescription = "Please provide a description"
			formDescriptionError = true
		}

		if media.Category == "" {
			validate = false
			formCategory = "Please provide a category"
			formCategoryError = true
		}

		if media.Location == "" {
			validate = false
			formLocation = "Please provide a location"
			formLocationError = true
		}

		if validate == true {

			// Create Record
			err = media.UpdateMedia()
			if err != nil {
				fmt.Println(err)
			}

			// Redirect on success otherwise fall through the form
			// and display any errors
			http.Redirect(w, r, fmt.Sprintf("/admin/media/view/%s", vars["id"]), http.StatusSeeOther)
			return
		}
	}

	// HTTP Template
	template, err := SiteTemplate("/admin/mediaedit.html")
	//template := "templates/admin/mediaedit.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":                "Edit Media",
		"media":                media,
		"user":                 sess.User,
		"formTitle":            formTitle,
		"formTitleError":       formTitleError,
		"formKeywords":         formKeywords,
		"formKeywordsError":    formKeywordsError,
		"formDescription":      formDescription,
		"formDescriptionError": formDescriptionError,
		"formCategory":         formCategory,
		"formCategoryError":    formCategoryError,
		"formLocation":         formLocation,
		"formLocationError":    formLocationError,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// MediaDelete Delete media from the database and s3
func MediaDelete(w http.ResponseWriter, r *http.Request) {

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		fmt.Printf("Error getting url variable, id: %s", val)
	}

	var media models.MediaModel

	err := media.GetMedia(vars["id"])
	if err != nil {
		fmt.Println(err)
		return
	}

	deleteS3Object(media.S3Location)

	deleteS3Object(media.S3Thumbnail)

	deleteS3Object(media.S3LargeView)

	err = media.DeleteMedia()
	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
}

func deleteS3Object(key string) {

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("could not get configuration object %s", (err))
		return
	}

	// Create a single AWS session (we can re use this if we're uploading many files)
	s, err := awsSession.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	if err != nil {
		fmt.Printf("Error creating session to s3 with error %s\n", err)
		return
	}

	svc := s3.New(s)

	_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(cfg.S3Bucket), Key: aws.String(key)})
	if err != nil {
		fmt.Printf("Unable to delete object %q from bucket %q, %v", key, cfg.S3Bucket, err)
		return
	}

	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(cfg.S3Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		fmt.Printf("Unable to wait on delete of object %q from bucket %q, %v", key, cfg.S3Bucket, err)
		return
	}

	return
}

func jsonError(err error, w http.ResponseWriter) {

	var jerror jsonErrorMessage

	jerror.Message = err.Error()
	jerror.Status = "error"

	byteError, err := json.Marshal(jerror)
	if err != nil {
		fmt.Printf("Could not marshal error into json string with error %s\n", err)
	}

	errorString := string(byteError)

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, errorString)

	return
}

func s3KeyGenerator(media *models.MediaModel) {

	year, month, day := media.DateTime.Date()
	minute := media.DateTime.Minute()
	second := media.DateTime.Second()
	hour := media.DateTime.Hour()
	media.S3Location = fmt.Sprintf("/media/%d/%d/%d/%d/%d/%d/%s/%s", year, month, day, hour, minute, second, media.MediaID, media.FileName)
	media.S3Thumbnail = fmt.Sprintf("/media/%d/%d/%d/%d/%d/%d/%s/thumb.jpeg", year, month, day, hour, minute, second, media.MediaID)
	media.S3LargeView = fmt.Sprintf("/media/%d/%d/%d/%d/%d/%d/%s/largeview.jpeg", year, month, day, hour, minute, second, media.MediaID)
}

// AddFileToS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func addFileToS3(filepath string, media models.MediaModel) {

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("could not get configuration object %s", (err))
		return
	}

	// Generate a random 10 character string
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	// Random string is appended to view and thumbnail
	// images because if we do a multiple file upload,
	// the files will be overwritten.
	randString := string(b)

	// Create a single AWS session (we can re use this if we're uploading many files)
	s, err := awsSession.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	if err != nil {
		fmt.Printf("Error creating session to s3 with error %s\n", err)
		return
	}

	start := time.Now()

	dThumb := fmt.Sprintf("temp/thumbnail-%s.jpeg", randString)

	// Create thumbnail
	err = GetThumbnail(filepath, dThumb)
	if err != nil {
		fmt.Printf("Error creating thumbnail %s with error %s\n", dThumb, err)
	}

	file, err := os.OpenFile(dThumb, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Printf("Error uploading file %s to s3 with error %s\n", dThumb, err)
		return
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(cfg.S3Bucket),
		Key:                  aws.String(media.S3Thumbnail),
		ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})

	if err != nil {
		fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
		return
	}

	end := time.Now()

	elapsed := end.Sub(start)

	fmt.Printf("Upload of thumb %s to s3 was completed in %f seconds\n", media.S3Thumbnail, elapsed.Seconds())

	start = time.Now()

	// Create Viewer Image
	dView := fmt.Sprintf("temp/view-%s.jpeg", randString)
	err = GetViewerImage(filepath, dView)
	if err != nil {
		fmt.Printf("Error creating view image %s with error %s\n", dView, err)
	}

	file, err = os.OpenFile(dView, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
		return
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ = file.Stat()
	size = fileInfo.Size()
	buffer = make([]byte, size)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(cfg.S3Bucket),
		Key:                  aws.String(media.S3LargeView),
		ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})

	if err != nil {
		fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
		return
	}

	end = time.Now()

	elapsed = end.Sub(start)

	fmt.Printf("Upload of view image %s to s3 was completed in %f seconds\n", media.S3LargeView, elapsed.Seconds())

	// Original Image
	start = time.Now()

	file, err = os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
		return
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ = file.Stat()
	size = fileInfo.Size()
	buffer = make([]byte, size)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		//Bucket:               aws.String("vi-goblog"),
		Bucket:               aws.String(cfg.S3Bucket),
		Key:                  aws.String(media.S3Location),
		ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})

	if err != nil {
		fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
		return
	}

	media.S3Uploaded = "true"
	err = media.SetS3Uploaded()
	if err != nil {
		fmt.Printf("Failed to set media s3 status for id %s with error: %s\n", media.MediaID, err)
	}

	end = time.Now()

	elapsed = end.Sub(start)

	fmt.Printf("Upload of full size image %s to s3 was completed in %f seconds\n", media.S3Location, elapsed.Seconds())

	// Remove the images we do not need
	err = os.Remove(filepath)
	if err != nil {
		fmt.Printf("Failed to delete file %s with error: %s\n", filepath, err)
	}

	// Remove the images we do not need
	err = os.Remove(dThumb)
	if err != nil {
		fmt.Printf("Failed to delete file %s with error: %s\n", dThumb, err)
	}

	// Remove the images we do not need
	err = os.Remove(dView)
	if err != nil {
		fmt.Printf("Failed to delete file %s with error: %s\n", dView, err)
	}

	return
}



// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

// ServerImage proxy image requests through a handler to obfuscate
// the s3 bucket location
func ServerImage(wr http.ResponseWriter, req *http.Request) {
	//log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)

	slug := ""
	mediaType := ""

	// HTTP URL Parameters
	vars := mux.Vars(req)
	if val, ok := vars["slug"]; ok {
		slug = vars["slug"]
	} else {
		fmt.Printf("Error getting url variable, slug: %s\n", val)
	}

	// HTTP URL Parameters
	if val, ok := vars["type"]; ok {
		mediaType = vars["type"]
	} else {
		fmt.Printf("Error getting url variable, type: %s\n", val)
	}

	var media models.MediaModel

	err := media.GetMediaBySlug(slug)
	if err != nil {
		fmt.Printf("error getting media by slug: %s", err)
	}

	s3Path := ""

	if mediaType == "thumb" {
		s3Path = media.S3Thumbnail
	}

	if mediaType == "large" {
		s3Path = media.S3LargeView
	}

	if mediaType == "original" {
		s3Path = media.S3Location
	}

	// Generate S3 URL
	var mediaRequest http.Request
	mediaURL, err := url.Parse("https://vi-goblog.s3-us-west-2.amazonaws.com" + s3Path)
	if err != nil {
		log.Printf("ServeHTTP: %s", err)
	}

	mediaRequest.URL = mediaURL

	// Create client
	client := &http.Client{}

	//delHopHeaders(req.Header)

	if clientIP, err := requestfilter.GetIPAddress(req); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	resp, err := client.Do(&mediaRequest)
	if err != nil {

		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Printf("ServeHTTP: %s", err)
		return
	}

	defer resp.Body.Close()

	log.Println(req.RemoteAddr, " ", resp.Status)

	delHopHeaders(resp.Header)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}
