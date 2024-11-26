<template>
  <div class='config-detail-info'>
    <div v-for="(tableItem, tableIndex) in totalPageConfig" :key="tableIndex + 'f'">
      <Card style="margin-bottom: 16px;">
        <div v-if="tableItem.endpoint_group"  class="w-header" slot="title">
          <div class="title">
            {{tableItem.display_name}}
            <span class="underline"></span>
          </div>
          <Tag color="gold" v-if="tableItem.service_group === ''">{{ $t('m_base_group') }}</Tag>
          <Tag color="blue" v-else>{{ $t('m_field_resourceLevel') }}</Tag>
        </div>
        <span slot="extra" v-if="isEditState">
          <Button type="success" @click="addAlarmItem(tableItem, tableIndex)">{{ $t('m_button_add') }}</Button>
          <Button type="primary" @click="updateNotify(tableItem)">{{ $t('m_button_save') }}</Button>
        </span>
        <span style="font-weight: 700;">{{$t('m_alarm_list')}}</span>
        <Table
          size="small"
          :columns="alarmItemTableColumns"
          :data="tableItem.tableData"
        />
        <div class="alarm-tips" style="margin-top:16px">
          <span style="font-weight: 700;">{{$t('m_alarm_schedulingNotification')}}({{$t('m_all') + $t('m_menu_alert')}})</span>
          <Tooltip :max-width="400" placement="right">
            <p slot=content>
              {{ $t('m_alarm_tips') }}
            </p>
            <Icon type="ios-help-circle-outline" style="margin-left:4px" />
          </Tooltip>
        </div>
        <div>
          <template v-if="tableItem.notify.length > 0">
            <div v-for="(item, index) in tableItem.notify" :key="index + 'S'" style="margin: 4px 0">
              <Row>
                <Col span="2">
                <span style="margin-right: 8px;line-height: 32px;">{{$t('m_' + item.alarm_action)}}</span>
                </Col>
                <Col span="6" style="">
                <Select v-model="item.notify_roles" :disabled="!isEditState" :max-tag-count="2" style="width: 99%;" multiple filterable :placeholder="$t('m_field_role')">
                  <Option v-for="item in allRole" :value="item.name" :key="item.value">{{ item.display_name }}</Option>
                </Select>
                </Col>
                <Col span="5">
                <Select v-model="item.proc_callback_key" clearable :disabled="!isEditState" @on-change="procCallbackKeyChange(item.proc_callback_key, tableIndex, index)" style="width:99%;" :placeholder="$t('m_proc_callback_key')">
                  <Option v-for="(flow, flowIndex) in flows" :key="flowIndex" :value="flow.procDefKey" :label="flow.procDefName + ' [' + flow.procDefVersion + ']'"><span>{{ flow.procDefName }} [{{ flow.procDefVersion }}]</span></Option>
                </Select>
                </Col>
                <Col span="5">
                <Select v-model="item.proc_callback_mode" clearable :disabled="!isEditState" style="width:99%" :placeholder="$t('m_callback_mode')">
                  <Option v-for="item in callbackMode" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
                </Select>
                </Col>
                <Col span="5">
                <Input
                  v-model="item.description"
                  clearable
                  :disabled="!isEditState"
                  style="width:99%"
                  type="text"
                  maxlength="50"
                  :placeholder="$t('m_tableKey_description')"
                />
                </Col>
              </Row>
            </div>
          </template>
        </div>
      </Card>
    </div>
    <div></div>
    <!-- 新增告警列表 -->
    <Modal
      :width="1200"
      :fullscreen="isfullscreen"
      v-model="isShowAddEditModal"
    >
      <div slot="header" class="custom-modal-header">
        <span>
          {{ (modelConfig.isAdd ? $t('m_button_add') : $t('m_button_edit')) + $t('m_metric_threshold') }}
        </span>
        <!-- <Icon v-if="isfullscreen" @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-expand" /> -->
      </div>
      <div class="extentClass">
        <div v-if="isModalShow" class="left-content">
          <div class="use-underline-title mb-3">
            {{$t('m_alarm_config')}}
            <span class="underline"></span>
          </div>
          <Form ref="formData" :label-width="100">
            <FormItem>
              <span slot="label">
                <span style="color:red">*</span>
                {{ $t('m_alarmName') }}
              </span>
              <Input v-model.trim="formData.name" :disabled="!isEditState" :maxlength="50" show-word-limit></Input>
            </FormItem>
            <FormItem>
              <span slot="label">
                <span style="color:red">*</span>
                {{ $t('m_tableKey_s_priority') }}
              </span>
              <Select
                v-model="formData.priority"
                :disabled="!isEditState"
              >
                <Option
                  v-for="item in modelConfig.priorityList"
                  :value="item.value"
                  :key="item.value"
                >
                  {{ $t(item.label) }}
                </Option>
              </Select>
            </FormItem>
            <FormItem>
              <span slot="label">
                <span style="color:red">*</span>
                {{ $t('m_notification') }}
              </span>
              <Select
                :disabled="!isEditState"
                v-model="formData.notify_enable"
              >
                <Option
                  v-for="item in modelConfig.notifyEnableOption"
                  :value="item.value"
                  :key="item.value"
                >
                  {{ item.label }}
                </Option>
              </Select>
            </FormItem>
            <FormItem>
              <span slot="label">
                <span style="color:red">*</span>
                {{ $t('m_delay') }}
              </span>
              <Select
                filterable
                :disabled="!isEditState"
                v-model="formData.notify_delay_second"
              >
                <Option
                  v-for="item in modelConfig.notifyDelayOption"
                  :value="item.value"
                  :key="item.value"
                >
                  {{ item.label }}
                </Option>
              </Select>
            </FormItem>
            <FormItem :label="$t('m_active_window')" prop="active_window_list">
              <activeWindowTime
                class='mb-3'
                :activeWindowList='formData.active_window_list'
                :isDisabled="!isEditState"
                @timeChange="onActiveTimeChange"
              >
              </activeWindowTime>
            </FormItem>
            <FormItem :label="$t('m_tableKey_content')" prop="content">
              <Input
                type="textarea"
                :disabled="!isEditState"
                v-model.trim="formData.content"
                :maxlength="200"
              >
              </Input>
            </FormItem>
            <div style="margin:12px 0 8px">
              <span style="font-weight: 700;">
                {{$t('m_alarm_schedulingNotification')}}({{$t('m_current_alarm')}})
              </span>
              <Tooltip :max-width="400" placement="right">
                <p slot=content>
                  {{ $t('m_alarm_tips') }}
                </p>
                <Icon type="ios-help-circle-outline" style="margin-left:4px" />
              </Tooltip>
            </div>
            <div
              v-for="(item, index) in formData.notify"
              :key="index + 'S'"
              style="margin: 4px 0"
            >
              <Row>
                <Col span="3">
                <span style="margin-right: 8px;line-height: 32px;">{{$t('m_' + item.alarm_action)}}</span>
                </Col>
                <Col span="6" style="">
                <Select v-model="item.notify_roles" :disabled="!isEditState" :max-tag-count="1" style="width: 97%;" multiple filterable :placeholder="$t('m_field_role')">
                  <Option v-for="item in allRole" :value="item.name" :key="item.value">{{ item.display_name }}</Option>
                </Select>
                </Col>
                <Col span="5">
                <Select v-model="item.proc_callback_key" clearable :disabled="!isEditState" @on-change="procCallbackKeyChangeForm(item.proc_callback_key, index)" style="width:97%;" :placeholder="$t('m_proc_callback_key')">
                  <Option v-for="(flow, flowIndex) in flows" :key="flowIndex" :value="flow.procDefKey" :label="flow.procDefName + ' [' + flow.procDefVersion + ']'"><span>{{ flow.procDefName }} [{{ flow.procDefVersion }}]</span></Option>
                </Select>
                </Col>
                <Col span="5">
                <Select v-model="item.proc_callback_mode" clearable :disabled="!isEditState" style="width:97%" :placeholder="$t('m_callback_mode')">
                  <Option v-for="item in callbackMode" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
                </Select>
                </Col>
                <Col span="5">
                <Input
                  v-model.trim="item.description"
                  clearable
                  :disabled="!isEditState"
                  style="width:97%"
                  type="text"
                  maxlength="50"
                  :placeholder="$t('m_tableKey_description')"
                />
                </Col>
              </Row>
            </div>
          </Form>
        </div>
        <div class="right-content">
          <div class="use-underline-title mb-3">
            {{$t('m_field_metric')}}{{$t('m_field_threshold')}}
            <span class="underline"></span>
          </div>
          <Table
            :key="refreshKey"
            class='metric-table'
            style="width:100%;"
            :border="false"
            size="small"
            :columns="metricItemTableColumns"
            :data="formData.conditions"
          />
          <div>
            <Button
              @click.stop="onAddIconClick"
              :disabled="!isEditState"
              type="success"
              size="small"
              style="float:right;margin: 6px 12px;"
              ghost
              icon="md-add"
            ></Button>
          </div>
        </div>
      </div>
      <div slot="footer">
        <Button v-if="isEditState" class="modal-button-save" style="float:right" type="primary" @click="submitContent">{{$t('m_button_save')}}</Button>
        <Button style="float:right" @click="cancelModal">{{$t('m_button_cancel')}}</Button>
      </div>
    </Modal>
  </div>
