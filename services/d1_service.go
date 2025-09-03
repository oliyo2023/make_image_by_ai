package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"make_image_by_ai/config"
	"make_image_by_ai/models"
)

// D1Service Cloudflare D1 数据库服务
type D1Service struct {
	config *config.Config
	client *http.Client
}

// D1Response D1 API响应结构
type D1Response struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Result []struct {
		Meta struct {
			Duration     float64 `json:"duration"`
			ChangesCount int     `json:"changes"`
			LastRowID    int     `json:"last_row_id"`
			RowsRead     int     `json:"rows_read"`
			RowsWritten  int     `json:"rows_written"`
			SizeAfter    int     `json:"size_after"`
		} `json:"meta"`
		Results []map[string]interface{} `json:"results"`
	} `json:"result"`
}

// NewD1Service 创建D1服务实例
func NewD1Service(cfg *config.Config) (*D1Service, error) {
	if cfg.D1AccountID() == "" || cfg.D1APIToken() == "" || cfg.D1DatabaseID() == "" {
		return nil, fmt.Errorf("D1配置不完整，请检查账户ID、API令牌和数据库ID")
	}

	service := &D1Service{
		config: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout()) * time.Second,
		},
	}

	// 初始化数据库表
	if err := service.initDatabase(); err != nil {
		log.Printf("警告: 初始化D1数据库失败: %v", err)
	}

	return service, nil
}

// initDatabase 初始化数据库表
func (d *D1Service) initDatabase() error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS image_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			original_prompt TEXT NOT NULL,
			english_prompt TEXT NOT NULL,
			local_path TEXT,
			r2_url TEXT,
			file_size INTEGER DEFAULT 0,
			width INTEGER DEFAULT 0,
			height INTEGER DEFAULT 0,
			format TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	// 执行建表语句
	if err := d.executeSQL(createTableSQL); err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	// 添加缺失的列（如果不存在）
	alterStatements := []string{
		"ALTER TABLE image_records ADD COLUMN width INTEGER DEFAULT 0;",
		"ALTER TABLE image_records ADD COLUMN height INTEGER DEFAULT 0;",
		"ALTER TABLE image_records ADD COLUMN file_size INTEGER DEFAULT 0;",
		"ALTER TABLE image_records ADD COLUMN format TEXT DEFAULT '';",
	}

	for _, alterSQL := range alterStatements {
		if err := d.executeSQL(alterSQL); err != nil {
			// 忽略"column already exists"错误
			if !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "duplicate column") {
				log.Printf("警告: 添加列失败: %v", err)
			}
		}
	}

	indexSQL := `
		CREATE INDEX IF NOT EXISTS idx_created_at ON image_records(created_at);
		CREATE INDEX IF NOT EXISTS idx_original_prompt ON image_records(original_prompt);
	`

	// 执行索引创建
	if err := d.executeSQL(indexSQL); err != nil {
		log.Printf("警告: 创建索引失败: %v", err)
	}

	// 初始化编辑相关表
	if err := d.initEditDatabase(); err != nil {
		log.Printf("警告: 初始化编辑数据库表失败: %v", err)
	}

	log.Println("D1数据库表初始化完成")
	return nil
}

