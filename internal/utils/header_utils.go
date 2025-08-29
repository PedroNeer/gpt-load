package utils

import (
	"gpt-load/internal/models"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// HeaderVariableContext holds context data for variable resolution
type HeaderVariableContext struct {
	ClientIP string
	Group    *models.Group
	APIKey   *models.APIKey
}

// ResolveHeaderVariables resolves dynamic variables in header values
func ResolveHeaderVariables(value string, ctx *HeaderVariableContext) string {
	if ctx == nil {
		return value
	}

	now := time.Now()
	result := value

	// 首先处理新的变量语法：${API_KEY}{param_name}
	result = resolveAPIKeyParams(result, ctx)

	// Replace all supported variables
	variables := map[string]string{
		"${CLIENT_IP}":    ctx.ClientIP,
		"${TIMESTAMP_MS}": strconv.FormatInt(now.UnixMilli(), 10),
		"${TIMESTAMP_S}":  strconv.FormatInt(now.Unix(), 10),
	}

	if ctx.Group != nil {
		variables["${GROUP_NAME}"] = ctx.Group.Name
	}

	if ctx.APIKey != nil {
		variables["${API_KEY}"] = ctx.APIKey.KeyValue
	}

	// Replace variables in the value
	for variable, replacement := range variables {
		result = strings.ReplaceAll(result, variable, replacement)
	}

	return result
}

// resolveAPIKeyParams 解析 ${API_KEY}{param_name} 语法
func resolveAPIKeyParams(value string, ctx *HeaderVariableContext) string {
	if ctx == nil || ctx.APIKey == nil || ctx.APIKey.ParsedKey == nil {
		return value
	}

	// 使用正则表达式匹配 ${API_KEY}{param_name} 模式
	re := regexp.MustCompile(`\$\{API_KEY\}\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(value, func(match string) string {
		matches := re.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match // 如果没有匹配到参数名，返回原字符串
		}

		paramName := matches[1]
		paramValue := ctx.APIKey.ParsedKey.GetParam(paramName)

		// 如果参数不存在，返回空字符串
		if paramValue == "" {
			return ""
		}

		return paramValue
	})
}

// ApplyHeaderRules applies header rules to the HTTP request
func ApplyHeaderRules(req *http.Request, rules []models.HeaderRule, ctx *HeaderVariableContext) {
	if req == nil || len(rules) == 0 {
		return
	}

	for _, rule := range rules {
		canonicalKey := http.CanonicalHeaderKey(rule.Key)

		switch rule.Action {
		case "remove":
			req.Header.Del(canonicalKey)
		case "set":
			resolvedValue := ResolveHeaderVariables(rule.Value, ctx)
			req.Header.Set(canonicalKey, resolvedValue)
		}
	}
}

// NewHeaderVariableContextFromGin creates HeaderVariableContext from Gin context
func NewHeaderVariableContextFromGin(c *gin.Context, group *models.Group, apiKey *models.APIKey) *HeaderVariableContext {
	if c == nil {
		return nil
	}

	return &HeaderVariableContext{
		ClientIP: c.ClientIP(),
		Group:    group,
		APIKey:   apiKey,
	}
}

// NewHeaderVariableContext creates HeaderVariableContext without Gin context
func NewHeaderVariableContext(group *models.Group, apiKey *models.APIKey) *HeaderVariableContext {
	return &HeaderVariableContext{
		ClientIP: "127.0.0.1",
		Group:    group,
		APIKey:   apiKey,
	}
}
