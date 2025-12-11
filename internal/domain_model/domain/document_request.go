package domain

type DocumentCreateRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	FileURL     string  `json:"file_url" validate:"required,url"`
	FileType    string  `json:"file_type,omitempty"`
	Size        int64   `json:"size,omitempty"`
}

type DocumentUpdateRequest struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	FileURL     string  `json:"file_url"`
	FileType    *string `json:"file_type,omitempty"`
	Size        *int64  `json:"size,omitempty"`
}
