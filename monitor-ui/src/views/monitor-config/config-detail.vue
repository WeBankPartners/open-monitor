<template>
  <div>
    <template v-for="(tableItem, tableIndex) in totalPageConfig">
      <section :key="tableIndex + 'f'">
        <Card>
          <div v-if="tableItem.endpoint_group"  class="w-header" slot="title">
            <div class="title">
              {{tableItem.display_name}}
              <span class="underline"></span>
            </div>
          </div>
          <span slot="extra" v-if="isEditState">
            <Button type="success" size="small" @click="addAlarmItem(tableItem, tableIndex)">{{ $t('button.add') }}</Button>
            <Button type="primary" size="small" @click="updateNotify(tableItem)">{{ $t('button.save') }}</Button>
          </span>
          <span>{{$t('m_alarm_list')}}</span>
          <Table
            size="small"
            :columns="alarmItemTableColumns"
            :data="tableItem.tableData"
            :span-method="(e) => handleMergeSpan(e, tableIndex)"
          />
          <div class="alarm-tips" style="margin-top:16px">
            <span>{{$t('m_alarm_schedulingNotification')}}({{$t('m_all') + $t('menu.alert')}})</span>
            <Alert style="width:60%;display: inline-block;margin-left: 12px;">{{$t('m_alarm_tips')}}</Alert>
          </div>
          <div>
            <template v-if="tableItem.notify.length > 0">
              <div v-for="(item, index) in tableItem.notify" :key="index + 'S'" style="margin: 4px 0">
                <Row>
                  <Col span="2">
                    <span style="float: right;margin-right: 8px;line-height: 32px;">{{$t(item.alarm_action)}}</span>
                  </Col>
                  <Col span="6" style="">
                    <!-- <Tooltip :content="$t('resourceLevel.role')" :delay="1000"> -->
                      <Select v-model="item.notify_roles" :disabled="!isEditState" :max-tag-count="2" style="width: 99%;" multiple filterable :placeholder="$t('field.role')">
                        <Option v-for="item in allRole" :value="item.name" :key="item.value">{{ item.name }}</Option>
                      </Select>
                    <!-- </Tooltip> -->
                  </Col>
                  <Col span="5">
                    <!-- <Tooltip :content="$t('proc_callback_key')" :delay="1000"> -->
                      <Select v-model="item.proc_callback_key" clearable :disabled="!isEditState" @on-change="procCallbackKeyChange(item.proc_callback_key, tableIndex, index)" style="width:99%;" :placeholder="$t('proc_callback_key')">
                        <Option v-for="(flow, flowIndex) in flows" :key="flowIndex" :value="flow.procDefKey" :label="flow.procDefName + ' [' + flow.procDefVersion + ']'"><span>{{ flow.procDefName }} [{{ flow.procDefVersion }}]</span></Option>
                      </Select>
                    <!-- </Tooltip> -->
                  </Col>
                  <Col span="5">
                    <!-- <Tooltip :content="$t('m_callback_mode')" :delay="1000"> -->
                      <Select v-model="item.proc_callback_mode" clearable :disabled="!isEditState" style="width:99%" :placeholder="$t('m_callback_mode')">
                        <Option v-for="item in callbackMode" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
                      </Select>
                    <!-- </Tooltip> -->
                  </Col>
                  <Col span="5">
                    <!-- <Tooltip :content="$t('tableKey.description')" :delay="1000"> -->
                      <Input 
                        v-model="item.description"
                        clearable 
                        :disabled="!isEditState"
                        style="width:99%"
                        type="text" 
                        maxlength="50"
                        :placeholder="$t('tableKey.description')"/>
                    <!-- </Tooltip> -->
                  </Col>
                </Row>
              </div>
            </template>
          </div>
        </Card>
      </section>
    </template>
    <!-- 已废弃 -->
    <ModalComponent :modelConfig="modelConfig">
      <div slot="metricSelect" class="extentClass">  
        <div v-if="isModalShow" class="left-content">
          <div class="use-underline-title mb-3">
            {{$t('m_alarm_config')}}
            <span class="underline"></span>
          </div>
          <Form ref="formData" :model="formData" :rules="ruleValidate" :label-width="100">
            <FormItem :label="$t('m_alarmName')" prop="name">
                <Input v-model="formData.name" :disabled="!isEditState" :maxlength="10"></Input>
            </FormItem>
            <FormItem :label="$t('tableKey.s_priority')" prop="priority">
                <Select 
                  v-model="formData.priority"
                  :disabled="!isEditState"
                  filterable>
                  <Option 
                    v-for="item in modelConfig.priorityList" 
                    :value="item.value" 
                    :key="item.value">
                    {{ $t(item.label) }}
                  </Option>
                </Select>
            </FormItem>
            <FormItem :label="$t('tableKey.status')" prop="notify_enable">
                <Select 
                  filterable
                  :disabled="!isEditState"
                  v-model="formData.notify_enable" >
                  <Option 
                    v-for="item in modelConfig.notifyEnableOption" 
                    :value="item.value" 
                    :key="item.value">
                    {{ item.label }}
                  </Option>
                </Select>
            </FormItem>
            <FormItem :label="$t('delay')" prop="notify_delay_second">
                <Select 
                  filterable 
                  :disabled="!isEditState"
                  v-model="formData.notify_delay_second">
                  <Option 
                    v-for="item in modelConfig.notifyDelayOption" 
                    :value="item.value" 
                    :key="item.value">
                    {{ item.label }}
                  </Option>
                </Select>
            </FormItem>
            <FormItem :label="$t('m_active_window')" prop="active_window">
                <TimePicker 
                  v-model="formData.active_window" 
                  :clearable="false" 
                  format="HH:mm" 
                  :disabled="!isEditState"
                  type="timerange" 
                  placement="bottom-end" 
                  style="width: 168px">
                </TimePicker>
            </FormItem>
            <FormItem :label="$t('tableKey.content')" prop="content">
                <Input 
                  type="textarea" 
                  :disabled="!isEditState"
                  v-model="formData.content"
                  :maxlength="200">
                </Input>
            </FormItem>
            <span class='arrange-config-Title'>{{$t('m_alarm_schedulingNotification')}}({{$t('m_current_alarm')}})</span>
            <div 
              v-for="(item, index) in formData.notify"
              :key="index + 'S'"
              class="arrange-item"
              >
              <span class="mr-1 mt-1" style="font-size: 12px">{{$t(item.alarm_action)}}</span>
              <Tooltip :content="$t('resourceLevel.role')" :delay="1000">
                <Select v-model="item.notify_roles" :disabled="!isEditState" :max-tag-count="2" style="width: 150px" multiple filterable :placeholder="$t('field.role')">
                  <Option v-for="item in allRole" :value="item.name" :key="item.value">{{ item.name }}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('proc_callback_key')" :delay="1000">
                <Select v-model="item.proc_callback_key" :disabled="!isEditState" @on-change="procCallbackKeyChangeForm(item.proc_callback_key, index)" style="width: 160px" :placeholder="$t('proc_callback_key')">
                  <Option v-for="(flow, flowIndex) in flows" :key="flowIndex" :value="flow.procDefKey" :label="flow.procDefName + ' [' + flow.procDefVersion + ']'"><span>{{ flow.procDefName }} [{{ flow.procDefVersion }}]</span></Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('m_callback_mode')" :delay="1000">
                <Select v-model="item.proc_callback_mode" :disabled="!isEditState" style="width: 100px" :placeholder="$t('m_callback_mode')">
                  <Option v-for="item in callbackMode" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('tableKey.description')" :delay="1000">
                <input 
                  v-model="item.description" 
                  :disabled="!isEditState"
                  style="width: 100px"
                  type="text" 
                  maxlength="50"
                  :placeholder="$t('tableKey.description')"
                  class="form-control model-input search-input c-dark"/>
              </Tooltip>
            </div>
          </Form>
        </div>
        <div class="right-content">
          <div class="use-underline-title mb-3">
            {{$t('field.metric')}}{{$t('field.threshold')}}
            <span class="underline"></span>
          </div>
          <Table
            style="width:100%;"
            :border="false"
            size="small"
            :columns="metricItemTableColumns"
            :data="formData.conditions"
          />
          <Button
            @click.stop="onAddIconClick"
            :disabled="!isEditState"
            type="success"
            size="small"
            class="mt-2"
            ghost
            icon="md-add"
          ></Button>
        </div>
      </div>
      <div slot="btn">
        <Button v-if="isEditState" class="modal-button-save" style="float:right" type="primary" @click="submitContent">{{$t('button.save')}}</Button>
        <Button style="float:right" @click="cancelModal">{{$t('button.cancel')}}</Button>
      </div>
    </ModalComponent>
    <!-- 新增告警列表 -->
    <Modal
      :width="1100"
      :fullscreen="isfullscreen"
      v-model="isShowAddEditModal">
      <div slot="header" class="custom-modal-header">
        <span>
          {{ (isEditState ? $t('button.add') : $t('button.edit')) + $t('m_metric_threshold') }}
        </span>
        <Icon v-if="isfullscreen" @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-expand" />
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
              <Input v-model="formData.name" :disabled="!isEditState" :maxlength="10"></Input>
            </FormItem>
            <FormItem>
              <span slot="label">
                <span style="color:red">*</span>
                {{ $t('tableKey.s_priority') }}
              </span>
              <Select 
                v-model="formData.priority"
                :disabled="!isEditState"
                filterable>
                <Option 
                  v-for="item in modelConfig.priorityList" 
                  :value="item.value" 
                  :key="item.value">
                  {{ $t(item.label) }}
                </Option>
              </Select>
            </FormItem>
            <FormItem>
              <span slot="label">
                <span style="color:red">*</span>
                {{ $t('tableKey.status') }}
              </span>
              <Select 
                filterable
                :disabled="!isEditState"
                v-model="formData.notify_enable" >
                <Option 
                  v-for="item in modelConfig.notifyEnableOption" 
                  :value="item.value" 
                  :key="item.value">
                  {{ item.label }}
                </Option>
              </Select>
            </FormItem>
            <FormItem>
              <span slot="label">
                <span style="color:red">*</span>
                {{ $t('delay') }}
              </span>
              <Select 
                filterable 
                :disabled="!isEditState"
                v-model="formData.notify_delay_second">
                <Option 
                  v-for="item in modelConfig.notifyDelayOption" 
                  :value="item.value" 
                  :key="item.value">
                  {{ item.label }}
                </Option>
              </Select>
            </FormItem>
            <FormItem :label="$t('m_active_window')" prop="active_window">
              <span slot="label">
                <span style="color:red">*</span>
                {{ $t('m_active_window') }}
              </span>
              <TimePicker 
                v-model="formData.active_window" 
                :clearable="false" 
                format="HH:mm" 
                :disabled="!isEditState"
                type="timerange" 
                placement="bottom-end" 
                style="width: 168px">
              </TimePicker>
            </FormItem>
            <FormItem :label="$t('tableKey.content')" prop="content">
              <Input 
                type="textarea" 
                :disabled="!isEditState"
                v-model="formData.content"
                :maxlength="200">
              </Input>
            </FormItem>
            <span class='arrange-config-Title'>{{$t('m_alarm_schedulingNotification')}}({{$t('m_current_alarm')}})</span>
            <div 
              v-for="(item, index) in formData.notify"
              :key="index + 'S'"
              class="arrange-item"
              >
              <span class="mr-1 mt-1" style="font-size: 12px">{{$t(item.alarm_action)}}</span>
              <Tooltip :content="$t('resourceLevel.role')" :delay="1000">
                <Select v-model="item.notify_roles" :disabled="!isEditState" :max-tag-count="2" style="width: 150px" multiple filterable :placeholder="$t('field.role')">
                  <Option v-for="item in allRole" :value="item.name" :key="item.value">{{ item.name }}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('proc_callback_key')" :delay="1000">
                <Select v-model="item.proc_callback_key" clearable :disabled="!isEditState" @on-change="procCallbackKeyChangeForm(item.proc_callback_key, index)" style="width: 160px" :placeholder="$t('proc_callback_key')">
                  <Option v-for="(flow, flowIndex) in flows" :key="flowIndex" :value="flow.procDefKey" :label="flow.procDefName + ' [' + flow.procDefVersion + ']'"><span>{{ flow.procDefName }} [{{ flow.procDefVersion }}]</span></Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('m_callback_mode')" :delay="1000">
                <Select v-model="item.proc_callback_mode" clearable :disabled="!isEditState" style="width: 100px" :placeholder="$t('m_callback_mode')">
                  <Option v-for="item in callbackMode" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('tableKey.description')" :delay="1000">
                <input 
                  v-model="item.description"
                  clearable
                  :disabled="!isEditState"
                  style="width: 100px"
                  type="text" 
                  maxlength="50"
                  :placeholder="$t('tableKey.description')"
                  class="form-control model-input search-input c-dark"/>
              </Tooltip>
            </div>
          </Form>
        </div>
        <div class="right-content">
          <div class="use-underline-title mb-3">
            {{$t('field.metric')}}{{$t('field.threshold')}}
            <span class="underline"></span>
          </div>
          <Table
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
        <Button v-if="isEditState" class="modal-button-save" style="float:right" type="primary" @click="submitContent">{{$t('button.save')}}</Button>
        <Button style="float:right" @click="cancelModal">{{$t('button.cancel')}}</Button>
      </div>
    </Modal>

    <Modal
      v-model="isShowWarningDelete"
      :title="$t('delConfirm.title')"
      :ok-text="$t('m_confirm')"
      @on-ok="okDelRow"
      @on-cancel="cancleDelRow">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('delConfirm.tip')}}</p>
        </div>
      </div>
    </Modal>
  </div>
