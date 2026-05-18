// Package location 提供 652 签到所需的校区地理位置库.
// 数据来源于内嵌 JSON 文件,在 init 时解析为内存数组,运行时随机选取.
package location

import (
	_ "embed"
	"encoding/json"
	"math/rand"
	"time"
)

//go:embed 宜宾.json
var rawYibin []byte

//go:embed 李白河.json
var rawLibaihe []byte

//go:embed 汇东.json
var rawHuidong []byte

type SignInLocation struct {
	Qdddjtdz string         `json:"qdddjtdz"` // 签到地点具体地址文本
	Location LocationCoords `json:"location"` // 经纬度 + 地址详情
}

// LocationCoords 是 location 字段对应的 JSON 对象.
type LocationCoords struct {
	Point   [2]float64 `json:"point"`   // [longitude, latitude]
	Address string     `json:"address"` // 地址文本
}

// ToPayload 将地址数据转换为 652 签到接口所需的 map 格式.
// location 字段会被序列化为 JSON 字符串(后端要求 string 类型).
func (l SignInLocation) ToPayload() map[string]any {
	locBytes, _ := json.Marshal(l.Location)
	return map[string]any{
		"qdddjtdz": l.Qdddjtdz,
		"location": string(locBytes),
	}
}

// campusDB 校区 → 地址数组的内存数据库.
var campusDB map[string][]SignInLocation

// rnd 私有随机源,避免并发冲突和全局种子污染.
var rnd *rand.Rand

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

	campusDB = make(map[string][]SignInLocation, 3)

	var yibin, libaihe, huidong []SignInLocation
	_ = json.Unmarshal(rawYibin, &yibin)
	_ = json.Unmarshal(rawLibaihe, &libaihe)
	_ = json.Unmarshal(rawHuidong, &huidong)

	campusDB["宜宾"] = yibin
	campusDB["李白河"] = libaihe
	campusDB["汇东"] = huidong
}

func GetRandom(campus string) (SignInLocation, bool) {
	list, ok := campusDB[campus]
	if !ok || len(list) == 0 {
		return SignInLocation{}, false
	}
	return list[rnd.Intn(len(list))], true
}
