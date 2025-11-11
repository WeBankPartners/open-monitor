package models

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"reflect"
	"strings"
)

var (
	ErrorTemplateList []*ErrorTemplate
)

type CustomError struct {
	Key           string        `json:"key"`           // 错误编码
	PassEnable    bool          `json:"passEnable"`    // 透传其它服务报错，不用映射
	Code          int           `json:"code"`          // 错误码
	Message       string        `json:"message"`       // 错误信息模版
	DetailErr     error         `json:"detail"`        // 错误信息
	MessageParams []interface{} `json:"messageParams"` // 消息参数列表
}

func (c CustomError) Error() string {
	return c.Message
}

func (c CustomError) WithParam(params ...interface{}) CustomError {
	c.MessageParams = params
	return c
}

type ErrorTemplate struct {
	CodeMessageMap map[int]string `json:"-"`
	CodeKeyMap     map[int]string `json:"-"`

	Language           string      `json:"language"`
	Success            string      `json:"success"`
	ParamValidateError CustomError `json:"param_validate_error"`
	ParamEmptyError    CustomError `json:"param_empty_error"`
	ParamTypeError     CustomError `json:"param_validate_type"`

	RequestBodyError          CustomError `json:"request_body_error"`
	RequestJsonUnmarshalError CustomError `json:"request_json_unmarshal_error"`

	QueryTableError     CustomError `json:"query_table_error"`
	FetchTableDataError CustomError `json:"fetch_table_data_error"`

	UpdateTableError     CustomError `json:"update_table_error"`
	DeleteTableDataError CustomError `json:"delete_table_data_error"`

	HandleError CustomError `json:"handle_error"`

	PasswordError       CustomError `json:"password_error"`
	TokenError          CustomError `json:"token_error"`
	TokenAuthorityError CustomError `json:"token_authority_error"`

	SaveDoneButSyncFail                       CustomError `json:"save_done_but_sync_fail"`
	MetricDuplicateError                      CustomError `json:"metric_duplicate_error"`
	MetricNotFound                            CustomError `json:"metric_not_found"`
	StrategyNameImportDuplicateError          CustomError `json:"strategy_name_import_duplicate_error"`
	LogGroupNameDuplicateError                CustomError `json:"log_group_name_duplicate_error"`
	LogGroupNameIllegalError                  CustomError `json:"log_group_name_illegal_error"`
	DashboardNameRepeatError                  CustomError `json:"dashboard_name_repeat_error"`
	ApiPermissionDeny                         CustomError `json:"api_permission_deny"`
	DashboardIdExistError                     CustomError `json:"dashboard_id_exist_error"`
	ImportDashboardNameExistError             CustomError `json:"import_dashboard_name_exist_error"`
	CreateDashboardNameExistError             CustomError `json:"create_dashboard_name_exist_error"`
	MetricNotSupportPreview                   CustomError `json:"metric_not_support_preview"`
	TypeConfigNameRepeatError                 CustomError `json:"type_config_name_repeat_error"`
	TypeConfigNameAssociationObjectError      CustomError `json:"type_config_association_obj_error"`
	TypeConfigNameAssociationObjectGroupError CustomError `json:"type_config_association_obj_group_error"`
	AlertNameRepeatError                      CustomError `json:"alert_name_repeat_error"`
	AlertKeywordRepeatError                   CustomError `json:"alert_keyword_repeat_error"`
	EndpointHostDeleteError                   CustomError `json:"endpoint_host_delete_error"`
	LogMonitorTemplateDeleteError             CustomError `json:"log_monitor_template_delete_error"`
	AddComparisonMetricRepeatError            CustomError `json:"add_comparison_metric_repeat_error"`
	AddMetricRepeatError                      CustomError `json:"add_metric_repeat_error"`
	DashboardChangedError                     CustomError `json:"dashboard_changed_error"`
}

