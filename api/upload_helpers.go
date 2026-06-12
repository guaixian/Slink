package api

import (
	"Slink/model"
	"Slink/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// resolveUploadStrategy 解析本次上传使用的存储策略：表单/JSON 的 strategy_id 优先，否则用户偏好中的 default_strategy，否则第一个已配置策略
func resolveUploadStrategy(db *gorm.DB, userConfig model.UserConfig, strategyIDRaw string) (*model.Strategies, *model.StrategyConfigData, error) {
	sid := userConfig.DefaultStrategy
	if strategyIDRaw != "" {
		if v, err := strconv.ParseUint(strategyIDRaw, 10, 32); err == nil && v > 0 {
			sid = int(v)
		}
	}
	all, err := model.GetAllStrategies(db)
	if err != nil {
		return nil, nil, err
	}
	if len(all) == 0 {
		return nil, nil, fmt.Errorf("未配置存储策略")
	}
	var st *model.Strategies
	for i := range all {
		if int(all[i].ID) == sid {
			st = &all[i]
			break
		}
	}
	if st == nil {
		st = &all[0]
	}
	scd, err := model.GetStrategyConfigData(st)
	if err != nil {
		return nil, nil, err
	}
	return st, scd, nil
}

func strategyUtilsConfigFromModel(strategy *model.Strategies) (*utils.StrategyConfig, error) {
	sc, err := model.GetStrategyConfig(strategy)
	if err != nil {
		return nil, err
	}
	baseURL := strings.TrimSpace(sc.URL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(sc.Domain)
	}
	return &utils.StrategyConfig{
		URL:      baseURL,
		Root:     sc.Root,
		Queries:  sc.Queries,
		AuthType: sc.AuthType,
	}, nil
}

func strategyUtilsConfigForStoredImage(db *gorm.DB, img *model.Images) (*utils.StrategyConfig, error) {
	st, err := model.GetStrategyForStoredImage(db, img.StrategyID)
	if err != nil {
		return nil, err
	}
	return strategyUtilsConfigFromModel(st)
}

func normalizePathnameForLinks(path string) string {
	pathname := path
	if len(pathname) > 7 && pathname[:7] == "static/" {
		pathname = pathname[7:]
	} else if len(pathname) > 8 && pathname[:8] == `static\` {
		pathname = pathname[8:]
	}
	return pathname
}

func readGlobalUploadPolicy(c *gin.Context) (*model.GroupConfig, bool) {
	pol, err := model.GetGlobalUploadPolicy(model.DB)
	if err != nil {
		c.JSON(500, gin.H{"error": "读取全局上传策略失败"})
		return nil, false
	}
	return pol, true
}
