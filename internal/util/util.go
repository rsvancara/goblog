package util

import (
	"fmt"
	"net/http"

	"goblog/internal/config"
	"goblog/internal/requestfilter"
	"goblog/internal/session"

	"github.com/disintegration/imaging"

	"github.com/h2non/bimg"
)

// CtxKey Context Key
type CtxKey string

// SiteTemplate loads the correct template directory for the site
func SiteTemplate(path string) (string, error) {

	cfg, err := config.GetConfig()
	if err != nil {
		return "", fmt.Errorf("error loading template directory %s", err)
	}
	return "sites/" + cfg.Site + path, nil

}

//GetViewerBImage get the image
func GetViewerBImage(srcFilePath string, dstFilePath string) error {

	buffer, err := bimg.Read(srcFilePath)
	if err != nil {
		fmt.Println(err)
	}

	newImage, err := bimg.NewImage(buffer).Resize(1440, 0)
	if err != nil {
		fmt.Println(err)
	}

	//size, err := bimg.NewImage(newImage).Size()
	//if size.Width == 1400 && size.Height == 1080 {
	//	fmt.Println("The image size is valid")
	//}

	bimg.Write(dstFilePath, newImage)

	return nil

}

// GetVeryLargeImage 4K image
func GetVeryLargeImage(srcFilePath string, dstFilePath string) error {
	// Open a test image.
	src, err := imaging.Open(srcFilePath)
	if err != nil {
		return err
	}

	// Resize the cropped image to width = 200px preserving the aspect ratio.
	src = imaging.Resize(src, 3840, 0, imaging.Lanczos)

	// Crop the original image to 300x300px size using the center anchor.
	//src = imaging.CropAnchor(src, 300, 300, imaging.Center)

	// Save the resulting image as JPEG.
	err = imaging.Save(src, dstFilePath)
	if err != nil {
		return err
	}

	return nil
}

// GetViewerImage 1080P image
func GetViewerImage(srcFilePath string, dstFilePath string) error {
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

// GetThumbnail Thumbnail generator
func GetThumbnail(srcFilePath string, dstFilePath string) error {
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

//GeoIPContext get the geoip object
func GeoIPContext(r *http.Request) (requestfilter.GeoIP, error) {

	var geoIP requestfilter.GeoIP

	// Attempt to extract additional information from a context
	//var geoIP requestfilter.GeoIP
	var ctxKey CtxKey
	ctxKey = "geoip"

	if result := r.Context().Value(ctxKey); result != nil {

		//fmt.Println("Found context")
		//fmt.Println(result)
		// Type Assertion....
		geoIP, ok := result.(requestfilter.GeoIP)
		if !ok {
			return geoIP, fmt.Errorf("could not perform type assertion on result to GeoIP type for ctxKey %s", ctxKey)
		}

		return geoIP, nil

	}

	return geoIP, fmt.Errorf("unable to find context for geoip %s", ctxKey)
}

//SessionContext get the session object
func SessionContext(r *http.Request) (session.Session, error) {

	var sess session.Session

	// Attempt to extract additional information from a context
	//var geoIP requestfilter.GeoIP
	var ctxKey CtxKey
	ctxKey = "session"

	if result := r.Context().Value(ctxKey); result != nil {

		//fmt.Println("Found context")
		//fmt.Println(result)
		// Type Assertion....
		sess, ok := result.(session.Session)
		if !ok {
			return sess, fmt.Errorf("could not perform type assertion on result to session.Session type for ctxKey %s", ctxKey)
		}

		return sess, nil

	}

	return sess, fmt.Errorf("unable to find context for session  %s on page %s", ctxKey, r.RequestURI)
}

// util.GetPageID get the page ID for a request
func GetPageID(r *http.Request) string {
	geoIP, err := GeoIPContext(r)
	if err != nil {
		fmt.Printf("error obtaining geoip context: %s\n", err)
	}

	return geoIP.PageID
}

// GetSession get session object for a request
func GetSession(r *http.Request) session.Session {
	sess, err := SessionContext(r)
	if err != nil {
		fmt.Printf("error obtaining session context: %s\n", err)
	}

	return sess
}
