// Copyright 2026

package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hanchuanchuan/goInception/session"
	"golang.org/x/net/context"
)

type auditRequest struct {
	Host             string `json:"host" binding:"required"`
	Port             int    `json:"port" binding:"required"`
	User             string `json:"user" binding:"required"`
	Password         string `json:"password"`
	DB               string `json:"db" binding:"required"`
	SQL              string `json:"sql" binding:"required"`
	Backup           bool   `json:"backup,omitempty"`
	IgnoreWarnings   bool   `json:"ignore_warnings,omitempty"`
	Sleep            int    `json:"sleep,omitempty"`
	SleepRows        int    `json:"sleep_rows,omitempty"`
	MiddlewareExtend string `json:"middleware_extend,omitempty"`
	MiddlewareDB     string `json:"middleware_db,omitempty"`
	ParseHost        string `json:"parse_host,omitempty"`
	ParsePort        int    `json:"parse_port,omitempty"`
	Fingerprint      bool   `json:"fingerprint,omitempty"`
	Print            bool   `json:"print,omitempty"`
	Masking          bool   `json:"masking,omitempty"`
	Split            bool   `json:"split,omitempty"`
	RealRowCount     bool   `json:"real_row_count,omitempty"`
	SSL              string `json:"ssl,omitempty"`
	SSLCa            string `json:"ssl_ca,omitempty"`
	SSLCert          string `json:"ssl_cert,omitempty"`
	SSLKey           string `json:"ssl_key,omitempty"`
	TranBatch        int    `json:"tran_batch,omitempty"`
}

type apiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type auditRecord struct {
	SeqNo          int    `json:"seq_no"`
	Stage          string `json:"stage"`
	StageStatus    string `json:"stage_status"`
	ErrLevel       uint8  `json:"err_level"`
	ErrorMessage   string `json:"error_message"`
	Sql            string `json:"sql"`
	AffectedRows   int64  `json:"affected_rows"`
	BackupDBName   string `json:"backup_db_name,omitempty"`
	ExecTime       string `json:"exec_time,omitempty"`
	BackupCostTime string `json:"backup_cost_time,omitempty"`
	Sqlsha1        string `json:"sqlsha1,omitempty"`
	DDLRollback    string `json:"ddl_rollback,omitempty"`
	OPID           string `json:"opid,omitempty"`
}

type auditResponse struct {
	Records []auditRecord `json:"records"`
}

// NewGinEngine returns a gin router configured for goInception SQL audit and execute APIs.
func NewGinEngine() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, apiResponse{Code: 0, Message: "ok"})
	})

	api := router.Group("/api/v1")
	api.POST("/audit", func(c *gin.Context) {
		handleRequest(c, false)
	})
	api.POST("/execute", func(c *gin.Context) {
		handleRequest(c, true)
	})

	return router
}

func handleRequest(c *gin.Context, execute bool) {
	var req auditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	records, err := processSQL(c, &req, execute)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiResponse{Code: 0, Data: auditResponse{Records: records}})
}

func processSQL(c *gin.Context, req *auditRequest, execute bool) ([]auditRecord, error) {
	s := session.NewInception()
	defer s.Close()

	opt := session.SourceOptions{
		Host:             req.Host,
		Port:             req.Port,
		User:             req.User,
		Password:         req.Password,
		DB:               req.DB,
		Backup:           req.Backup,
		IgnoreWarnings:   req.IgnoreWarnings,
		Sleep:            req.Sleep,
		SleepRows:        req.SleepRows,
		MiddlewareExtend: req.MiddlewareExtend,
		MiddlewareDB:     req.MiddlewareDB,
		ParseHost:        req.ParseHost,
		ParsePort:        req.ParsePort,
		Fingerprint:      req.Fingerprint,
		Print:            req.Print,
		Masking:          req.Masking,
		Split:            req.Split,
		RealRowCount:     req.RealRowCount,
		SSL:              req.SSL,
		SSLCa:            req.SSLCa,
		SSLCert:          req.SSLCert,
		SSLKey:           req.SSLKey,
		TranBatch:        req.TranBatch,
	}

	if execute {
		opt.Check = false
		opt.Execute = true
	} else {
		opt.Check = true
		opt.Execute = false
	}

	if err := s.LoadOptions(opt); err != nil {
		return nil, err
	}

	ctx := context.Background()
	var result []session.Record
	var err error
	if execute {
		result, err = s.RunExecute(ctx, req.SQL)
	} else {
		result, err = s.Audit(ctx, req.SQL)
	}
	if err != nil {
		return nil, err
	}

	records := make([]auditRecord, 0, len(result))
	for _, r := range result {
		records = append(records, convertRecord(r))
	}
	return records, nil
}

func convertRecord(r session.Record) auditRecord {
	stage := "UNKNOWN"
	if int(r.Stage) >= 0 && int(r.Stage) < len(session.StageList) {
		stage = session.StageList[r.Stage]
	}
	status := "UNKNOWN"
	if int(r.StageStatus) >= 0 && int(r.StageStatus) < len(session.StatusList) {
		status = session.StatusList[r.StageStatus]
	}

	return auditRecord{
		SeqNo:          r.SeqNo,
		Stage:          stage,
		StageStatus:    status,
		ErrLevel:       r.ErrLevel,
		ErrorMessage:   r.ErrorMessage,
		Sql:            r.Sql,
		AffectedRows:   r.AffectedRows,
		BackupDBName:   r.BackupDBName,
		ExecTime:       r.ExecTime,
		BackupCostTime: r.BackupCostTime,
		Sqlsha1:        r.Sqlsha1,
		DDLRollback:    r.DDLRollback,
		OPID:           r.OPID,
	}
}
