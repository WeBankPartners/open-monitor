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
          {{(isAdd ? $t('button.add') : $t('button.edit')) + $t('m_template')}}
        </span>
        <Icon v-if="isfullscreen" @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-expand" />
      </div>
      <div>
        <Row>
          <Col span="8" :style="{ height: isfullscreen ? '' : '510px' }" style="overflow: auto;">
            <Form :label-width="120">
              <FormItem :label="$t('m_template_name')">
                <Tooltip :content="configInfo.name" transfer :disabled="configInfo.name === ''" style="width: 100%;" max-width="200">
                  <Input
                    v-model="configInfo.name"
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
                <FormItem :label="$t('title.updateTime')">
                  {{ configInfo.update_time }}
                </FormItem>
              </template>
              <FormItem :label="$t('m_json_regular')" style="margin-bottom: 12px;">
                <Tooltip theme="light" transfer placement="bottom" style="width: 100%;" max-width="500">
                  <div slot="content">
                    <div v-html="regRes" style="word-break: break-all;max-height: 400px;overflow: auto;min-width:200px"></div>
                  </div>
                  <Input
                    v-model="configInfo.json_regular"
                    type="textarea"
                    style="width: 96%"
                  />
                  <span style="color: red">*</span>
                </Tooltip>
              </FormItem>
              <FormItem :label="$t('m_log_example')">
                <Input
                  v-model="configInfo.demo_log"
                  type="textarea"
                  :rows="12"
                  style="width: 96%"
                />
                <div v-if="isParmasChanged && configInfo.demo_log.length === 0" style="color: red">
                  {{ $t('m_log_example') }} {{ $t('tips.required') }}
                </div>
              </FormItem>
              <FormItem>
                <Button type="primary" @click="confirmGenerateBackstageTrial" ghost size="small" style="float:right;margin:12px;margin-right:4%" :disabled="configInfo.demo_log===''||configInfo.json_regular===''">{{ $t('m_match') }}</Button>
              </FormItem>
              <FormItem :label="$t('m_matching_result')">
                <Input
                  disabled
                  v-model="configInfo.calc_result.match_text"
                  type="textarea"
                  :rows="12"
                  style="width: 96%"
                />
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
              </div>
              <!-- 计算指标 -->
              <div>
                <Divider orientation="left" size="small">{{ $t('m_compute_metrics') }}</Divider>
                <Table
                  size="small"
                  :columns="columnsForComputeMetrics"
                  :data="configInfo.metric_list"
                  width="100%"
                ></Table>
                
              </div>
            </div>
          </Col>
        </Row>
      </div>
      <div slot="footer">
        <Button @click="showModal = false">{{ $t('button.cancel') }}</Button>
        <Button @click="saveConfig" type="primary">{{ $t('button.save') }}</Button>
      </div>
    </Modal>
    <Modal
      v-model="generateBackstageTrialWarning"
      :title="$t('m_is_start_match')"
      @on-ok="generateBackstageTrial"
      @on-cancel="generateBackstageTrialWarning = false">
      <div class="modal-body" style="padding:30px">
        <div>
          <p style="color: red">{{ $t('m_is_start_match_tip') }}</p>
        </div>
      </div>
    </Modal>
  </div>
</template>

