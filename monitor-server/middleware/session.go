package middleware

import (
	"bytes"
	"encoding/gob"
	"time"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/go-redis/redis"
	"strings"
	"encoding/base64"
	"encoding/json"
)

var RedisClient *redis.Client
var LocalMem = make(map[string]m.Session)
var onlyLocalStore bool
var localStoreLock = new(sync.RWMutex)
var expireTime = int64(3600)

func InitSession()  {
	sessionConfig := m.Config().Http.Session
	expireTime = m.Config().Http.Session.Expire
	onlyLocalStore = true
	if sessionConfig.Redis.Enable {
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", sessionConfig.Redis.Server, sessionConfig.Redis.Port),
			Password: sessionConfig.Redis.Pwd, // no password set
			DB:       sessionConfig.Redis.Db,  // use default DB
		})
		_, err := client.Ping().Result()
		if err!=nil {
			LogError("init session redis fail ", err)
			onlyLocalStore = true
		}else{
			LogInfo("init session redis success")
			onlyLocalStore = false
			RedisClient = client
		}
	}
}

func SaveSession(session m.Session) (isOk bool,sId string) {
	isOk = true
	session.Expire = time.Now().Unix() + expireTime
	serializeData,err := serialize(session)
	if err != nil {
		LogError("serialize session error", err)
		return false, sId
	}
	md := md5.New()
	md.Write(serializeData)
	sId = hex.EncodeToString(md.Sum(nil))
	if !onlyLocalStore {
		backCmd := RedisClient.Set(fmt.Sprintf("session_%s", sId), serializeData, time.Duration(expireTime) * time.Second)
		if !strings.Contains(backCmd.Val(), "OK") {
			LogError(fmt.Sprintf("save session to redis fail : %v ", err), nil)
			return false, sId
		}
	}
	localStoreLock.Lock()
	LocalMem[sId] = session
	localStoreLock.Unlock()
	return isOk, sId
}

func GetOperateUser(c *gin.Context) string {
	if !m.Config().Http.Session.Enable {
		return ""
	}
	auToken := c.GetHeader("X-Auth-Token")
	if auToken!= "" {
		if m.Config().Http.Session.ServerEnable {
			if auToken == m.Config().Http.Session.ServerToken {
				return "auth_server"
			}
		}
		session := GetSessionData(auToken)
		return fmt.Sprintf("%s", session.User)
	}else{
		Return(c, RespJson{Msg:"no auth token", Code:http.StatusUnauthorized})
		return ""
	}
}

func GetCoreRequestRoleList(c *gin.Context) []string {
	var result []string
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return result
	}
	authHeader = authHeader[strings.Index(authHeader, ".")+1:]
	authHeader = authHeader[:strings.LastIndex(authHeader, ".")]
	authHeader += "=="
	b,err := base64.StdEncoding.DecodeString(authHeader)
	if err != nil {
		LogError("decode core request base64 fail ", err)
		return result
	}
	var requestToke m.CoreRequestToken
	err = json.Unmarshal(b, &requestToke)
	if err != nil {
		LogError("get core token,json unmarchal fail ", err)
		return result
	}
	if requestToke.Authority != "" {
		requestToke.Authority = strings.Replace(requestToke.Authority, "[", "", -1)
		requestToke.Authority = strings.Replace(requestToke.Authority, "]", "", -1)
		result = strings.Split(requestToke.Authority, ",")
	}
	return result
}

func GetSessionData(sId string) m.Session {
	var result m.Session
	localContain := false
	localStoreLock.RLock()
	if v,i := LocalMem[sId];i {
		result = v
		localContain = true
	}
	localStoreLock.RUnlock()
	if !localContain && !onlyLocalStore {
		re := RedisClient.Get(fmt.Sprintf("session_%s", sId))
		if len(re.Val()) > 0 {
			deserialize([]byte(re.Val()), &result)
			LocalMem[sId] = result
		}
	}
	return result
}

func IsActive(sId string) bool {
	if m.Config().Http.Session.ServerEnable {
		if sId == m.Config().Http.Session.ServerToken {
			return true
		}
	}
	localContain := false
	localStoreLock.RLock()
	defer localStoreLock.RUnlock()
	if v,i := LocalMem[sId];i {
		if time.Now().Unix() > v.Expire {
			delete(LocalMem, sId)
			return false
		}
		localContain = true
	}
	if !localContain && !onlyLocalStore {
		var result m.Session
		re := RedisClient.Get(fmt.Sprintf("session_%s", sId))
		if len(re.Val()) > 0 {
			deserialize([]byte(re.Val()), &result)
			LocalMem[sId] = result
			localContain = true
		}
	}
	return localContain
}

func DelSession(sId string) {
	localStoreLock.Lock()
	if _,i := LocalMem[sId];i {
		delete(LocalMem, sId)
	}
	localStoreLock.Unlock()
	if !onlyLocalStore {
		RedisClient.Del(sId)
	}
}

// Serialize encodes a value using gob.
func serialize(src interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(src); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize decodes a value using gob.
func deserialize(src []byte, dst interface{}) error {
	dec := gob.NewDecoder(bytes.NewBuffer(src))
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}
