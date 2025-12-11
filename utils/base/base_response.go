package base

import "encoding/json"

type BaseResponse struct {
	Status    bool   `json:"status"`
	Message   string `json:"message"`
	Error     error  `json:"error"`
	ErrorCode string `json:"error_code"`
	Data      any    `json:"data"`
	Total     *int64 `json:"total,omitempty"`
}

func (b BaseResponse) MarshalJSON() ([]byte, error) {
	out := map[string]any{
		"status": b.Status,
	}

	if b.Message != "" {
		out["message"] = b.Message
	}
	if b.ErrorCode != "" {
		out["error_code"] = b.ErrorCode
	}
	if b.Error != nil {
		out["error"] = b.Error.Error()
	}
	if b.Data != nil {
		out["data"] = b.Data
	}
	if b.Total != nil {
		out["total"] = b.Total
	}

	return json.Marshal(out)
}
