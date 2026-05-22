package modelrouter

import (
	"encoding/json"
	"time"
)

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getFloat(m map[string]interface{}, key string) (float64, bool) {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return n, true
		case int:
			return float64(n), true
		}
	}
	return 0, false
}

func getInt(m map[string]interface{}, key string) (int, bool) {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n), true
		case int:
			return n, true
		case int64:
			return int(n), true
		}
	}
	return 0, false
}

func getIntDefault(m map[string]interface{}, key string) int {
	if v, ok := getInt(m, key); ok {
		return v
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return int64(n)
		case int64:
			return n
		case int:
			return int64(n)
		}
	}
	return 0
}

func getList(m map[string]interface{}, key string) ([]interface{}, bool) {
	if v, ok := m[key]; ok {
		if l, ok := v.([]interface{}); ok {
			return l, true
		}
	}
	return nil, false
}

func getStringList(m map[string]interface{}, key string) ([]string, bool) {
	if v, ok := m[key]; ok {
		switch l := v.(type) {
		case []interface{}:
			var ss []string
			for _, item := range l {
				if s, ok := item.(string); ok {
					ss = append(ss, s)
				}
			}
			return ss, true
		case []string:
			return l, true
		}
	}
	return nil, false
}

func contentToString(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var s string
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if t := getString(m, "type"); t == "text" {
					s += getString(m, "text")
				}
			}
		}
		return s
	default:
		b, _ := marshalJSON(content)
		return string(b)
	}
}

func parseJSONString(s string, target interface{}) {
	_ = json.Unmarshal([]byte(s), target)
}

func marshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func now() int64 {
	return time.Now().Unix()
}

func maybeNow(t int64) int64 {
	if t > 0 {
		return t
	}
	return now()
}