</template>

<script>
import {thresholdList, lastList, priorityList} from '@/assets/config/common-config.js'
import cloneDeep from 'lodash/cloneDeep';
import hasIn from 'lodash/hasIn';
import isEmpty from 'lodash/isEmpty';
import find from 'lodash/find';
import Vue from 'vue';

const initFormData = {
  name: '', // 告警名
  priority: 'low', // 级别
  notify_enable: 1, // 告警发送
  notify_delay_second: 0, // 延时
  active_window:  ['00:00','23:59'], // 告警时间段
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
      tags: {} // 有内容的标签值
    }
  ]
}

const alarmLevelMap = {
  low: {
    label: "m_low",
    buttonType: "info"
  },
  medium: {
    label: "m_medium",
    buttonType: "warning"
  },
  high: {
    label: "m_high",
    buttonType: "error"
  }
}

const statusButtonLabelMap = {
  1: "m_alarm_open",
  0: "m_alarm_close"
}


export default {
  name: '',
  props: {
    type: String // 为枚举类型，gourp为组，service为层级对象，endpoint为对象
  },
  data () {
    return {
      targetId: '', // 层级对象选中的
      totalPageConfig: [],
      selectedTableData: null,
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: 'button.add',
        modalStyle: 'min-width:98%',
        isAdd: true,
        noBtn: true,
        config: [
          {name:'metricSelect',type:'slot'},
          {name:'btn',type:'slot'}
        ],
        metricName: '',
        metricList: [],
        threshold: '>',
        thresholdList: thresholdList,
        thresholdValue: '',
        last: 's',
        lastList: lastList,
        lastValue: '',
        priority: 'low',
        priorityList: priorityList,
        notifyEnableOption: [
          {label: '启用', value: 1},
          {label: '关闭', value: 0}
        ],
        notifyDelayOption: [
          {label: '0s', value: 0},
          {label: '30s', value: 30},
          {label: '60s', value: 60},
          {label: '90s', value: 90},
          {label: '120s', value: 120},
          {label: '180s', value: 180},
          {label: '300s', value: 300},
          {label: '600s', value: 600}
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
      isShowWarningDelete: false,
      selectedData: {},
      notify: [],
      callbackMode: [
        {label: 'm_manual', value: 'manual'},
        {label: 'm_auto', value: 'auto'}
      ],
      formData: cloneDeep(initFormData),
      ruleValidate: {
        name: [
          {type: 'string', required: true, message: '请输入', trigger: 'blur' }
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
          title: this.$t('tableKey.metricName'),
          key: 'metric',
          align: 'left',
          minWidth: 150,
          render: (h, params) => {
            return (
              <Select
                value={params.row.metric}
                disabled={!this.isEditState}
                on-on-change={v => {
                  if (v) {
                    this.getTagList(v, params.index)
                  }
                  this.formData.conditions[params.index].metric = v
                }}
                filterable
                clearable
              >
                {this.modelConfig.metricList &&
                  this.modelConfig.metricList.map((i, index) => (
                    <Option value={i.guid} key={i.metric + index}>
                      {i.metric}
                    </Option>
                  ))}
              </Select>
            )
          }
        },
        {
          title: this.$t('m_label_value'),
          key: 'tags',
          align: 'left',
          width: 180,
          render: (h, params) => {
            return (
              <div>
                {!isEmpty(params.row.tagOptions) && !isEmpty(params.row.tags) && this.isModalShow ?
                  (params.row.tags.map((i, selectIndex) => (
                    <div class="tags-show" key={selectIndex + '' + JSON.stringify(i)}>
                      <span>{i.tagName}</span>
                      <Select
                        style="maxWidth: 130px"
                        value={i.tagValue}
                        disabled={!this.isEditState}
                        on-on-change={v => {
                          Vue.set(this.formData.conditions[params.index].tags[selectIndex], 'tagValue', v)
                        }}
                        filterable
                        multiple
                        clearable
                      >
                        {!isEmpty(this.formData.conditions[params.index].tagOptions) && 
                          !isEmpty(this.formData.conditions[params.index].tagOptions[i.tagName]) &&
                            this.formData.conditions[params.index].tagOptions[i.tagName].map((item, index) => (
                            <Option value={item.value} key={item.key + index}>
                              {item.key}
                            </Option>
                          ))}
                      </Select>
                    </div>
                  )) ) : '-' } 
              </div>
            )
          }
        },
        {
          title: this.$t('m_symbol'),
          key: 'threshold',
          align: 'left',
          minWidth: 50,
          render: (h, params) => {
            return (
              <Select
                value={params.row.threshold}
                disabled={!this.isEditState}
                on-on-change={v => {
                  this.formData.conditions[params.index].threshold = v
                }}
                filterable
                clearable
              >
                {this.modelConfig.thresholdList &&
                  this.modelConfig.thresholdList.map((i, index) => (
                    <Option value={i.value} key={index}>
                      {i.label}
                    </Option>
                  ))}
              </Select>
            )
          }
        },
        {
          title: this.$t('field.threshold'),
          key: 'thresholdValue',
          align: 'left',
          minWidth: 70,
          render: (h, params) => {
            return (
              <Input
                value={params.row.thresholdValue}
                disabled={!this.isEditState}
                maxlength="10"
                on-on-change={v => {
                  this.formData.conditions[params.index].thresholdValue = v.target.value
                }}
                clearable
              />
            )
          }
        },
        {
          title: this.$t('tableKey.s_last'),
          key: 'lastValue',
          align: 'left',
          minWidth: 70,
          render: (h, params) => {
            return (
              <Input
                value={params.row.lastValue}
                disabled={!this.isEditState}
                maxlength="10"
                on-on-change={v => {
                  this.formData.conditions[params.index].lastValue = v.target.value
                }}
                clearable
              />
            )
          }
        },
        {
          title: this.$t('m_time_unit'),
          key: 'lastSymbol',
          align: 'left',
          minWidth: 70,
          render: (h, params) => {
            return (
              <Select
                value={params.row.lastSymbol}
                disabled={!this.isEditState}
                on-on-change={v => {
                  this.formData.conditions[params.index].lastSymbol = v
                }}
                filterable
                clearable
              >
                {this.modelConfig.lastList &&
                  this.modelConfig.lastList.map(i => (
                    <Option value={i.value} key={i.value}>
                      {i.label}
                    </Option>
                  ))}
              </Select>
            )
          }
        },
        {
          title: this.$t('table.action'),
          key: 'index',
          align: 'center',
          width: 50,
          render: (h, params) => {
            return (
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
        }
      ],
      monitorType: "",
      // 主页中的表格
      alarmItemTableColumns: [
        {
            title: this.$t('m_alarmName'),
            align: 'center',
            width: 160,
            key: 'name'
        },
        {
          title: this.$t('m_alarmPriority'),
          key: 'priority',
          align: 'center',
          width: 100,
          render: (h, params) => {
            return (
              <Button
                type={alarmLevelMap[params.row.priority].buttonType}
                size="small"
                ghost
              >{this.$t(alarmLevelMap[params.row.priority].label)}</Button>
            )
          }
        },
        {
          title: this.$t('m_notification'),
          key: 'notify_enable',
          align: 'center',
          width: 100,
          render: (h, params) => {
            return (
              <Button
                type="success"
                disabled={params.row.notify_enable === 0}
                size="small"
                ghost
              >{this.$t(statusButtonLabelMap[params.row.notify_enable])}</Button>
            )
          }
        },
        {
            title: this.$t('field.relativeTime'),
            width: 130,
            align: 'center',
            key: 'active_window'
        },
        {
          title: this.$t('firing'),
          key: 'firing',
          align: 'left',
          render: (h, params) => {
            return !isEmpty(params.row.notify) ?
            (
              <div>
                <div>{this.$t('m_notification_role')}:{params.row.notify[0].notify_roles.join(';')} </div>
                <div>{this.$t('m_trigger_arrange')}: {params.row.notify[0].proc_callback_key}{params.row.notify[0].proc_callback_mode ? '(' + this.$t(find(this.callbackMode, {value: params.row.notify[0].proc_callback_mode}).label) + ')' : ""}</div>
                <div>{this.$t('tableKey.description')}:{params.row.notify[0].description} </div>
              </div>
            ) : <div>-</div>
          }
        },
        {
          title: this.$t('ok'),
          key: 'ok',
          align: 'left',
          render: (h, params) => {
            return !isEmpty(params.row.notify) && params.row.notify.length > 1 ? (
              <div>
                <div>{this.$t('m_notification_role')}:{params.row.notify[1].notify_roles.join(';')} </div>
                <div>{this.$t('m_trigger_arrange')}: {params.row.notify[1].proc_callback_key}{params.row.notify[1].proc_callback_mode ? '(' + this.$t(find(this.callbackMode, {value: params.row.notify[1].proc_callback_mode}).label) + ')' : ""}</div>
                <div>{this.$t('tableKey.description')}:{params.row.notify[1].description} </div>
              </div>
            ) : <div>-</div>
          }
        },
        {
          title: this.$t('tableKey.metricName'),
          key: 'metric',
          align: 'center',
          width: 250
        },
        {
          title: this.$t('field.threshold'),
          key: 'condition',
          align: 'center',
          width: 100
        },
        {
          title: this.$t('tableKey.s_last'),
          key: 'last',
          align: 'center',
          width: 100
        },
        {
          title: this.$t('table.action'),
          key: 'index',
          align: 'center',
          width: 150,
          render: (h, params) => {
            return (
              <div>
                <Button size="small" class="mr-1"  type="primary" on-click={() => this.editAlarmItem(params.row, params)}>
                  {
                    this.type==='endpoint' ? 
                    <Icon type="md-eye" size="16"></Icon> :
                    <Icon type="md-create" size="16"></Icon> 
                  }
                </Button>
                {
                  this.isEditState ? (
                    <Button size="small" type="error" on-click={() => this.deleteConfirmModal(params.row)}>
                      <Icon type="md-trash" size="16"></Icon>
                    </Button>
                  ) : null
                }
              </div>
            )
          }
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
    }
  },
  methods: {
    getUpdateNotifyApi(data) {
      return this.type === 'service' 
        ? `/monitor/api/v2/alarm/endpoint_group/${data.endpoint_group}/notify/update`
        : `/monitor/api/v2/alarm/endpoint_group/${this.targetId}/notify/update`
    },
    updateNotify (tableData) {
      const api = this.getUpdateNotifyApi(tableData);
      this.request('POST', api, tableData.notify, () => {
        this.$Message.success(this.$t('tips.success'))
        this.getDetail(this.targetId, this.type)
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
    deleteAlarmItem (rowData) {
      const api = `/monitor/api/v2/alarm/strategy/${rowData.guid}`
      this.request('DELETE', api, '', () => {
        this.$Message.success(this.$t('tips.success'))
        this.getDetail(this.targetId, this.type)
      })
    },
    deleteConfirmModal (rowData) {
      this.selectedData = rowData;
      this.isShowWarningDelete = true;
    },
    okDelRow () {
      this.deleteAlarmItem(this.selectedData)
    },
    cancleDelRow () {
      this.isShowWarningDelete = false
    },
    manageEditParams(addParams, rowparams){
      for (let key in addParams) {
        if (hasIn(rowparams, key)) {
          Vue.set(addParams, key, cloneDeep(rowparams[key]));
        }
      }
    },
    getMetricListPath(data) {
      return this.type === 'service' 
        ? `/monitor/api/v2/monitor/metric/list?monitorType=${data.monitor_type}&serviceGroup=${this.targetId}` 
        : `/monitor/api/v2/monitor/metric/list?monitorType=${this.monitorType}`;
    },
    getSymbolAndValue(str) {
      if (!str.length) return {};
      return {
        symbol: str.match(/[<>=smh]+/g)[0],
        value: str.match(/\d+/g)[0]
      }
    },
    getInitialTags() {
      return new Promise(resolve => {
        const initialMetricName = this.modelConfig.metricList[0].guid;
        this.findTagsByMetric(initialMetricName).then(res => {
          this.initialTagOptions = res;
          if (!isEmpty(this.initialTagOptions)) {
            this.initialTags = [];
            for(let key in this.initialTagOptions) {
              this.initialTags.push(
                {
                  tagName: key,
                  tagValue: []
                }
              )
            }
          }
          resolve()
        })
      })
    },
    async editAlarmItem (rowData, allParams) {
      this.isfullscreen = true
      this.currentAlarmListIndex = rowData.alarmListIndex;
      this.selectedData = rowData
      const api = this.getMetricListPath(rowData);
      this.request('GET', api, '', (responseData) => {
        this.modelConfig.metricList = responseData;
        this.modelConfig.isAdd = false
        this.modelConfig.modalTitle = 'button.edit',
        this.formData = cloneDeep(initFormData);
        this.manageEditParams(this.formData, rowData);
        this.formData.active_window = rowData.active_window === '' ? ['00:00', '23:59'] : rowData.active_window.split('-');
        const conditions = this.formData.conditions;
        conditions.length && conditions.forEach(async item => {
            const thresholdsAndSymbols = this.getSymbolAndValue(item.condition);
            const lastValueAndSymbol = this.getSymbolAndValue(item.last);
            const tagOptions = await this.findTagsByMetric(item.metric); // 指标下拉option获取
            if (isEmpty(item.tags) && !isEmpty(tagOptions)) {
                const tags = [];
                for(let key in tagOptions) {
                    tags.push(
                        {
                            tagName: key,
                            tagValue: []
                        }
                    )
                }
                Vue.set(item, 'tags', tags)
            }
            Vue.set(item, 'threshold', thresholdsAndSymbols.symbol)
            Vue.set(item, 'thresholdValue', thresholdsAndSymbols.value)
            Vue.set(item, 'lastSymbol', lastValueAndSymbol.symbol);
            Vue.set(item, 'lastValue', lastValueAndSymbol.value);
            Vue.set(item, 'tagOptions', tagOptions);
        })
        this.showAddEditModal();
      })
    },
    addAlarmItem (tableItem, index) {
      this.isfullscreen = true
      this.currentAlarmListIndex = index;
      this.selectedTableData = tableItem;
      const api = this.getMetricListPath(tableItem);
      this.request('GET', api, '', async responseData => {
          this.modelConfig.metricList = responseData;
          this.modelConfig.isAdd = true
          this.modelConfig.modalTitle = 'button.add',
          this.formData = cloneDeep(initFormData);
          this.formData.name = this.$t('m_alert') + new Date().getTime()
          await this.getInitialTags();
          this.formData.conditions[0].metric = this.modelConfig.metricList[0].guid; // 指标名
          this.formData.conditions[0].tags = cloneDeep(this.initialTags);
          this.formData.conditions[0].tagOptions = cloneDeep(this.initialTagOptions);
          this.showAddEditModal();
      })
    },
    getWorkFlow () {
      this.request('GET', '/monitor/api/v2/alarm/event/callback/list', '', (responseData) => {
        this.flows = responseData
      })
    },
    getAllRole () {
      this.request('GET', '/monitor/api/v1/user/role/list?page=1&size=1000', '', (responseData) => {
        this.allRole = responseData.data.map((_) => {
          return {
            ..._,
            value: _.id
          }
        })
      })
    },
    processConditions(conditions) {
      if (!conditions.length) return 
      conditions.forEach(item => {
        item.condition = item.threshold + item.thresholdValue;
        item.last = item.lastValue + item.lastSymbol;
        const needDeleteAttr = ['lastSymbol', 'lastValue', 'threshold', 'thresholdValue', 'tagOptions'];
        needDeleteAttr.forEach(key => {
          delete item[key];
        })
      })
    },
    validateConditions(conditions) {
        if (!conditions || isEmpty(conditions)) return false;
        const needValidateKey = ['metric', 'lastSymbol', 'lastValue', 'threshold', 'thresholdValue']
        for(let i=0; i<conditions.length; i++) {
            for(let k=0; k < needValidateKey.length; k++) {
                if (!conditions[i][needValidateKey[k]]) {
                    return false
                }
            } 
        }
        return true
    },
    validateDuplicateName(alarmName, guid = '') {
        const currentTableList =  this.totalPageConfig[this.currentAlarmListIndex].tableData;
        if (isEmpty(currentTableList)) return true
        for(let i=0; i<currentTableList.length; i++) {
            const item = currentTableList[i];
            if (item.guid !== guid && item.name === alarmName) {
                return false
            }
        }
        return true
    },
    submitContent() {
      if (this.formData.name.trim() === '') {
        return this.$Message.error(this.$t('m_alarmName') + this.$t('tips.required'));
      }
      if (!this.validateConditions(this.formData.conditions)) {
        return this.$Message.error(this.$t('m_metric_threshold') + this.$t('tips.emptyToSave'));
      }
      if (!this.validateDuplicateName(this.formData.name, this.selectedData.guid)) {
        return this.$Message.error(this.$t('m_alarmName') + this.$t('m_cannot_be_repeated'));
      }
      let params = cloneDeep(this.formData);
      const needMergeParams = {
        endpoint_group: this.modelConfig.isAdd ? this.selectedTableData.endpoint_group : this.selectedData.endpoint_group,
        guid: this.selectedData.guid,
        active_window: params.active_window.join('-')
      }
      Object.assign(params, needMergeParams)
      this.processConditions(params.conditions);
      const requestMethod = this.modelConfig.isAdd ? "POST" : "PUT";
      this.request(requestMethod, '/monitor/api/v2/alarm/strategy', params, () => {
        this.$Message.success(this.$t('tips.success'));
        this.closeAddEditModal()
        this.getDetail(this.targetId, this.type)
      })
    },
    showAddEditModal() {
      this.isShowAddEditModal = true
      this.isModalShow = true;
    },
    closeAddEditModal() {
      this.isShowAddEditModal = false
      this.isModalShow = false
    },
    setMonitorType(monitorType){
      this.monitorType = monitorType;
    },
    handleTableData(tempTableData, alarmListIndex) {
      const initialData = cloneDeep(tempTableData);
      const resData = [];
      let startRowIndex = 0;
      this.mergeSpanMap = {};
      initialData.forEach(item => {
        item.notify = !isEmpty(item.notify) ? item.notify : cloneDeep(initFormData.notify);
        if(item.conditions && item.conditions.length) {
          item.conditions.forEach(metricItem => {
            resData.push(Object.assign({}, item, metricItem, {alarmListIndex}))
          })
          this.mergeSpanMap[item.guid] = {
            startRowIndex,
            colSpan: item.conditions.length
          }
          startRowIndex += item.conditions.length;
        }
      })
      return resData;
    },
    getDetail (targetId) {
      this.targetId = targetId
      const api = '/monitor/api/v2/alarm/strategy/list' + `/${this.type}/` + targetId
      this.totalPageConfig = []
      this.request('GET', api, '', responseData => {
        const allConfigDetail = responseData;
        allConfigDetail.forEach((item, alarmIndex) => {
          let tempTableData = item.strategy.map(s => {
            s.monitor_type = item.monitor_type
            return s
          })
          const tableData = this.handleTableData(tempTableData, alarmIndex);
          this.totalPageConfig.push({
            tableData,
            endpoint_group: item.endpoint_group,
            display_name: item.display_name,
            monitor_type: item.monitor_type,
            notify: item.notify,
            mergeSpanMap: this.mergeSpanMap
          })
        })
      }, {isNeedloading:true})
      this.getAllRole()
      this.getWorkFlow()
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
    cancelModal () {
      this.closeAddEditModal();
    },
    handleMergeSpan ({ row, rowIndex, columnIndex }, index) {
      if ([6,7,8].includes(columnIndex)) return;
      const mergeSpanMap = this.totalPageConfig[index].mergeSpanMap;
      const spanMap = mergeSpanMap[row.guid];
      if (rowIndex === spanMap.startRowIndex) {
        return [spanMap.colSpan,1]
      }
      if (rowIndex !== spanMap.startRowIndex && (columnIndex <= 5 || columnIndex === 9)) {
        return [0,0]
      }
    },
    findTagsByMetric(metricId) {
      const api = '/monitor/api/v2/metric/tag/value-list';
      const params = {
        metricId,
        endpoint: this.targetId,
        serviceGroup: this.targetId
      }
      
      return new Promise(resolve => {
        this.request('POST', api, params, responseData => {
          let result = {}
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
      const item = this.formData.conditions[tableIndex];
      item.tagOptions = await this.findTagsByMetric(metricId);
      
      if (!isEmpty(item.tagOptions)) {
        const tags = [];
        for(let key in item.tagOptions) {
          tags.push(
            {
              tagName: key,
              tagValue: []
            }
          )
        }
        Vue.set(item, 'tags', tags)
      }
    }
  },
  computed: {
    isEditState() {
      return this.type !== 'endpoint'
    }
  }
}
</script>

<style lang="less">
.ivu-table-cell {
  padding: 4px;
}
.ivu-table-wrapper {
  overflow: inherit;
}
.ivu-form-item {
  margin-bottom: 0;
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
  justify-content: center;
} 
.tags-show > span {
  width: 70px;
  overflow: scroll;
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

</style>
<style scoped lang="less">
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
    .arrange-config-Title {
      display: inline-block;
      font-size: 14px;
      color: #515a6e;
      margin: 5px 0px 20px 2px;
    }
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
</style>