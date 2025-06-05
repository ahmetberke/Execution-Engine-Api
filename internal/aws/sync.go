package aws

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

// MinIO yapılandırma değişkenleri
var (
	minioEndpoint string
	accessKey     string
	secretKey     string
	region        string
	bucketName    string
)

// MinIO erişim bilgilerini yükleme
func LoadAWSCredentials() error {
	godotenv.Load()
	minioEndpoint = os.Getenv("MINIO_ENDPOINT") // MinIO Server
	accessKey = os.Getenv("MINIO_ACCESS_KEY")   // MinIO Access Key
	secretKey = os.Getenv("MINIO_SECRET_KEY")   // MinIO Secret Key
	region = os.Getenv("MINIO_REGION")          // Varsayılan Bölge
	bucketName = os.Getenv("MINIO_BUCKET")      // Bucket Adı

	if minioEndpoint == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		return fmt.Errorf("MinIO environment variables are not set properly")
	}

	return nil
}

// MinIO S3 istemcisini oluştur
func getS3Client() (*s3.S3, error) {
	err := LoadAWSCredentials()
	if err != nil {
		return nil, err
	}

	// AWS SDK üzerinden MinIO bağlantısını oluştur
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(minioEndpoint), // MinIO URL
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(true), // MinIO için gerekli
		DisableSSL:       aws.Bool(true), // HTTP ile bağlanıyoruz
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO session: %w", err)
	}

	return s3.New(sess), nil
}

// Kullanıcının dosyalarını senkronize et (S3'ten indir)
func SyncUserFiles(userID string) error {
	s3Client, err := getS3Client()
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %w", err)
	}

	prefix := fmt.Sprintf("user_clouds/%s/", userID) // Kullanıcının dosyalarının olduğu path
	tmpDir := fmt.Sprintf("tmp/%s", userID)          // Geçici dizin

	log.Println("## 1 - Temporary directory:", tmpDir)

	// Geçici klasörü oluştur (Eğer yoksa)
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	log.Println("## 2 - Fetching file list from S3")

	// S3'ten dosya listesini al
	resp, err := s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	log.Println("## 3 - Found", len(resp.Contents), "files")

	if len(resp.Contents) == 0 {
		fmt.Println("No files found in S3 bucket for user:", userID)
		return nil
	}

	for _, obj := range resp.Contents {
		key := *obj.Key
		fileName := filepath.Base(key)
		filePath := filepath.Join(tmpDir, fileName)

		log.Println("Downloading file:", key, "->", filePath)

		// Dosyayı oluştur
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()

		// S3'ten dosyayı al
		getObj, err := s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		if err != nil {
			return fmt.Errorf("failed to get object: %w", err)
		}
		defer getObj.Body.Close()

		// Dosyayı kaydet
		_, err = io.Copy(file, getObj.Body)
		if err != nil {
			return fmt.Errorf("failed to copy object data: %w", err)
		}

		log.Println("File successfully downloaded:", filePath)
	}

	return nil
}

// Belirli bir alt dizini (örneğin "src/") indirir
func SyncUserSubPath(userID string, subPath string) error {
	s3Client, err := getS3Client()
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %w", err)
	}

	fmt.Printf(">>>> SUB PATH: %v", subPath)

	prefix := fmt.Sprintf("user_clouds/%s/%s", userID, subPath) // Örn: user_clouds/abc123/src/
	tmpDir := fmt.Sprintf("tmp/%s", userID)                     // Geçici yerel klasör

	log.Println("## 1 - Temporary directory:", tmpDir)
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	log.Println("## 2 - Fetching file list from S3 with prefix:", prefix)

	resp, err := s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	log.Println("## 3 - Found", len(resp.Contents), "files")

	if len(resp.Contents) == 0 {
		fmt.Println("No files found in S3 bucket for path:", prefix)
		return nil
	}

	for _, obj := range resp.Contents {
		key := *obj.Key
		relPath := key[len(prefix):] // `subPath` sonrasını al
		localFilePath := filepath.Join(tmpDir, relPath)

		// Klasörleri oluştur (varsa)
		if err := os.MkdirAll(filepath.Dir(localFilePath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create local subdirectories: %w", err)
		}

		log.Println("Downloading file:", key, "->", localFilePath)

		file, err := os.Create(localFilePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()

		getObj, err := s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		if err != nil {
			return fmt.Errorf("failed to get object: %w", err)
		}
		defer getObj.Body.Close()

		_, err = io.Copy(file, getObj.Body)
		if err != nil {
			return fmt.Errorf("failed to copy object data: %w", err)
		}

		log.Println("File successfully downloaded:", localFilePath)
	}

	return nil
}
