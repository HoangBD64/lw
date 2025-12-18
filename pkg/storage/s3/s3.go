package s3

import (
	"bytes"
	"context"
	"core/internal/domain_model/model"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type PutObjectUpload struct {
	Acl             string `json:"acl"`
	ContentType     string `json:"content_type"`
	Title           string `json:"title,omitempty"`
	ContentEncoding string `json:"content_encoding"`
	Key             string `json:"key"`
	UploadThumbnail bool   `json:"upload_thumbnail"`
	ParentFolder    string `json:"parent_folder"`
}

func GetFolder(contentType string) string {
	if len(contentType) < 1 {
		return "other"
	}
	var folder string
	switch contentType {
	case "image/png", "image/jpeg", "image/bmp":
		folder = "images"
	case "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "application/pdf":
		folder = "documents"
	case "video/mp4", "video/mpeg", "audio/mpeg":
		folder = "media"
	default:
		folder = "other"
	}
	return folder
}

func GetFileExtension(contentType string) string {
	if len(contentType) < 1 {
		return ""
	}
	var ext string
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/jpeg":
		ext = ".jpg"
	case "image/bmp":
		ext = ".bmp"
	case "application/msword":
		ext = ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		ext = ".docx"
	case "application/vnd.ms-excel":
		ext = ".xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		ext = ".xlsx"
	case "application/pdf":
		ext = ".pdf"
	case "video/mp4":
		ext = ".mp4"
	case "video/mpeg":
		ext = ".mpeg"
	case "audio/mpeg":
		ext = ".mp3"
	case "text/plain":
		ext = ".txt"
	case "image/svg+xml":
		ext = ".svg"
	case "application/zip":
		ext = ".zip"
	case "image/webp":
		ext = ".webp"
	case "application/x-zip-compressed":
		ext = ".zip"
	case "image/vnd.adobe.photoshop":
		ext = ".psd"
	case "emb":
		ext = ".emb"
	}
	return ext
}

type KeyS3 struct {
	Bucket      string `json:"bucket"`
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
	EndPointApi string `json:"end_point_api"`
	EndPointCdn string `json:"end_point_cdn"`
	IsSSL       bool   `json:"is_ssl"`
}

func GetUrlS3(data PutObjectUpload, user model.User, key KeyS3) (string, string, error) {
	accessKey := key.AccessKey
	secKey := key.SecretKey
	endpoint := key.EndPointApi
	endpoint_cdn := key.EndPointCdn
	bucketName := key.Bucket // Space names must be globally unique
	isSSL := key.IsSSL       // Space names must be globally unique
	// location := "us-east-1"
	// ssl := true
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secKey, ""),
		Secure: isSSL,
	})
	if err != nil {
		slog.Error("failed to create minio client", "error", err)
		return "", "", err
	}

	if data.Title == "" {
		var ext = GetFileExtension(data.ContentType)
		uuidKey, _ := uuid.NewUUID()
		data.Title = strings.TrimSpace(fmt.Sprintf("%s%s", uuidKey, ext))
	} else {
		data.Title = GenerateHandle(data.Title)
	}
	uuidKey, _ := uuid.NewUUID()
	objectKey := fmt.Sprintf("%s/%s/%s/%s/%s", data.ParentFolder, fmt.Sprint(user.Name), time.Now().Format("200601"), uuidKey, data.Title)
	presignedURL, err := minioClient.PresignedPutObject(context.Background(), bucketName, objectKey, time.Duration(1000)*time.Second)
	if err != nil {
		return "", "", err
	}
	var direct_url string
	url := presignedURL.String()
	if endpoint_cdn != "" && endpoint_cdn != endpoint {
		if strings.HasPrefix(presignedURL.Path, "/"+bucketName) || strings.HasPrefix(presignedURL.Path, bucketName) {
			presignedURL.Path = strings.TrimPrefix(presignedURL.Path, "/"+bucketName)
		}
		endpoint = endpoint_cdn
	}
	if isSSL {
		direct_url = fmt.Sprintf("https://%s%s", endpoint, presignedURL.Path)
	} else {
		direct_url = fmt.Sprintf("http://%s%s", endpoint, presignedURL.Path)
	}

	return url, direct_url, nil
}
func AdminGetUrlS3(data PutObjectUpload, admin model.Admin, key KeyS3) (string, string, error) {
	accessKey := key.AccessKey
	secKey := key.SecretKey
	endpoint := key.EndPointApi
	endpoint_cdn := key.EndPointCdn
	bucketName := key.Bucket // Space names must be globally unique
	isSSL := key.IsSSL       // Space names must be globally unique
	// location := "us-east-1"
	// ssl := true
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secKey, ""),
		Secure: isSSL,
	})
	if err != nil {
		slog.Error("failed to create minio client", "error", err)
		return "", "", err
	}

	var ext = GetFileExtension(data.ContentType)

	uuidKey, _ := uuid.NewUUID()
	// var folder = GetFolder(data.ContentType)
	objectKey := fmt.Sprintf("%s/%s/%s%s", "admin", time.Now().Format("200601"), uuidKey, ext)
	presignedURL, err := minioClient.PresignedPutObject(context.Background(), bucketName, objectKey, time.Duration(1000)*time.Second)
	if err != nil {
		return "", "", err
	}
	var direct_url string
	url := presignedURL.String()
	if endpoint_cdn != "" && endpoint_cdn != endpoint {
		if strings.HasPrefix(presignedURL.Path, "/"+bucketName) || strings.HasPrefix(presignedURL.Path, bucketName) {
			presignedURL.Path = strings.TrimPrefix(presignedURL.Path, "/"+bucketName)
		}
		endpoint = endpoint_cdn
	}
	if isSSL {
		direct_url = fmt.Sprintf("https://%s%s", endpoint, presignedURL.Path)
	} else {
		direct_url = fmt.Sprintf("http://%s%s", endpoint, presignedURL.Path)
	}

	return url, direct_url, nil
}