func InitErrorTemplateList(dirPath string) (err error) {
	if !strings.HasSuffix(dirPath, "/") {
		dirPath = dirPath + "/"
	}
	fs, readDirErr := os.ReadDir(dirPath)
	if readDirErr != nil {
		return readDirErr
	}
	if len(fs) == 0 {
		return fmt.Errorf("dirPath:%s is empty dir", dirPath)
	}
	for _, v := range fs {
		if !strings.HasSuffix(v.Name(), ".json") {
			continue
		}
		tmpFileBytes, _ := os.ReadFile(dirPath + v.Name())
		tmpErrorTemplate := ErrorTemplate{}
		tmpErr := json.Unmarshal(tmpFileBytes, &tmpErrorTemplate)
		if tmpErr != nil {
			err = fmt.Errorf("unmarshal json file :%s fail,%s ", v.Name(), tmpErr.Error())
			continue
		}
		tmpErrorTemplate.Language = strings.Replace(v.Name(), ".json", "", -1)
		tmpErrorTemplate.CodeMessageMap = make(map[int]string)
		tmpErrorTemplate.CodeKeyMap = make(map[int]string)
		tmpRt := reflect.TypeOf(tmpErrorTemplate)
		tmpVt := reflect.ValueOf(tmpErrorTemplate)
		for i := 0; i < tmpRt.NumField(); i++ {
			if tmpRt.Field(i).Type.Name() == "CustomError" {
				tmpC := tmpVt.Field(i).Interface().(CustomError)
				tmpErrorTemplate.CodeMessageMap[tmpC.Code] = tmpC.Message
				tmpErrorTemplate.CodeKeyMap[tmpC.Code] = tmpRt.Field(i).Tag.Get("json")
			}
		}
		ErrorTemplateList = append(ErrorTemplateList, &tmpErrorTemplate)
	}
	if err == nil && len(ErrorTemplateList) == 0 {
		err = fmt.Errorf("i18n error template list empty")
	}
	return err
}

func GetMessageMap(c *gin.Context) *ErrorTemplate {
	acceptLanguage := c.GetHeader("Accept-Language")
	if len(ErrorTemplateList) == 0 {
		return &ErrorTemplate{}
	}
	if acceptLanguage != "" {
		acceptLanguage = strings.Replace(acceptLanguage, ";", ",", -1)
		for _, v := range strings.Split(acceptLanguage, ",") {
			if strings.HasPrefix(v, "q=") {
				continue
			}
			lowerV := strings.ToLower(v)
			for _, vv := range ErrorTemplateList {
				if vv.Language == lowerV {
					return vv
				}
			}
		}
	}
	for _, v := range ErrorTemplateList {
		if v.Language == Config().Http.DefaultLanguage {
			return v
		}
	}
	return ErrorTemplateList[0]
}

func GetErrorResult(headerLanguage string, err error, defaultErrorCode int) (errorCode int, errorKey, errorMessage string) {
	customErr, b := err.(CustomError)
	if !b {
		return defaultErrorCode, "ERROR", err.Error()
	} else {
		errorCode = customErr.Code
		if headerLanguage == "" || customErr.PassEnable {
			errorMessage = buildErrMessage(customErr.Message, customErr.MessageParams)
			if customErr.DetailErr != nil {
				errorMessage = fmt.Sprintf("%s (%s)", errorMessage, customErr.DetailErr.Error())
			}
			return
		}
		headerLanguage = strings.Replace(headerLanguage, ";", ",", -1)
		for _, lang := range strings.Split(headerLanguage, ",") {
			if strings.HasPrefix(lang, "q=") {
				continue
			}
			lang = strings.ToLower(lang)
			for _, template := range ErrorTemplateList {
				if template.Language == lang {
					if message, exist := template.CodeMessageMap[errorCode]; exist {
						errorMessage = buildErrMessage(message, customErr.MessageParams)
						errorKey = template.CodeKeyMap[errorCode]
					}
					break
				}
			}
			if errorMessage != "" {
				break
			}
		}
		if errorMessage == "" {
			errorMessage = buildErrMessage(customErr.Message, customErr.MessageParams)
		}
	}
	if customErr.DetailErr != nil {
		errorMessage = fmt.Sprintf("%s (%s)", errorMessage, customErr.DetailErr.Error())
	}
	return
}

func buildErrMessage(templateMessage string, params []interface{}) (message string) {
	message = templateMessage
	if strings.Count(templateMessage, "%") == 0 {
		return
	}
	message = fmt.Sprintf(message, params...)
	return
}

func IsBusinessErrorCode(errorCode int) bool {
	return strings.HasPrefix(fmt.Sprintf("%d", errorCode), "2")
}
