package middleware

import (
	"bytes"
	"encoding/gob"
	"time"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"github.com/gin-gonic/gin"
	"net/http"
)

//var RedisClient *redis.Client
var LocalMem map[string]m.Session
var onlyLocalStore bool
var localStoreLock = new(sync.RWMutex)

//func InitRedisPool(host string, pwd string, db int) (bool, error) {
//	client := redis.NewClient(&redis.Options{
//		Addr:     host,
//		Password: pwd, // no password set
//		DB:       db,  // use default DB
//	})
//	_, err := client.Ping().Result()
//	if err!=nil {
//		onlyLocalStore = true
//	}else{
//		onlyLocalStore = false
//		RedisClient = client
//	}
//	return !onlyLocalStore, err
//}

func InitLocalSession()  {
	onlyLocalStore = true
	LocalMem = make(map[string]m.Session)
}

func SaveSession(session m.Session) (isOk bool,sId string) {
	isOk = true
	session.Expire = time.Now().Unix() + m.Config().Http.Alive
	serializeData,err := serialize(session)
	if err != nil {
		LogError("serialize session error", err)
		return false, sId
	}
	md := md5.New()
	md.Write(serializeData)
	sId = hex.EncodeToString(md.Sum(nil))
	//if !onlyLocalStore {
	//	backCmd := RedisClient.Set(fmt.Sprintf("session_%s", sId), serializeData, time.Duration(m.Config().Http.Alive) * time.Second)
	//	if !strings.Contains(backCmd.Val(), "OK") {
	//		LogError(fmt.Sprintf("save session to redis fail : %v ", err), nil)
	//		return false, sId
	//	}
	//}
	localStoreLock.Lock()
	LocalMem[sId] = session
	localStoreLock.Unlock()
	return isOk, sId
}

func GetOperateUser(c *gin.Context) string {
	auToken := c.GetHeader("X-Auth-Token")
	if auToken!= "" {
		session := GetSessionData(auToken)
		return fmt.Sprintf("%s", session.User)
	}else{
		Return(c, RespJson{Msg:"no auth token", Code:http.StatusUnauthorized})
		return ""
	}
}

func GetSessionData(sId string) m.Session {
	var result m.Session
	//localContain := false
	localStoreLock.RLock()
	if v,i := LocalMem[sId];i {
		result = v
		//localContain = true
	}
	localStoreLock.RUnlock()
	//if !localContain && !onlyLocalStore {
	//	re := RedisClient.Get(fmt.Sprintf("session_%s", sId))
	//	if len(re.Val()) > 0 {
	//		deserialize([]byte(re.Val()), &result)
	//		LocalMem[sId] = result
	//	}
	//}
	return result
}

func IsActive(sId string) bool {
	localContain := false
	localStoreLock.RLock()
	if _,i := LocalMem[sId];i {
		localContain = true
	}
	localStoreLock.RUnlock()
	//if !localContain && !onlyLocalStore {
	//	var result m.Session
	//	re := RedisClient.Get(fmt.Sprintf("session_%s", sId))
	//	if len(re.Val()) > 0 {
	//		deserialize([]byte(re.Val()), &result)
	//		LocalMem[sId] = result
	//		localContain = true
	//	}
	//}
	return localContain
}

func DelSession(sId string) {
	localStoreLock.Lock()
	if _,i := LocalMem[sId];i {
		delete(LocalMem, sId)
	}
	localStoreLock.Unlock()
	//if !onlyLocalStore {
	//	RedisClient.Del(sId)
	//}
}

// Serialization --------------------------------------------------------------

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
