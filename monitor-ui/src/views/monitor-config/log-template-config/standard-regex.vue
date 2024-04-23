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
          <Col span="8" style="border-right: 2px solid rgb(232 234 236)">
            <Form :label-width="120">
              <FormItem :label="$t('m_template_name')">
                <Input
                  v-model="configInfo.name"
                  maxlength="30"
                  show-word-limit
                  style="width: 96%"
                />
                <span style="color: red">*</span>
                <div
                  v-if="isParmasChanged && (configInfo.name.length === 0 || configInfo.name.length > 30)"
                  style="color: red"
                >
                  {{ $t('m_template_name') }}{{ $t('tw_limit_30') }}
                </div>
              </FormItem>
              <template v-if="!isAdd">
                <FormItem :label="$t('m_updatedBy')">
                  {{ configInfo.update_user }}
                </FormItem>
                <FormItem :label="$t('title.updateTime')">
                  {{ configInfo.update_time }}
                </FormItem>
              </template>
              <FormItem :label="$t('m_log_example')">
                <Input
                  v-model="configInfo.demo_log"
                  type="textarea"
                  :rows="6"
                  style="width: 96%"
                />
                <div v-if="isParmasChanged && configInfo.demo_log.length === 0" style="color: red">
                  {{ $t('m_log_example') }} {{ $t('tips.required') }}
                </div>
              </FormItem>
            </Form>
          </Col>
          <Col span="16">
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
                <Button type="primary" @click="generateBackstageTrial" ghost size="small" style="float:right;margin:12px">{{ $t('m_match') }}</Button>
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
  </div>
</template>

<script>
export default {
  name: "standard-regex",
  data() {
    return {
      showModal: false,
      isfullscreen: false,
      isParmasChanged: false,
      isAdd: true,
      configInfo: {
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
              'etcode'
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
        ],
        // create_user: '',
        // create_time: '',
        // update_user: '',
        // update_time: ''
      },
      columnsForParameterCollection: [
        {
          title: this.$t('field.displayName'),
          key: 'display_name',
        },
        {
          title: this.$t('m_parameter_key'),
          key: 'name',
        },
        {
          title: this.$t('m_extract_regular'),
          key: 'regular',
          render: (h, params) => {
            return (
              <Input
                value={params.row.regular}
                onInput={v => {
                  this.changeRegex(params.index, v)
                }}
              />
            )
          }
        },
        {
          title: this.$t('m_matching_result'),
          ellipsis: true,
          tooltip: true,
          key: 'demo_match_value',
        },
      ],
      columnsForComputeMetrics: [
        {
          title: this.$t('field.displayName'),
          key: 'display_name',
        },
        {
          title: this.$t('m_metric_key'),
          key: 'metric',
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
        }
      ]
    }
  },
  methods: {
    loadPage (guid) {
      if (guid) {
        this.isAdd = false
        this.getConfigDetail(guid)
      } else {
        this.showModal = true
        this.isAdd = true
        console.log('新增')
      }
    },
    saveConfig () {
      let tmpData = JSON.parse(JSON.stringify(this.configInfo))
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
    changeRegex (index, val) {
      this.configInfo.param_list[index].regular = val
    },
    generateBackstageTrial () {
      const params = {
        demo_log: this.configInfo.demo_log,
        param_list: this.configInfo.param_list
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.standardLogRegexMatch, params, (responseData) => {
        this.configInfo.param_list = responseData
      }, {isNeedloading:false})
    },
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