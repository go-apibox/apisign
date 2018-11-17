package apisign

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-apibox/api"
	"github.com/go-apibox/utils"
	"github.com/go-xorm/xorm"
)

type Sign struct {
	app      *api.App
	disabled bool
	inited   bool

	signKey         string
	expireTime      int
	allowTimeOffset int
	actionMatcher   *utils.Matcher

	appEnabled         bool
	appDb              *xorm.Engine
	appTable           string
	appIdColumn        string
	appIdType          string
	appSignKeyColumn   string
	appStatusColumn    string
	appAdminAppEnabled bool
	appAdminAppId      string
	appAdminSignKey    string
}

func NewSign(app *api.App) *Sign {
	app.Error.RegisterGroupErrors("sign", ErrorDefines)

	sign := new(Sign)
	sign.app = app

	cfg := app.Config
	disabled := cfg.GetDefaultBool("apisign.disabled", false)
	sign.disabled = disabled
	if disabled {
		return sign
	}

	sign.init()
	return sign
}

func (s *Sign) init() {
	if s.inited {
		return
	}

	app := s.app
	cfg := app.Config
	signKey := cfg.GetDefaultString("apisign.sign_key", "")
	expireTime := cfg.GetDefaultInt("apisign.expire_time", 600)
	allowTimeOffset := cfg.GetDefaultInt("apisign.allow_time_offset", 300)
	actionWhitelist := cfg.GetDefaultStringArray("apisign.actions.whitelist", []string{"*"})
	actionBlacklist := cfg.GetDefaultStringArray("apisign.actions.blacklist", []string{})

	// 多应用签名配置
	appEnabled := cfg.GetDefaultBool("apisign.app.enabled", false)
	appDbType := cfg.GetDefaultString("apisign.app.db_type", "mysql")
	appDbAlias := cfg.GetDefaultString("apisign.app.db_alias", "default")
	appTable := cfg.GetDefaultString("apisign.app.table", "app")
	appIdColumn := cfg.GetDefaultString("apisign.app.app_id_column", "app_id")
	appIdType := cfg.GetDefaultString("apisign.app.app_id_type", "int")
	if appIdType != "int" && appIdType != "string" {
		appIdType = "int"
	}
	appSignKeyColumn := cfg.GetDefaultString("apisign.app.sign_key_column", "sign_key")
	appStatusColumn := cfg.GetDefaultString("apisign.app.app_status_column", "status")
	appAdminAppEnabled := cfg.GetDefaultBool("apisign.app.admin_app_enabled", false)
	appAdminAppId := cfg.GetDefaultString("apisign.app.admin_app_id", "admin")
	appAdminSignKey := cfg.GetDefaultString("apisign.app.admin_sign_key", "")

	matcher := utils.NewMatcher()
	matcher.SetWhiteList(actionWhitelist)
	matcher.SetBlackList(actionBlacklist)

	s.signKey = signKey
	s.expireTime = expireTime
	s.allowTimeOffset = allowTimeOffset
	s.actionMatcher = matcher

	s.appEnabled = appEnabled
	s.appTable = appTable
	s.appIdColumn = appIdColumn
	s.appIdType = appIdType
	s.appSignKeyColumn = appSignKeyColumn
	s.appStatusColumn = appStatusColumn
	s.appAdminAppEnabled = appAdminAppEnabled
	s.appAdminAppId = appAdminAppId
	s.appAdminSignKey = appAdminSignKey
	if appDbType == "mysql" {
		s.appDb, _ = s.app.DB.GetMysql(appDbAlias)
	} else {
		s.appDb, _ = s.app.DB.GetSqlite3(appDbAlias)
	}
	s.inited = true
}

func (s *Sign) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if s.disabled {
		next(w, r)
		return
	}

	c, err := api.NewContext(s.app, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// check if sign is enable
	if !s.appEnabled && s.signKey == "" {
		next(w, r)
		return
	}

	// check if action not required sign check
	action := c.Input.GetAction()
	if !s.actionMatcher.Match(action) {
		next(w, r)
		return
	}

	var signKey string
	if !s.appEnabled {
		signKey = s.signKey
	} else {
		// read sign key from database
		appId := c.Input.Get("api_appid")
		if appId == "" {
			api.WriteResponse(c, c.Error.NewGroupError("sign", errorMissingAppId))
			return
		}

		// check if it is admin app
		if s.appAdminAppEnabled && appId == s.appAdminAppId {
			signKey = s.appAdminSignKey
		} else {
			var v interface{}
			if s.appIdType == "int" {
				intVal, err := strconv.ParseInt(appId, 10, 32)
				if err != nil {
					api.WriteResponse(c, c.Error.NewGroupError("sign", errorInvalidAppId))
					return
				}
				v = intVal
			} else {
				v = appId
			}
			result, err := s.appDb.Query(
				fmt.Sprintf(
					"SELECT `%s`, `%s` FROM `%s` WHERE `%s`=? LIMIT 1",
					s.appSignKeyColumn,
					s.appStatusColumn,
					s.appTable,
					s.appIdColumn,
				), v)
			if err != nil {
				api.WriteResponse(c, c.Error.New(api.ErrorInternalError, "DBFailed"))
				return
			}
			if len(result) == 0 {
				// no app found, return sign error
				api.WriteResponse(c, c.Error.NewGroupError("sign", errorSignError))
				return
			}
			row := result[0]
			if v, ok := row[s.appStatusColumn]; ok {
				status := string(v)
				if status != "normal" {
					api.WriteResponse(c, c.Error.NewGroupError("sign", errorAppStatusError))
					return
				}
			}
			if v, ok := row[s.appSignKeyColumn]; ok {
				signKey = string(v)
			}
		}
	}

	// check if sign is enable
	if signKey == "" {
		next(w, r)
		return
	}

	// check timestamp
	tstr := c.Input.Get("api_timestamp")
	if tstr == "" {
		api.WriteResponse(c, c.Error.NewGroupError("sign", errorMissingTimestamp))
		return
	}
	ts, err := strconv.ParseInt(tstr, 10, 64)
	if err != nil {
		api.WriteResponse(c, c.Error.NewGroupError("sign", errorInvalidTimestamp))
		return
	}
	// 检测时间是否超前过多
	now := time.Now().Unix()
	offset := ts - now
	if offset > int64(s.allowTimeOffset) {
		api.WriteResponse(c, c.Error.NewGroupError("sign", errorInvalidTimestamp))
		return
	}
	// 超时检测
	if now-ts > int64(s.expireTime) {
		api.WriteResponse(c, c.Error.NewGroupError("sign", errorSignExpired))
		return
	}

	// check sign
	signStr := c.Input.Get("api_sign")
	if signStr == "" {
		api.WriteResponse(c, c.Error.NewGroupError("sign", errorMissingSign))
		return
	}
	values := c.Input.GetForm()
	if !CheckSign(values, signKey, []byte(signStr)) {
		api.WriteResponse(c, c.Error.NewGroupError("sign", errorSignError))
		return
	}

	// next middleware
	next(w, r)
}

// GetSignKey return current sign key as string.
func (s *Sign) GetSignKey() string {
	return s.signKey
}

// SetSignKey allow you update sign key dynamically.
func (s *Sign) SetSignKey(signKey string) {
	s.signKey = signKey
}

// Enable enable the middle ware.
func (s *Sign) Enable() {
	s.disabled = false
	s.init()
}

// Disable disable the middle ware.
func (s *Sign) Disable() {
	s.disabled = true
}