</template>

<script>
import {thresholdList, lastList, priorityList} from '@/assets/config/common-config.js'
import cloneDeep from 'lodash/cloneDeep'
import hasIn from 'lodash/hasIn'
import isEmpty from 'lodash/isEmpty'
import find from 'lodash/find'
import Vue from 'vue'
import TagShow from '@/components/Tag-show.vue'
import activeWindowTime from '@/components/active-window-time.vue'

const initFormData = {
  name: '', // 告警名
  priority: 'medium', // 级别
  notify_enable: 1, // 告警发送
  notify_delay_second: 0, // 延时
  active_window_list: ['00:00-23:59'], // 告警时间段
  content: '', // 通知内容
  notify: [
    {
      alarm_action: 'firing',
      proc_callback_key: '',
      notify_roles: [],
      proc_callback_mode: '',
      description: ''
    },
    {
      alarm_action: 'ok',
      proc_callback_key: '',
      notify_roles: [],
      proc_callback_mode: '',
      description: ''
    }
  ], // 编排配置
  conditions: [
    {
      metric: '', // 指标名
      threshold: '', // 符号
      thresholdValue: null, // 阈值
      lastSymbol: null, // 时间单位
      lastValue: null, // 持续时间,
      tagOptions: [], // 标签值options
      tags: {}, // 有内容的标签值
    }
  ]
}

const alarmLevelMap = {
  low: {
    label: 'm_low',
    buttonType: 'green'
  },
  medium: {
    label: 'm_medium',
    buttonType: 'gold'
  },
  high: {
    label: 'm_high',
    buttonType: 'red'
  }
}

const statusButtonLabelMap = {
  1: 'm_alarm_open',
  0: 'm_alarm_close'
}

const equalOptionList = [
  {
    name: 'm_include',
    value: 'in'
  },
  {
    name: 'm_not_include',
    value: 'notin'
  }
]

