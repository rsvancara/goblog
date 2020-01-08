package views

import (
	"blog/blog/models"
	"blog/blog/session"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"crypto/sha256"

	"github.com/aws/aws-sdk-go/aws"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dsoprea/go-exif"
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
	err := sess.Session(r)
	if err != nil {
		fmt.Printf("Session not available %s\n", err)
	}

	template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User.Username})
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

	errorMessage := "{\"status\":\"error\", \"message\": \"error: %s\",\"file\":\"error\"}\n"

	vars := mux.Vars(r)
	var media models.MediaModel

	var sess session.Session
	err := sess.Session(r)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err)
		return
	}

	//err = r.ParseForm()
	err = r.ParseMultipartForm(128 << 20)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		//fmt.Fprintf(w, errorMessage, err)

		fmt.Println(err)
		return
	}

	keywords := r.FormValue("keywords")
	description := r.FormValue("description")

	fmt.Printf("keywords: %s\n", keywords)
	//fmt.Println(r.Form)

	file, handler, err := r.FormFile("file") // Retrieve the file from form data
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err)
		return
	}

	defer file.Close() // Close the file when we finish

	// This is path which we want to store the file
	f, err := os.OpenFile("temp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err)
		return
	}

	defer f.Close()

	// Copy the file to the destination path
	_, err = io.Copy(f, file)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err)
		return
	}

	rf, err := os.OpenFile("temp/"+handler.Filename, os.O_RDONLY, 0666)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err)
		return
	}
	defer rf.Close()

	// Get exif
	err = exifExtractor(rf, &media)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err)
		return
	}

	h := sha256.New()
	if _, err := io.Copy(h, rf); err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err)
		return
	}
	sha256 := hex.EncodeToString(h.Sum(nil))

	media.Keywords = keywords
	media.Checksum = string(sha256)
	media.Description = description
	media.FileName = handler.Filename
	media.S3Uploaded = "false"

	err = media.InsertMedia()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err)
		return
	}

	// Get s3 key
	s3KeyGenerator(&media)

	go addFileToS3(media.S3Location, "temp/"+handler.Filename, media)
	//fmt.Println("File uploaded")

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"file %s uploaded\",\"file\":\"%s\"}\n", vars["id"], handler.Filename)
	fmt.Printf("{\"status\":\"success\", \"message\": \"file %s uploaded\",\"file\":\"%s\"}\n", vars["id"], handler.Filename)
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

// Extract EXIF Information from image
func exifExtractor(f *os.File, media *models.MediaModel) error {

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	//log.PanicIf(err)

	exifData, err := exif.SearchAndExtractExif(data)
	if err != nil {
		if err == exif.ErrNoExif {
			return err
		}
		return err
	}

	// parse exif information
	im := exif.NewIfdMappingWithStandard()
	ti := exif.NewTagIndex()

	visitor := func(fqIfdPath string, ifdIndex int, tagId uint16, tagType exif.TagType, valueContext exif.ValueContext) (err error) {
		ifdPath, err := im.StripPathPhraseIndices(fqIfdPath)
		if err != nil {
			return err
		}

		it, err := ti.Get(ifdPath, tagId)
		if err != nil {
			return err
		}

		valueString := ""
		if tagType.Type() == exif.TypeUndefined {
			value, err := exif.UndefinedValue(ifdPath, tagId, valueContext, tagType.ByteOrder())
			if err == exif.ErrUnhandledUnknownTypedTag {
				valueString = "!UNDEFINED!"
			} else {
				return err
			}
			valueString = fmt.Sprintf("%v", value)
		} else {
			valueString, err = tagType.ResolveAsString(valueContext, true)
			if err != nil {
				return err
			}
		}

		// Obtain the various components and add exif information
		if it.Name == "Make" {
			media.Make = valueString
		}

		if it.Name == "Model" {
			media.Model = valueString
		}

		if it.Name == "Software" {
			media.Software = valueString
		}

		if it.Name == "DateTime" {
			layOut := "2006:01:02 15:04:05 MST"
			//"2019:12:23 18:46:27"
			timeStamp, _ := time.Parse(layOut, fmt.Sprintf("%s PST", valueString))
			media.DateTime = timeStamp
		}

		if it.Name == "Artist" {
			media.Artist = valueString
		}

		if it.Name == "ExposureTime" {
			media.ExposureTime = valueString
		}

		if it.Name == "FNumber" {
			media.FNumber = valueString
		}

		if it.Name == "ISOSpeedRatings" {
			media.ISOSpeedRatings = valueString
		}

		if it.Name == "LightSource" {
			media.LightSource = valueString
		}

		if it.Name == "FocalLength" {
			media.FocalLength = valueString
		}

		if it.Name == "PixelXDimension" {
			media.PixelXDimension = valueString
		}

		if it.Name == "PixelYDimension" {
			media.PixelYDimension = valueString
		}

		if it.Name == "FocalLengthIn35mmFilm" {
			media.FocalLengthIn35mmFilm = valueString
		}

		if it.Name == "LensModel" {
			media.LensModel = valueString
		}

		//fmt.Printf("FQ-IFD-PATH=[%s] ID=(0x%04x) NAME=[%s] COUNT=(%d) TYPE=[%s] VALUE=[%s]\n", fqIfdPath, tagId, it.Name, valueContext.UnitCount, tagType.Name(), valueString)
		return nil
	}

	_, err = exif.Visit(exif.IfdStandard, im, ti, exifData, visitor)

	if err != nil {
		return err
	}
	return nil
}

func s3KeyGenerator(media *models.MediaModel) {

	year, month, day := media.DateTime.Date()
	minute := media.DateTime.Minute()
	second := media.DateTime.Second()
	hour := media.DateTime.Hour()
	media.S3Location = fmt.Sprintf("/%d/%d/%d/%d/%d/%d/%s/%s", year, month, day, hour, minute, second, media.MediaID, media.FileName)
}

// AddFileToS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func addFileToS3(key string, filepath string, media models.MediaModel) {

	fmt.Printf("uploading file %s to path %s for media %s\n", filepath, key, media.MediaID)

	start := time.Now()

	file, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
		return
	}
	defer file.Close()

	// Create a single AWS session (we can re use this if we're uploading many files)
	s, err := awsSession.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	if err != nil {
		fmt.Printf("Error creating session to s3 with error %s\n", err)
		return
	}

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String("vi-goblog"),
		Key:                  aws.String(key),
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

	err = media.SetS3Uploaded("true", key)
	if err != nil {
		fmt.Printf("Failed to set media s3 status for id %s with error: %s\n", media.MediaID, err)
	}

	end := time.Now()

	elapsed := end.Sub(start)

	fmt.Printf("Upload of media %s to s3 was completed in %f seconds\n", media.MediaID, elapsed.Seconds())

	return
}
