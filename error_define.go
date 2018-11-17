// 错误定义

package apisign

import (
	"github.com/go-apibox/api"
)

// error type
const (
	errorAppNotExist = iota
	errorMissingSign
	errorSignError
	errorSignExpired
	errorMissingTimestamp
	errorInvalidTimestamp
	errorMissingAppId
	errorInvalidAppId
	errorAppStatusError
)

var ErrorDefines = map[api.ErrorType]*api.ErrorDefine{
	errorAppNotExist: api.NewErrorDefine(
		"AppNotExist",
		[]int{0},
		map[string]map[int]string{
			"en_us": {
				0: "Application does not exist!",
			},
			"zh_cn": {
				0: "应用不存在！",
			},
		},
	),
	errorMissingSign: api.NewErrorDefine(
		"MissingSign",
		[]int{0},
		map[string]map[int]string{
			"en_us": {
				0: "Missing signature!",
			},
			"zh_cn": {
				0: "缺少签名！",
			},
		},
	),
	errorSignError: api.NewErrorDefine(
		"SignError",
		[]int{0},
		map[string]map[int]string{
			"en_us": {
				0: "Signature check failed!",
			},
			"zh_cn": {
				0: "签名校验失败！",
			},
		},
	),
	errorSignExpired: api.NewErrorDefine(
		"SignExpired",
		[]int{0},
		map[string]map[int]string{
			"en_us": {
				0: "Request has expired!",
			},
			"zh_cn": {
				0: "请求已超时！",
			},
		},
	),
	errorMissingTimestamp: api.NewErrorDefine(
		"MissingTimestamp",
		[]int{0},
		map[string]map[int]string{
			"en_us": {
				0: "Missing timestamp!",
			},
			"zh_cn": {
				0: "缺少时间戳！",
			},
		},
	),
	errorInvalidTimestamp: api.NewErrorDefine(
		"InvalidTimestamp",
		[]int{0},
		map[string]map[int]string{
			"en_us": {
				0: "Timestamp is invalid!",
			},
			"zh_cn": {
				0: "时间戳不合法！",
			},
		},
	),
	errorMissingAppId: api.NewErrorDefine(
		"MissingAppId",
		[]int{0},
		map[string]map[int]string{
			"en_us": {
				0: "Missing app id!",
			},
			"zh_cn": {
				0: "缺少应用ID！",
			},
		},
	),
	errorInvalidAppId: api.NewErrorDefine(
		"InvalidAppId",
		[]int{0},
		map[string]map[int]string{
			"en_us": {
				0: "App id is invalid!",
			},
			"zh_cn": {
				0: "应用ID不合法！",
			},
		},
	),
	errorAppStatusError: api.NewErrorDefine(
		"AppStatusError",
		[]int{0},
		map[string]map[int]string{
			"en_us": {
				0: "App status error!",
			},
			"zh_cn": {
				0: "应用状态错误！",
			},
		},
	),
}
