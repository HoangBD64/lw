package handler

import (
	"context"
	"core/internal/domain_model/model"
	"core/internal/repo"
	base "core/utils/base"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

func AdminGetSystemStats(c *fiber.Ctx) error {
	_, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{
			Status:  false,
			Message: "Please login again as admin",
		})
	}

	var q repo.Query
	q.Parse(c)
	if q.StartTime == nil || q.EndTime == nil {
		return c.JSON(base.BaseResponse{
			Status:  false,
			Message: "must specify time range",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		statusCounts     map[string]int64
		stepStatusCounts map[string]int64
		templateUsages   any
		newUsers         any
	)

	var wg sync.WaitGroup
	errCh := make(chan error, 4)

	wg.Add(1)
	go func() {
		defer wg.Done()
		sc, err := repo.GetUserProcedureStatusCounts(ctx, q)
		if err != nil {
			errCh <- err
			return
		}
		statusCounts = sc
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ssc, err := repo.GetUserProcedureStepStatusCounts(ctx, q)
		if err != nil {
			errCh <- err
			return
		}
		stepStatusCounts = ssc
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		tu, err := repo.GetTemplateUsages(ctx, q)
		if err != nil {
			errCh <- err
			return
		}
		templateUsages = tu
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		nu, err := repo.GetAnalyticsNewUser(ctx, q)
		if err != nil {
			errCh <- err
			return
		}
		newUsers = nu
	}()

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "failed to get system stats", Error: err})
		}
	}

	upCounts := 0
	for _, count := range statusCounts {
		upCounts += int(count)
	}
	data := fiber.Map{
		"user_procedure_status":      statusCounts,
		"user_procedure_count":       upCounts,
		"user_procedure_step_status": stepStatusCounts,
		"template_usage":             templateUsages,
		"new_users":                  newUsers,
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: data})
}
