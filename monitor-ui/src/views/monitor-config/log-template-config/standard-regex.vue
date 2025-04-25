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
          {{(isAdd ? $t('m_button_add') : $t('m_button_edit')) + $t('m_template')}}
        </span>
        <Icon v-if="isfullscreen" @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-expand" />
      </div>
      <div>
        <Row>
          <Col span="8" :style="{height: isfullscreen ? '' : '550px'}" style="overflow: auto;">
          <Form :label-width="120">
            <FormItem :label="$t('m_template_name')">
              <Tooltip :content="configInfo.name" transfer :disabled="configInfo.name === ''" style="width: 100%;" max-width="200">
                <Input
                  v-model.trim="configInfo.name"
                  maxlength="30"
                  show-word-limit
                  style="width: 96%"
                />
                <span style="color: red">*</span>
              </Tooltip>
            </FormItem>
            <template v-if="!isAdd">
              <FormItem :label="$t('m_updatedBy')">
                {{ configInfo.update_user }}
              </FormItem>
              <FormItem :label="$t('m_title_updateTime')">
                {{ configInfo.update_time }}
              </FormItem>
            </template>
            <FormItem :label="$t('m_log_example')">
              <Input
                v-model.trim="configInfo.demo_log"
                type="textarea"
                :rows="24"
                style="width: 96%"
              />
              <div v-if="isParmasChanged && configInfo.demo_log.length === 0" style="color: red">
                {{ $t('m_log_example') }} {{ $t('m_tips_required') }}
              </div>
            </FormItem>
          </Form>
          </Col>
          <Col span="16" style="border-left: 2px solid rgb(232 234 236);">
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
              <Button type="primary" :disabled="configInfo.demo_log === ''" @click="generateBackstageTrial" ghost size="small" style="float:right;margin:12px">{{ $t('m_match') }}</Button>
            </div>
            <!-- 计算指标 -->
            <div class='calculation-indicators'>
              <Divider orientation="left" size="small">
                {{ $t('m_compute_metrics') }}
                <Tooltip content="" placement="top" :max-width="500" transfer>
                  <div slot="content" style="white-space: normal;">
                    <div>{{ $t('m_calculation_title') }}</div>
                    <div>{{ $t('m_calculation_tip1') }}</div>
                    <div>{{ $t('m_calculation_tip2') }}</div>
                    <div>{{ $t('m_calculation_tip3') }}</div>
                    <div>{{ $t('m_calculation_tip4') }}</div>
                    <div>{{ $t('m_calculation_tip5') }}</div>
                    <div>{{ $t('m_calculation_tip6') }}</div>
                    <div>{{ $t('m_calculation_tip7') }}</div>
                    <div>{{ $t('m_calculation_tip8') }}</div>
                  </div>
                  <Icon style="cursor: pointer;" type="ios-alert-outline" :size="18" />
                </Tooltip>
              </Divider>
              <Table
                size="small"
                :columns="columnsForComputeMetrics"
                :data="configInfo.metric_list"
                width="100%"
              ></Table>
            </div>
            <Divider orientation="left" size="small">{{ $t('m_success_code') }}</Divider>
            <Row class="mb-2">
              <Col offset="3"  span="5">{{$t('m_match_type')}}</Col>
              <Col span="5">{{$t('m_source_value')}}</Col>
              <Col span="5">{{$t('m_match_value')}}</Col>
              <Col span="4">{{$t('m_type')}}</Col>
            </Row>
            <Row class="mb-3">
              <Col span="3">{{$t('m_return_code')}}</Col>
              <Col span="5">
              <Select style="width:90%" v-model="successCode.regulative">
                <Option :value="1" key="m_regular_match">{{ $t('m_regular_match') }}</Option>
                <Option :value="0" key="m_irregular_matching">{{ $t('m_irregular_matching') }}</Option>
              </Select>
              </Col>
              <Col span="5">
              <Input style="width:90%" v-model.trim="successCode.source_value" />
              </Col>
              <Col span="5">
              <Input style="width:90%" v-model.trim="successCode.target_value" />
              </Col>
              <Col span="4">
              <div>{{$t('m_success')}}</div>
              </Col>
            </Row>
          </div>
          </Col>
        </Row>
      </div>
      <div slot="footer">
        <Button @click="showModal = false">{{ $t('m_button_cancel') }}</Button>
        <Button @click="saveConfig" type="primary">{{ $t('m_button_save') }}</Button>
      </div>
    </Modal>
  </div>
