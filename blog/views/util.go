package views

import (
	"blog/blog/config"
	"fmt"

	"github.com/disintegration/imaging"

	"github.com/h2non/bimg"
)

// SiteTemplate loads the correct template directory for the site
func SiteTemplate(path string) (string, error) {

	cfg, err := config.GetConfig()
	if err != nil {
		return "", fmt.Errorf("error loading template directory %s", err)
	}
	return cfg.Site + path, nil

}

func GetViewerBImage(srcFilePath string, dstFilePath string) error {
	buffer, err := bimg.Read(srcFilePath)
	if err != nil {
		fmt.Fprintln(err)
	}

	newImage, err := bimg.NewImage(buffer).Resize(1400, 1080)
	if err != nil {
		fmt.Fprintln(err)
	}

	size, err := bimg.NewImage(newImage).Size()
	if size.Width == 800 && size.Height == 600 {
		fmt.Println("The image size is valid")
	}

	bimg.Write(dstFilePath, newImage)

	return nil

}

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
