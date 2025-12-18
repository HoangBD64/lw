package domain

type DocumentCreateRequest struct {
	Name            string `json:"name" validate:"required"`
	Description     string `json:"description,omitempty"`
	FileURL         string `json:"file_url" validate:"required,url"`
	FileType        string `json:"file_type,omitempty"`
	DocumentType    string `json:"document_type,omitempty"`
	Size            int64  `json:"size,omitempty"`
	UserProcedureID int    `json:"user_procedure_id,omitempty"`
}

type DocumentUpdateRequest struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	FileURL      string `json:"file_url"`
	FileType     string `json:"file_type,omitempty"`
	DocumentType string `json:"document_type,omitempty"`
	Size         int64  `json:"size,omitempty"`
}

type DocumentDeleteRequest struct {
	IDs []int `json:"ids" validate:"required,gt=0"`
}

type DocumentAssignRequest struct {
	Assignments []DocumentAssign `json:"assignments" validate:"required,dive"`
}

type DocumentUnassignRequest struct {
	Unassignments []DocumentAssign `json:"unassignments" validate:"required,dive"`
}

type DocumentAssign struct {
	UserID     string `json:"user_id" validate:"required"`
	DocumentID int    `json:"document_id" validate:"required,gt=0"`
}
