//Package simplestorageservice simple storage service functions S3
package simplestorageservice

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"goblog/internal/config"
	mediadao "goblog/internal/dao/media"
	"goblog/internal/models"
	"goblog/internal/util"

	"github.com/aws/aws-sdk-go/aws"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

// DeleteS3Object Deletes and object from s3
func DeleteS3Object(key string) {

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

// S3KeyGenerator Generates an S3Key
func S3KeyGenerator(media *models.MediaModel) {

	year, month, day := media.DateTime.Date()
	minute := media.DateTime.Minute()
	second := media.DateTime.Second()
	hour := media.DateTime.Hour()
	media.S3Location = fmt.Sprintf("/media/%d/%d/%d/%d/%d/%d/%s/%s", year, month, day, hour, minute, second, media.MediaID, media.FileName)
	media.S3Thumbnail = fmt.Sprintf("/media/%d/%d/%d/%d/%d/%d/%s/thumb.jpeg", year, month, day, hour, minute, second, media.MediaID)
	media.S3LargeView = fmt.Sprintf("/media/%d/%d/%d/%d/%d/%d/%s/largeview.jpeg", year, month, day, hour, minute, second, media.MediaID)
	media.S3VeryLarge = fmt.Sprintf("/media/%d/%d/%d/%d/%d/%d/%s/verylargeview.jpeg", year, month, day, hour, minute, second, media.MediaID)
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// AddFileToS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func AddFileToS3(filepath string, media *models.MediaModel, mongoclient *mongo.Client, cfg *config.AppConfig) {

	var mediaDAO mediadao.MediaDAO

	err := mediaDAO.Initialize(mongoclient, cfg)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
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
		log.Error().Err(err).Msgf("Error creating thumbnail, could not create aws session")
		return
	}

	// media File stores temporary file locations
	type mediaFile struct {
		imagetype   string
		path        string
		destination string
	}

	var mfThumb mediaFile
	mfThumb.imagetype = "thumb"
	mfThumb.path = fmt.Sprintf("temp/thumbnail-%s.jpeg", randString)
	mfThumb.destination = media.S3Thumbnail

	log.Info().Msgf("Generating thumbnail image %s", mfThumb.path)
	err = util.GetThumbnail(filepath, mfThumb.path)
	if err != nil {
		log.Error().Err(err).Msgf("Error creating %s for path %s\n", mfThumb.imagetype, mfThumb.path)
	}

	var mfLargeView mediaFile
	mfLargeView.imagetype = "largeview"
	mfLargeView.path = fmt.Sprintf("temp/largeview-%s.jpeg", randString)
	mfLargeView.destination = media.S3LargeView

	log.Info().Msgf("Generating largeview image %s", mfLargeView.path)
	err = util.GetViewerImage(filepath, mfLargeView.path)
	if err != nil {
		log.Error().Err(err).Msgf("Error creating %s for path %s\n", mfLargeView.imagetype, mfLargeView.path)
	}

	var mfVeryLargeView mediaFile
	mfVeryLargeView.imagetype = "verylargeview"
	mfVeryLargeView.path = fmt.Sprintf("temp/verylargeview-%s.jpeg", randString)
	mfVeryLargeView.destination = media.S3VeryLarge

	log.Info().Msgf("Generating verlargeview image %s", mfVeryLargeView.path)
	err = util.GetVeryLargeImage(filepath, mfVeryLargeView.path)
	if err != nil {
		log.Error().Err(err).Msgf("Error creating %s for path %s\n", mfVeryLargeView.imagetype, mfVeryLargeView.path)
	}

	var mfOriginal mediaFile
	mfOriginal.imagetype = "original"
	mfOriginal.path = fmt.Sprintf("temp/original-%s.jpeg", randString)
	mfOriginal.destination = media.S3Location
	log.Info().Msgf("Copying original to temporary file %s", mfOriginal.path)

	size, err := copy(filepath, mfOriginal.path)
	if err != nil {
		log.Error().Err(err).Msgf("Error copying %s for path %s\n", mfOriginal.imagetype, mfOriginal.path)
	}
	log.Info().Msgf("Copied %d bytes for original file %s", size, mfOriginal.path)

	var mediaFiles = [4]mediaFile{mfThumb, mfLargeView, mfVeryLargeView, mfOriginal}

	for _, mf := range mediaFiles {

		start := time.Now()

		file, err := os.OpenFile(mf.path, os.O_RDONLY, 0666)
		if err != nil {
			log.Error().Err(err).Msgf("Error opening file %s to s3 bucket %s\n", mf.path, cfg.S3Bucket)
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
			Key:                  aws.String(mf.destination),
			ACL:                  aws.String("public-read"),
			Body:                 bytes.NewReader(buffer),
			ContentLength:        aws.Int64(size),
			ContentType:          aws.String(http.DetectContentType(buffer)),
			ContentDisposition:   aws.String("attachment"),
			ServerSideEncryption: aws.String("AES256"),
		})

		if err != nil {
			log.Error().Err(err).Msgf("Error uploading %s with path %s to s3 bucket %s", mf.imagetype, mf.destination, cfg.S3Bucket)
			return
		}

		// Remove the temporary file
		err = os.Remove(mf.path)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to delete temporary file %s file %s\n", mf.imagetype, mf.path)
		}

		end := time.Now()

		elapsed := end.Sub(start)
		log.Info().Msgf("Uploaded %s to s3 in %f seconds", mf.imagetype, elapsed.Seconds())

	}

	// Remove the original file
	err = os.Remove(filepath)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to delete file %s\n", filepath)
	}

	// Set media uploaded in the database
	media.S3Uploaded = "true"
	err = mediaDAO.SetS3Uploaded(media)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to set media s3 status for id %s", media.MediaID)
	}

	// start := time.Now()

	// dThumb := fmt.Sprintf("temp/thumbnail-%s.jpeg", randString)

	// // Create thumbnail
	// err = util.GetThumbnail(filepath, dThumb)
	// if err != nil {
	// 	log.Error().Err(err).Msgf("Error creating thumbnail %s\n", dThumb)
	// }

	// file, err := os.OpenFile(dThumb, os.O_RDONLY, 0666)
	// if err != nil {
	// 	log.Error().Err(err).Msgf("Error uploading file %s to s3 bucket %s\n", dThumb, cfg.S3Bucket)
	// 	return
	// }
	// defer file.Close()

	// // Get file size and read the file content into a buffer
	// fileInfo, _ := file.Stat()
	// size := fileInfo.Size()
	// buffer := make([]byte, size)
	// file.Read(buffer)

	// // Config settings: this is where you choose the bucket, filename, content-type etc.
	// // of the file you're uploading.
	// _, err = s3.New(s).PutObject(&s3.PutObjectInput{
	// 	Bucket:               aws.String(cfg.S3Bucket),
	// 	Key:                  aws.String(media.S3Thumbnail),
	// 	ACL:                  aws.String("public-read"),
	// 	Body:                 bytes.NewReader(buffer),
	// 	ContentLength:        aws.Int64(size),
	// 	ContentType:          aws.String(http.DetectContentType(buffer)),
	// 	ContentDisposition:   aws.String("attachment"),
	// 	ServerSideEncryption: aws.String("AES256"),
	// })

	// if err != nil {
	// 	fmt.Printf("Error uploading file %s to s3 bucket %s with error %s\n", filepath, cfg.S3Bucket, err)
	// 	return
	// }

	// end := time.Now()

	// elapsed := end.Sub(start)

	// fmt.Printf("Upload of thumb %s to s3 was completed in %f seconds\n", media.S3Thumbnail, elapsed.Seconds())

	// start = time.Now()

	// // Create Viewer Image
	// dView := fmt.Sprintf("temp/view-%s.jpeg", randString)
	// err = util.GetViewerBImage(filepath, dView)
	// if err != nil {
	// 	fmt.Printf("Error creating view image %s with error %s\n", dView, err)
	// }

	// file, err = os.OpenFile(dView, os.O_RDONLY, 0666)
	// if err != nil {
	// 	fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
	// 	return
	// }
	// defer file.Close()

	// // Get file size and read the file content into a buffer
	// fileInfo, _ = file.Stat()
	// size = fileInfo.Size()
	// buffer = make([]byte, size)
	// file.Read(buffer)

	// // Config settings: this is where you choose the bucket, filename, content-type etc.
	// // of the file you're uploading.
	// _, err = s3.New(s).PutObject(&s3.PutObjectInput{
	// 	Bucket:               aws.String(cfg.S3Bucket),
	// 	Key:                  aws.String(media.S3LargeView),
	// 	ACL:                  aws.String("public-read"),
	// 	Body:                 bytes.NewReader(buffer),
	// 	ContentLength:        aws.Int64(size),
	// 	ContentType:          aws.String(http.DetectContentType(buffer)),
	// 	ContentDisposition:   aws.String("attachment"),
	// 	ServerSideEncryption: aws.String("AES256"),
	// })

	// if err != nil {
	// 	fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
	// 	return
	// }

	// end = time.Now()

	// elapsed = end.Sub(start)

	// fmt.Printf("Upload of view image %s to s3 was completed in %f seconds\n", media.S3LargeView, elapsed.Seconds())

	// // 4K image
	// start = time.Now()

	// dlarge := fmt.Sprintf("temp/verylarge-%s.jpeg", randString)
	// err = util.GetVeryLargeImage(filepath, dlarge)
	// if err != nil {
	// 	fmt.Printf("Error creating view image %s with error %s\n", dlarge, err)
	// }

	// file, err = os.OpenFile(dlarge, os.O_RDONLY, 0666)
	// if err != nil {
	// 	fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
	// 	return
	// }
	// defer file.Close()

	// // Get file size and read the file content into a buffer
	// fileInfo, _ = file.Stat()
	// size = fileInfo.Size()
	// buffer = make([]byte, size)
	// file.Read(buffer)

	// // Config settings: this is where you choose the bucket, filename, content-type etc.
	// // of the file you're uploading.
	// _, err = s3.New(s).PutObject(&s3.PutObjectInput{
	// 	Bucket:               aws.String(cfg.S3Bucket),
	// 	Key:                  aws.String(media.S3VeryLarge),
	// 	ACL:                  aws.String("public-read"),
	// 	Body:                 bytes.NewReader(buffer),
	// 	ContentLength:        aws.Int64(size),
	// 	ContentType:          aws.String(http.DetectContentType(buffer)),
	// 	ContentDisposition:   aws.String("attachment"),
	// 	ServerSideEncryption: aws.String("AES256"),
	// })

	// if err != nil {
	// 	fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
	// 	return
	// }

	// end = time.Now()

	// elapsed = end.Sub(start)

	// fmt.Printf("Upload of very large image %s to s3 was completed in %f seconds\n", media.S3VeryLarge, elapsed.Seconds())

	// // Original Image
	// start = time.Now()

	// file, err = os.OpenFile(filepath, os.O_RDONLY, 0666)
	// if err != nil {
	// 	fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
	// 	return
	// }
	// defer file.Close()

	// // Get file size and read the file content into a buffer
	// fileInfo, _ = file.Stat()
	// size = fileInfo.Size()
	// buffer = make([]byte, size)
	// file.Read(buffer)

	// // Config settings: this is where you choose the bucket, filename, content-type etc.
	// // of the file you're uploading.
	// _, err = s3.New(s).PutObject(&s3.PutObjectInput{
	// 	//Bucket:               aws.String("vi-goblog"),
	// 	Bucket:               aws.String(cfg.S3Bucket),
	// 	Key:                  aws.String(media.S3Location),
	// 	ACL:                  aws.String("public-read"),
	// 	Body:                 bytes.NewReader(buffer),
	// 	ContentLength:        aws.Int64(size),
	// 	ContentType:          aws.String(http.DetectContentType(buffer)),
	// 	ContentDisposition:   aws.String("attachment"),
	// 	ServerSideEncryption: aws.String("AES256"),
	// })

	// if err != nil {
	// 	fmt.Printf("Error uploading file %s to s3 with error %s\n", filepath, err)
	// 	return
	// }

	// media.S3Uploaded = "true"
	// err = mediaDAO.SetS3Uploaded(media)
	// if err != nil {
	// 	fmt.Printf("Failed to set media s3 status for id %s with error: %s\n", media.MediaID, err)
	// }

	// end = time.Now()

	// elapsed = end.Sub(start)

	// fmt.Printf("Upload of full size image %s to s3 was completed in %f seconds\n", media.S3Location, elapsed.Seconds())

	// Remove the images we do not need
	// err = os.Remove(filepath)
	// if err != nil {
	// 	fmt.Printf("Failed to delete file %s with error: %s\n", filepath, err)
	// }

	// // Remove the images we do not need
	// err = os.Remove(dThumb)
	// if err != nil {
	// 	fmt.Printf("Failed to delete file %s with error: %s\n", dThumb, err)
	// }

	// // Remove the images we do not need
	// err = os.Remove(dView)
	// if err != nil {
	// 	fmt.Printf("Failed to delete file %s with error: %s\n", dView, err)
	// }

	// // Remove the images we do not need
	// err = os.Remove(dlarge)
	// if err != nil {
	// 	fmt.Printf("Failed to delete file %s with error: %s\n", dlarge, err)
	// }

	return
}
