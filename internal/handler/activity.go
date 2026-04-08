package handler

import (
	"core/internal/domain_model/model"
	"core/internal/repo"
	"encoding/json"
	"log/slog"
	"time"

	"gorm.io/datatypes"
)

func activityMetadata(value any) datatypes.JSON {
	if value == nil {
		return nil
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return datatypes.JSON(bytes)
}

func adminActivity(admin model.Admin, action, description string, metadata any) model.ActivityLog {
	return model.ActivityLog{
		Action:      action,
		ActorType:   "admin",
		ActorID:     admin.ID.String(),
		ActorName:   admin.Name,
		Description: description,
		Metadata:    activityMetadata(metadata),
		CreatedAt:   time.Now(),
	}
}

func userActivity(user model.User, action, description string, metadata any) model.ActivityLog {
	return model.ActivityLog{
		Action:      action,
		ActorType:   "user",
		ActorID:     user.ID.String(),
		ActorName:   user.Name,
		Description: description,
		Metadata:    activityMetadata(metadata),
		CreatedAt:   time.Now(),
	}
}

func systemActivity(action, description string, metadata any) model.ActivityLog {
	return model.ActivityLog{
		Action:      action,
		ActorType:   "system",
		ActorName:   "system",
		Description: description,
		Metadata:    activityMetadata(metadata),
		CreatedAt:   time.Now(),
	}
}

func appendProcedureActivity(userProcedureID int, entry model.ActivityLog) {
	if err := repo.AppendUserProcedureActivity(userProcedureID, entry); err != nil {
		slog.Warn("failed to append user procedure activity", "user_procedure_id", userProcedureID, "error", err)
	}
}
