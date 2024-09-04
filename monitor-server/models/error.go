package models

type ErrorMessageObj struct {
	Language string `json:"language"`
	Success  string `json:"success"`

	ParamValidateError string `json:"param_validate_error"`
	ParamEmptyError    string `json:"param_empty_error"`
	ParamTypeError     string `json:"param_validate_type"`

	RequestBodyError          string `json:"request_body_error"`
	RequestJsonUnmarshalError string `json:"request_json_unmarshal_error"`

	QueryTableError     string `json:"query_table_error"`
	FetchTableDataError string `json:"fetch_table_data_error"`

	UpdateTableError     string `json:"update_table_error"`
	DeleteTableDataError string `json:"delete_table_data_error"`

	HandleError string `json:"handle_error"`

	PasswordError       string `json:"password_error"`
	TokenError          string `json:"token_error"`
	TokenAuthorityError string `json:"token_authority_error"`

	SaveDoneButSyncFail                       string `json:"save_done_but_sync_fail"`
	MetricDuplicateError                      string `json:"metric_duplicate_error"`
	MetricNotFound                            string `json:"metric_not_found"`
	StrategyNameImportDuplicateError          string `json:"strategy_name_import_duplicate_error"`
	LogGroupNameDuplicateError                string `json:"log_group_name_duplicate_error"`
	LogGroupNameIllegalError                  string `json:"log_group_name_illegal_error"`
	DashboardNameRepeatError                  string `json:"dashboard_name_repeat_error"`
	DashboardIdExistError                     string `json:"dashboard_id_exist_error"`
	ImportDashboardNameExistError             string `json:"import_dashboard_name_exist_error"`
	CreateDashboardNameExistError             string `json:"create_dashboard_name_exist_error"`
	MetricNotSupportPreview                   string `json:"metric_not_support_preview"`
	TypeConfigNameRepeatError                 string `json:"type_config_name_repeat_error"`
	TypeConfigNameAssociationObjectError      string `json:"type_config_association_obj_error"`
	TypeConfigNameAssociationObjectGroupError string `json:"type_config_association_obj_group_error"`
	AlertNameRepeatError                      string `json:"alert_name_repeat_error"`
	AlertKeywordRepeatError                   string `json:"alert_keyword_repeat_error"`
	EndpointHostDeleteError                   string `json:"endpoint_host_delete_error"`
	LogMonitorTemplateDeleteError             string `json:"log_monitor_template_delete_error"`
	AddComparisonMetricRepeatError            string `json:"add_comparison_metric_repeat_error"`
	AddMetricRepeatError                      string `json:"add_metric_repeat_error"`
}
