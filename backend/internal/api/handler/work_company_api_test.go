package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"rolecraft-ai/internal/api/handler"
	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/models"
	workspaceSvc "rolecraft-ai/internal/service/workspace"
)

func setupWorkCompanyAPITestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("work-company-%d.db", time.Now().UnixNano()))
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.Company{},
		&models.Role{},
		&models.Work{},
		&models.AgentRun{},
		&models.CompanyExport{},
		&models.Document{},
	))
	return db
}

func TestWorkHandlerBatchRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupWorkCompanyAPITestDB(t)
	runner := workspaceSvc.NewRunner(db, &config.Config{})
	workHandler := handler.NewWorkHandler(db, runner)

	user := models.User{
		ID:           models.NewUUID(),
		Email:        "batch@test.local",
		PasswordHash: "hashed",
	}
	require.NoError(t, db.Create(&user).Error)

	company := models.Company{
		ID:      models.NewUUID(),
		OwnerID: user.ID,
		Name:    "Batch Company",
	}
	require.NoError(t, db.Create(&company).Error)

	workA := models.Work{
		ID:          models.NewUUID(),
		UserID:      user.ID,
		CompanyID:   company.ID,
		Name:        "Batch Task A",
		TriggerType: "manual",
		Timezone:    "Asia/Shanghai",
		AsyncStatus: "idle",
		Status:      "todo",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	workB := models.Work{
		ID:          models.NewUUID(),
		UserID:      user.ID,
		CompanyID:   company.ID,
		Name:        "Batch Task B",
		TriggerType: "manual",
		Timezone:    "Asia/Shanghai",
		AsyncStatus: "idle",
		Status:      "todo",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, db.Create(&workA).Error)
	require.NoError(t, db.Create(&workB).Error)

	body, _ := json.Marshal(map[string]interface{}{
		"ids":         []string{workA.ID, workB.ID},
		"maxParallel": 2,
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/workspaces/batch/run", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("userId", user.ID)

	workHandler.BatchRun(ctx)
	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			SuccessCount int `json:"successCount"`
			FailedCount  int `json:"failedCount"`
			Items        []struct {
				WorkID string `json:"workId"`
				Status string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, 200, resp.Code)
	require.Equal(t, 2, resp.Data.SuccessCount)
	require.Equal(t, 0, resp.Data.FailedCount)
	require.Len(t, resp.Data.Items, 2)

	var runCount int64
	require.NoError(t, db.Model(&models.AgentRun{}).Where("user_id = ?", user.ID).Count(&runCount).Error)
	require.Equal(t, int64(2), runCount)
}

func TestCompanyExportAPIs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupWorkCompanyAPITestDB(t)
	companyHandler := handler.NewCompanyHandler(db)

	user := models.User{
		ID:           models.NewUUID(),
		Email:        "export@test.local",
		PasswordHash: "hashed",
	}
	require.NoError(t, db.Create(&user).Error)

	company := models.Company{
		ID:          models.NewUUID(),
		OwnerID:     user.ID,
		Name:        "Export Company",
		Description: "export tests",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, db.Create(&company).Error)

	work := models.Work{
		ID:          models.NewUUID(),
		UserID:      user.ID,
		CompanyID:   company.ID,
		Name:        "Export Task",
		TriggerType: "manual",
		Timezone:    "Asia/Shanghai",
		AsyncStatus: "completed",
		Status:      "done",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, db.Create(&work).Error)

	run := models.AgentRun{
		ID:          models.NewUUID(),
		WorkID:      work.ID,
		UserID:      user.ID,
		CompanyID:   company.ID,
		Status:      "completed",
		Summary:     "export summary",
		FinalAnswer: "export answer",
		Confidence:  0.91,
		Trace: models.ToJSON(map[string]interface{}{
			"steps": []map[string]interface{}{
				{"agent": "Planner", "purpose": "split", "output": "ok", "durationMs": 10},
			},
			"nextActions": []string{"action-1"},
			"evidence":    []string{"evidence-1"},
		}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&run).Error)

	// create export
	createBody, _ := json.Marshal(map[string]interface{}{
		"format":        "json",
		"keyword":       "export",
		"minConfidence": 0.6,
	})
	createReq, _ := http.NewRequest(http.MethodPost, "/api/v1/companies/"+company.ID+"/exports", bytes.NewBuffer(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	createCtx, _ := gin.CreateTestContext(createW)
	createCtx.Request = createReq
	createCtx.Set("userId", user.ID)
	createCtx.Params = []gin.Param{{Key: "id", Value: company.ID}}

	companyHandler.CreateExport(createCtx)
	require.Equal(t, http.StatusCreated, createW.Code)

	var createResp struct {
		Code int `json:"code"`
		Data struct {
			ID            string `json:"id"`
			CompanyID     string `json:"companyId"`
			DeliveryCount int    `json:"deliveryCount"`
			Content       string `json:"content"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(createW.Body.Bytes(), &createResp))
	require.Equal(t, 200, createResp.Code)
	require.NotEmpty(t, createResp.Data.ID)
	require.Equal(t, company.ID, createResp.Data.CompanyID)
	require.Equal(t, 1, createResp.Data.DeliveryCount)
	require.Contains(t, createResp.Data.Content, "deliveries")

	// list exports
	listReq, _ := http.NewRequest(http.MethodGet, "/api/v1/companies/"+company.ID+"/exports", nil)
	listW := httptest.NewRecorder()
	listCtx, _ := gin.CreateTestContext(listW)
	listCtx.Request = listReq
	listCtx.Set("userId", user.ID)
	listCtx.Params = []gin.Param{{Key: "id", Value: company.ID}}

	companyHandler.ListExports(listCtx)
	require.Equal(t, http.StatusOK, listW.Code)

	var listResp struct {
		Code int `json:"code"`
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(listW.Body.Bytes(), &listResp))
	require.Equal(t, 200, listResp.Code)
	require.Len(t, listResp.Data, 1)
	require.Equal(t, createResp.Data.ID, listResp.Data[0].ID)

	// get export
	getReq, _ := http.NewRequest(http.MethodGet, "/api/v1/companies/"+company.ID+"/exports/"+createResp.Data.ID, nil)
	getW := httptest.NewRecorder()
	getCtx, _ := gin.CreateTestContext(getW)
	getCtx.Request = getReq
	getCtx.Set("userId", user.ID)
	getCtx.Params = []gin.Param{
		{Key: "id", Value: company.ID},
		{Key: "exportId", Value: createResp.Data.ID},
	}

	companyHandler.GetExport(getCtx)
	require.Equal(t, http.StatusOK, getW.Code)

	var getResp struct {
		Code int `json:"code"`
		Data struct {
			ID      string `json:"id"`
			Content string `json:"content"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(getW.Body.Bytes(), &getResp))
	require.Equal(t, 200, getResp.Code)
	require.Equal(t, createResp.Data.ID, getResp.Data.ID)
	require.NotEmpty(t, getResp.Data.Content)
}
