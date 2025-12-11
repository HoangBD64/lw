package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Model struct {
	ID        uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (u *Model) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

// JsonData is a generic JSON wrapper for any type T.
type JsonData[T any] struct {
	lenByte int   // length of marshaled data
	Data    T     // extracted data
	Error   error // marshal error
}

// Value implements driver.Valuer — convert Value to JSON before saving.
func (j JsonData[T]) Value() (driver.Value, error) {
	return json.Marshal(j.Data)
}

// Scan implements sql.Scanner — parse JSON back into Value.
func (j *JsonData[T]) Scan(src any) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %s", src)
	}

	var result T
	if err := json.Unmarshal(bytes, &result); err != nil {
		j.Error = err
	}

	j.Data = result
	j.lenByte = len(bytes)

	return nil
}

func (j JsonData[T]) IsZero() bool {
	return j.lenByte == 0
}

func (j JsonData[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Data)
}

func (j *JsonData[T]) UnmarshalJSON(bytes []byte) error {
	var result T
	if err := json.Unmarshal(bytes, &result); err != nil {
		j.Error = err
	}

	j.Data = result
	j.lenByte = len(bytes)

	return nil
}

func (j JsonData[T]) GormDataType() string {
	return "json"
}

func (j JsonData[T]) DatatypesJSON() datatypes.JSON {
	bytes, _ := j.MarshalJSON()
	return bytes
}

func NewJsonData[T any](data T) JsonData[T] {
	return JsonData[T]{
		Data:  data,
		Error: nil,
	}
}

// JsonSlice is a placeholder type for slices stored as JSON in the database without any type constraints, allow flexible appending of elements.
// targeted for HISTORY/LOG fields where we just need to append new entries without reading existing ones.
//
// its core is []any so only WRITE operations are supported.
//
// Experimential: use with caution.
type JsonSlice[T any] struct {
	arr   []any
	Bytes []byte
}

func (j *JsonSlice[T]) Append(val T) {
	j.arr = append(j.arr, val)
}

func (JsonSlice[T]) GormDataType() string {
	return "json"
}

func (j *JsonSlice[T]) Scan(src any) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %s", src)
	}

	j.Bytes = bytes

	var val []any
	if err := json.Unmarshal(bytes, &val); err != nil {
		return err
	}

	j.arr = val

	return nil
}
