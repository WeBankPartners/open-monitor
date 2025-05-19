<template>
  <div>
    <Modal v-model="showModel" :title="$t('m_menu_configuration')" :mask-closable="false" :width="1100" :fullscreen="isfullscreen">
      <div slot="header" class="custom-modal-header">
        <span>
          {{(view ? $t('m_button_view') : (isAdd ? $t('m_button_add') : $t('m_modify'))) + $t('m_menu_configuration')}}
        </span>
        <Icon v-if="isfullscreen" @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-expand" />
      </div>
      <div :class="isfullscreen ? 'modal-container-fullscreen' : 'modal-container-normal'">
        <div class="w-header" slot="title">
          <div class="title">
            {{ $t('m_configuration_information') }}
            <span class="underline"></span>
          </div>
        </div>
        <Row>
          <Col span="8">
          <Form :label-width="120">
            <FormItem :label="$t('m_tableKey_name')">
              <Tooltip :content="businessConfig.name" transfer :disabled="businessConfig.name === ''" max-width="200" style="width: 100%;">
                <Input v-model.trim="businessConfig.name" :disabled="view" maxlength="30" show-word-limit style="width:96%"></Input>
                <span style="color: red">*</span>
              </Tooltip>
            </FormItem>
            <FormItem :label="$t('m_metric_code') ">
              <Input
                v-model.trim="businessConfig.metric_prefix_code"
                maxlength="15"
                :disabled="!isAdd || view"
                show-word-limit
                :placeholder="$t('m_metric_code_placeholder')"
                style="width:96%"
              >
              </Input>
              <span style="color: red">*</span>
            </FormItem>
          </Form>
          </Col>
          <Col span="16" style="border-left: 2px solid rgb(232 234 236)">
          <div style="padding: 0 8px">
            <div>
              <Row>
                <Col span="3" style="margin-top: 30px">
                <span style="color:#5cadff">{{ $t('m_service_code') }}</span>
                </Col>
                <Col span="21">
                <Row>
                  <Col span="3">{{ $t('m_match_type') }}</Col>
                  <Col span="3">
                  <span style="color:red">*</span>
                  {{ $t('m_source_value') }}</Col>
                  <Col span="4">
                  {{ $t('m_matching_source_value') }}
                  </Col>
                  <Col span="4">
                  {{ $t('m_matching_result_test') }}
                  </Col>
                  <Col span="4">
                  <span style="color:red">*</span>
                  {{ $t('m_match_value') }}
                  </Col>
                  <Col span="2">
                  {{ $t('m_field_type') }}</Col>
                  <Col span="1"></Col>
                </Row>
                <Row v-for="(item, itemIndex) in businessConfig.code_string_map" :key="itemIndex" class='action-row'>
                  <Col span="3">
                  <Select v-model="item.regulative"
                          :disabled="view"
                          style="width:90%"
                          @on-change='onRegulativeChange(item)'
                  >
                    <Option :value="1" key="m_regular_match">{{ $t('m_regular_match') }}</Option>
                    <Option :value="0" key="m_irregular_matching">{{ $t('m_irregular_matching') }}</Option>
                  </Select>
                  </Col>
                  <Col span="3">
                  <Input v-model.trim="item.source_value"
                         :disabled="view"
                         @on-change="(e) => onSourceValueChange(e, item)"
                         @on-blur="refreshPage"
                         style="width:90%"
                  >
                  </Input>
                  </Col>
                  <Col span="4">
                  <Input v-model.trim="item.matchingSourceValue"
                         :disabled="view || item.regulative === 0"
                         style="width:90%"
                         @on-change="debounceRefresh"
                         @on-blur="refreshPage"
                  >
                  </Input>
                  </Col>

                  <Col span="4" :key='rowKey'>
                  <span>{{ item.regulative === 0 ? '-' : (item.matchingResult ? $t('m_matching_success') : $t('m_matching_failed'))}}</span>
                  <Button
                    v-if="item.regulative === 1"
                    type="info"
                    ghost
                    size="small"
                    :disabled="view || !item.source_value || !item.matchingSourceValue || item.regulative === 0"
                    @click="onMatchButtonClick(item)"
                  >
                    {{$t('m_match')}}
                  </Button>
                  </Col>

                  <Col span="4">
                  <Input
                    :disabled="view || (!item.matchingResult && item.regulative === 1)"
                    v-model.trim="item.target_value"
                    style="width:90%"
                  >
                  </Input>
                  </Col>
                  <Col span="2" offset="2">
                  <Button
                    type="error"
                    ghost
                    @click="deleteItem('code_string_map',itemIndex)"
                    size="small"
                    style="vertical-align: sub;cursor: pointer"
                    icon="md-trash"
                    :disabled="view"
                  ></Button>
                  </Col>
                </Row>
                </Col>
              </Row>
              <Row>
                <Col span="21" offset="3">
                <Row>
                  <Col span="2" offset="20">
                  <div style="cursor: pointer">
                    <Button type="success" :disabled="view" ghost @click="addItem('code_string_map')" size="small" icon="md-add"></Button>
                  </div>
                  </Col>
                </Row>
                </Col>
              </Row>
            </div>
            <div>
              <Row>
                <Col span="3" style="margin-top: 12px">
                <span style="color:#5cadff">{{ $t('m_return_code') }}</span>
                </Col>
                <Col span="21">
                <Row v-for="(item, itemIndex) in businessConfig.retcode_string_map" :key="itemIndex" class='action-row'>
                  <Col span="3">
                  <Select v-model="item.regulative"
                          :disabled="view || retcodeItemDisabled(item, itemIndex)"
                          style="width:90%"
                          @on-change='onRegulativeChange(item)'
                  >
                    <Option :value="1" key="m_regular_match">{{ $t('m_regular_match') }}</Option>
                    <Option :value="0" key="m_irregular_matching">{{ $t('m_irregular_matching') }}</Option>
                  </Select>
                  </Col>
                  <Col span="3">
                  <Input v-model.trim="item.source_value"
                         :disabled="view || retcodeItemDisabled(item, itemIndex)"
                         @on-change="(e) => onSourceValueChange(e, item)"
                         style="width:90%"
                         @on-blur="refreshPage"
                  >
                  </Input>
                  </Col>

                  <Col span="4">
                  <Input v-model.trim="item.matchingSourceValue"
                         :disabled="view || retcodeItemDisabled(item, itemIndex) || item.regulative === 0"
                         style="width:90%"
                         @on-change="debounceRefresh"
                         @on-blur="refreshPage"
                  >
                  </Input>
                  </Col>
                  <Col span="4" :key='rowKey'>
                  <span>{{item.regulative === 0 ? '-' : (item.matchingResult ? $t('m_matching_success') : $t('m_matching_failed'))}}</span>
                  <Button
                    v-if="item.regulative === 1"
                    type="info"
                    ghost
                    size="small"
                    :disabled="view || retcodeItemDisabled(item, itemIndex) || !item.source_value || !item.matchingSourceValue || item.regulative === 0"
                    @click="onMatchButtonClick(item)"
                  >
                    {{$t('m_match')}}
                  </Button>
                  </Col>

                  <Col span="4">
                  <Input v-model.trim="item.target_value" :disabled="view || retcodeItemDisabled(item, itemIndex) || (!item.matchingResult && item.regulative === 1)" style="width:90%"></Input>
                  </Col>
                  <Col span="2">
                  <span style="line-height: 32px;">{{ $t('m_' + item.value_type) }}</span>
                  </Col>
                  <Col span="2">
                  <Button
                    type="error"
                    ghost
                    v-if="itemIndex !== 0"
                    @click="deleteItem('retcode_string_map',itemIndex)"
                    size="small"
                    :disabled="view"
                    style="cursor: pointer"
                    icon="md-trash"
                  ></Button>
                  </Col>
                </Row>
                </Col>
              </Row>
              <Row>
                <Col span="21" offset="3">
                <Row>
                  <Col span="2" offset="20">
                  <div style="cursor: pointer">
                    <Button type="success" :disabled="view" ghost @click="addItem('retcode_string_map')" size="small" icon="md-add"></Button>
                  </div>
                  </Col>
                </Row>
                </Col>
              </Row>
            </div>
          </div>
          </Col>
        </Row>
        <div class="w-header" slot="title">
          <div class="title">
            {{ $t('m_associated_template') }}
            <span class="underline"></span>
          </div>
          <Button type="primary" @click="changeTemplateStatus" ghost size="small" >{{templeteStatus ? $t('m_hide_template') : $t('m_expand_template')}}</Button>
        </div>
        <div>
          <StandardRegexDisplay
            ref="standardRegexDisplayRef"
            v-if="configInfo.log_type === 'regular'"
            :prefixCode="businessConfig.metric_prefix_code"
            :configInfo="configInfo"
          ></StandardRegexDisplay>
          <JsonRegexDisplay ref="jsonRegexDisplayRef"
                            v-if="configInfo.log_type === 'json'"
                            :prefixCode="businessConfig.metric_prefix_code"
                            :configInfo="configInfo"
          ></JsonRegexDisplay>
        </div>
      </div>
      <template slot='footer'>
        <Checkbox v-if="['add', 'copy'].includes(actionType)" v-model="auto_create_warn">{{$t('m_auto_create_warn')}}</Checkbox>
        <Checkbox v-if="['add', 'copy'].includes(actionType)" v-model="auto_create_dashboard">{{$t('m_auto_create_dashboard')}}</Checkbox>
        <Button @click="showModel = false">{{ $t('m_button_cancel') }}</Button>
        <Button :disabled="view" @click="saveConfig" type="primary">{{ $t('m_button_save') }}</Button>
      </template>
    </Modal>
  </div>
