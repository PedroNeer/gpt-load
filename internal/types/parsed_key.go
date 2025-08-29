package types

// ParsedKey 解析后的密钥信息
type ParsedKey struct {
	RawKey    string            // 原始密钥字符串
	ActualKey string            // 解析后的实际密钥
	Params    map[string]string // URL参数
}

// GetParam 获取指定参数的值
func (pk *ParsedKey) GetParam(name string) string {
	if pk.Params == nil {
		return ""
	}
	return pk.Params[name]
}

// HasParam 检查是否包含指定参数
func (pk *ParsedKey) HasParam(name string) bool {
	if pk.Params == nil {
		return false
	}
	_, exists := pk.Params[name]
	return exists
}
