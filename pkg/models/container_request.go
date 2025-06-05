// models/container_request.go
package models

type ContainerInitRequest struct {
	UserID  string `json:"userId"`
	RootDir string `json:"rootDir"` // Örn: "my-project/src"
}
