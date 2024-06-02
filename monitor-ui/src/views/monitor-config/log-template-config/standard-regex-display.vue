<template>
  <Row>
    <Col span="8" style="border-right: 2px solid rgb(232 234 236)">
      <Form :label-width="120">
        <FormItem :label="$t('m_template_name')">
          <Tooltip :content="configInfo.name" transfer :disabled="configInfo.name === ''" style="width: 100%;" max-width="200">
            <Input
              v-model="configInfo.name"
              maxlength="30"
              show-word-limit
              style="width: 96%"
              disabled
            />
            <span style="color: red">*</span>
          </Tooltip>
        </FormItem>
        <FormItem>
          <Button type="primary" @click="showTemplate = !showTemplate" ghost size="small" style="bottom: 72px;right: 36px;position: relative;">{{showTemplate ? $t('m_hide_template'):$t('m_expand_template')}}</Button>
        </FormItem>
        <template v-if="showTemplate===true">
          <FormItem :label="$t('m_updatedBy')">
            {{ configInfo.update_user }}
          </FormItem>
          <FormItem :label="$t('title.updateTime')">
            {{ configInfo.update_time }}
          </FormItem>
        </template>
        <FormItem :label="$t('m_log_example')" v-if="showTemplate===true">
          <Input
            v-model="configInfo.demo_log"
            type="textarea"
            :rows="11"
            style="width: 96%"
            disabled
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
        <div  v-if="showTemplate===true">
          <Divider orientation="left" size="small">{{ $t('m_parameter_collection') }}</Divider>
          <Table
            style="position: inherit;"
            size="small"
            :columns="columnsForParameterCollection"
            :data="configInfo.param_list"
            width="100%"
          ></Table>
          <!-- <Button type="primary" @click="generateBackstageTrial" ghost size="small" style="float:right;margin:12px">{{ $t('m_match') }}</Button> -->
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
</template>

<script>
export default {
  name: "standard-regex",
  data() {
    return {
      isParmasChanged: false,
      showTemplate: false,
      isAdd: true,
      // configInfo: {
      //   guid: '',
      //   name: '',
      //   log_type: 'regular',
      //   demo_log: '',
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
      //         'retcode'
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
          renderHeader: () => {
            return (
              <span>
                <span style="color:red">*</span>
                <span>{this.$t('m_extract_regular')}</span>
              </span>
            )
          },
          render: (h, params) => {
            return (
              <Input
                value={params.row.regular}
                disabled
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
        },
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
          width: 140,
          render: (h, params) => {
            return (
              <span>
                <span>{this.prefixCode}_{params.row.metric}</span>
              </span>
            )
          }
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
      ]
    }
  },
  props: {
    configInfo: Object,
    prefixCode: String
  },
  mounted () {},
  methods: {
    hideTemplate () {
      this.showTemplate = false
      this.isfullscreen = false
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