<template>
  <div>
    <Modal
      v-model="showModal"
      :mask-closable="false"
      :fullscreen="isfullscreen"
      :width="1100"
    >
      <div slot="header" class="custom-modal-header">
        <span>
          {{(view ? $t('m_button_view') : (isAdd ? $t('m_button_add') : $t('m_button_edit'))) + $t('m_custom_regex')}}
        </span>
        <Icon v-if="isfullscreen" @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-expand" />
      </div>
      <div :class="isfullscreen ? 'modal-container-fullscreen' : 'modal-container-normal'">
        <Row>
          <Col span="8">
          <Divider orientation="left" size="small">{{ $t('m_configuration_information') }}</Divider>
          <Form :label-width="120">
            <FormItem :label="$t('m_configuration_name')">
              <Tooltip :content="configInfo.name" transfer :disabled="configInfo.name === ''" style="width: 100%;" max-width="200">
                <Input
                  v-model="configInfo.name"
                  maxlength="30"
                  show-word-limit
                  style="width: 96%"
                  :disabled="view"
                />
                <span style="color: red">*</span>
              </Tooltip>
            </FormItem>
            <FormItem
              v-if='!isInTemplatePage'
              :label="$t('m_metric_code')"
            >
              <Input
                v-model.trim="metricPrefixCode"
                maxlength="15"
                :disabled="!isAdd || view"
                show-word-limit
                :placeholder="$t('m_metric_code_placeholder')"
                style="width:96%"
              >
              </Input>
              <span style="color: red">*</span>
            </FormItem>
            <FormItem :label="$t('m_log_example')">
              <Input
                v-model="configInfo.demo_log"
                type="textarea"
                :rows="15"
                style="width: 96%"
                :disabled="view"
              />
              <div v-if="isParmasChanged && configInfo.demo_log.length === 0" style="color: red">
                {{ $t('m_log_example') }} {{ $t('m_tips_required') }}
              </div>
            </FormItem>
          </Form>
          </Col>
          <Col span="16" style="border-left: 2px solid rgb(232 234 236)">
          <div style="margin-left: 8px">
            <!-- 采集参数 -->
            <div>
              <Divider orientation="left" size="small">{{ $t('m_parameter_collection') }}</Divider>
              <Table
                style="position: inherit;"
                size="small"
                :columns="columnsForParameterCollection"
                :data="configInfo.param_list"
                width="100%"
              ></Table>
              <div style="text-align: right;margin: 0 33px;">
                <Button type="primary" :disabled="configInfo.demo_log === '' || configInfo.param_list.length === 0 || view" @click="generateBackstageTrial" ghost size="small" style="margin:12px">{{ $t('m_match') }}</Button>
                <Button type="success" :disabled="view" @click="addParameterCollection" ghost size="small" icon="md-add"></Button>
              </div>
            </div>
            <!-- 计算指标 -->
            <div>
              <Divider orientation="left" size="small">{{ $t('m_compute_metrics') }}</Divider>
              <Table
                class='compute-metrics-style'
                size="small"
                :columns="columnsForComputeMetrics"
                :data="configInfo.metric_list"
                width="100%"
              ></Table>
              <div style="text-align: right;margin: 0 24px;">
                <Button type="success" :disabled="view" @click="addComputeMetrics" ghost size="small" icon="md-add" style="margin: 12px;"></Button>
              </div>
            </div>
          </div>
          </Col>
        </Row>
      </div>
      <div slot="footer">
        <Checkbox v-if="isInBusinessConfigAdd" v-model="auto_create_warn">{{$t('m_auto_create_warn')}}</Checkbox>
        <Checkbox v-if="isInBusinessConfigAdd" v-model="auto_create_dashboard">{{$t('m_auto_create_dashboard')}}</Checkbox>
        <Button @click="showModal = false">{{ $t('m_button_cancel') }}</Button>
        <Button :disabled="view" @click="saveConfig" type="primary">{{ $t('m_button_save') }}</Button>
      </div>
    </Modal>
    <TagMapConfig ref="tagMapConfigRef" @setTagMap="setTagMap"></TagMapConfig>
  </div>