</template>

<script>
import {
  cloneDeep, isEmpty, hasIn, debounce, countBy, some, keys, pickBy
} from 'lodash'
import Vue from 'vue'
import StandardRegexDisplay from '@/views/monitor-config/log-template-config/standard-regex-display.vue'
import JsonRegexDisplay from '@/views/monitor-config/log-template-config/json-regex-display.vue'

export const custom_api_enum = [
  {
    getConfigDetailById: 'get'
  },
  {
    getLogMetricConfigById: 'get'
  }
]
const initRangeConfig = {
  operator: '>',
  threshold: 3,
  time: 60,
  time_unit: 's'
}

export default {
  name: '',
  data() {
    return {
      showModel: false,
      isAdd: true,
      view: false,
      isfullscreen: true,
      parentGuid: '', // 上级唯一标识
      configInfo: {
        log_type: ''
      },
      businessConfig: {},
      templeteStatus: false,
      rowKey: '',
      template_version: '',
      auto_create_warn: true,
      auto_create_dashboard: true,
      templateRetCode: {},
      actionType: '',
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  methods: {
    loadPage(actionType, templateGuid, parentGuid, configGuid) {
      this.templeteStatus = false
      this.isfullscreen = true
      this.parentGuid = parentGuid
      // actionType add/edit
      // templateGuid, 模版id
      // parentGuid, 上级唯一标识
      // configGuid, 配置唯一标志
      this.isAdd = ['add', 'copy'].includes(actionType)
      this.actionType = actionType
      this.view = actionType === 'view'
      this.businessConfig.log_monitor_template_guid = templateGuid
      this.businessConfig.log_metric_monitor_guid = parentGuid
      // this.auto_create_warn = true
      // this.auto_create_dashboard = true
      if (configGuid) {
        this.getConfig(configGuid)
      } else {
        this.businessConfig = {
          name: '', // 名称
          metric_prefix_code: '', // 指标编码 1到6个字符的字母、数字、下划线或短横线
          log_metric_monitor_guid: '',
          log_monitor_template_guid: '',
          code_string_map: [
            {
              regulative: 0,
              source_value: '',
              target_value: ''
            }
          ],
          retcode_string_map: [
            {
              regulative: 0,
              source_value: '',
              target_value: '',
              value_type: 'success'
            }
          ]
        }
        if (templateGuid) {
          this.businessConfig.log_monitor_template_guid = templateGuid
          this.businessConfig.log_metric_monitor_guid = parentGuid
          this.getTemplateDetail(templateGuid)
        }
      }
      this.$refs.standardRegexDisplayRef && this.$refs.standardRegexDisplayRef.hideTemplate()
      this.$refs.jsonRegexDisplayRef && this.$refs.jsonRegexDisplayRef.hideTemplate()
      this.showModel = true
    },
    changeTemplateStatus() {
      this.$refs.standardRegexDisplayRef && this.$refs.standardRegexDisplayRef.changeTemplateStatus()
      this.$refs.jsonRegexDisplayRef && this.$refs.jsonRegexDisplayRef.changeTemplateStatus()
      this.templeteStatus = this.getTemplateStatus() || false
    },
    getTemplateStatus() {
      const displayRef = this.configInfo.log_type==='regular' ? 'standardRegexDisplayRef' : 'jsonRegexDisplayRef'
      const res = this.$refs[displayRef] && this.$refs[displayRef].returnCurrentStatus()
      return res
    },
    processConfigInfo(res) {
      this.configInfo = res
      !isEmpty(this.configInfo.metric_list) && this.configInfo.metric_list.forEach(item => {
        Vue.set(item, 'range_config', isEmpty(item.range_config) ? cloneDeep(initRangeConfig) : JSON.parse(item.range_config))
      })
      this.templateRetCode = isEmpty(res.success_code) ? {} : JSON.parse(res.success_code)
    },
    getTemplateDetail(guid) {
      const api = this.apiCenter.getConfigDetailByGuid + guid
      this.request('GET', api, {}, resp => {
        this.auto_create_dashboard = hasIn(resp, 'auto_create_dashboard') ? resp.auto_create_dashboard : true
        this.auto_create_warn = hasIn(resp, 'auto_create_warn') ? resp.auto_create_warn : true
        this.processConfigInfo(resp)
        Object.assign(this.businessConfig.retcode_string_map[0], JSON.parse(resp.success_code))
      })
    },
    getConfig(guid) {
      const api = this.apiCenter.getLogMetricConfig + guid
      this.request('GET', api, {}, resp => {
        this.businessConfig = resp
        this.processConfigInfo(resp.log_monitor_template)
        if (this.actionType === 'copy') {
          this.businessConfig.name += '1'
          this.businessConfig.metric_prefix_code += '1'
        }
        this.configInfo.log_monitor_template_version = resp.log_monitor_template_version

        Array.isArray(this.businessConfig.code_string_map) && this.businessConfig.code_string_map.forEach((item, index) => {
          Vue.set(this.businessConfig.code_string_map[index], 'matchingSourceValue', '')
          Vue.set(this.businessConfig.code_string_map[index], 'matchingResult', false)
          Vue.set(this.businessConfig.code_string_map[index], 'matchingResultText', this.businessConfig.code_string_map[index].source_value)
        })
        Array.isArray(this.businessConfig.retcode_string_map) && this.businessConfig.retcode_string_map.forEach((item, index) => {
          Vue.set(this.businessConfig.retcode_string_map[index], 'matchingSourceValue', '')
          Vue.set(this.businessConfig.retcode_string_map[index], 'matchingResult', false)
          Vue.set(this.businessConfig.retcode_string_map[index], 'matchingResultText', this.businessConfig.retcode_string_map[index].source_value)
        })
      })
    },
    paramsValidate(tmpData) {
      if (tmpData.name === '') {
        this.$Message.warning(`${this.$t('m_tableKey_name')}: ${this.$t('m_cannot_be_empty')}`)
        return true
      }
      if (tmpData.metric_prefix_code === '') {
        this.$Message.warning(`${this.$t('m_metric_code')}: ${this.$t('m_cannot_be_empty')}`)
        return true
      }
      // eslint-disable-next-line no-useless-escape
      // const regex = /^[a-zA-Z0-9_\-]{1,6}$/;
      const regex = /^[A-Za-z][A-Za-z0-9]{0,14}$/

      if (!regex.test(tmpData.metric_prefix_code)) {
        this.$Message.warning(`${this.$t('m_metric_code')}: ${this.$t('m_metric_prefix_code_validate')}`)
        return true
      }
      const isCodeMapEmpty = tmpData.code_string_map.some(element => element.source_value === '' || element.target_value === '')
      if (isCodeMapEmpty) {
        this.$Message.warning(`${this.$t('m_service_code')}: ${this.$t('m_fields_cannot_be_empty')}`)
        return true
      }

      const isRetcodeMapEmpty = tmpData.retcode_string_map.some(element => element.source_value === '' || element.target_value === '')
      if (isRetcodeMapEmpty) {
        this.$Message.warning(`${this.$t('m_return_code')}: ${this.$t('m_fields_cannot_be_empty')}`)
        return true
      }
      if (!isEmpty(tmpData.code_string_map)) {
        const targetValueCount = countBy(tmpData.code_string_map, 'target_value')
        const targetValueHasDuplicates = some(targetValueCount, count => count > 1)
        if (targetValueHasDuplicates) {
          const hasDuplicatedKeys = keys(pickBy(targetValueCount, value => value > 1))
          const warningTips = this.$t('m_service_code') + this.$t('m_match_value_pure') + this.$t('m_cannot_be_repeated') + '; ' + (hasDuplicatedKeys.length ? (this.$t('m_repeated_value') + ': ' + hasDuplicatedKeys.join(';')) : '')
          this.$Message.warning(warningTips)
          return true
        }

        const sourceValueCount = countBy(tmpData.code_string_map, 'source_value')
        const sourceValueHasDuplicates = some(sourceValueCount, count => count > 1)
        if (sourceValueHasDuplicates) {
          const hasDuplicatedKeys = keys(pickBy(sourceValueCount, value => value > 1))
          const warningTips = this.$t('m_service_code') + this.$t('m_source_value') + this.$t('m_cannot_be_repeated') + '; ' + (hasDuplicatedKeys.length ? (this.$t('m_repeated_value') + ': ' + hasDuplicatedKeys.join(';')) : '')
          this.$Message.warning(warningTips)
          return true
        }
      }

      if (!isEmpty(tmpData.retcode_string_map)) {
        const targetValueCount = countBy(tmpData.retcode_string_map, 'target_value')
        const targetValueHasDuplicates = some(targetValueCount, count => count > 1)
        if (targetValueHasDuplicates) {
          const hasDuplicatedKeys = keys(pickBy(targetValueCount, value => value > 1))
          const warningTips = this.$t('m_return_code') + this.$t('m_match_value_pure') + this.$t('m_cannot_be_repeated') + '; ' + (hasDuplicatedKeys.length ? (this.$t('m_repeated_value') + ': ' + hasDuplicatedKeys.join(';')) : '')
          this.$Message.warning(warningTips)
          return true
        }

        const sourceValueCount = countBy(tmpData.retcode_string_map, 'source_value')
        const sourceValueHasDuplicates = some(sourceValueCount, count => count > 1)
        if (sourceValueHasDuplicates) {
          const hasDuplicatedKeys = keys(pickBy(sourceValueCount, value => value > 1))
          const warningTips = this.$t('m_return_code') + this.$t('m_source_value') + this.$t('m_cannot_be_repeated') + '; ' + (hasDuplicatedKeys.length ? (this.$t('m_repeated_value') + ': ' + hasDuplicatedKeys.join(';')) : '')
          this.$Message.warning(warningTips)
          return true
        }
      }
      return false
    },
    processUpdateData(data) {
      !isEmpty(data.log_monitor_template) && !isEmpty(data.log_monitor_template.metric_list) && data.log_monitor_template.metric_list.forEach(item => {
        item.range_config = JSON.stringify(item.range_config)
      })
      if (['add', 'copy'].includes(this.actionType)) {
        data.auto_create_warn = this.auto_create_warn
        data.auto_create_dashboard = this.auto_create_dashboard
      }
    },
    async saveConfig() {
      const tmpData = cloneDeep(this.businessConfig)
      this.processUpdateData(tmpData)
      if (this.paramsValidate(tmpData)) {
        return
      }
      // const methodType = this.isAdd ? 'POST' : 'PUT'
      let res
      if (this.isAdd) {
        res = await this.request('POST', this.apiCenter.logMetricGroup, tmpData)
      } else {
        res = await this.request('PUT', this.apiCenter.logMetricGroup, tmpData)
      }
      // const res = await this.request(methodType, this.apiCenter.logMetricGroup, tmpData)
      const messageTips = this.$t('m_tips_success')
      if (!isEmpty(res) && hasIn(res, 'alarm_list') && hasIn(res, 'custom_dashboard')) {
        const tipOne = isEmpty(res.alarm_list) ? '' : '<br/>' + res.alarm_list.join('<br/>')
        const tipTwo = isEmpty(res.custom_dashboard) ? '' : res.custom_dashboard
        this.$Message.success({
          render: h => h('div', { class: 'add-business-config' }, [
            h('div', {class: 'add-business-config-item'}, [
              h('div', { class: 'add-business-config-item-title' }, this.$t('m_has_create_dashboard') + ':'),
              h('div', {
                domProps: {
                  innerHTML: tipTwo
                }
              })
            ]),
            h('div', { class: 'add-business-config-item' }, [
              h('div', { class: 'add-business-config-item-title' }, this.$t('m_has_create_warn') + ':'),
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
      this.showModel = false
      this.$emit('reloadMetricData', this.parentGuid)

    },
    addItem(key) {
      const params = key === 'code_string_map' ? {
        regulative: 0,
        source_value: '',
        target_value: '',
        matchingSourceValue: '',
        matchingResultText: '',
        matchingResult: false
      } : {
        regulative: 0,
        source_value: '',
        target_value: '',
        value_type: 'fail',
        matchingSourceValue: '',
        matchingResultText: '',
        matchingResult: false
      }
      this.businessConfig[key].push(params)
    },
    deleteItem(key, itemIndex) {
      this.businessConfig[key].splice(itemIndex, 1)
    },
    onMatchButtonClick(item) {
      // const api = '/monitor/api/v2/service/log_metric/data_map/regexp/match'
      const params = {
        content: item.matchingSourceValue,
        regexp: item.source_value,
        is_regexp: item.regulative === 1 ? true : false
      }
      this.request('POST', this.apiCenter.logMetricDataMapMatch, params, res => {
        Vue.set(item, 'matchingResult', res.match)
        this.refreshPage()
      })
    },
    onSourceValueChange(event, item) {
      item.matchingResultText = event.target.value
      item.matchingResult = false
      this.debounceRefresh()
    },
    debounceRefresh: debounce(function (){
      this.refreshPage()
    }, 1000),
    retcodeItemDisabled(item, index = -1) {
      return item.regulative === this.templateRetCode.regulative
        && item.source_value === this.templateRetCode.source_value
          && item.target_value === this.templateRetCode.target_value
            && index === 0
    },
    onRegulativeChange(item) {
      item.matchingSourceValue = ''
      item.matchingResult = false
      this.refreshPage()
    },
    refreshPage() {
      this.rowKey = +new Date() + ''
    }
  },
  components: {
    StandardRegexDisplay,
    JsonRegexDisplay
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

.w-header {
  display: flex;
  align-items: center;
  .title {
    font-size: 16px;
    font-weight: bold;
    margin: 0 10px;
    .underline {
      display: block;
      margin-top: -10px;
      margin-left: -6px;
      width: 100%;
      padding: 0 6px;
      height: 12px;
      border-radius: 12px;
      background-color: #c6eafe;
      box-sizing: content-box;
    }
  }
}

.action-row {
  margin:6px 0;
  display: flex;
  align-items: center;
}

</style>

<style lang="less">
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
    .add-business-config-item-title {
      min-width: 80px;
    }
  }
}
.add-business-config > div {
  max-width: 850px;
  word-wrap: break-word;
  white-space: normal;
}

</style>