export default {
  name: '',
  props: {
    type: String, // 为枚举类型，gourp为组，service为层级对象，endpoint为对象
    alarmName: String,
    onlyShowCreated: String // me我创建的，all所有
  },
  data() {
    return {
      targetId: '', // 层级对象选中的
      totalPageConfig: [],
      originTotalPageConfig: [], // 保存原始数据
      selectedTableData: null,
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: 'm_button_add',
        modalStyle: 'min-width:98%',
        isAdd: true,
        noBtn: true,
        config: [
          {
            name: 'metricSelect',
            type: 'slot'
          },
          {
            name: 'btn',
            type: 'slot'
          }
        ],
        metricName: '',
        metricList: [],
        threshold: '>',
        thresholdList,
        thresholdValue: '',
        last: 's',
        lastList,
        lastValue: '',
        priority: 'low',
        priorityList,
        notifyEnableOption: [
          {
            label: '启用',
            value: 1
          },
          {
            label: '关闭',
            value: 0
          }
        ],
        notifyDelayOption: [
          {
            label: '0s',
            value: 0
          },
          {
            label: '30s',
            value: 30
          },
          {
            label: '60s',
            value: 60
          },
          {
            label: '90s',
            value: 90
          },
          {
            label: '120s',
            value: 120
          },
          {
            label: '180s',
            value: 180
          },
          {
            label: '300s',
            value: 300
          },
          {
            label: '600s',
            value: 600
          }
        ],
        slotConfig: {
          resourceSelected: [],
          resourceOption: []
        }
      },
      allRole: [],
      flows: [],
      modelTip: {
        key: '',
        value: ''
      },
      selectedData: {},
      notify: [],
      callbackMode: [
        {
          label: 'm_manual',
          value: 'manual'
        },
        {
          label: 'm_auto',
          value: 'auto'
        }
      ],
      formData: cloneDeep(initFormData),
      ruleValidate: {
        name: [
          {
            type: 'string',
            required: true,
            message: '请输入',
            trigger: 'blur'
          }
        ],
        // priority: [
        //   {type: 'string', required: true, message: '请输入', trigger: 'blur' }
        // ],
        // notify_enable: [
        //   {type: 'number', required: true, message: '请输入', trigger: 'blur' }
        // ],
        // notify_delay_second: [
        //   {type: 'number', required: true, message: '请输入', trigger: 'blur' }
        // ],
        // active_window: [
        //   {type: 'array', required: true, message: '请输入', trigger: 'blur' }
        // ]
      },
      // 模态框中的表格
      metricItemTableColumns: [
        {
          title: this.$t('m_tableKey_metricName'),
          key: 'metric',
          align: 'left',
          minWidth: 150,
          render: (h, params) => {
            const typeList = [
              {
                label: this.$t('m_basic_type'),
                value: 'common',
                color: '#2d8cf0'
              },
              {
                label: this.$t('m_business_configuration'),
                value: 'business',
                color: '#81b337'
              },
              {
                label: this.$t('m_metric_list'),
                value: 'custom',
                color: '#b886f8'
              }
            ]
            return (
              <Select
                value={params.row.metric}
                disabled={!this.isEditState}
                on-on-open-change={e => {
                  if (e) {
                    this.getMetricList()
                  }
                }}
                on-on-change={v => {
                  if (v) {
                    this.getTagList(v, params.index)
                  }
                  this.formData.conditions[params.index].metric = v
                }}
                filterable
                clearable
              >
                {this.modelConfig.metricList
                  && this.modelConfig.metricList.map((i, index) => {
                    const find = typeList.find(item => item.value === i.metric_type) || {}
                    return (
                      <Option value={i.guid} key={i.metric + index} label={i.metric}>
                        <Tag size="medium" type="border" color={find.color}>{find.label || '-'}</Tag>
                        {i.metric}
                      </Option>
                    )
                  })}
              </Select>
            )
          }
        },
        {
          title: this.$t('m_label_value'),
          key: 'tags',
          align: 'left',
          width: 300,
          render: (h, params) => (
            <div>
              {!isEmpty(params.row.tagOptions) && !isEmpty(params.row.tags) && this.isModalShow
                ? (params.row.tags.map((i, selectIndex) => (
                  <div class="tags-show" key={selectIndex + '' + JSON.stringify(i)}>
                    <span>{i.tagName}</span>
                    <Select
                      style="width: 70px"
                      value={i.equal}
                      disabled={!this.isEditState}
                      on-on-change={v => {
                        Vue.set(this.formData.conditions[params.index].tags[selectIndex], 'equal', v)
                      }}
                      filterable>
                      {equalOptionList.map((item, index) => (
                        <Option value={item.value} key={item.value + index}>
                          {this.$t(item.name)}
                        </Option>
                      ))}
                    </Select>
                    <Select
                      style="maxWidth: 130px"
                      value={i.tagValue}
                      disabled={!this.isEditState}
                      on-on-change={v => {
                        Vue.set(this.formData.conditions[params.index].tags[selectIndex], 'tagValue', v)
                      }}
                      filterable
                      multiple
                      clearable>
                      {!isEmpty(this.formData.conditions[params.index].tagOptions)
                          && !isEmpty(this.formData.conditions[params.index].tagOptions[i.tagName])
                            && this.formData.conditions[params.index].tagOptions[i.tagName].map((item, index) => (
                              <Option value={item.value} key={item.key + index}>
                                {item.key}
                              </Option>
                            ))}
                    </Select>
                  </div>
                ))) : '-' }
            </div>
          )
        },
        {
          title: this.$t('m_symbol'),
          key: 'threshold',
          align: 'left',
          minWidth: 50,
          render: (h, params) => (
            <Select
              value={params.row.threshold}
              disabled={!this.isEditState}
              on-on-change={v => {
                this.formData.conditions[params.index].threshold = v
              }}
              filterable
              clearable
            >
              {this.modelConfig.thresholdList
                  && this.modelConfig.thresholdList.map((i, index) => (
                    <Option value={i.value} key={index}>
                      {i.label}
                    </Option>
                  ))}
            </Select>
          )
        },
        {
          title: this.$t('m_field_threshold'),
          key: 'thresholdValue',
          align: 'left',
          width: 70,
          render: (h, params) => (
            <Input
              value={params.row.thresholdValue}
              disabled={!this.isEditState}
              on-on-change={v => {
                this.formData.conditions[params.index].thresholdValue = v.target.value
              }}
              clearable
            />
          )
        },
        {
          title: this.$t('m_tableKey_s_last'),
          key: 'lastValue',
          align: 'left',
          width: 70,
          render: (h, params) => (
            <Input
              value={params.row.lastValue}
              disabled={!this.isEditState}
              on-on-change={v => {
                this.formData.conditions[params.index].lastValue = v.target.value
              }}
              clearable
            />
          )
        },
        {
          title: this.$t('m_time_unit'),
          key: 'lastSymbol',
          align: 'left',
          minWidth: 70,
          render: (h, params) => (
            <Select
              value={params.row.lastSymbol}
              disabled={!this.isEditState}
              on-on-change={v => {
                this.formData.conditions[params.index].lastSymbol = v
              }}
              filterable
              clearable
            >
              {this.modelConfig.lastList
                  && this.modelConfig.lastList.map(i => (
                    <Option value={i.value} key={i.value}>
                      {i.label}
                    </Option>
                  ))}
            </Select>
          )
        },
        {
          title: this.$t('m_table_action'),
          key: 'index',

          width: 50,
          render: (h, params) => (
            <Button
              disabled={this.formData.conditions.length < 2 || !this.isEditState}
              on-click={() => {
                this.onDeleteIconClick(params.index)
              }}
              type="error"
              size="small"
              ghost
            >
              <Icon type="md-trash" size="16"></Icon>
            </Button>
          )
        }
      ],
      monitorType: '',
      // 主页中的表格
      alarmItemTableColumns: [
        {
          title: this.$t('m_alarmName'),
          width: 250,
          key: 'name',
          render: (h, params) => params.row.name ? (<div style='display: flex; align-items:center'>
            {params.row.log_metric_group ? <Tag class='auto-tag-style' color='green'>auto</Tag> : <div></div>}
            <Tooltip class='table-alarm-name' placement="right" max-width="400" content={params.row.name}>
              {params.row.name || '-'}
            </Tooltip>
          </div>) : (<div>-</div>)
        },
        {
          title: this.$t('m_alarmPriority'),
          key: 'priority',
          width: 100,
          render: (h, params) => (
            <Tag color={alarmLevelMap[params.row.priority].buttonType}>{this.$t(alarmLevelMap[params.row.priority].label)}</Tag>
          )
        },
        {
          title: this.$t('m_notification'),
          key: 'notify_enable',
          width: 100,
          render: (h, params) => {
            const notify_enable = params.row.notify_enable === 0
            return (
              <Tag color={notify_enable ? 'default' : 'green'}>{this.$t(statusButtonLabelMap[params.row.notify_enable])}</Tag>
            )
          }
        },
        {
          title: this.$t('m_field_relativeTime'),
          width: 130,
          key: 'active_window_list',
          render: (h, params) => {
            const text = !isEmpty(params.row.active_window_list) ? params.row.active_window_list.join(';') : '-'
            return (<Tooltip class='table-active-time' placement="top" max-width="400" content={text}>
              {text}
            </Tooltip>)
          }
        },
        {
          title: this.$t('m_firing'),
          key: 'firing',
          width: 150,
          align: 'left',
          render: (h, params) => {
            const { showBtn, result } = this.mgmtConfigDetail(params.row.notify[0])
            return showBtn
              ? (
                <div>
                  <Tooltip placement="right" max-width="400">
                    <div slot="content" style="white-space: normal;">
                      <p>{this.$t('m_notification_role')}: {result.role}</p>
                      <p>{this.$t('m_trigger_arrange')}: {result.arrange}</p>
                      <p>{this.$t('m_tableKey_description')}: {result.description}</p>
                    </div>
                    <Tag color="geekblue" style="cursor:pointer">{this.$t('m_config_view')}</Tag>
                  </Tooltip>
                </div>
              ) : <div>-</div>
          }
        },
        {
          title: this.$t('m_ok'),
          key: 'ok',
          width: 150,
          align: 'left',
          render: (h, params) => {
            const { showBtn, result } = this.mgmtConfigDetail(params.row.notify[1])
            return showBtn
              ? (
                <div>
                  <Tooltip placement="right" max-width="400">
                    <div slot="content" style="white-space: normal;">
                      <p>{this.$t('m_notification_role')}: {result.role}</p>
                      <p>{this.$t('m_trigger_arrange')}: {result.arrange}</p>
                      <p>{this.$t('m_tableKey_description')}: {result.description}</p>
                    </div>
                    <Tag color="geekblue" style="cursor:pointer">{this.$t('m_config_view')}</Tag>
                  </Tooltip>
                </div>
              ) : <div>-</div>
          }
        },
        {
          title: this.$t('m_tableKey_metricName'),
          key: 'metric_name',
          minWidth: 300
        },
        {
          title: this.$t('m_field_threshold'),
          key: 'condition',

          width: 100
        },
        {
          title: this.$t('m_tableKey_s_last'),
          key: 'last',

          width: 100
        },
        {
          title: this.$t('m_update_time'), // 更新时间
          key: 'update_time',
          width: 160,
          render: (h, params) => <span>{params.row.update_time || '-'}</span>
        },
        {
          title: this.$t('m_updatedBy'), // 更新人
          key: 'update_user',
          width: 160,
          render: (h, params) => <span>{params.row.update_user || '-'}</span>
        },
        {
          title: this.$t('m_table_action'),
          key: 'index',
          fixed: 'right',
          width: 150,
          render: (h, params) => (
            <div style='display: flex'>
              {this.type==='endpoint'
                ? (
                  <Tooltip max-width={400} placement="top" transfer content={this.$t('m_preview')}>
                    <Button size="small" class="mr-1" type="primary" on-click={() => this.editAlarmItem(params.row, params)}>
                      <Icon type="md-eye" size="16"></Icon>
                    </Button>
                  </Tooltip>
                )
                : (<span style='display: flex'>
                  <Tooltip max-width={400} placement="top" transfer content={this.$t('m_copy')}>
                    <Button size="small" class="mr-1" type="success" on-click={() => this.copySingleItem(params.row)}>
                      <Icon type="md-document" size="16"></Icon>
                    </Button>
                  </Tooltip>
                  <Tooltip max-width={400} placement="top" transfer content={this.$t('m_button_edit')}>
                    <Button size="small" class="mr-1" type="primary" on-click={() => this.editAlarmItem(params.row, params)}>
                      <Icon type="md-create" size="16"></Icon>
                    </Button>
                  </Tooltip>
                </span>
                )
              }
              {
                this.isEditState ? (

                  <Tooltip max-width={400} placement="top" transfer content={this.$t('m_button_remove')}>
                    <Poptip
                      confirm
                      transfer
                      title={this.$t('m_delConfirm_tip')}
                      placement="left-end"
                      on-on-ok={() => {
                        this.okDelRow(params.row)
                      }}>
                      <Button size="small" type="error">
                        <Icon type="md-trash" size="16" />
                      </Button>
                    </Poptip>
                  </Tooltip>
                ) : null
              }
            </div>
          )
        }
      ],
      mergeSpanMap: {},
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      initialTags: [],
      initialTagOptions: {},
      isModalShow: false, // 为了解决select组件渲染出错问题,
      currentAlarmListIndex: null, // 解决告警名称重名校验

      isShowAddEditModal: false, // 告警配置弹窗
      isfullscreen: true,
      refreshKey: ''
    }
  },
  methods: {
    // 根据告警名称模糊搜索
    // filterData(alarmName) {
    //   this.totalPageConfig = cloneDeep(this.originTotalPageConfig)
    //   this.totalPageConfig = this.totalPageConfig.map(item => {
    //     item.tableData = item.tableData.filter(row => row.name.toLowerCase().indexOf(alarmName.toLowerCase()) > -1)
    //     return item
    //   })
    // },
    mgmtConfigDetail(val) {
      const res = {
        showBtn: true,
        result: {
          role: !val.notify_roles || val.notify_roles.length === 0 ? '-' : val.notify_roles.join(';'),
          arrange: (!val.proc_callback_name && !val.proc_callback_mode) ? '-' : (val.proc_callback_name ? val.proc_callback_name : '-') + (val.proc_callback_mode ? '(' + this.$t(find(this.callbackMode, {value: val.proc_callback_mode}).label) + ')' : ''),
          description: val.description || '-'
        }
      }
      if (val.notify_roles.length === 0 && val.proc_callback_key === '' && val.proc_callback_mode === '') {
        res.showBtn = false
      }
      return res
    },
    getUpdateNotifyApi(data) {
      return this.type === 'service'
        ? `/monitor/api/v2/alarm/endpoint_group/${data.endpoint_group}/notify/update`
        : `/monitor/api/v2/alarm/endpoint_group/${this.targetId}/notify/update`
    },
    updateNotify(tableData) {
      const api = this.getUpdateNotifyApi(tableData)
      this.request('POST', api, tableData.notify, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getDetail(this.targetId)
      })
    },
    procCallbackKeyChange(proc_callback_key, tableIndex, index) {
      const findFlow = this.flows.find(f => f.procDefKey === proc_callback_key)
      if (findFlow) {
        this.totalPageConfig[tableIndex].notify[index].proc_callback_name = `${findFlow.procDefName}[${findFlow.procDefVersion}]`
      } else {
        this.totalPageConfig[tableIndex].notify[index].proc_callback_name = ''
      }
    },
    procCallbackKeyChangeForm(proc_callback_key, index) {
      const findFlow = this.flows.find(f => f.procDefKey === proc_callback_key)
      if (findFlow) {
        this.formData.notify[index].proc_callback_name = `${findFlow.procDefName}[${findFlow.procDefVersion}]`
      } else {
        this.formData.notify[index].proc_callback_name = ''
      }
    },
    deleteAlarmItem(rowData) {
      const api = `/monitor/api/v2/alarm/strategy/${rowData.guid}`
      this.request('DELETE', api, '', () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getDetail(this.targetId)
      })
    },
    okDelRow(rowData) {
      this.deleteAlarmItem(rowData)
    },
    manageEditParams(addParams, rowparams){
      for (const key in addParams) {
        if (hasIn(rowparams, key)) {
          Vue.set(addParams, key, cloneDeep(rowparams[key]))
        }
      }
    },
    getMetricListPath(data) {
      return this.type === 'service'
        ? `/monitor/api/v2/monitor/metric/list?monitorType=${data.monitor_type}&serviceGroup=${this.targetId}&query=all`
        : `/monitor/api/v2/monitor/metric/list?monitorType=${this.monitorType}&query=all`
    },
    getSymbolAndValue(str) {
      if (!str.length) {
        return {}
      }
      return {
        symbol: str.match(/[!<>=smh]+/g)[0],
        value: str.match(/[-+]?[\d.]+/)[0]
      }
    },
    getInitialTags() {
      return new Promise(resolve => {
        const initialMetricName = this.modelConfig.metricList[0].guid
        this.findTagsByMetric(initialMetricName).then(res => {
          this.initialTagOptions = res
          if (!isEmpty(this.initialTagOptions)) {
            this.initialTags = []
            for (const key in this.initialTagOptions) {
              this.initialTags.push(
                {
                  tagName: key,
                  tagValue: [],
                  equal: 'in'
                }
              )
            }
          }
          resolve()
        })
      })
    },
    async editAlarmItem(rowData) {
      this.isfullscreen = true
      this.currentAlarmListIndex = rowData.alarmListIndex
      this.selectedData = rowData
      const api = this.getMetricListPath(rowData)
      this.request('GET', api, '', responseData => {
        this.modelConfig.metricList = responseData
        this.modelConfig.isAdd = false
        this.modelConfig.modalTitle = 'm_button_edit',
        this.formData = cloneDeep(initFormData)
        this.manageEditParams(this.formData, rowData)
        const conditions = this.formData.conditions
        conditions.length && conditions.forEach(async item => {
          const thresholdsAndSymbols = this.getSymbolAndValue(item.condition)
          const lastValueAndSymbol = this.getSymbolAndValue(item.last)
          const tagOptions = await this.findTagsByMetric(item.metric) // 指标下拉option获取
          if (isEmpty(item.tags) && !isEmpty(tagOptions)) {
            const tags = []
            for (const key in tagOptions) {
              tags.push(
                {
                  tagName: key,
                  tagValue: [],
                  equal: 'in'
                }
              )
            }
            Vue.set(item, 'tags', tags)
          }
          Vue.set(item, 'threshold', thresholdsAndSymbols.symbol)
          Vue.set(item, 'thresholdValue', thresholdsAndSymbols.value)
          Vue.set(item, 'lastSymbol', lastValueAndSymbol.symbol)
          Vue.set(item, 'lastValue', lastValueAndSymbol.value)
          Vue.set(item, 'tagOptions', tagOptions)
          !isEmpty(item.tags) && item.tags.forEach(single => {
            if (!single.equal) {
              single.equal = 'in'
            }
          })
        })
        this.showAddEditModal()
      })
    },
    async copySingleItem(rowData) {
      this.isfullscreen = true
      this.currentAlarmListIndex = rowData.alarmListIndex
      this.selectedData = rowData
      delete this.selectedData.guid
      this.selectedTableData = rowData
      const api = this.getMetricListPath(rowData)
      this.request('GET', api, '', responseData => {
        this.modelConfig.metricList = responseData
        this.modelConfig.isAdd = true
        this.modelConfig.modalTitle = 'm_button_add',
        this.formData = cloneDeep(initFormData)
        this.manageEditParams(this.formData, rowData)
        this.formData.name = this.formData.name + '1'
        const conditions = this.formData.conditions
        conditions.length && conditions.forEach(async item => {
          const thresholdsAndSymbols = this.getSymbolAndValue(item.condition)
          const lastValueAndSymbol = this.getSymbolAndValue(item.last)
          const tagOptions = await this.findTagsByMetric(item.metric) // 指标下拉option获取
          if (isEmpty(item.tags) && !isEmpty(tagOptions)) {
            const tags = []
            for (const key in tagOptions) {
              tags.push(
                {
                  tagName: key,
                  tagValue: [],
                  equal: 'in'
                }
              )
            }
            Vue.set(item, 'tags', tags)
          }
          Vue.set(item, 'threshold', thresholdsAndSymbols.symbol)
          Vue.set(item, 'thresholdValue', thresholdsAndSymbols.value)
          Vue.set(item, 'lastSymbol', lastValueAndSymbol.symbol)
          Vue.set(item, 'lastValue', lastValueAndSymbol.value)
          Vue.set(item, 'tagOptions', tagOptions)
          !isEmpty(item.tags) && item.tags.forEach(single => {
            if (!single.equal) {
              single.equal = 'in'
            }
          })
        })
        this.showAddEditModal()
      })
    },
    addAlarmItem(tableItem, index) {
      this.isfullscreen = true
      this.currentAlarmListIndex = index
      this.selectedTableData = tableItem
      this.selectedData = tableItem
      const api = this.getMetricListPath(tableItem)
      this.request('GET', api, '', async responseData => {
        this.modelConfig.metricList = responseData
        this.modelConfig.isAdd = true
        this.modelConfig.modalTitle = 'm_button_add',
        this.formData = cloneDeep(initFormData)
        this.formData.name = this.$t('m_alert') + new Date().getTime()
        await this.getInitialTags()
        this.formData.conditions[0].metric = this.modelConfig.metricList[0].guid // 指标名
        this.formData.conditions[0].tags = cloneDeep(this.initialTags)
        this.formData.conditions[0].tagOptions = cloneDeep(this.initialTagOptions)
        this.showAddEditModal()
      })
    },
    getWorkFlow() {
      this.request('GET', '/monitor/api/v2/alarm/event/callback/list', '', responseData => {
        this.flows = responseData
      })
    },
    getAllRole() {
      this.request('GET', '/monitor/api/v1/user/role/list?page=1&size=1000', '', responseData => {
        this.allRole = responseData.data.map(_ => ({
          ..._,
          value: _.id
        }))
      })
    },
    processConditions(conditions) {
      if (!conditions.length) {
        return
      }
      conditions.forEach(item => {
        item.condition = item.threshold + item.thresholdValue
        item.last = item.lastValue + item.lastSymbol
        const needDeleteAttr = ['lastSymbol', 'lastValue', 'threshold', 'thresholdValue', 'tagOptions']
        needDeleteAttr.forEach(key => {
          delete item[key]
        })
      })
    },
    validateConditions(conditions) {
      if (!conditions || isEmpty(conditions)) {
        return false
      }
      const needValidateKey = ['metric', 'lastSymbol', 'lastValue', 'threshold', 'thresholdValue']
      for (let i=0; i<conditions.length; i++) {
        for (let k=0; k < needValidateKey.length; k++) {
          if (!conditions[i][needValidateKey[k]]) {
            return false
          }
          if (needValidateKey[k] === 'thresholdValue' && !this.isNumericString(conditions[i][needValidateKey[k]])) {
            return false
          }
          if (needValidateKey[k] === 'lastValue' && !this.isPositiveNumericString(conditions[i][needValidateKey[k]])) {
            return false
          }
        }
      }
      return true
    },
    isNumericString(str) {
      return !isNaN(str) && !isNaN(parseFloat(str))
    },
    isPositiveNumericString(str) {
      return /^\d+$/.test(str) && parseFloat(str) >= 0
    },
    validateDuplicateName(alarmName, guid = '') {
      const currentTableList = this.totalPageConfig[this.currentAlarmListIndex].tableData
      if (isEmpty(currentTableList)) {
        return true
      }
      for (let i=0; i<currentTableList.length; i++) {
        const item = currentTableList[i]
        if (item.guid !== guid && item.name === alarmName) {
          return false
        }
      }
      return true
    },
    submitContent() {
      if (this.formData.name.trim() === '') {
        return this.$Message.error(this.$t('m_alarmName') + this.$t('m_tips_required'))
      }
      if (!this.validateConditions(this.formData.conditions)) {
        return this.$Message.error(this.$t('m_metric_threshold') + this.$t('m_tips_emptyToSave') + this.$t('m_threshold_tips')+ ';' + this.$t('m_time_tips'))
      }
      if (!this.validateDuplicateName(this.formData.name, this.selectedData.guid)) {
        return this.$Message.error(this.$t('m_alarmName') + this.$t('m_cannot_be_repeated'))
      }
      const params = cloneDeep(this.formData)
      const needMergeParams = {
        endpoint_group: this.modelConfig.isAdd ? this.selectedTableData.endpoint_group : this.selectedData.endpoint_group,
        guid: this.selectedData.guid
      }
      Object.assign(params, needMergeParams)
      this.processConditions(params.conditions)
      const requestMethod = this.modelConfig.isAdd ? 'POST' : 'PUT'
      this.request(requestMethod, '/monitor/api/v2/alarm/strategy', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.closeAddEditModal()
        this.getDetail(this.targetId)
      })
    },
    showAddEditModal() {
      this.isShowAddEditModal = true
      this.isModalShow = true
    },
    closeAddEditModal() {
      this.isShowAddEditModal = false
      this.isModalShow = false
    },
    setMonitorType(monitorType){
      this.monitorType = monitorType
    },
    handleTableData(tempTableData, alarmListIndex) {
      const initialData = cloneDeep(tempTableData)
      const resData = []
      let startRowIndex = 0
      this.mergeSpanMap = {}
      initialData.forEach(item => {
        item.notify = !isEmpty(item.notify) ? item.notify : cloneDeep(initFormData.notify)
        if (item.conditions && item.conditions.length) {
          item.conditions.forEach(metricItem => {
            resData.push(Object.assign({}, item, metricItem, {alarmListIndex}))
          })
          this.mergeSpanMap[item.guid] = {
            startRowIndex,
            colSpan: item.conditions.length
          }
          startRowIndex += item.conditions.length
        }
      })
      return resData
    },
    getDetail(targetId) {
      if (targetId) {
        this.targetId = targetId
        const api = '/monitor/api/v2/alarm/strategy/query'
        const params = {
          queryType: this.type,
          guid: this.targetId,
          show: this.onlyShowCreated,
          alarmName: this.alarmName
        }
        this.totalPageConfig = []
        this.request('post', api, params, responseData => {
          this.$emit('feedbackInfo', responseData.length === 0)
          const allConfigDetail = responseData
          allConfigDetail.forEach((item, alarmIndex) => {
            const strategy = item.strategy || []
            const tempTableData = strategy.map(s => {
              s.monitor_type = item.monitor_type
              return s
            })
            const tableData = this.handleTableData(tempTableData, alarmIndex)
            this.totalPageConfig.push({
              tableData,
              endpoint_group: item.endpoint_group,
              display_name: item.display_name,
              service_group: item.service_group,
              monitor_type: item.monitor_type,
              notify: item.notify,
              mergeSpanMap: this.mergeSpanMap
            })
            this.originTotalPageConfig = cloneDeep(this.totalPageConfig)
          })
        }, {isNeedloading: true})
        this.getAllRole()
        this.getWorkFlow()
      } else {
        this.totalPageConfig = []
      }
    },
    onAddIconClick() {
      const metricItem = Object.assign({}, cloneDeep(initFormData.conditions[0]), {
        metric: this.modelConfig.metricList[0].guid,
        tags: cloneDeep(this.initialTags),
        tagOptions: cloneDeep(this.initialTagOptions)
      })
      this.formData.conditions.push(metricItem)
    },
    onDeleteIconClick(index) {
      this.formData.conditions.splice(index, 1)
    },
    cancelModal() {
      this.closeAddEditModal()
    },
    // handleMergeSpan({ row, rowIndex, columnIndex }, index) {
    //   if ([6,7,8,11].includes(columnIndex)) {
    //     return
    //   }
    //   const mergeSpanMap = this.totalPageConfig[index].mergeSpanMap
    //   const spanMap = mergeSpanMap[row.guid]
    //   if (rowIndex === spanMap.startRowIndex) {
    //     return [spanMap.colSpan,1]
    //   }
    //   if (rowIndex !== spanMap.startRowIndex && (columnIndex <= 5 || [9,10,11].includes(columnIndex))) {
    //     return [0,0]
    //   }
    // },
    findTagsByMetric(metricId) {
      const api = '/monitor/api/v2/metric/tag/value-list'
      const params = {
        metricId
      }

      return new Promise(resolve => {
        this.request('POST', api, params, responseData => {
          const result = {}
          if (!isEmpty(responseData)) {
            responseData.forEach(item => {
              result[item.tag] = item.values
            })
          }
          resolve(result)
        })
      })
    },
    async getTagList(metricId, tableIndex) {
      const item = this.formData.conditions[tableIndex]
      Vue.set(item, 'tagOptions', await this.findTagsByMetric(metricId))
      // item.tagOptions = await this.findTagsByMetric(metricId)

      if (!isEmpty(item.tagOptions)) {
        const tags = []
        for (const key in item.tagOptions) {
          tags.push(
            {
              tagName: key,
              tagValue: [],
              equal: 'in'
            }
          )
        }
        Vue.set(item, 'tags', tags)
      }
      this.refreshKey = +new Date()
    },
    getMetricList() {
      const api = this.getMetricListPath(this.selectedData)
      this.request('GET', api, '', res => {
        this.modelConfig.metricList = res
      })
    },
    onActiveTimeChange(time) {
      this.formData.active_window_list = time
    }
  },
  computed: {
    isEditState() {
      return this.type !== 'endpoint'
    }
  },
  components: {
  // eslint-disable-next-line vue/no-unused-components
    TagShow,
    activeWindowTime
  }
}
</script>