</template>

<script>
import {
  isEmpty, hasIn, cloneDeep
} from 'lodash'
import Vue from 'vue'
import TagMapConfig from './tag-map-config.vue'
import {thresholdList, lastList} from '@/assets/config/common-config.js'
import {isStringFromNumber, isPositiveNumericString, getRandomColor} from '@/assets/js/utils.js'

const initRangeConfigMap = {
  operator: '',
  threshold: '',
  time: '',
  time_unit: ''
}

const initMetricItem = {
  log_param_name: '',
  metric: '',
  display_name: '',
  agg_type: '',
  tag_config: [],
  color_group: '#1a94bc',
  auto_alarm: true,
  range_config: cloneDeep(initRangeConfigMap)
}

export default {
  name: 'standard-regex',
  data() {
    return {
      showModal: false,
      isfullscreen: true,
      isParmasChanged: false,
      parentGuid: '', // 上级唯一标识
      isAdd: true,
      view: false, // 仅查看，供对象类型查看使用
      configInfo: {
        param_list: []
      },
      columnsForParameterCollection: [
        {
          title: this.$t('m_parameter_key'),
          key: 'name',
          renderHeader: () => (
            <span>
              <span style="color:red">*</span>
              <span>{this.$t('m_parameter_key')}</span>
            </span>
          ),
          render: (h, params) => (
            <Input
              value={params.row.name}
              disabled={this.view}
              placeholder={this.$t('m_metric_key_placeholder')}
              onInput={v => {
                this.changeVal('param_list', params.index, 'name', v)
                this.paramKeyChange()
              }}
            />
          )
        },
        {
          title: this.$t('m_extract_regular'),
          key: 'regular',
          renderHeader: () => (
            <span>
              <span style="color:red">*</span>
              <span>{this.$t('m_extract_regular')}</span>
            </span>
          ),
          render: (h, params) => (
            <Tooltip transfer placement="bottom" theme="light" style="width: 100%;" max-width="500">
              <div slot="content">
                <div domPropsInnerHTML={params.row.regular_font_result} style="word-break: break-all;max-height: 400px;overflow: auto;min-width:200px"></div>
              </div>
              <Input
                value={params.row.regular}
                disabled={this.view}
                onInput={v => {
                  this.changeVal('param_list', params.index, 'regular', v)
                }}
              />
            </Tooltip>
          )
        },
        {
          title: this.$t('m_matching_result'),
          ellipsis: true,
          tooltip: true,
          renderHeader: () => (
            <span>
              <span style="color:red">*</span>
              <span>{this.$t('m_matching_result')}</span>
            </span>
          ),
          key: 'demo_match_value',
          render: (h, params) => {
            const demo_match_value = params.row.demo_match_value
            const notEmpty = demo_match_value !== ''
            return (
              <Tooltip content={demo_match_value} max-width="300" >
                <span style={notEmpty?'':'color:#c5c8ce'}>{notEmpty ? demo_match_value : this.$t('m_no_matching')}</span>
              </Tooltip>
            )
          }
        },
        {
          title: this.$t('m_match_value_pure'),
          ellipsis: true,
          tooltip: true,
          key: 'string_map',
          render: (h, params) => {
            const val = !isEmpty(params.row.string_map) && params.row.string_map.map(item => item.target_value).join(',') || ''
            return (
              <div>
                <Input disabled style="width:80%" value={val}/>
                <Button
                  size="small"
                  type="primary"
                  disabled={this.view}
                  icon="md-create"
                  onClick={() => this.editTagMapping(params.index)}
                >
                </Button>
              </div>
            )
          }
        },
        {
          title: this.$t('m_table_action'),
          key: 'action',
          width: 80,
          align: 'left',
          render: (h, params) => (
            <div style="text-align: left; cursor: pointer;display: inline-flex;">
              <Button
                size="small"
                type="error"
                style="margin-right:5px;"
                disabled={this.configInfo.param_list.length === 1}
                onClick={() => this.deleteAction('param_list', params.index)}
              >
                <Icon type="md-trash" size="16"></Icon>
              </Button>
            </div>
          )
        }
      ],
      columnsForComputeMetrics: [
        {
          title: this.$t('m_color_system'),
          key: 'color_group',
          width: 80,
          render: (h, params) => (
            <div class="color_system">
              <ColorPicker value={params.row.color_group || ''}
                disabled={this.view}
                on-on-open-change={
                  isShow => this.changeColorGroup(isShow, this.configInfo.metric_list[params.index], 'color_group')
                }
              />
            </div>
          )
        },
        {
          title: this.$t('m_field_displayName'),
          key: 'display_name',
          width: 120,
          render: (h, params) => (
            <Input
              value={params.row.display_name}
              disabled={this.view}
              placeholder={this.$t('m_metric_key_placeholder')}
              onInput={v => {
                this.changeVal('metric_list', params.index, 'display_name', v)
              }}
            />
          )
        },
        {
          title: this.$t('m_metric_key'),
          key: 'metric',
          width: 120,
          renderHeader: () => (
            <span>
              <span style="color:red">*</span>
              <span>{this.$t('m_metric_key')}</span>
            </span>
          ),
          render: (h, params) => (
            <Input
              value={params.row.metric}
              disabled={this.view}
              placeholder={this.$t('m_metric_key_placeholder')}
              onInput={v => {
                this.changeVal('metric_list', params.index, 'metric', v)
              }}
            />
          )
        },
        {
          title: this.$t('m_metric_key'),
          key: 'index',
          width: 120,
          renderHeader: () => (
            <span>
              <span>{this.$t('m_final_metric_key')}</span>
            </span>
          ),
          render: (h, params) => (
            <div>{this.metricPrefixCode ? this.metricPrefixCode + '_' + params.row.metric : params.row.metric}</div>
          )
        },
        {
          title: this.$t('m_statistical_parameters'),
          key: 'log_param_name',
          width: 130,
          renderHeader: () => (
            <span>
              <span style="color:red">*</span>
              <span>{this.$t('m_statistical_parameters')}</span>
            </span>
          ),
          render: (h, params) => {
            const keys = this.configInfo.param_list.map(p => p.name)
            const selectOptions = [...new Set(keys)]
            return (
              <Select
                filterable
                value={params.row.log_param_name}
                disabled={this.view}
                on-on-change={v => {
                  this.changeVal('metric_list', params.index, 'log_param_name', v)
                }}
              >
                {selectOptions.map(option => (
                  <Option key={option} value={option}>
                    {option}
                  </Option>
                ))}
              </Select>
            )
          }
        },
        {
          title: this.$t('m_filter_label'),
          key: 'tag_config',
          width: 120,
          render: (h, params) => {
            const keys = this.configInfo.param_list.map(p => p.name)
            const selectOptions = [...new Set(keys)]
            return (
              <Select
                filterable
                value={params.row.tag_config}
                disabled={this.view}
                multiple
                on-on-change={v => {
                  this.changeVal('metric_list', params.index, 'tag_config', v)
                }}
              >
                {selectOptions.map(option => (
                  <Option key={option} value={option}>
                    {option}
                  </Option>
                ))}
              </Select>
            )
          }
        },
        {
          title: this.$t('m_computed_type'),
          key: 'agg_type',
          width: 120,
          renderHeader: () => (
            <span>
              <span style="color:red">*</span>
              <span>{this.$t('m_computed_type')}</span>
            </span>
          ),
          render: (h, params) => {
            const canOnlySelectCount = this.isNumericValue[params.row.log_param_name]
            const selectOptions = [
              {
                labelAndValue: 'avg',
                disabled: canOnlySelectCount
              },
              {
                labelAndValue: 'count',
                disabled: false
              },
              {
                labelAndValue: 'max',
                disabled: canOnlySelectCount
              },
              {
                labelAndValue: 'min',
                disabled: canOnlySelectCount
              },
              {
                labelAndValue: 'sum',
                disabled: canOnlySelectCount
              }
            ]
            return (
              <Select
                filterable
                disabled={params.row.log_param_name==='' || this.view}
                value={params.row.agg_type}
                on-on-change={v => {
                  this.changeVal('metric_list', params.index, 'agg_type', v)
                }}
              >
                {selectOptions.map(option => (
                  <Option disabled={option.disabled} key={option.labelAndValue} value={option.labelAndValue}>
                    {option.labelAndValue}
                  </Option>
                ))}
              </Select>
            )
          }
        },
        {
          title: this.$t('m_automatic_alert'),
          key: 'auto_alarm',
          width: 80,
          render: (h, params) =>
            (
              <i-switch value={params.row.auto_alarm}
                disabled={this.view}
                on-on-change={val => {
                  if (!val) {
                    Vue.set(this.configInfo.metric_list[params.index], 'range_config', cloneDeep(initRangeConfigMap))
                  }
                  Vue.set(this.configInfo.metric_list[params.index], 'auto_alarm', val)
                  // this.configInfo.metric_list[params.index].auto_alarm = val
                }} />
            )
        },
        {
          title: this.$t('m_symbol'),
          key: 'operator',
          align: 'left',
          minWidth: 100,
          render: (h, params) => params.row.auto_alarm
            ? (
              <Select
                value={params.row.range_config.operator}
                disabled={this.view}
                on-on-change={v => {
                  this.configInfo.metric_list[params.index].range_config.operator = v
                }}
                filterable
                clearable
              >
                {thresholdList.map((i, index) => (
                  <Option value={i.value} key={index}>
                    {i.label}
                  </Option>
                ))}
              </Select>
            ) : <div></div>
        },
        {
          title: this.$t('m_field_threshold'),
          key: 'threshold',
          align: 'left',
          width: 70,
          render: (h, params) => params.row.auto_alarm ? (
            <Input
              value={params.row.range_config.threshold}
              disabled={this.view}
              on-on-change={v => {
                this.configInfo.metric_list[params.index].range_config.threshold = v.target.value
              }}
              clearable
            />
          ) : <div/>
        },
        {
          title: this.$t('m_tableKey_s_last'),
          key: 'time',
          align: 'left',
          width: 70,
          render: (h, params) => params.row.auto_alarm ? (
            <Input
              value={params.row.range_config.time}
              disabled={this.view}
              on-on-change={v => {
                this.configInfo.metric_list[params.index].range_config.time = v.target.value
              }}
              clearable
            />
          ) : <div/>
        },
        {
          title: this.$t('m_time_unit'),
          key: 'time_unit',
          align: 'left',
          minWidth: 70,
          render: (h, params) => params.row.auto_alarm ? (
            <Select
              value={params.row.range_config.time_unit}
              disabled={this.view}
              on-on-change={v => {
                this.configInfo.metric_list[params.index].range_config.time_unit = v
              }}
              filterable
              clearable
            >
              {lastList.map(i => (
                <Option value={i.value} key={i.value}>
                  {i.label}
                </Option>
              ))}
            </Select>
          ) : <div/>
        },
        {
          title: this.$t('m_table_action'),
          key: 'action',
          width: 80,
          fixed: 'right',
          render: (h, params) => (
            <div style="text-align: left; cursor: pointer;display: inline-flex;">
              <Button
                disabled={this.configInfo.metric_list.length === 1 || this.view}
                size="small"
                type="error"
                style="margin-right:5px;"
                onClick={() => this.deleteAction('metric_list', params.index)}
              >
                <Icon type="md-trash" size="16"></Icon>
              </Button>
            </div>
          )
        }
      ],
      editTagMappingIndex: -1, // 正在编辑的参数采集
      isNumericValue: {}, // 缓存后参数key对应的匹配结果能否转成数字
      actionType: '',
      isLogTemplate: false, // 该组件在业务配置和日志模板中均使用，true代表在日志模板中， false为业务配置中
      isEmpty,
      auto_create_warn: true,
      auto_create_dashboard: true,
      metricPrefixCode: ''
    }
  },
  computed: {
    isInBusinessConfigAdd() {
      return this.actionType === 'add' && !isEmpty(this.parentGuid)
    },
    isBaseCustomeTemplateAdd() { // 在业务配置页面新增
      return this.actionType === 'add' && this.isLogTemplate && !isEmpty(this.parentGuid)
    },
    isInTemplatePage() { // 在模板配置也新增or修改
      return this.isLogTemplate && isEmpty(this.parentGuid)
    }
  },
  methods: {
    loadPage(actionType, templateGuid, parentGuid, configGuid, isLogTemplate = false) {
      this.isLogTemplate = isLogTemplate
      this.isfullscreen = true
      this.parentGuid = parentGuid
      this.metricPrefixCode = ''
      this.auto_create_dashboard = true
      this.auto_create_warn = true
      // actionType add/edit
      // templateGuid, 模版id
      // parentGuid, 上级唯一标识
      // configGuid, 配置唯一标志
      this.isAdd = ['add', 'copy'].includes(actionType)
      this.actionType = actionType
      this.view = actionType === 'view'
      if (configGuid) {
        this.getConfig(configGuid)
      } else {
        this.configInfo = {
          guid: '',
          log_metric_monitor: '',
          name: '',
          log_type: 'custom',
          demo_log: '',
          param_list: [
            {
              guid: '',
              name: '',
              display_name: '',
              json_key: '',
              regular: '',
              demo_match_value: '',
              string_map: [
                {
                  regulative: 0, // 匹配类型： 0 是非正则，1是正则
                  source_value: '', // 源值
                  target_value: '', // 映射值
                }
              ]
            }
          ],
          metric_list: [cloneDeep(initMetricItem)]
        }
        this.configInfo.log_metric_monitor = parentGuid
      }
      this.showModal = true
    },
    regularCheckValue(arr = [], key) {
      const regex = /^[A-Za-z][A-Za-z0-9_]{0,48}[A-Za-z0-9]$/
      for (let i=0; i<arr.length; i++) {
        if (!regex.test(arr[i][key])) {
          return false
        }
      }
      return true
    },
    paramsValidate(tmpData) {
      if (this.isAdd && !this.isInTemplatePage) {
        if (this.metricPrefixCode === '') {
          this.$Message.warning(`${this.$t('m_metric_code')}: ${this.$t('m_cannot_be_empty')}`)
          return true
        }
        const regex = /^[A-Za-z][A-Za-z0-9]{0,14}$/

        if (!regex.test(this.metricPrefixCode)) {
          this.$Message.warning(`${this.$t('m_metric_code')}: ${this.$t('m_metric_prefix_code_validate')}`)
          return true
        }
      }
      if (!this.regularCheckValue(tmpData.param_list, 'name')) {
        return this.$Message.warning(`${this.$t('m_parameter_key')}: ${this.$t('m_regularization_check_failed_tips')}`)
      }
      if (!this.regularCheckValue(tmpData.metric_list, 'metric')) {
        return this.$Message.warning(`${this.$t('m_metric_key')}: ${this.$t('m_regularization_check_failed_tips')}`)
      }
      if (tmpData.name === '') {
        this.$Message.warning(`${this.$t('m_tableKey_name')}${this.$t('m_cannot_be_empty')}`)
        return true
      }
      if (tmpData.param_list.length === 0) {
        this.$Message.warning(`${this.$t('m_add_one_tip')}: ${this.$t('m_parameter_collection')}`)
        return true
      }
      const is_param_list_empty = tmpData.param_list.some(element => element.name === '' || element.regular === '')
      if (is_param_list_empty) {
        this.$Message.warning(`${this.$t('m_parameter_collection')}: ${this.$t('m_fields_cannot_be_empty')}`)
        return true
      }
      const is_demo_match_value_empty = tmpData.param_list.some(element => element.demo_match_value === '')
      if (is_demo_match_value_empty) {
        this.$Message.warning(`${this.$t('m_matching_result')}: ${this.$t('m_cannot_be_empty')}`)
        return true
      }

      const hasDuplicatesParamList = tmpData.param_list.some((element, index) => tmpData.param_list.findIndex(item => item.name === element.name) !== index)
      if (hasDuplicatesParamList) {
        this.$Message.warning(`${this.$t('m_parameter_key')}${this.$t('m_cannot_be_repeated')}`)
        return true
      }

      if (tmpData.metric_list.length === 0) {
        this.$Message.warning(`${this.$t('m_add_one_tip')}: ${this.$t('m_compute_metrics')}`)
        return true
      }

      const is_metric_list_empty = tmpData.metric_list.some(element => element.metric === '' || element.log_param_name === '' || element.agg_type === '')
      if (is_metric_list_empty) {
        this.$Message.warning(`${this.$t('m_compute_metrics')}: ${this.$t('m_fields_cannot_be_empty')}`)
        return true
      }
      const hasDuplicatesMetricList = tmpData.metric_list.some((element, index) => tmpData.metric_list.findIndex(item => item.metric === element.metric) !== index)
      if (hasDuplicatesMetricList) {
        this.$Message.warning(`${this.$t('m_metric_key')}${this.$t('m_cannot_be_repeated')}`)
        return true
      }
      if (!isEmpty(tmpData.metric_list)) {
        const list = tmpData.metric_list
        for (let i=0; i<list.length; i++) {
          const item = list[i]
          if (item.auto_alarm === true) {
            if (!item.range_config.operator || !item.range_config.threshold || !item.range_config.time || !item.range_config.time_unit) {
              this.$Message.warning(`${this.$t('m_threshold_property')}${this.$t('m_cannot_be_empty')}`)
              return true
            }
            if (!isStringFromNumber(item.range_config.threshold + '')) {
              this.$Message.warning(`${this.$t('m_threshold_tips')}`)
              return true
            }
            if (!isPositiveNumericString(item.range_config.time + '')) {
              this.$Message.warning(`${this.$t('m_time_tips')}`)
              return true
            }
          }
        }
      }
      return false
    },
    // 更新筛选标签中的值
    paramKeyChange() {
      const paramsNameArray = this.configInfo.param_list.map(p => p.name)
      this.configInfo.metric_list.forEach(metric => {
        const tmpTag = []
        metric.tag_config.forEach(tag => {
          if (paramsNameArray.includes(tag)) {
            tmpTag.push(tag)
          }
        })
        metric.tag_config = tmpTag
      })
    },
    saveConfig() {
      const tmpData = JSON.parse(JSON.stringify(this.configInfo))
      if (this.paramsValidate(tmpData)) {
        return
      }
      delete tmpData.create_user
      delete tmpData.create_time
      delete tmpData.update_user
      delete tmpData.update_time
      const methodType = this.isAdd ? 'POST' : 'PUT'
      let api = ''
      if (this.isInTemplatePage) { // 在模板配置页面
        api = this.$root.apiCenter.logTemplateConfig
        tmpData.calc_result = {
          match_text: '',
          json_key_list: [],
          json_obj: {}
        }
        !isEmpty(tmpData.param_list) && tmpData.param_list.forEach(item => {
          item.display_name = item.name
        })
        this.processSaveData(tmpData)
        // !isEmpty(tmpData.metric_list) && tmpData.metric_list.forEach(item => {
        //   item.display_name = item.log_param_name
        // })
      } else {
        // api = this.$root.apiCenter.logMetricGroup
        api = this.$root.apiCenter.customLogMetricConfig
        if (this.isAdd) {
          tmpData.log_monitor_template_guid = tmpData.guid
          tmpData.guid = ''
          if (hasIn(tmpData, 'calc_result')) {
            delete tmpData.calc_result
          }
          if (hasIn(tmpData, 'success_code')) {
            delete tmpData.success_code
          }
          tmpData.log_metric_monitor = this.parentGuid
          tmpData.metric_prefix_code = this.metricPrefixCode
        }
      }
      if (this.isInBusinessConfigAdd) {
        tmpData.auto_create_dashboard = this.auto_create_dashboard
        tmpData.auto_create_warn = this.auto_create_warn
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(methodType, api, tmpData, res => {
        const messageTips = this.$t('m_tips_success')
        if (!isEmpty(res) && hasIn(res, 'alarm_list') && hasIn(res, 'custom_dashboard')) {
          const tipOne = isEmpty(res.alarm_list) ? '' : '<br/>' + res.alarm_list.join('<br/>')
          const tipTwo = isEmpty(res.custom_dashboard) ? '' : res.custom_dashboard
          this.$Message.success({
            render: h => h('div', { class: 'add-business-config' }, [
              h('div', {class: 'add-business-config-item'}, [
                h('div', this.$t('m_has_create_dashboard') + ':'),
                h('div', {
                  domProps: {
                    innerHTML: tipTwo
                  }
                })
              ]),
              h('div', { class: 'add-business-config-item' }, [
                h('div', this.$t('m_has_create_warn') + ':'),
                h('div', {
                  class: 'create_warn_text',
                  domProps: {
                    innerHTML: tipOne
                  }
                })
              ])
            ]),
            duration: 5
          })
        } else {
          this.$Message.success({
            content: messageTips,
            duration: 2
          })
        }
        this.showModal = false
        this.$emit('reloadMetricData', this.parentGuid)
      })
    },
    processSaveData(data){
      if (isEmpty(data)) {return}
      !isEmpty(data.metric_list) && data.metric_list.forEach(item => {
        if (!isEmpty(item.range_config)) {
          item.range_config.threshold += ''
          item.range_config.time += ''
          item.range_config = JSON.stringify(item.range_config)
        }
      })
    },
    processConfigInfo() {
      !isEmpty(this.configInfo.metric_list) && this.configInfo.metric_list.forEach(item => {
        Vue.set(item, 'range_config', isEmpty(item.range_config) ? cloneDeep(initRangeConfigMap.other) : JSON.parse(item.range_config))
        Vue.set(item, 'auto_alarm', hasIn(item, 'auto_alarm') ? item.auto_alarm : false)
        Vue.set(item, 'color_group', isEmpty(item.color_group) ? '' : item.color_group)
      })
      this.metricPrefixCode = this.configInfo.metric_prefix_code
    },
    getConfig(guid) {
      const api = this.isLogTemplate ? this.$root.apiCenter.getConfigDetailByGuid + guid : this.$root.apiCenter.customLogMetricConfig + '/' + guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, {}, resp => {
        this.configInfo = resp
        this.processConfigInfo()
        if (this.actionType === 'copy') {
          this.configInfo.name = this.configInfo.name + '1'
          this.metricPrefixCode = ''
        }
        if (this.isBaseCustomeTemplateAdd) {
          this.configInfo.name = ''
        }
        {/* 整理计算类型可选项 */}
        const param_list = this.configInfo.param_list || []
        param_list.forEach(item => {
          this.isNumericValue[item.name] = !this.isNumericString(item.demo_match_value)
        })
        this.configInfo.param_list.forEach(r => {
          r.regular_font_result = this.regRes(r.regular)
        })
        this.showModal = true
      })
    },
    changeVal(params, index, key, val) {
      this.configInfo[params][index][key] = val
      if (params === 'param_list' && key === 'regular') {
        this.configInfo[params][index].regular_font_result = this.regRes(val)
      }
      if (key === 'log_param_name') {
        this.configInfo[params][index]['agg_type'] = ''
      }
    },
    regRes(val) {
      try {
        const reg = new RegExp(val, 'g')
        const match = reg.exec(this.configInfo.demo_log)
        if (match) {
          return this.configInfo.demo_log.replace(match[1], '<span style=\'color:red\'>' + match[1] + '</span>')
        }
        return `<span style='color:#c5c8ce'>${this.$t('m_no_matching')}</span>`
      } catch (err) {
        return `<span style='color:#c5c8ce'>${this.$t('m_no_matching')}</span>`
      }
    },
    generateBackstageTrial() {
      if (this.configInfo.demo_log === '') {
        this.$Message.warning(`${this.$t('m_log_example')}${this.$t('m_cannot_be_empty')}`)
        return
      }
      {/* const hasDuplicatesParamList = this.configInfo.param_list.some((element, index) => {
        return this.configInfo.param_list.findIndex((item) => item.name === element.name) !== index
      })
      if (hasDuplicatesParamList) {
        this.$Message.warning(`${this.$t('m_parameter_key')}${this.$t('m_cannot_be_repeated')}`)
        return true
      } */}
      const params = {
        demo_log: this.configInfo.demo_log,
        param_list: this.configInfo.param_list
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.standardLogRegexMatch, params, responseData => {
        this.$Message.success(this.$t('m_success'))
        this.configInfo.param_list = responseData || []
        responseData.forEach(item => {
          this.isNumericValue[item.name] = !this.isNumericString(item.demo_match_value)
        })
        this.configInfo.metric_list.forEach(item => {
          item.agg_type = ''
        })
      }, {isNeedloading: false})
    },
    isNumericString(str) {
      return !isNaN(parseFloat(str)) && isFinite(str)
    },
    // #region 参数采集
    addParameterCollection() {
      this.configInfo.param_list.push({
        guid: '',
        name: '',
        display_name: '',
        json_key: '',
        regular: '',
        demo_match_value: '',
        string_map: [
          // {
          //   regulative: 1,  //匹配类型： 0 是非正则，1是正则
          //   source_value: '', // 源值
          //   target_value: '', // 映射值
          //   value_type: '', //值类型： success 成功，fail 失败
          // }
        ]
      })
    },
    // 编辑标签映射
    editTagMapping(index) {
      this.editTagMappingIndex = index
      const tagMap = this.configInfo.param_list[index].string_map || []
      this.$refs.tagMapConfigRef.loadPage(tagMap)
    },
    setTagMap(arr) {
      this.configInfo.param_list[this.editTagMappingIndex].string_map = arr
    },
    // #endregion
    // #region 计算指标
    addComputeMetrics() {
      const item = cloneDeep(initMetricItem)
      item.color_group = getRandomColor()
      this.configInfo.metric_list.push(item)
    },
    // #endregion
    deleteAction(key, index) {
      this.configInfo[key].splice(index, 1)
    },
    changeColorGroup(isShow = true, data, key) {
      if (isShow) {
        this.$nextTick(() => {
          const confirmButtonList = document.querySelectorAll('.ivu-color-picker-confirm .ivu-btn-primary')
          const resetButtonList = document.querySelectorAll('.ivu-color-picker-confirm .ivu-btn-default')
          if (isEmpty(confirmButtonList)) {
            return
          }
          confirmButtonList[0].addEventListener('click', () => {
            const inputList = document.querySelectorAll('.ivu-color-picker-confirm .ivu-input')
            if (isEmpty(inputList)) {
              return
            }
            const color = inputList[0].value
            Vue.set(data, key, color)
          })
          if (isEmpty(resetButtonList)) {
            return
          }
          resetButtonList[0].addEventListener('click', () => {
            Vue.set(data, key, '')
          })
        })
      }
    }
  },
  components: {
    TagMapConfig
  }
}
</script>

<style lang="less" scoped>
.modal-container-normal {
  height: ~"calc(100vh - 280px)";
  overflow: auto;
}
.modal-container-fullscreen {
  height: ~"calc(100vh - 150px)";
  overflow: auto;
}
.custom-modal-header {
  line-height: 20px;
  font-size: 16px;
  color: #17233d;
  font-weight: 500;
  .fullscreen-icon {
    float: right;
    margin-right: 28px;
    font-size: 18px;
    cursor: pointer;
  }
}
.ivu-form-item {
  margin-bottom: 0px;
}
</style>

<style lang='less'>
.compute-metrics-style {
  position: inherit;
  .ivu-table-cell {
    padding-left: 5px;
    padding-right: 5px;
  }
}

.add-business-config {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  max-width: 900px;
  max-height: 600px;
  overflow-y: auto;
  .add-business-config-item {
    display: flex;
    flex-direction: row;
    .create_warn_text {
      text-align: left
    }
  }
}
.add-business-config > div {
  max-width: 850px;
  word-wrap: break-word;
  white-space: normal;
}

</style>
