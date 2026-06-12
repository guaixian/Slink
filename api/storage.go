package api

import (
	"Slink/model"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// StorageResponse 存储响应结构
type StorageResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// StrategyResponse 策略响应结构（不包含时间字段）。Roles 保留字段名以兼容前端，固定为空：策略全局可用。
type StrategyResponse struct {
	ID           uint                   `json:"id"`
	Name         string                 `json:"name"`
	Introduction string                 `json:"introduction"`
	Key          string                 `json:"key"`
	Configs      map[string]interface{} `json:"configs"`
	Roles        []uint                 `json:"roles"`
}

// StrategyRequest 策略请求结构（支持JSON对象格式的configs）
type StrategyRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Introduction string                 `json:"introduction"`
	Key          string                 `json:"key" binding:"required"`
	Configs      map[string]interface{} `json:"configs" binding:"required"`
}

// Storage 获取存储信息
func Storage(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取用户信息
	user, err := model.GetUserByID(model.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 获取用户的所有图片
	images, err := model.GetImagesByUserID(model.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取图片列表失败"})
		return
	}

	// 计算总存储大小
	var totalSize int64
	for _, img := range images {
		totalSize += img.Size
	}

	// 获取存储策略
	strategies, err := model.GetAllStrategies(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取存储策略失败"})
		return
	}

	// 构建不包含时间字段的策略响应
	var strategyResponses []StrategyResponse
	for _, strategy := range strategies {
		// 将字符串格式的configs转换为JSON对象
		var configs map[string]interface{}
		if err := json.Unmarshal([]byte(strategy.Configs), &configs); err != nil {
			// 如果解析失败，使用空对象
			configs = make(map[string]interface{})
		}

		strategyResponse := StrategyResponse{
			ID:           strategy.ID,
			Name:         strategy.Name,
			Introduction: strategy.Introduction,
			Key:          strategy.StrategyKey,
			Configs:      configs,
			Roles:        []uint{},
		}
		strategyResponses = append(strategyResponses, strategyResponse)
	}

	// 构建存储信息
	storageInfo := map[string]interface{}{
		"user_id":      user.ID,
		"user_name":    user.Name,
		"total_images": len(images),
		"total_size":   totalSize,
		"used_size_mb": float64(totalSize) / (1024 * 1024),
		"strategies":   strategyResponses,
	}

	response := &StorageResponse{
		Status:  true,
		Message: "获取成功",
		Data:    storageInfo,
	}

	c.JSON(http.StatusOK, response)
}

// GetStrategies 获取所有存储策略

func GetStrategyByID(c *gin.Context) {
	// 从URL参数获取策略ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的策略ID"})
		return
	}

	// 根据ID获取策略
	strategy, err := model.GetStrategyByID(model.DB, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "策略不存在"})
		return
	}

	// 将字符串格式的configs转换为JSON对象
	var configs map[string]interface{}
	if err := json.Unmarshal([]byte(strategy.Configs), &configs); err != nil {
		// 如果解析失败，使用空对象
		configs = make(map[string]interface{})
	}

	// 构建不包含时间字段的响应
	strategyResponse := &StrategyResponse{
		ID:           strategy.ID,
		Name:         strategy.Name,
		Introduction: strategy.Introduction,
		Key:          strategy.StrategyKey,
		Configs:      configs,
		Roles:        []uint{},
	}

	response := &StorageResponse{
		Status:  true,
		Message: "获取成功",
		Data:    strategyResponse,
	}

	c.JSON(http.StatusOK, response)
}

func GetStrategies(c *gin.Context) {
	strategies, err := model.GetStrategyStats(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取存储策略失败"})
		return
	}

	response := &StorageResponse{
		Status:  true,
		Message: "获取成功",
		Data:    strategies,
	}

	c.JSON(http.StatusOK, response)
}

// CreateStrategy 创建存储策略
func CreateStrategy(c *gin.Context) {
	var request StrategyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 将configs对象转换为JSON字符串
	configsJSON, err := json.Marshal(request.Configs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置格式错误"})
		return
	}

	// 创建策略对象
	strategy := model.Strategies{
		Name:         request.Name,
		Introduction: request.Introduction,
		StrategyKey:  request.Key,
		Configs:      string(configsJSON),
	}

	if err := model.CreateStrategy(model.DB, &strategy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建策略失败"})
		return
	}

	// 构建不包含时间字段的响应
	strategyResponse := &StrategyResponse{
		ID:           strategy.ID,
		Name:         strategy.Name,
		Introduction: strategy.Introduction,
		Key:          strategy.StrategyKey,
		Configs:      request.Configs,
	}

	response := &StorageResponse{
		Status:  true,
		Message: "创建成功",
		Data:    strategyResponse,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateStrategy 更新存储策略
func UpdateStrategy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的策略ID"})
		return
	}

	var request StrategyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 将configs对象转换为JSON字符串
	configsJSON, err := json.Marshal(request.Configs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置格式错误"})
		return
	}

	// 创建策略对象
	strategy := model.Strategies{
		ID:           uint(id),
		Name:         request.Name,
		Introduction: request.Introduction,
		StrategyKey:  request.Key,
		Configs:      string(configsJSON),
	}

	if err := model.UpdateStrategy(model.DB, &strategy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新策略失败"})
		return
	}

	// 构建不包含时间字段的响应
	strategyResponse := &StrategyResponse{
		ID:           strategy.ID,
		Name:         strategy.Name,
		Introduction: strategy.Introduction,
		Key:          strategy.StrategyKey,
		Configs:      request.Configs,
	}

	response := &StorageResponse{
		Status:  true,
		Message: "更新成功",
		Data:    strategyResponse,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteStrategy 删除存储策略
func DeleteStrategy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的策略ID"})
		return
	}

	// 检查是否为默认策略
	strategy, err := model.GetStrategyByID(model.DB, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "策略不存在"})
		return
	}

	if strategy.StrategyKey == "1" {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能删除默认策略"})
		return
	}

	if err := model.DeleteStrategy(model.DB, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除策略失败"})
		return
	}

	if err := model.DeleteGroupStrategiesByStrategyID(model.DB, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除策略关联失败"})
		return
	}

	response := &StorageResponse{
		Status:  true,
		Message: "删除成功",
		Data:    nil,
	}

	c.JSON(http.StatusOK, response)
}
