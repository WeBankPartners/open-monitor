<template>
  <div>
    <Row>
      <Col span="8" style="max-height: 510px;overflow: auto;">
        <Form :label-width="120">
          <FormItem :label="$t('m_template_name')">
            <Input
              v-model="configInfo.name"
              maxlength="30"
              show-word-limit
              style="width: 96%"
              disabled
            />
            <span style="color: red">*</span>
            <div
              v-if="(configInfo.name.length === 0 || configInfo.name.length > 30)"
              style="color: red"
            >
              {{ $t('m_template_name') }}{{ $t('tw_limit_30') }}
            </div>
          </FormItem>
          <FormItem>
            <Button type="primary" @click="showTemplate = !showTemplate" ghost size="small" style="float:right;margin:12px">{{showTemplate ? $t('m_hide_template'):$t('m_expand_template')}}</Button>
          </FormItem>
          <template v-if="showTemplate===true" >
            <FormItem :label="$t('m_updatedBy')">
              {{ configInfo.update_user }}
            </FormItem>
            <FormItem :label="$t('title.updateTime')">
              {{ configInfo.update_time }}
            </FormItem>
          </template>
          <FormItem  v-if="showTemplate===true" :label="$t('m_json_regular')" style="margin-bottom: 12px;">
            <Input
              v-model="configInfo.json_regular"
              maxlength="200"
              show-word-limit
              type="textarea"
              style="width: 96%"
              disabled
            />
            <div v-if="isParmasChanged && configInfo.json_regular.length > 200" style="color: red">
              {{ $t('m_json_regular') }}{{ $t('tw_limit_200') }}
            </div>
          </FormItem>
          <FormItem  v-if="showTemplate===true" :label="$t('m_log_example')">
            <Input
              v-model="configInfo.demo_log"
              type="textarea"
              :rows="6"
              style="width: 96%"
              disabled
            />
            <div v-if="isParmasChanged && configInfo.demo_log.length === 0" style="color: red">
              {{ $t('m_log_example') }} {{ $t('tips.required') }}
            </div>
          </FormItem>
          <FormItem>
            <!-- <Button type="primary" @click="confirmGenerateBackstageTrial" ghost size="small" style="float:right;margin:12px" :disabled="configInfo.demo_log===''||configInfo.json_regular===''">{{ $t('m_match') }}</Button> -->
          </FormItem>
          <FormItem  v-if="showTemplate===true" :label="$t('m_matching_result')" style="margin-top: 12px;">
            <Input
              disabled
              v-model="configInfo.calc_result.match_text"
              type="textarea"
              :rows="6"
              style="width: 96%"
            />
          </FormItem>
        </Form>
      </Col>
      <Col span="16" style="border-left: 2px solid rgb(232 234 236)">
        <div style="margin-left: 8px;">
          <!-- 采集参数 -->
          <div v-if="showTemplate===true">
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
</template>

<script>
export default {
  name: "json-regex",
  data() {
    return {
      isParmasChanged: false,
      isAdd: false,
      showTemplate: false,
      // configInfo: {
      //   guid: '',
      //   name: '',
      //   log_type: 'json',
      //   demo_log: '',
      //   json_regular: '',
      //   calc_result: {
      //     match_text: '',
      //     json_key_list: [],
      //     json_obj: {}
      //   },
      //   param_list: [
      //     {
      //       guid: '',
      //       name: 'code',
      //       display_name: this.$t('m_service_code'),
      //       json_key: '',
      //       regular: '',
      //       demo_match_value: '',
      //     },
      //     {
      //       guid: '',
      //       name: 'retcode',
      //       display_name: this.$t('m_return_code'),
      //       json_key: '',
      //       regular: '',
      //       demo_match_value: '',
      //     },
      //     {
      //       guid: '',
      //       name: 'costtime',
      //       display_name: this.$t('m_time_consuming'),
      //       json_key: '',
      //       regular: '',
      //       demo_match_value: '',
      //     }
      //   ],
      //   metric_list: [
      //     {
      //       log_param_name: 'code',
      //       metric: 'req_count',
      //       display_name: this.$t('m_request_volume'),
      //       agg_type: 'count',
      //       tag_config: [
      //         'code'
      //       ]
      //     },
      //     {
      //       log_param_name: 'code',
      //       metric: 'req_suc_count',
      //       display_name: this.$t('m_success_volume'),
      //       agg_type: 'count',
      //       tag_config: [
      //         'code',
      //         'retcode'
      //       ]
      //     },
      //     {
      //       log_param_name: 'code',
      //       metric: 'req_suc_rate',
      //       display_name: this.$t('m_success_rate'),
      //       agg_type: '100*{req_suc_count}/{req_count}',
      //       tag_config: [
      //         'code',
      //         'retcode'
      //       ]
      //     },
      //     {
      //       log_param_name: 'code',
      //       metric: 'req_fail_rate',
      //       display_name: this.$t('m_failure_rate'),
      //       agg_type: '100-100*{req_suc_count}/{req_count}',
      //       tag_config: [
      //         'code',
      //         'etcode'
      //       ]
      //     },
      //     {
      //       log_param_name: 'costtime',
      //       metric: 'req_costtime_avg',
      //       display_name: this.$t('m_average_time'),
      //       agg_type: 'avg',
      //       tag_config: [
      //         'code',
      //         'retcode'
      //       ]
      //     },
      //     {
      //       log_param_name: 'costtime',
      //       metric: 'req_costtime_max',
      //       display_name: this.$t('m_max_costtime'),
      //       agg_type: 'max',
      //       tag_config: [
      //         'code',
      //         'retcode'
      //       ]
      //     }
      //   ],
      //   // create_user: '',
      //   // create_time: '',
      //   // update_user: '',
      //   // update_time: ''
      // },
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
          title: this.$t('m_json_key'),
          key: 'json_key',
          render: (h, params) => {
            const selectOptions = this.configInfo.calc_result.json_key_list
            return (
              <Select
                style="'z-index: 1000'"
                disabled
                value={params.row.json_key}
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
      ],
      generateBackstageTrialWarning: false
    }
  },
  props: {
    configInfo: Object,
  },
  mounted () {
    this.showTemplate = false
  },
  methods: {
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