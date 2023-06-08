package funcs

import (
	"encoding/json"
	"fmt"
)

type HttpResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (h *HttpResponse) Byte() []byte {
	d, err := json.Marshal(h)
	if err == nil {
		return d
	} else {
		return []byte(fmt.Sprintf("{\"code\":%d,\"message\":\"%s\",\"data\":%v}", h.Code, h.Message, h.Data))
	}
}
