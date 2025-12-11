package domain

type DocumentResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	FileURL     string  `json:"file_url"`
	FileType    string  `json:"file_type,omitempty"`
	Size        int64   `json:"size,omitempty"`
}
