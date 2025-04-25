<template>
  <Row>
    <Col span="8" style="border-right: 2px solid rgb(232 234 236)">
    <Form :label-width="120" style="margin-top:12px">
      <FormItem :label="$t('m_template_name')">
        <Tooltip :content="configInfo.name" transfer :disabled="configInfo.name === ''" style="width: 100%;" max-width="200">
          <span>{{configInfo.name}}</span>
          <!-- <Input
            v-model.trim="configInfo.name"
            maxlength="30"
            show-word-limit
            style="width: 96%"
            disabled
          />
          <span style="color: red">*</span> -->
        </Tooltip>
      </FormItem>
      <FormItem :label="$t('m_template_version')">
        <span>{{configInfo.log_monitor_template_version}}</span>
      </FormItem>
      <FormItem :label="$t('m_updatedBy')">
        {{ configInfo.update_user }}
      </FormItem>
      <FormItem :label="$t('m_title_updateTime')">
        {{ configInfo.update_time }}
      </FormItem>
      <!-- <FormItem>
          <Button type="primary" @click="showTemplate = !showTemplate" ghost size="small" style="bottom: 72px;right: 36px;position: relative;">{{showTemplate ? $t('m_hide_template'):$t('m_expand_template')}}</Button>
        </FormItem> -->
      <FormItem :label="$t('m_log_example')" v-if="showTemplate === true">
        <Input
          v-model.trim="configInfo.demo_log"
          type="textarea"
          :rows="11"
          style="width: 96%"
          disabled
        />
        <div v-if="isParmasChanged && configInfo.demo_log.length === 0" style="color: red">
          {{ $t('m_log_example') }} {{ $t('m_tips_required') }}
        </div>
      </FormItem>
    </Form>
    </Col>
    <Col span="16">
    <div style="margin-left: 8px">
      <!-- 采集参数 -->
      <div  v-if="showTemplate === true">
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
    </div>
    </Col>
  </Row>
</template>

<script>
import {isEmpty} from 'lodash'
import Vue from 'vue'
import {renderDisplayName} from '@/assets/js/utils'
import {thresholdList, lastList} from '@/assets/config/common-config.js'
export default {
  name: 'standard-regex',
  data() {
    return {
      isParmasChanged: false,
      showTemplate: false,
      isAdd: true,
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
            <Input
              value={params.row.regular}
              disabled
              onInput={v => {
                this.changeRegex(params.index, v)
              }}
            />
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
              <ColorPicker value={params.row.color_group || ''} disabled
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
          width: 140,
          render: (h, params) => (
            <span>
              <span>{this.prefixCode}_{params.row.metric}</span>
            </span>
          )
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
                disabled
                on-on-change={val => {
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
                disabled
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
              disabled
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
              disabled
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
              disabled
            >
              {lastList.map(i => (
                <Option value={i.value} key={i.value}>
                  {i.label}
                </Option>
              ))}
            </Select>
          ) : <div/>
        }
      ]
    }
  },
  props: {
    configInfo: Object,
    prefixCode: String
  },
  computed: {
    isAddState() {
      return this.$parent.$parent.isAdd
    }
  },
  methods: {
    hideTemplate() {
      this.showTemplate = false
      this.isfullscreen = false
    },
    changeTemplateStatus() {
      this.showTemplate = !this.showTemplate
    },
    returnCurrentStatus() {
      return this.showTemplate
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

<style lang="less">
.calculation-indicators {
  .ivu-table-cell {
    padding-left: 2px;
    padding-right: 2px;
  }
}
</style>