<style lang="less">
.metric-table {
  .ivu-table-cell {
    padding: 4px;
  }
}

.ivu-table-wrapper {
  overflow: inherit;
}

.right-content {
  .ivu-table-cell {
    padding-left: 2px;
    padding-right: 2px;
  }
}
.alarm-tips {
  .ivu-alert {
    margin-bottom: 0
  }
}

.tags-show {
  display: flex;
  align-items: center;
  justify-content: left;
}
.tags-show > span {
  width: 70px;
}

.tags-show .ivu-select-item.ivu-select-item-selected::after {
    top: 6px;
    right: -6px
}

.modal-button-save.modal-button-save {
  background-color: #2d8cf0!important;
}

.modal-dialog[data-v-0eaeaf66] {
  top: 10%;
}
// .ivu-select-dropdown {
//   max-height: 300px !important;
// }
.table-alarm-name {
  .ivu-tooltip-rel {
    width: 170px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.table-active-time {
  .ivu-tooltip-rel {
    width: 110px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.auto-tag-style {
  .ivu-tag-text.ivu-tag-color-white {
    display: inline-block;
    min-width: 40px;
  }
}

</style>
<style scoped lang="less">
  /deep/ .ivu-form-item {
    margin-bottom: 0;
  }
  .use-underline-title {
    display: inline-block;
    font-size: 16px;
    font-weight: 700;
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
      -webkit-box-sizing: content-box;
      box-sizing: content-box;
    }
  }
  .extentClass {
    display: flex;
    flex-direction: row;
    .arrange-item {
      display: flex;
      flex-direction: row;
      justify-content: space-between;
      margin-bottom: 10px;
    }
    .left-content {
      flex-direction: column;
      width: 45%;
      margin-right: 15px;
      border-right: 1px solid #dcdee2;
      padding-right: 15px;
    }
    .reset-padding-right {
      padding-right: 0px;
    }
    .metric-title {
      font-size: 14px;
    }
    .right-content {
      flex-direction: column;
      width: 53%;
      .add-circle-icon {
        margin-left: 98%;
      }
    }
  }
  .search-input {
    height: 32px;
    padding: 4px 7px;
    font-size: 14px;
    border: 1px solid #dcdee2;
    border-radius: 4px;
    width: 230px;
  }
  .search-input:focus {
    outline: 0;
    border-color: #57a3f3;
  }

  .search-input-content {
    display: inline-block;
    vertical-align: middle;
  }
  .receiver-config {
    margin: 8px 20px;
    .alarm-tips {
      display: flex;
      align-items: center;
    }
  }
  .receiver-config-set {
    display: flex;
  }
  .receiver-header {
    font-weight: 500;
    margin-right: 8px;
  }
  .form-control {
    // 解决IE11 输入框无法与按钮同行问题
    display: inline-block;
    padding: 4px 18px 4px 7px;
  }
  /*取消选中样式*/
  .form-control:focus {
    box-shadow: none;
  }
    /*input框样式*/
  .research-input {
    width: 200px;
    height: 32px;
    font-size: 12px;
    margin-left: 8px;
    margin-right: 8px;
  }
  .partition {
    height: 2px;
    background: @blue-lingt;
  }
  .success-btn {
    color: #fff;
    background-color: #19be6b;
    border-color: #19be6b;
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

/deep/ .ivu-card-extra {
  top: 6px;
}

</style>
