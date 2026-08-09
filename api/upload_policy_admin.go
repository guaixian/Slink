package api

import (
	"Slink/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ---- 上传策略组管理 ----

// GetUploadPolicyGroups 获取所有上传策略组列表
func GetUploadPolicyGroups(c *gin.Context) {
	groups, err := model.GetAllUploadPolicyGroups(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "读取策略组失败"})
		return
	}
	// 统计每个策略组下的用户数
	type GroupWithStats struct {
		model.UploadPolicyGroup
		UserCount int64 `json:"user_count"`
	}
	result := make([]GroupWithStats, 0, len(groups))
	for _, g := range groups {
		var count int64
		model.DB.Model(&model.User{}).Where("policy_group_id = ?", g.ID).Count(&count)
		result = append(result, GroupWithStats{UploadPolicyGroup: g, UserCount: count})
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "data": result})
}

// GetUploadPolicyGroup 获取单个上传策略组详情
func GetUploadPolicyGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "无效的ID"})
		return
	}
	g, err := model.GetUploadPolicyGroupByID(model.DB, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "策略组不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "data": g.ToGroupConfig()})
}

// CreateUploadPolicyGroup 创建上传策略组
func CreateUploadPolicyGroup(c *gin.Context) {
	var body struct {
		Name        string             `json:"name" binding:"required"`
		Description string             `json:"description"`
		Config      model.GroupConfig  `json:"config"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "参数错误"})
		return
	}
	model.NormalizeWatermarkConfig(body.Config.WatermarkConfigs)
	g := &model.UploadPolicyGroup{
		Name:        body.Name,
		Description: body.Description,
	}
	g.SetGroupConfig(&body.Config)
	if err := model.CreateUploadPolicyGroup(model.DB, g); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "创建成功", "data": g})
}

// UpdateUploadPolicyGroup 更新上传策略组
func UpdateUploadPolicyGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "无效的ID"})
		return
	}
	existing, err := model.GetUploadPolicyGroupByID(model.DB, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "策略组不存在"})
		return
	}
	var body struct {
		Name        string             `json:"name"`
		Description string             `json:"description"`
		Config      *model.GroupConfig `json:"config"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "参数错误"})
		return
	}
	if body.Name != "" {
		existing.Name = body.Name
	}
	existing.Description = body.Description
	if body.Config != nil {
		model.NormalizeWatermarkConfig(body.Config.WatermarkConfigs)
		existing.SetGroupConfig(body.Config)
	}
	if err := model.UpdateUploadPolicyGroup(model.DB, existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "保存失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "保存成功"})
}

// DeleteUploadPolicyGroup 删除上传策略组
func DeleteUploadPolicyGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "无效的ID"})
		return
	}
	g, err := model.GetUploadPolicyGroupByID(model.DB, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "策略组不存在"})
		return
	}
	if g.IsDefault == 1 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "不能删除默认策略组"})
		return
	}
	// 将该策略组下的用户移回默认策略组
	defaultGroup, _ := model.GetDefaultUploadPolicyGroup(model.DB)
	if defaultGroup != nil {
		model.DB.Model(&model.User{}).Where("policy_group_id = ?", g.ID).Update("policy_group_id", defaultGroup.ID)
	}
	if err := model.DeleteUploadPolicyGroup(model.DB, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "删除成功"})
}

// ---- 用户上传策略组分配 ----

// SetDefaultPolicyGroup 设置默认上传策略组
func SetDefaultPolicyGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "无效的ID"})
		return
	}
	g, err := model.GetUploadPolicyGroupByID(model.DB, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "策略组不存在"})
		return
	}
	// 先将所有策略组设为非默认
	model.DB.Model(&model.UploadPolicyGroup{}).Where("is_default = ?", 1).Update("is_default", 0)
	// 设置当前策略组为默认
	g.IsDefault = 1
	model.DB.Save(g)
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "已设为默认策略组"})
}

// SetUserPolicyGroup 设置用户的上传策略组
func SetUserPolicyGroup(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "无效的用户ID"})
		return
	}
	var body struct {
		PolicyGroupID uint `json:"policy_group_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "参数错误"})
		return
	}
	// 验证策略组存在
	if _, err := model.GetUploadPolicyGroupByID(model.DB, body.PolicyGroupID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "策略组不存在"})
		return
	}
	model.DB.Model(&model.User{}).Where("id = ?", userID).Update("policy_group_id", body.PolicyGroupID)
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "设置成功"})
}

// ---- 兼容旧接口 ----

// GetUploadPolicy 获取当前用户的上传策略（兼容）
func GetUploadPolicy(c *gin.Context) {
	userID, _ := c.Get("userID")
	if userID != nil {
		pg, err := model.GetUserUploadPolicyGroup(model.DB, userID.(uint))
		if err == nil && pg != nil {
			c.JSON(http.StatusOK, gin.H{"status": true, "message": "获取成功", "data": pg.ToGroupConfig()})
			return
		}
	}
	pol, err := model.GetGlobalUploadPolicy(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "读取上传策略失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "获取成功", "data": pol})
}

// UpdateUploadPolicy 更新默认策略组的上传策略（兼容）
func UpdateUploadPolicy(c *gin.Context) {
	var body model.GroupConfig
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "请求参数错误"})
		return
	}
	model.NormalizeWatermarkConfig(body.WatermarkConfigs)
	defaultGroup, err := model.GetDefaultUploadPolicyGroup(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "默认策略组不存在"})
		return
	}
	defaultGroup.SetGroupConfig(&body)
	if err := model.UpdateUploadPolicyGroup(model.DB, defaultGroup); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "保存失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "保存成功"})
}
