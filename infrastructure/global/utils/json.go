package utils

import "encoding/json"

// JsonUnmarshal 将json数据解析到v中，忽略解析错误
func JsonUnmarshal(data []byte, v interface{}) {
	e := json.Unmarshal(data, v)
	if e != nil {
		return
	}
}

// JsonMarshal 将v转换为json数据，忽略转换错误
func JsonMarshal(v interface{}) []byte {
	bts, marshalErr := json.Marshal(v)
	if marshalErr != nil {
		return nil
	}
	return bts
}