<script>
export default {
  name: "json-regex",
  data() {
    return {
      showModal: false,
      isfullscreen: true,
      isParmasChanged: false,
      isAdd: true,
      configInfo: {
        calc_result: {
          match_text: '',
          json_key_list: [],
          json_obj: {}
        }
      },
      columnsForParameterCollection: [
        {
          title: this.$t('field.displayName'),
          key: 'display_name',
          width: 120
        },
        {
          title: this.$t('m_parameter_key'),
          key: 'name',
          width: 140
        },
        {
          title: this.$t('m_json_key'),
          key: 'json_key',
          renderHeader: () => {
            return (
              <span>
                <span style="color:red">*</span>
                <span>{this.$t('m_json_key')}</span>
              </span>
            )
          },
          render: (h, params) => {
            const selectOptions = this.configInfo.calc_result.json_key_list
            return (
                <Select
                  filterable
                  value={params.row.json_key}
                  on-on-change={(val) => {
                    this.changeJSONKey(params.index, val)
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
          title: this.$t('m_matching_result'),
          ellipsis: true,
          tooltip: true,
          renderHeader: () => {
            return (
              <span>
                <span style="color:red">*</span>
                <span>{this.$t('m_matching_result')}</span>
              </span>
            )
          },
          key: 'demo_match_value',
          render: (h, params) => {
            const demo_match_value = params.row.demo_match_value
            return (
              <Tooltip content={demo_match_value} max-width="300" >
                <span style={demo_match_value?'':'color:#c5c8ce'}>{demo_match_value || this.$t('m_no_matching')}</span>
              </Tooltip>
            )
          }
        }
      ],
      columnsForComputeMetrics: [
        {
          title: this.$t('field.displayName'),
          key: 'display_name',
          width: 120
        },
        {
          title: this.$t('m_metric_key'),
          key: 'metric',
          width: 140
        },
        {
          title: this.$t('m_statistical_parameters'),
          key: 'log_param_name',
        },
        {
          title: this.$t('m_filter_label'),
          key: 'tag_config',
          render: (h, params) => {
            return (
              <span>
                {params.row.tag_config.join(',')}
              </span>
            )
          }
        },
        {
          title: this.$t('m_computed_type'),
          key: 'agg_type',
          render: (h, params) => {
            const agg_type = params.row.agg_type
            return (
              <Tooltip content={agg_type} max-width="300" >
                <span>{agg_type}</span>
              </Tooltip>
            )
          }
        }
      ],
      generateBackstageTrialWarning: false
    }
  },
  computed: {
    regRes: function () {
      try {
        const reg = new RegExp(this.configInfo.json_regular, 'g')
        const match = reg.exec(this.configInfo.demo_log)
        if (match) {
          return this.configInfo.demo_log.replace(match[1], "<span style='color:red'>" + match[1] + '</span>')
        }
        return `<span style='color:#c5c8ce'>${this.$t('m_no_matching')}</span>`
      } catch (err) {
        return `<span style='color:#c5c8ce'>${this.$t('m_no_matching')}</span>`
      }
    }
  },
  methods: {
    loadPage (guid) {
      this.isfullscreen = true
      if (guid) {
        this.isAdd = false
        this.getConfigDetail(guid)
      } else {
        this.configInfo = {
          guid: '',
          name: '',
          log_type: 'json',
          demo_log: '',
          json_regular: '',
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
              ]
            },
            {
              log_param_name: 'code',
              metric: 'req_suc_count',
              display_name: this.$t('m_success_volume'),
              agg_type: 'count',
              tag_config: [
                'code',
                'retcode'
              ]
            },
            {
              log_param_name: 'code',
              metric: 'req_suc_rate',
              display_name: this.$t('m_success_rate'),
              agg_type: '100*{req_suc_count}/{req_count}',
              tag_config: [
                'code',
                'retcode'
              ]
            },
            {
              log_param_name: 'code',
              metric: 'req_fail_rate',
              display_name: this.$t('m_failure_rate'),
              agg_type: '100-100*{req_suc_count}/{req_count}',
              tag_config: [
                'code',
                'retcode'
              ]
            },
            {
              log_param_name: 'costtime',
              metric: 'req_costtime_avg',
              display_name: this.$t('m_average_time'),
              agg_type: 'avg',
              tag_config: [
                'code',
                'retcode'
              ]
            },
            {
              log_param_name: 'costtime',
              metric: 'req_costtime_max',
              display_name: this.$t('m_max_costtime'),
              agg_type: 'max',
              tag_config: [
                'code',
                'retcode'
              ]
            }
          ]
        }
        this.showModal = true
        this.isAdd = true
      }
    },
    paramsValidate (tmpData) {
      if (tmpData.name === '') {
        this.$Message.warning(`${this.$t('m_template_name')}${this.$t('m_cannot_be_empty')}`)
        return true
      }
      if (tmpData.json_regular === '') {
        this.$Message.warning(`${this.$t('m_json_regular')}${this.$t('m_cannot_be_empty')}`)
        return true
      }
      const isJsonKeyEmpty = tmpData.param_list.some((element) => {
        return element.json_key === ''
      })
      if (isJsonKeyEmpty) {
        this.$Message.warning(`${this.$t('m_json_key')}${this.$t('m_cannot_be_empty')}`)
        return true
      }
      const is_demo_match_value = tmpData.param_list.some((element) => {
        return element.demo_match_value === ''
      })
      if (is_demo_match_value) {
        this.$Message.warning(`${this.$t('m_matching_result')}${this.$t('m_cannot_be_empty')}`)
        return true
      }
      
      const hasDuplicatesJsonKey = tmpData.param_list.some((element, index) => {
        return tmpData.param_list.findIndex((item) => item.json_key === element.json_key) !== index
      })
      if (hasDuplicatesJsonKey) {
        this.$Message.warning(`${this.$t('m_json_key')}${this.$t('m_cannot_be_repeated')}`)
        return true
      }
      return false
    },

    saveConfig () {
      let tmpData = JSON.parse(JSON.stringify(this.configInfo))
      if (this.paramsValidate(tmpData)) return
      delete tmpData.create_user
      delete tmpData.create_time
      delete tmpData.update_user
      delete tmpData.update_time
      let methodType = this.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(methodType, this.$root.apiCenter.logTemplateConfig, tmpData, () => {
        this.$Message.success(this.$t('tips.success'))
        this.showModal = false
        this.$emit('refreshData')
      })
    },
    getConfigDetail(guid) {
      const api = this.$root.apiCenter.getConfigDetailByGuid + guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, {}, (resp) => {
        this.configInfo = resp
        this.showModal = true
      })
    },
    // 后台试算
    confirmGenerateBackstageTrial () {
      if (this.isAdd) {
        this.generateBackstageTrial()
      } else {
        this.generateBackstageTrialWarning = true
      }
    },
    generateBackstageTrial () {
      this.configInfo.calc_result.match_text = ''
      this.configInfo.calc_result.json_key_list = []
      this.configInfo.calc_result.json_obj = {}
      const params = {
        reg_string: this.configInfo.json_regular,
        test_context: this.configInfo.demo_log,
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v2/regexp/test/match', params, (responseData) => {
        this.configInfo.calc_result.match_text = responseData.match_text
        this.configInfo.calc_result.json_key_list = responseData.json_key_list || []
        this.configInfo.calc_result.json_obj = responseData.json_obj || {}
        this.configInfo.param_list.forEach(item => {
          item.json_key = ''
          item.demo_match_value = ''
        })
        this.$Message.success(this.$t('m_success'))
      }, {isNeedloading:false})
    },
    changeJSONKey (index, val) {
      this.configInfo.param_list[index].json_key = val
      this.configInfo.param_list[index].demo_match_value = ((this.configInfo.calc_result.json_obj[val] || '') + '')
    }
  }
}
</script>

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
</style>