func GetStoreInvoiceUrlS3(data PutObjectUpload, storeID uuid.UUID, key KeyS3) (string, string, error) {
	var directURL string

	accessKey := key.AccessKey
	secKey := key.SecretKey
	endpoint := key.EndPointApi
	endpointCdn := key.EndPointCdn
	bucketName := key.Bucket // Space names must be globally unique
	isSSL := key.IsSSL       // Space names must be globally unique

	minioClient, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(accessKey, secKey, ""), Secure: isSSL})

	if err != nil {
		slog.Error("failed to create minio client", "error", err)
		return "", "", err
	}

	var ext = GetFileExtension(data.ContentType)

	now := time.Now()
	fileName := fmt.Sprintf("invoice_%s%s", now.Format("20060102_150405"), ext)
	objectKey := fmt.Sprintf("%s/%s/%s/%s", "invoices", storeID, now.Format("200601"), fileName)

	presignedURL, err := minioClient.PresignedPutObject(context.Background(), bucketName, objectKey, time.Duration(1000)*time.Second)
	if err != nil {
		slog.Error("failed to create presigned upload URL", "error", err)
		return "", "", err
	}

	url := presignedURL.String()
	if endpointCdn != "" && endpointCdn != endpoint {
		if strings.HasPrefix(presignedURL.Path, "/"+bucketName) || strings.HasPrefix(presignedURL.Path, bucketName) {
			presignedURL.Path = strings.TrimPrefix(presignedURL.Path, "/"+bucketName)
		}
		endpoint = endpointCdn
	}

	if isSSL {
		directURL = fmt.Sprintf("https://%s%s", endpoint, presignedURL.Path)
	} else {
		directURL = fmt.Sprintf("http://%s%s", endpoint, presignedURL.Path)
	}

	return url, directURL, nil
}

