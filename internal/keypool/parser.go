package keypool

import (
	"fmt"
	"gpt-load/internal/types"
	"net/url"
	"strings"
)

// KeyParser 密钥解析器
type KeyParser struct {
	method string
}

// NewKeyParser 创建新的密钥解析器
func NewKeyParser(method string) *KeyParser {
	return &KeyParser{
		method: method,
	}
}

// ParseKey 解析密钥
func (p *KeyParser) ParseKey(rawKey string) (*types.ParsedKey, error) {
	switch p.method {
	case "urlencode":
		return p.parseURLEncodedKey(rawKey)
	case "none", "":
		fallthrough
	default:
		return &types.ParsedKey{
			RawKey:    rawKey,
			ActualKey: rawKey,
			Params:    make(map[string]string),
		}, nil
	}
}

// parseURLEncodedKey 解析URL编码的密钥
func (p *KeyParser) parseURLEncodedKey(rawKey string) (*types.ParsedKey, error) {
	// 检查是否包含URL编码字符
	if !strings.Contains(rawKey, "=") && !strings.Contains(rawKey, "&") {
		// 不包含URL编码字符，直接返回原始密钥
		return &types.ParsedKey{
			RawKey:    rawKey,
			ActualKey: rawKey,
			Params:    make(map[string]string),
		}, nil
	}

	// 尝试解析为URL查询字符串
	// 首先检查是否包含key参数
	if strings.Contains(rawKey, "key=") {
		// 如果包含key=，则按URL查询字符串解析
		// 直接使用原始字符串，不需要添加前缀
		values, err := url.ParseQuery(rawKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL encoded key: %w", err)
		}

		// 提取key参数作为实际密钥
		keyValues := values["key"]
		if len(keyValues) == 0 {
			return nil, fmt.Errorf("no 'key' parameter found in URL encoded string")
		}
		actualKey := keyValues[0]

		// 构建参数字典
		params := make(map[string]string)
		for k, v := range values {
			if k != "key" && len(v) > 0 {
				params[k] = v[0]
			}
		}

		return &types.ParsedKey{
			RawKey:    rawKey,
			ActualKey: actualKey,
			Params:    params,
		}, nil
	}

	// 如果没有key参数，尝试解析为key=value&key2=value2格式
	// 假设第一个参数是密钥
	parts := strings.Split(rawKey, "&")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid URL encoded key format")
	}

	// 解析第一个参数作为密钥
	firstPart := parts[0]
	if !strings.Contains(firstPart, "=") {
		// 如果没有=，则整个字符串就是密钥
		return &types.ParsedKey{
			RawKey:    rawKey,
			ActualKey: rawKey,
			Params:    make(map[string]string),
		}, nil
	}

	// 解析所有参数
	params := make(map[string]string)
	var actualKey string

	for i, part := range parts {
		if !strings.Contains(part, "=") {
			continue
		}

		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) != 2 {
			continue
		}

		key := strings.TrimSpace(keyValue[0])
		value := strings.TrimSpace(keyValue[1])

		if i == 0 {
			// 第一个参数作为实际密钥
			actualKey = value
		} else {
			// 其他参数作为附加参数
			params[key] = value
		}
	}

	if actualKey == "" {
		return nil, fmt.Errorf("failed to extract actual key from URL encoded string")
	}

	return &types.ParsedKey{
		RawKey:    rawKey,
		ActualKey: actualKey,
		Params:    params,
	}, nil
}
