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

	SaveDoneButSyncFail              string `json:"save_done_but_sync_fail"`
	MetricDuplicateError             string `json:"metric_duplicate_error"`
	MetricNotFound                   string `json:"metric_not_found"`
	StrategyNameImportDuplicateError string `json:"strategy_name_import_duplicate_error"`
	LogGroupNameDuplicateError       string `json:"log_group_name_duplicate_error"`
	LogGroupNameIllegalError         string `json:"log_group_name_illegal_error"`
}
