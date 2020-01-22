package views

import (
	"blog/blog/models"
	"blog/blog/session"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"crypto/sha256"

	_ "blog/blog/filters" //import pongo  plugins

	"github.com/aws/aws-sdk-go/aws"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"

	"github.com/disintegration/imaging"
)

type jsonErrorMessage struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Media media
func Media(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r)
	if err != nil {
		fmt.Printf("Session not available %s\n", err)
	}

	// Get List
	media, err := models.AllMediaSortedByDate()
	if err != nil {
		fmt.Println(err)
	}

	template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Index",
		"media":     media,
		"user":      sess.User.Username,
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
	err := sess.Session(r)
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

	template := "templates/admin/mediaview.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":           "View Media",
		"media":           media,
		"user":            sess.User.Username,
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
	err := sess.Session(r)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	template := "templates/admin/mediaadd.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User.Username})
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
	err := sess.Session(r)
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

	//http Session
	var sess session.Session
	err := sess.Session(r)
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
	template := "templates/admin/mediaedit.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":                "Edit Media",
		"media":                media,
		"user":                 sess.User.Username,
		"formTitle":            formTitle,
		"formTitleError":       formTitleError,
		"formKeywords":         formKeywords,
		"formKeywordsError":    formKeywordsError,
		"formDescription":      formDescription,
		"formDescriptionError": formDescriptionError,
		"formCategory":         formCategory,
		"formCategoryError":    formCategoryError,
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

	deleteS3Object("vi-goblog", media.S3Location)

	deleteS3Object("vi-goblog", media.S3Thumbnail)

	deleteS3Object("vi-goblog", media.S3LargeView)

	err = media.DeleteMedia()
	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, "/admin/media", http.StatusSeeOther)

}

func deleteS3Object(bucket string, key string) {

	// Create a single AWS session (we can re use this if we're uploading many files)
	s, err := awsSession.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	if err != nil {
		fmt.Printf("Error creating session to s3 with error %s\n", err)
		return
	}

	svc := s3.New(s)

	_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil {
		fmt.Printf("Unable to delete object %q from bucket %q, %v", key, bucket, err)
		return
	}

	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		fmt.Printf("Unable to wait on delete of object %q from bucket %q, %v", key, bucket, err)
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

	// Create a single AWS session (we can re use this if we're uploading many files)
	s, err := awsSession.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	if err != nil {
		fmt.Printf("Error creating session to s3 with error %s\n", err)
		return
	}

	start := time.Now()

	// Create thumbnail
	err = getThumbnail(filepath, "temp/thumbnail.jpeg")
	if err != nil {
		fmt.Printf("Error creating thumbnail %s with error %s\n", filepath, err)
	}

	file, err := os.OpenFile("temp/thumbnail.jpeg", os.O_RDONLY, 0666)
	if err != nil {
		fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
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
		Bucket:               aws.String("vi-goblog"),
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
	err = getViewerImage(filepath, "temp/view.jpeg")
	if err != nil {
		fmt.Printf("Error creating view image %s with error %s\n", filepath, err)
	}

	file, err = os.OpenFile("temp/view.jpeg", os.O_RDONLY, 0666)
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
		Bucket:               aws.String("vi-goblog"),
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

	// Full Size
	fmt.Printf("uploading full size image file %s to path %s for media %s\n", filepath, media.S3Location, media.MediaID)

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
		Bucket:               aws.String("vi-goblog"),
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

	return
}

func getViewerImage(srcFilePath string, dstFilePath string) error {
	// Open a test image.
	src, err := imaging.Open(srcFilePath)
	if err != nil {
		return err
	}

	// Resize the cropped image to width = 200px preserving the aspect ratio.
	src = imaging.Resize(src, 1440, 0, imaging.Lanczos)

	// Crop the original image to 300x300px size using the center anchor.
	//src = imaging.CropAnchor(src, 300, 300, imaging.Center)

	// Save the resulting image as JPEG.
	err = imaging.Save(src, dstFilePath)
	if err != nil {
		return err
	}

	return nil
}

func getThumbnail(srcFilePath string, dstFilePath string) error {
	// Open a test image.
	src, err := imaging.Open(srcFilePath)
	if err != nil {
		return err
	}

	// Resize the cropped image to width = 200px preserving the aspect ratio.
	src = imaging.Resize(src, 300, 0, imaging.Lanczos)

	// Crop the original image to 300x300px size using the center anchor.
	src = imaging.CropAnchor(src, 300, 300, imaging.Center)

	// Save the resulting image as JPEG.
	err = imaging.Save(src, dstFilePath)
	if err != nil {
		return err
	}

	return nil
}
