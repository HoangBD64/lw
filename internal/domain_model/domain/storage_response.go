package domain

type UploadResponse struct {
	Url       string `json:"url"`
	DirectUrl string `json:"direct_url"`
}

type RenameFileRequest struct {
	// FileURL string `json:"file_url" validate:"required,url"`
	ID      int    `json:"id" validate:"required"`
	NewName string `json:"new_name" validate:"required"`
}

type RenameFileResponse struct {
	FileURL string `json:"file_url"`
	Key     string `json:"key"`
}
