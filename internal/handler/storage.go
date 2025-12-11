package handler

import (
	"core/internal/common"
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	"core/pkg/storage/s3"
	"core/utils/base"

	"github.com/gofiber/fiber/v2"
)

func GetUploadUrl(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false})
	}
	var data s3.PutObjectUpload
	if err := c.BodyParser(&data); err != nil {
		return c.Status(500).JSON(base.BaseResponse{Status: false, Error: err})
	}
	data.ParentFolder = "assets"

	s3Key := common.S3Credential(c.UserContext())
	url, direct_url, err := s3.GetUrlS3(data, user, s3Key)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetUrlS3", "error": err.Error()})
	}

	resp := domain.UploadResponse{
		Url:       url,
		DirectUrl: direct_url,
	}
	return c.JSON(base.BaseResponse{Status: true, Data: resp})
}
func AdminGetUploadUrl(c *fiber.Ctx) error {
	admin, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again", Error: nil})
	}

	var data s3.PutObjectUpload
	if err := c.BodyParser(&data); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": false, "data": err})
	}
	s3Key := common.S3Credential(c.UserContext())
	url, direct_url, err := s3.AdminGetUrlS3(data, admin, s3Key)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetUrlS3", Error: err})
	}
	resp := domain.UploadResponse{
		Url:       url,
		DirectUrl: direct_url,
	}
	return c.JSON(base.BaseResponse{Status: true, Data: resp})

}