</template>

<script>
import {isEmpty, cloneDeep, hasIn} from 'lodash'
import Vue from 'vue'
import {thresholdList, lastList} from '@/assets/config/common-config.js'
import {renderDisplayName} from '@/assets/js/utils'

export const custom_api_enum = [
  {
    getConfigDetailById: 'get'
  }
]

const initRangeConfigMap = {
  req_fail_rate: {
    operator: '>',
    threshold: '10',
    time: '0',
    time_unit: 's'
  },
  req_costtime_avg: {
    operator: '>',
    threshold: '500',
    time: '60',
    time_unit: 's'
  },
  other: {
    operator: '',
    threshold: '',
    time: '',
    time_unit: ''
  }
}

const initSuccessCode = {
  regulative: 0,
  source_value: 200,
  target_value: 'success'
}

export default {
  name: 'standard-regex',
  data() {
    return {
      showModal: false,
      isfullscreen: true,
      isParmasChanged: false,
      isAdd: true,
      configInfo: {},
      columnsForParameterCollection: [
        {
          title: this.$t('m_field_displayName'),
          key: 'display_name',
          width: 120
        },
        {
          title: this.$t('m_parameter_key'),
          key: 'name',
          width: 140
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
                <div domPropsInnerHTML={this.regRes(params.row.regular)} style="word-break: break-all;max-height: 400px;overflow: auto;min-width:200px"></div>
              </div>
              <Input
                value={params.row.regular}
                onInput={v => {
                  this.changeRegex(params.index, v)
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
      ],
      columnsForComputeMetrics: [
        {
          title: this.$t('m_color_system'),
          key: 'color_group',
          width: 100,
          render: (h, params) => (
            <div class="color_system">
              <ColorPicker value={params.row.color_group || ''}
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
          render: (h, params) => (<span>{renderDisplayName(params.row.display_name)}</span>)
        },
        {
          title: this.$t('m_metric_key'),
          key: 'metric',
          width: 140
        },
        {
          title: this.$t('m_statistical_parameters'),
          key: 'log_param_name',
          width: 150,
        },
        {
          title: this.$t('m_filter_label'),
          key: 'tag_config',
          width: 150,
          render: (h, params) => (
            <span>
              {params.row.tag_config.join(',')}
            </span>
          )
        },
        {
          title: this.$t('m_computed_type'),
          key: 'agg_type',
          width: 200,
          render: (h, params) => {
            const agg_type = params.row.agg_type
            return (
              <Tooltip content={agg_type} max-width="300" >
                <span>{agg_type}</span>
              </Tooltip>
            )
          }
        },
        {
          title: this.$t('m_automatic_alert'),
          key: 'auto_alarm',
          width: 60,
          render: (h, params) =>
            (
              <i-switch value={params.row.auto_alarm}
                on-on-change={val => {
                  if (!val) {
                    const key = ['req_fail_rate', 'req_costtime_avg'].includes(params.row.metric) ? params.row.metric : 'other'
                    Vue.set(this.configInfo.metric_list[params.index], 'range_config', cloneDeep(initRangeConfigMap[key]))
                  }
                  this.configInfo.metric_list[params.index].auto_alarm = val
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
        }
      ],
      successCode: cloneDeep(initSuccessCode),
      actionType: '',
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  methods: {
    loadPage(guid, actionType) {
      this.isfullscreen = true
      this.successCode = cloneDeep(initSuccessCode)
      if (guid) {
        if (actionType === 'copy') {
          this.actionType = actionType
          this.isAdd = true
        } else {
          this.isAdd = false
          this.actionType = ''
        }
        this.getConfigDetail(guid)
      } else {
        this.configInfo = {
          guid: '',
          name: '',
          log_type: 'regular',
          demo_log: '',
          calc_result: {
            match_text: '',
            json_key_list: [],
            json_obj: {}
          },
          param_list: [
            {
              guid: '',
              name: 'code',
              display_name: this.$t('m_service_code'),
              json_key: '',
              regular: '',
              demo_match_value: '',
            },
            {
              guid: '',
              name: 'retcode',
              display_name: this.$t('m_return_code'),
              json_key: '',
              regular: '',
              demo_match_value: '',
            },
            {
              guid: '',
              name: 'costtime',
              display_name: this.$t('m_time_consuming'),
              json_key: '',
              regular: '',
              demo_match_value: '',
            }
          ],
          metric_list: [
            {
              log_param_name: 'code',
              metric: 'req_count',
              display_name: this.$t('m_request_volume'),
              agg_type: 'count',
              tag_config: [
                'code'
              ],
              color_group: '#1a94bc',
              auto_alarm: false,
              range_config: cloneDeep(initRangeConfigMap.other)
            },
            {
              log_param_name: 'code',
              metric: 'req_suc_count',
              display_name: this.$t('m_success_volume'),
              agg_type: 'count',
              tag_config: [
                'code'
              ],
              color_group: '#bec936',
              auto_alarm: false,
              range_config: cloneDeep(initRangeConfigMap.other)
            },
            {
              log_param_name: 'code',
              metric: 'req_suc_rate',
              display_name: this.$t('m_success_rate'),
              agg_type: '100*{req_suc_count}/{req_count}',
              tag_config: [
                'code'
              ],
              color_group: '#20a162',
              auto_alarm: false,
              range_config: cloneDeep(initRangeConfigMap.other)
            },
            {
              log_param_name: 'code',
              metric: 'req_fail_count',
              display_name: this.$t('m_failure_volume'),
              agg_type: '{req_count}-{req_suc_count}',
              tag_config: [
                'code'
              ],
              color_group: '#ee3f4d',
              auto_alarm: false,
              range_config: cloneDeep(initRangeConfigMap.other)
            },
            {
              log_param_name: 'code',
              metric: 'req_fail_rate',
              display_name: this.$t('m_failure_rate'),
              agg_type: '100*{req_fail_count_detail}/{req_count}',
              tag_config: [
                'code'
              ],
              color_group: '#7c1823',
              auto_alarm: true,
              range_config: cloneDeep(initRangeConfigMap.req_fail_rate)
            },
            {
              log_param_name: 'code',
              metric: 'req_fail_count_detail',
              display_name: this.$t('m_categorized_failure_count'),
              agg_type: 'count',
              tag_config: [
                'code',
                'retcode'
              ],
              color_group: '#ee3f4d',
              auto_alarm: false,
              range_config: cloneDeep(initRangeConfigMap.other)
            },
            {
              log_param_name: 'costtime',
              metric: 'req_costtime_avg',
              display_name: this.$t('m_average_time'),
              agg_type: 'avg',
              tag_config: [
                'code',
                'retcode'
              ],
              color_group: '#d6a01d',
              auto_alarm: true,
              range_config: cloneDeep(initRangeConfigMap.req_costtime_avg)
            },
            {
              log_param_name: 'costtime',
              metric: 'req_costtime_max',
              display_name: this.$t('m_max_costtime'),
              agg_type: 'max',
              tag_config: [
                'code',
                'retcode'
              ],
              color_group: '#815c94',
              auto_alarm: false,
              range_config: cloneDeep(initRangeConfigMap.other)
            }
          ]
        }
        this.showModal = true
        this.isAdd = true
      }
    },
    // 1
    paramsValidate(tmpData) {
      if (tmpData.name === '') {
        this.$Message.warning(`${this.$t('m_template_name')}${this.$t('m_cannot_be_empty')}`)
        return true
      }

      const isRegularEmpty = tmpData.param_list.some(element => element.regular === '')
      if (isRegularEmpty) {
        this.$Message.warning(`${this.$t('m_extract_regular')}${this.$t('m_cannot_be_empty')}`)
        return true
      }

      const isMatchValueEmpty = tmpData.param_list.some(element => element.demo_match_value === '')
      if (isMatchValueEmpty) {
        this.$Message.warning(`${this.$t('m_matching_result')}${this.$t('m_cannot_be_empty')}`)
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
            if (!this.isNumericString(item.range_config.threshold + '')) {
              this.$Message.warning(`${this.$t('m_threshold_tips')}`)
              return true
            }
            if (!this.isPositiveNumericString(item.range_config.time + '')) {
              this.$Message.warning(`${this.$t('m_time_tips')}`)
              return true
            }
          }
        }
      }
      if (!isEmpty(this.successCode) && (isEmpty(this.successCode.regulative + '') || isEmpty(this.successCode.source_value + '') || isEmpty(this.successCode.target_value))) {
        this.$Message.warning(`${this.$t('m_return_code')}${this.$t('m_cannot_be_empty')}`)
        return true
      }
      // if (!hasIn(tmpData, 'calc_result.match_text') || isEmpty(tmpData.calc_result.match_text)) {
      //   this.$Message.warning(`${this.$t('m_matching_result')}${this.$t('m_cannot_be_empty')}`)
      //   return true
      // }

      return false
    },
    async saveConfig() {
      const tmpData = JSON.parse(JSON.stringify(this.configInfo))
      if (this.paramsValidate(tmpData)) {
        return
      }
      this.processSaveData(tmpData)
      delete tmpData.create_user
      delete tmpData.create_time
      delete tmpData.update_user
      delete tmpData.update_time
      // const methodType = this.isAdd ? 'POST' : 'PUT'
      if (this.actionType === 'copy') {
        delete tmpData.guid
      }

      if (this.isAdd) {
        await this.request('POST', this.apiCenter.logTemplateConfig, tmpData)
      } else {
        await this.request('PUT', this.apiCenter.logTemplateConfig, tmpData)
      }
      this.$Message.success(this.$t('m_tips_success'))
      this.showModal = false
      this.$emit('refreshData')
    },
    processSaveData(data){
      if (isEmpty(data)) {return}
      this.successCode.source_value = this.successCode.source_value + ''
      data.success_code = JSON.stringify(this.successCode)
      !isEmpty(data.metric_list) && data.metric_list.forEach(item => {
        item.range_config.threshold += ''
        item.range_config.time += ''
        item.range_config = JSON.stringify(item.range_config)
      })
    },
    getConfigDetail(guid) {
      const api = this.apiCenter.getConfigDetailByGuid + guid
      this.request('GET', api, {}, resp => {
        this.configInfo = resp
        this.configInfo.param_list.forEach(r => {
          r.regular_font_result = this.regRes(r.regular)
        })
        this.processConfigInfo()
        if (this.actionType === 'copy') {
          this.configInfo.name += '(1)'
        }
        this.showModal = true
      })
    },
    processConfigInfo() {
      !isEmpty(this.configInfo.metric_list) && this.configInfo.metric_list.forEach(item => {
        Vue.set(item, 'range_config', isEmpty(item.range_config) ? cloneDeep(initRangeConfigMap.other) : JSON.parse(item.range_config))
        Vue.set(item, 'auto_alarm', hasIn(item, 'auto_alarm') ? item.auto_alarm : false)
        Vue.set(item, 'color_group', isEmpty(item.color_group) ? '' : item.color_group)
      })
      this.successCode = !isEmpty(this.configInfo.success_code) ? JSON.parse(this.configInfo.success_code) : cloneDeep(initSuccessCode)
    },
    changeRegex(index, val) {
      this.configInfo.param_list[index].regular = val
      this.configInfo.param_list[index].regular_font_result = this.regRes(val)
    },
    regRes(val) {
      // 测试数据
      {/* {"time":"2024-08-12 10:04:48","url":"api/v1/ser/user/update","service_code":"updateUser","return_code":"200","cost_ms":0.819}
      .*"service_code":"(.*)","retu.* */}
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
      const params = {
        demo_log: this.configInfo.demo_log,
        param_list: this.configInfo.param_list
      }
      this.request('POST', this.apiCenter.standardLogRegexMatch, params, responseData => {
        this.configInfo.param_list = responseData
        this.$Message.success(this.$t('m_success'))
      }, {isNeedloading: false})
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
    },
    isNumericString(str) {
      return !isNaN(str) && !isNaN(parseFloat(str))
    },
    isPositiveNumericString(str) {
      return /^\d+$/.test(str) && parseFloat(str) >= 0
    }
  }
}
</script>

<style lang="less">
.calculation-indicators {
  .ivu-table-cell {
    padding-left: 2px;
    padding-right: 2px;
  }
}
.color_system {
  width: 40px
}
.ivu-table-wrapper {
  overflow: inherit;
}
</style>

<style lang="less" scoped>
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
.success-code {
  display: flex;
  align-items: center;
  justify-content: flex-start
}
.success-code > div {
  margin-right: 10px;
  margin-bottom: 20px
}
</style>