// SaveImageRecord 保存图片记录
func (d *D1Service) SaveImageRecord(record *models.ImageRecord) error {
	// 先尝试使用完整的SQL
	fullSQL := `
		INSERT INTO image_records 
		(original_prompt, english_prompt, local_path, r2_url, file_size, width, height, format)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	fullParams := []interface{}{
		record.OriginalPrompt,
		record.EnglishPrompt,
		record.LocalPath,
		record.R2URL,
		record.FileSize,
		record.Width,
		record.Height,
		record.Format,
	}

	if err := d.executeSQLWithParams(fullSQL, fullParams); err != nil {
		// 如果包含"no column named"错误，尝试使用简化的SQL
		if strings.Contains(err.Error(), "no column named") {
			log.Printf("警告: 表结构不完整，使用简化插入: %v", err)
			return d.saveImageRecordSimple(record)
		}
		return fmt.Errorf("保存图片记录失败: %v", err)
	}

	log.Printf("图片记录已保存到D1: %s", record.OriginalPrompt)
	return nil
}

// saveImageRecordSimple 使用简化的SQL保存记录（处理缺失列的情况）
func (d *D1Service) saveImageRecordSimple(record *models.ImageRecord) error {
	simpleSQL := `
		INSERT INTO image_records 
		(original_prompt, english_prompt, local_path, r2_url)
		VALUES (?, ?, ?, ?)
	`

	simpleParams := []interface{}{
		record.OriginalPrompt,
		record.EnglishPrompt,
		record.LocalPath,
		record.R2URL,
	}

	if err := d.executeSQLWithParams(simpleSQL, simpleParams); err != nil {
		return fmt.Errorf("保存简化图片记录失败: %v", err)
	}

	log.Printf("图片记录已保存到D1（简化模式）: %s", record.OriginalPrompt)
	return nil
}

// GetImageRecords 获取图片记录列表
func (d *D1Service) GetImageRecords(req *models.ImageRecordRequest) (*models.ImageRecordResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// 构建WHERE条件
	var conditions []string
	var params []interface{}

	if req.Keyword != "" {
		conditions = append(conditions, "(original_prompt LIKE ? OR english_prompt LIKE ?)")
		keyword := "%" + req.Keyword + "%"
		params = append(params, keyword, keyword)
	}

	if req.DateFrom != "" {
		conditions = append(conditions, "created_at >= ?")
		params = append(params, req.DateFrom)
	}

	if req.DateTo != "" {
		conditions = append(conditions, "created_at <= ?")
		params = append(params, req.DateTo+" 23:59:59")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) as total FROM image_records %s", whereClause)
	countResult, err := d.querySQLWithParams(countSQL, params)
	if err != nil {
		return nil, fmt.Errorf("查询总数失败: %v", err)
	}

	total := 0
	if len(countResult) > 0 {
		if totalVal, ok := countResult[0]["total"].(float64); ok {
			total = int(totalVal)
		}
	}

	// 查询数据
	offset := (req.Page - 1) * req.Limit
	dataSQL := fmt.Sprintf(`
		SELECT * FROM image_records %s 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`, whereClause)

	dataParams := append(params, req.Limit, offset)
	results, err := d.querySQLWithParams(dataSQL, dataParams)
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %v", err)
	}

	// 转换结果
	var records []models.ImageRecord
	for _, result := range results {
		record := models.ImageRecord{}
		if id, ok := result["id"].(float64); ok {
			record.ID = int(id)
		}
		if val, ok := result["original_prompt"].(string); ok {
			record.OriginalPrompt = val
		}
		if val, ok := result["english_prompt"].(string); ok {
			record.EnglishPrompt = val
		}
		if val, ok := result["local_path"].(string); ok {
			record.LocalPath = val
		}
		if val, ok := result["r2_url"].(string); ok {
			record.R2URL = val
		}
		if val, ok := result["file_size"].(float64); ok {
			record.FileSize = int64(val)
		}
		if val, ok := result["width"].(float64); ok {
			record.Width = int(val)
		}
		if val, ok := result["height"].(float64); ok {
			record.Height = int(val)
		}
		if val, ok := result["format"].(string); ok {
			record.Format = val
		}
		if val, ok := result["created_at"].(string); ok {
			record.CreatedAt = val
		}
		if val, ok := result["updated_at"].(string); ok {
			record.UpdatedAt = val
		}

		records = append(records, record)
	}

	pages := (total + req.Limit - 1) / req.Limit

	return &models.ImageRecordResponse{
		Records: records,
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
		Pages:   pages,
	}, nil
}

// GetImageRecordByID 根据ID获取图片记录
func (d *D1Service) GetImageRecordByID(id int) (*models.ImageRecord, error) {
	sql := "SELECT * FROM image_records WHERE id = ?"
	results, err := d.querySQLWithParams(sql, []interface{}{id})
	if err != nil {
		return nil, fmt.Errorf("查询图片记录失败: %v", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("图片记录不存在")
	}

	result := results[0]
	record := &models.ImageRecord{}

	if id, ok := result["id"].(float64); ok {
		record.ID = int(id)
	}
	if val, ok := result["original_prompt"].(string); ok {
		record.OriginalPrompt = val
	}
	if val, ok := result["english_prompt"].(string); ok {
		record.EnglishPrompt = val
	}
	if val, ok := result["local_path"].(string); ok {
		record.LocalPath = val
	}
	if val, ok := result["r2_url"].(string); ok {
		record.R2URL = val
	}
	if val, ok := result["file_size"].(float64); ok {
		record.FileSize = int64(val)
	}
	if val, ok := result["width"].(float64); ok {
		record.Width = int(val)
	}
	if val, ok := result["height"].(float64); ok {
		record.Height = int(val)
	}
	if val, ok := result["format"].(string); ok {
		record.Format = val
	}
	if val, ok := result["created_at"].(string); ok {
		record.CreatedAt = val
	}
	if val, ok := result["updated_at"].(string); ok {
		record.UpdatedAt = val
	}

	return record, nil
}

// executeSQL 执行SQL语句
func (d *D1Service) executeSQL(sql string) error {
	return d.executeSQLWithParams(sql, nil)
}

// executeSQLWithParams 执行带参数的SQL语句
func (d *D1Service) executeSQLWithParams(sql string, params []interface{}) error {
	requestBody := map[string]interface{}{
		"sql":    sql,
		"params": params,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %v", err)
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s/query",
		d.config.D1AccountID(), d.config.D1DatabaseID())

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.config.D1APIToken())
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		// 限制响应体日志长度，避免打印大量数据
		responsePreview := string(body)
		if len(responsePreview) > 500 {
			responsePreview = responsePreview[:500] + "... (响应内容被截断)"
		}
		return fmt.Errorf("D1 API请求失败，状态码: %d, 响应: %s", resp.StatusCode, responsePreview)
	}

	var d1Resp D1Response
	if err := json.Unmarshal(body, &d1Resp); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if !d1Resp.Success {
		if len(d1Resp.Errors) > 0 {
			return fmt.Errorf("D1错误: %s", d1Resp.Errors[0].Message)
		}
		return fmt.Errorf("D1请求失败")
	}

	return nil
}

// querySQLWithParams 查询带参数的SQL语句
func (d *D1Service) querySQLWithParams(sql string, params []interface{}) ([]map[string]interface{}, error) {
	requestBody := map[string]interface{}{
		"sql":    sql,
		"params": params,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s/query",
		d.config.D1AccountID(), d.config.D1DatabaseID())

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.config.D1APIToken())
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		// 限制响应体日志长度，避免打印大量数据
		responsePreview := string(body)
		if len(responsePreview) > 500 {
			responsePreview = responsePreview[:500] + "... (响应内容被截断)"
		}
		return nil, fmt.Errorf("D1 API请求失败，状态码: %d, 响应: %s", resp.StatusCode, responsePreview)
	}

	var d1Resp D1Response
	if err := json.Unmarshal(body, &d1Resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !d1Resp.Success {
		if len(d1Resp.Errors) > 0 {
			return nil, fmt.Errorf("D1错误: %s", d1Resp.Errors[0].Message)
		}
		return nil, fmt.Errorf("D1请求失败")
	}

	if len(d1Resp.Result) > 0 {
		return d1Resp.Result[0].Results, nil
	}

	return []map[string]interface{}{}, nil
}

// ===== 图片编辑记录相关方法 =====

// SaveEditRecord 保存图片编辑记录
func (d *D1Service) SaveEditRecord(record *models.ImageEditRecord) error {
	// 生成任务ID
	if record.TaskID == "" {
		record.TaskID = fmt.Sprintf("edit_%d", time.Now().UnixNano())
	}

	sql := `
		INSERT INTO image_edit_records
		(original_image_id, edit_type, edit_prompt, input_image_urls, status, task_id, created_at, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	params := []interface{}{
		record.OriginalImageID,
		string(record.EditType),
		record.EditPrompt,
		record.InputImageURLs,
		string(record.Status),
		record.TaskID,
		record.CreatedAt,
		record.ErrorMessage,
	}

	if err := d.executeSQLWithParams(sql, params); err != nil {
		return fmt.Errorf("保存编辑记录失败: %v", err)
	}

	log.Printf("编辑记录已保存到D1: TaskID=%s, Type=%s", record.TaskID, record.EditType)
	return nil
}

// UpdateEditRecord 更新图片编辑记录
func (d *D1Service) UpdateEditRecord(record *models.ImageEditRecord) error {
	sql := `
		UPDATE image_edit_records
		SET status = ?, result_image_url = ?, local_path = ?, r2_url = ?,
			file_size = ?, width = ?, height = ?, format = ?,
			error_message = ?, completed_at = ?
		WHERE task_id = ?
	`

	params := []interface{}{
		string(record.Status),
		record.ResultImageURL,
		record.LocalPath,
		record.R2URL,
		record.FileSize,
		record.Width,
		record.Height,
		record.Format,
		record.ErrorMessage,
		record.CompletedAt,
		record.TaskID,
	}

	if err := d.executeSQLWithParams(sql, params); err != nil {
		return fmt.Errorf("更新编辑记录失败: %v", err)
	}

	return nil
}

// GetEditRecordByTaskID 根据任务ID获取编辑记录
func (d *D1Service) GetEditRecordByTaskID(taskID string) (*models.EditTaskStatusResponse, error) {
	sql := "SELECT * FROM image_edit_records WHERE task_id = ?"
	results, err := d.querySQLWithParams(sql, []interface{}{taskID})
	if err != nil {
		return nil, fmt.Errorf("查询编辑记录失败: %v", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("编辑记录不存在")
	}

	result := results[0]
	response := &models.EditTaskStatusResponse{}

	if id, ok := result["id"].(float64); ok {
		response.ID = int(id)
	}
	if val, ok := result["task_id"].(string); ok {
		response.TaskID = val
	}
	if val, ok := result["status"].(string); ok {
		response.Status = models.EditStatus(val)
	}
	if val, ok := result["edit_type"].(string); ok {
		response.EditType = models.EditType(val)
	}
	if val, ok := result["edit_prompt"].(string); ok {
		response.EditPrompt = val
	}
	if val, ok := result["result_image_url"].(string); ok {
		response.ResultImageURL = val
	}
	if val, ok := result["r2_url"].(string); ok {
		response.R2URL = val
	}
	if val, ok := result["error_message"].(string); ok {
		response.ErrorMessage = val
	}
	if val, ok := result["created_at"].(string); ok {
		response.CreatedAt = val
	}
	if val, ok := result["completed_at"].(string); ok {
		response.CompletedAt = val
	}

	// 计算进度百分比
	switch response.Status {
	case models.EditStatusPending:
		response.Progress = 0
	case models.EditStatusProcessing:
		response.Progress = 50
	case models.EditStatusCompleted:
		response.Progress = 100
	case models.EditStatusFailed:
		response.Progress = -1
	}

	return response, nil
}

// GetEditRecords 获取编辑记录列表
func (d *D1Service) GetEditRecords(req *models.EditRecordsRequest) (*models.EditRecordsResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// 构建WHERE条件
	var conditions []string
	var params []interface{}

	if req.EditType != "" {
		conditions = append(conditions, "edit_type = ?")
		params = append(params, string(req.EditType))
	}

	if req.Status != "" {
		conditions = append(conditions, "status = ?")
		params = append(params, string(req.Status))
	}

	if req.Keyword != "" {
		conditions = append(conditions, "edit_prompt LIKE ?")
		keyword := "%" + req.Keyword + "%"
		params = append(params, keyword)
	}

	if req.DateFrom != "" {
		conditions = append(conditions, "created_at >= ?")
		params = append(params, req.DateFrom)
	}

	if req.DateTo != "" {
		conditions = append(conditions, "created_at <= ?")
		params = append(params, req.DateTo+" 23:59:59")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) as total FROM image_edit_records %s", whereClause)
	countResult, err := d.querySQLWithParams(countSQL, params)
	if err != nil {
		return nil, fmt.Errorf("查询总数失败: %v", err)
	}

	total := 0
	if len(countResult) > 0 {
		if totalVal, ok := countResult[0]["total"].(float64); ok {
			total = int(totalVal)
		}
	}

	// 查询数据
	offset := (req.Page - 1) * req.Limit
	dataSQL := fmt.Sprintf(`
		SELECT * FROM image_edit_records %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	dataParams := append(params, req.Limit, offset)
	results, err := d.querySQLWithParams(dataSQL, dataParams)
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %v", err)
	}

	// 转换结果
	var records []models.ImageEditRecord
	for _, result := range results {
		record := models.ImageEditRecord{}
		if id, ok := result["id"].(float64); ok {
			record.ID = int(id)
		}
		if val, ok := result["original_image_id"].(float64); ok && val != 0 {
			id := int(val)
			record.OriginalImageID = &id
		}
		if val, ok := result["edit_type"].(string); ok {
			record.EditType = models.EditType(val)
		}
		if val, ok := result["edit_prompt"].(string); ok {
			record.EditPrompt = val
		}
		if val, ok := result["input_image_urls"].(string); ok {
			record.InputImageURLs = val
		}
		if val, ok := result["result_image_url"].(string); ok {
			record.ResultImageURL = val
		}
		if val, ok := result["local_path"].(string); ok {
			record.LocalPath = val
		}
		if val, ok := result["r2_url"].(string); ok {
			record.R2URL = val
		}
		if val, ok := result["status"].(string); ok {
			record.Status = models.EditStatus(val)
		}
		if val, ok := result["error_message"].(string); ok {
			record.ErrorMessage = val
		}
		if val, ok := result["task_id"].(string); ok {
			record.TaskID = val
		}
		if val, ok := result["file_size"].(float64); ok {
			record.FileSize = int64(val)
		}
		if val, ok := result["width"].(float64); ok {
			record.Width = int(val)
		}
		if val, ok := result["height"].(float64); ok {
			record.Height = int(val)
		}
		if val, ok := result["format"].(string); ok {
			record.Format = val
		}
		if val, ok := result["created_at"].(string); ok {
			record.CreatedAt = val
		}
		if val, ok := result["completed_at"].(string); ok {
			record.CompletedAt = val
		}

		records = append(records, record)
	}

	pages := (total + req.Limit - 1) / req.Limit

	return &models.EditRecordsResponse{
		Records: records,
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
		Pages:   pages,
	}, nil
}

// initEditDatabase 初始化编辑相关的数据库表
func (d *D1Service) initEditDatabase() error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS image_edit_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			original_image_id INTEGER,
			edit_type TEXT NOT NULL,
			edit_prompt TEXT NOT NULL,
			input_image_urls TEXT NOT NULL,
			result_image_url TEXT,
			local_path TEXT,
			r2_url TEXT,
			status TEXT DEFAULT 'pending',
			error_message TEXT,
			task_id TEXT UNIQUE,
			file_size INTEGER DEFAULT 0,
			width INTEGER DEFAULT 0,
			height INTEGER DEFAULT 0,
			format TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (original_image_id) REFERENCES image_records(id)
		);
	`

	// 执行建表语句
	if err := d.executeSQL(createTableSQL); err != nil {
		return fmt.Errorf("创建编辑表失败: %v", err)
	}

	// 创建索引
	indexSQL := `
		CREATE INDEX IF NOT EXISTS idx_edit_task_id ON image_edit_records(task_id);
		CREATE INDEX IF NOT EXISTS idx_edit_status ON image_edit_records(status);
		CREATE INDEX IF NOT EXISTS idx_edit_type ON image_edit_records(edit_type);
		CREATE INDEX IF NOT EXISTS idx_edit_created_at ON image_edit_records(created_at);
	`

	if err := d.executeSQL(indexSQL); err != nil {
		log.Printf("警告: 创建编辑表索引失败: %v", err)
	}

	log.Println("D1编辑数据库表初始化完成")
	return nil
}