func GenerateHandle(title string) string {
	handle := strings.ToLower(title)
	handle = strings.ReplaceAll(handle, "_", "-")
	dotIdx := strings.LastIndex(handle, ".")
	ext := ""
	if dotIdx > 0 && dotIdx < len(handle)-1 {
		ext = handle[dotIdx:]
		handle = handle[:dotIdx]
	}
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s-]+`)
	handle = reg.ReplaceAllString(handle, "")
	handle = strings.ReplaceAll(handle, " ", "-")
	handle = strings.Trim(handle, "-")
	if ext != "" {
		handle = handle + ext
	}
	return handle
}

func UploadLocalFile(ctx context.Context, key KeyS3, objectName string, filePath string) (string, error) {
	accessKey := key.AccessKey
	secKey := key.SecretKey
	endpoint := key.EndPointApi
	bucketName := key.Bucket
	isSSL := key.IsSSL

	client, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(accessKey, secKey, ""), Secure: isSSL})
	if err != nil {
		slog.Error("failed to create minio client", "error", err)
		return "", err
	}

	info, err := client.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		slog.Error("failed to upload object", "error", err, "object_name", objectName)
		return "", err
	}

	slog.Info("file uploaded successfully", "size", info.Size, "bucket", info.Bucket, "object_name", info.Key)
	url := getPublicObjectURL(endpoint, key.EndPointCdn, info.Bucket, fmt.Sprintf("/%s/%s", info.Bucket, info.Key), isSSL)

	return url, nil
}

// doesn't work if bucket has private access
func getPublicObjectURL(endpoint string, endpointCdn string, bucketName string, path string, isSSL bool) string {
	// trim bucket name from path if exists
	if endpointCdn != "" && endpointCdn != endpoint {
		bucketPath := "/" + bucketName
		if strings.HasPrefix(path, bucketPath) || strings.HasPrefix(path, bucketName) {
			path = strings.TrimPrefix(path, bucketPath)
		}
		endpoint = endpointCdn
	}

	var directUrl string
	if isSSL {
		directUrl = fmt.Sprintf("https://%s%s", endpoint, path)
	} else {
		directUrl = fmt.Sprintf("http://%s%s", endpoint, path)
	}

	return directUrl
}

func GetPresignedObject(ctx context.Context, key KeyS3, objectName string) (string, error) {
	accessKey := key.AccessKey
	secKey := key.SecretKey
	endpoint := key.EndPointApi
	endpointCdn := key.EndPointCdn
	bucketName := key.Bucket
	isSSL := key.IsSSL

	client, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(accessKey, secKey, ""), Secure: isSSL})
	if err != nil {
		slog.Error("failed to create minio client", "error", err)
		return "", err
	}

	presignedURL, err := client.PresignedGetObject(ctx, bucketName, objectName, time.Duration(1000)*time.Second, nil)
	if err != nil {
		slog.Error("failed to create presigned url", "error", err)
		return "", err
	}
	url := getPublicObjectURL(endpoint, endpointCdn, bucketName, presignedURL.Path, isSSL)

	return url, nil
}

type PutObjectRepsonse struct {
	PublicUrl *string `json:"public_url"`
	SignedUrl string  `json:"signed_url"`
}

func EntityPutPresignedObject(ctx context.Context, request PutObjectUpload, entityKey string, key KeyS3) (*PutObjectRepsonse, error) {
	credential := credentials.NewStaticV4(key.AccessKey, key.SecretKey, "")
	minioClient, err := minio.New(key.EndPointApi, &minio.Options{Creds: credential, Secure: key.IsSSL})
	if err != nil {
		return nil, err
	}

	var ext = GetFileExtension(request.ContentType)

	uuidKey := uuid.New()

	objectKey := fmt.Sprintf("%s/%s_%s%s", entityKey, time.Now().Format("200601"), uuidKey, ext)

	presignedURL, err := minioClient.PresignedPutObject(ctx, key.Bucket, objectKey, time.Duration(1000)*time.Second)
	if err != nil {
		return nil, err
	}

	publicUrl := getPublicObjectURL(key.EndPointApi, key.EndPointCdn, key.Bucket, presignedURL.Path, key.IsSSL)
	signedUrl := presignedURL.String()

	out := &PutObjectRepsonse{
		PublicUrl: &publicUrl,
		SignedUrl: signedUrl,
	}

	return out, nil
}

func GetURLS3ExcelLot(ctx context.Context, content []byte, fileName string, key KeyS3) (string, error) {
	endpoint := key.EndPointApi
	endpoint_cdn := key.EndPointCdn
	accessKey := key.AccessKey
	secKey := key.SecretKey
	isSSL := key.IsSSL
	bucketName := key.Bucket

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secKey, ""),
		Secure: isSSL,
	})
	if err != nil {
		slog.Error("failed to create minio client", "error", err)
		return "", err
	}
	keyURL := fmt.Sprintf("lots/%s/%s", time.Now().Format("20060102"), fileName)
	_, err = client.PutObject(ctx, bucketName, keyURL, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	})
	if err != nil {
		slog.Error("failed to upload object", "error", err)
		return "", err
	}
	url := ""
	if endpoint_cdn != "" && endpoint_cdn != endpoint {
		endpoint = endpoint_cdn
	}
	file_path := fmt.Sprintf("/%s/%s", bucketName, keyURL)
	if isSSL {
		url = fmt.Sprintf("https://%s%s", endpoint, file_path)
	} else {
		url = fmt.Sprintf("http://%s%s", endpoint, file_path)
	}

	return url, nil
}
