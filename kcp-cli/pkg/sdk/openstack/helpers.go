package openstack

import (
	"encoding/json"
	"fmt"
)

// extractList 는 JSON 응답에서 지정된 키의 배열을 추출한다.
// 예: {"servers": [...]} 에서 key="servers"로 배열을 추출
func extractList(data []byte, key string) ([]json.RawMessage, error) {
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("응답 JSON 파싱 실패: %w", err)
	}

	raw, ok := wrapper[key]
	if !ok {
		return nil, fmt.Errorf("응답에 '%s' 키가 없습니다", key)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("'%s' 배열 파싱 실패: %w", key, err)
	}

	return items, nil
}

// extractSingle 은 JSON 응답에서 지정된 키의 단일 객체를 추출한다.
// 예: {"server": {...}} 에서 key="server"로 객체를 추출
func extractSingle(data []byte, key string) (json.RawMessage, error) {
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("응답 JSON 파싱 실패: %w", err)
	}

	raw, ok := wrapper[key]
	if !ok {
		return nil, fmt.Errorf("응답에 '%s' 키가 없습니다", key)
	}

	return raw, nil
}
