package domain

type UserDashboardResponse struct {
	InformationRequired int64 `json:"information_required"`
	InformationReview   int64 `json:"information_review"`
	Pending             int64 `json:"pending"`
	Completed           int64 `json:"completed"`
}
