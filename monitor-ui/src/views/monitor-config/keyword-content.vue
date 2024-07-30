<template>
  <div class=" ">
    <section class="section-content">
      <div class="upload-content">
        <Button
          type="info"
          class="btn-left"
          v-if="isEditState"
          @click="exportData"
        >
          <img src="../../assets/img/export.png" class="btn-img" alt="" />
          {{ $t('m_export') }}
        </Button>
        <Upload
          v-if="isEditState"
          :action="uploadUrl"
          :show-upload-list="false"
          :max-size="1000"
          with-credentials
          :headers="{'Authorization': token}"
          :on-success="uploadSucess"
          :on-error="uploadFailed"
        >
          <!-- <Button icon="ios-cloud-upload-outline">{{$t('m_import')}}</Button> -->
          <Button type="primary" class="btn-left">
            <img src="../../assets/img/import.png" class="btn-img" alt="" />
            {{ $t('m_import') }}
          </Button>
        </Upload>
      </div>
      <div class="content-title">
        <div class="use-underline-title">
          {{$t('m_log_file')}}
          <span class="underline" ></span>
        </div>
        <Button v-if="isEditState" type="success" class="mr-4" @click="addLogFileConfig">
          {{ $t('m_button_add') }}
        </Button>
      </div>

      <Collapse v-model="keywordCollapseValue">
        <Panel v-for="(item, index) in keywordCollapseData"
               :key="index"
               :name="index + ''"
        >
          <div class="keyword-collapse-content">
            <div>
              <div class="use-underline-title mr-4">
                {{item.log_path}}
                <span class="underline"></span>
              </div>
              <Tag color="blue">{{ item.monitor_type }}</Tag>
            </div>
            <div v-if="isEditState" class="key-word-collapse-button" @click="(e) => {e.stopPropagation()}">
              <Tooltip :content="$t('m_title_logAdd')" placement="bottom" transfer>
                <Button
                  size="small"
                  type="success"
                  @click.stop="addCustomMetric(item)"
                >
                  <Icon type="ios-add-circle" size="16" />
                </Button>
              </Tooltip>
              <Tooltip :content="$t('m_keyword_edit')" placement="bottom" transfer>
                <Button size="small" class="mr-1"  type="primary" @click.stop="editF(item)">
                  <Icon type="md-create" size="16" />
                </Button>
              </Tooltip>
              <Tooltip :content="$t('m_alarm_configuration_save')" placement="bottom" transfer>
                <Button size="small" class="mr-1"  type="success" @click.stop="saveFileDetail(item)">
                  <Icon type="ios-folder-open" size="16" />
                </Button>
              </Tooltip>
              <Poptip
                confirm
                :title="$t('m_delConfirm_tip')"
                placement="left-end"
                @on-ok="deleteLogFileItem(item)"
              >
                <Button size="small" type="error" class="mr-2">
                  <Icon type="md-trash" size="16" />
                </Button>
              </Poptip>
            </div>
            <Button v-else style="margin-right: 124px" size="small" type="info" @click.stop="editF(item)">
              <Icon type="md-eye" size="16" />
            </Button>
          </div>
          <template slot='content'>
            <Table
              class="keyword-table"
              size="small"
              :columns="keywordTableColumns"
              :data="item.keyword_list"
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
              <template>
                <div style="margin: 4px 0">
                  <Row v-if="item.notify">
                    <Col span="2">
                    <span style="margin-right: 8px;line-height: 32px;">{{$t('firing')}}</span>
                    </Col>
                    <Col span="6" style="">
                    <Select v-model="item.notify.notify_roles" :disabled="!isEditState" :max-tag-count="2" style="width: 99%;" multiple filterable :placeholder="$t('m_field_role')">
                      <Option v-for="role in allRoles" :value="role.name" :key="role.value">{{ role.name }}</Option>
                    </Select>
                    </Col>
                    <Col span="5">
                    <Select v-model="item.notify.proc_callback_key"
                            clearable
                            :disabled="!isEditState"
                            @on-change="procCallbackKeyChangeForm(item.notify.proc_callback_key, item.notify)"
                            style="width:99%;"
                            :placeholder="$t('proc_callback_key')"
                    >
                      <Option v-for="(flow, flowIndex) in allFlows" :key="flowIndex" :value="flow.procDefKey" :label="flow.procDefName + ' [' + flow.procDefVersion + ']'"><span>{{ flow.procDefName }} [{{ flow.procDefVersion }}]</span></Option>
                    </Select>
                    </Col>
                    <Col span="5">
                    <Select v-model="item.notify.proc_callback_mode" clearable :disabled="!isEditState" style="width:99%" :placeholder="$t('m_callback_mode')">
                      <Option v-for="mode in callbackMode" :value="mode.value" :key="mode.value">{{ $t(mode.label) }}</Option>
                    </Select>
                    </Col>
                    <Col span="5">
                    <Input
                      v-model="item.notify.description"
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
          </template>
        </Panel>
      </Collapse>
    </section>

    <section style="margin-top: 16px; padding-bottom: 30px">
      <div class="content-title">
        <div class="use-underline-title">
          {{$t('m_db')}}
          <span class="underline"></span>
        </div>
        <Button v-if="isEditState" type="success" class="mr-4" @click="addDataBase" style="margin: 8px 0">
          {{ $t('m_button_add') }}
        </Button>
      </div>

      <Table
        size="small"
        :columns="dataBaseTableColumns"
        :data="dataBaseTableData"
      />
    </section>
    <Modal
      v-model="addAndEditModal.isShow"
      :title="addAndEditModal.isAdd ? $t('m_button_add') : $t('')"
      :mask-closable="false"
      :width="730"
    >
      <div :style="{'max-height': MODALHEIGHT + 'px', overflow: 'auto'}">
        <div>
          <span>{{$t('m_field_type')}}:</span>
          <Select v-model="addAndEditModal.dataConfig.monitor_type"
                  :disabled="!isEditState"
                  @on-change="getEndpoint(addAndEditModal.dataConfig.monitor_type, 'host')"
                  style="width: 640px"
          >
            <Option v-for="type in monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
          </Select>
        </div>
        <div v-if="addAndEditModal.isAdd" style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;width:680px">
          <template v-for="(item, index) in addAndEditModal.pathOptions">
            <p :key="index + 5">
              <Button
                v-if="addAndEditModal.isAdd"
                :disabled="!isEditState"
                @click="deleteItem('path', index)"
                size="small"
                type="error"
                icon="md-trash"
              ></Button>
              <Tooltip :content="$t('m_tableKey_logPath')" :delay="1000">
                <Input v-model="item.path"
                       :disabled="!isEditState"
                       style="width: 620px"
                       :placeholder="$t('m_tableKey_logPath')"
                />
              </Tooltip>
            </p>
          </template>
          <Button
            @click="addEmptyItem('path')"
            :disabled="!isEditState"
            type="success"
            size="small"
            style="width:650px"
            long
          >{{ $t('m_button_add') }}{{$t('m_tableKey_logPath')}}</Button>
        </div>
        <div v-else style="margin: 8px 0">
          <span>{{$t('m_tableKey_path')}}:</span>
          <Input :disabled="!isEditState" style="width: 640px" v-model="addAndEditModal.dataConfig.log_path" />
        </div>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;width:680px">
          <template v-for="(item, index) in addAndEditModal.dataConfig.endpoint_rel">
            <p :key="index + 'c'">
              <Button
                :disabled="!isEditState"
                @click="deleteItem('relate', index)"
                size="small"
                type="error"
                icon="md-trash"
              ></Button>
              <Tooltip :content="$t('m_business_object')" :delay="1000">
                <Select v-model="item.target_endpoint"
                        :disabled="!isEditState"
                        style="width: 310px"
                        :placeholder="$t('m_business_object')"
                >
                  <Option v-for="type in targetEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('m_log_server')" :delay="1000">
                <Select
                  v-model="item.source_endpoint"
                  :disabled="!isEditState"
                  style="width: 310px"
                  :placeholder="$t('m_log_server') + '333'"
                >
                  <Option v-for="type in sourceEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select>
              </Tooltip>
            </p>
          </template>
          <Button
            @click="addEmptyItem('relate')"
            type="success"
            :disabled="!isEditState"
            size="small"
            style="width:650px"
            long
          >{{$t('addStringMap')}}</Button>
        </div>
      </div>
      <div slot="footer">
        <Button @click="cancelAddAndEdit" :disabled="!isEditState">{{$t('m_button_cancel')}}</Button>
        <Button @click="okAddAndEdit" :disabled="!isEditState" type="primary">{{$t('m_button_save')}}</Button>
      </div>
    </Modal>

    <Drawer :title="isAddState ? $t('m_add') : $t('m_button_edit')"
            v-model="isTableChangeFormShow"
            :width="70"
            @on-close="onTableChangeFormClose"
            :mask-closable="false"
            class="config-drawer"
    >
      <div>
        <div class="file-log-form">
          <Form ref="formData" :model="formData" :rules="ruleValidate" :label-width="130">
            <FormItem :label="$t('m_alarmName')" prop="name">
              <Input v-model="formData.name" :disabled="!isEditState" :maxlength="10"></Input>
            </FormItem>
            <FormItem v-if="isLogFile" :label="$t('m_field_log')" prop="keyword">
              <Input v-model="formData.keyword" :disabled="!isEditState" :maxlength="10"></Input>
            </FormItem>
            <FormItem v-if="isLogFile" :label="$t('m_regular')" prop="regulative">
              <Select
                v-model="formData.regulative"
                :disabled="!isEditState"
              >
                <Option
                  v-for="item in notifyEnableOption"
                  :value="item.value"
                  :label="item.label"
                  :key="item.value"
                >
                  {{ $t(item.label) }}</Option>
              </Select>
            </FormItem>
            <FormItem v-if="!isLogFile" :label="$t('m_sql_script')" prop="query_sql">
              <Input
                type="textarea"
                :disabled="!isEditState"
                v-model="formData.query_sql"
                :maxlength="200"
              >
              </Input>
            </FormItem>
            <FormItem v-if="!isLogFile" :label="$t('m_type')" prop="monitor_type">
              <Select v-model="formData.monitor_type"
                      :disabled="!isEditState"
                      @on-change="getSqlSourceOptions"
              >
                <Option v-for="type in monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
              </Select>
            </FormItem>
            <FormItem v-if="!isLogFile" :label="$t('m_database_map')" prop="endpoint_rel">
              <div class="database-map-content">
                <template v-for="(item, index) in formData.endpoint_rel">
                  <div :key="index + 'S'" class="database-map-item">
                    <Tooltip :content="$t('m_db')" :delay="1000">
                      <Select v-model="item.target_endpoint" :disabled="!isEditState" clearable  :placeholder="$t('m_target_value')">
                        <Option v-for="type in sqlTargetEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                      </Select>
                    </Tooltip>
                    <Tooltip :content="$t('m_source_value')" :delay="1000">
                      <Select v-model="item.source_endpoint" :disabled="!isEditState" clearable :placeholder="$t('m_source_value')">
                        <Option v-for="type in sqlSourceEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                      </Select>
                    </Tooltip>
                    <Button
                      @click="deleteDatabaseMapItem(index)"
                      :disabled="!isEditState"
                      size="small"
                      type="error"
                      icon="md-trash"
                    ></Button>
                  </div>
                </template>
                <Button
                  @click="addMetricConfig"
                  :disabled="!isEditState"
                  type="success"
                  size="small"
                  long
                >{{ $t('addMetricConfig') }}</Button>
              </div>
            </FormItem>
            <FormItem :label="$t('m_tableKey_s_priority')" prop="priority">
              <Select
                v-model="formData.priority"
                :disabled="!isEditState"
              >
                <Option
                  v-for="item in priorityList"
                  :value="item.value"
                  :key="item.value"
                >
                  {{ $t(item.label) }}</Option>
              </Select>
            </FormItem>
            <FormItem :label="$t('m_tableKey_status')" prop="notify_enable">
              <Select
                :disabled="!isEditState"
                v-model="formData.notify_enable"
              >
                <Option
                  v-for="item in notifyEnableOption"
                  :value="item.value"
                  :label="item.label"
                  :key="item.value"
                >
                  {{ item.label }}</Option>
              </Select>
            </FormItem>
            <FormItem v-if="!isLogFile" :label="$t('delay')" prop="notify_delay_second">
              <Select
                filterable
                :disabled="!isEditState"
                v-model="formData.notify_delay_second"
              >
                <Option
                  v-for="item in notifyDelayOption"
                  :value="item.value"
                  :key="item.value"
                  :label="item.label"
                >
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
                style="width: 168px"
              >
              </TimePicker>
            </FormItem>
            <FormItem :label="$t('m_tableKey_content')" prop="content">
              <Input
                type="textarea"
                :disabled="!isEditState"
                v-model="formData.content"
                :maxlength="200"
              >
              </Input>
            </FormItem>
          </Form>
        </div>
        <div style="margin:20px 0 8px">
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
        <div class="arrange-item">
          <span class="mr-1 mt-1" style="font-size: 12px">{{$t('firing')}}</span>
          <Tooltip :content="$t('m_resourceLevel_role')" :delay="1000">
            <Select v-model="formData.notify.notify_roles" :disabled="!isEditState" clearable :max-tag-count="2" style="width: 200px" multiple filterable :placeholder="$t('m_field_role')">
              <Option v-for="item in allRoles" :value="item.name" :key="item.value">{{ item.name }}</Option>
            </Select>
          </Tooltip>
          <Tooltip :content="$t('proc_callback_key')" :delay="1000">
            <Select v-model="formData.notify.proc_callback_key" :disabled="!isEditState" clearable @on-change="procCallbackKeyChangeForm(formData.notify.proc_callback_key, formData.notify)" style="width: 160px" :placeholder="$t('proc_callback_key')">
              <Option v-for="(flow, flowIndex) in allFlows" :key="flowIndex" :value="flow.procDefKey" :label="flow.procDefName + ' [' + flow.procDefVersion + ']'"><span>{{ flow.procDefName }} [{{ flow.procDefVersion }}]</span></Option>
            </Select>
          </Tooltip>
          <Tooltip :content="$t('m_callback_mode')" :delay="1000">
            <Select v-model="formData.notify.proc_callback_mode" :disabled="!isEditState" clearable style="width: 180px" :placeholder="$t('m_callback_mode')">
              <Option v-for="item in callbackMode" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
            </Select>
          </Tooltip>
          <Tooltip :content="$t('m_tableKey_description')" :delay="1000">
            <Input
              v-model="formData.notify.description"
              :disabled="!isEditState"
              style="width: 150px"
              type="text"
              clearable
              maxlength="50"
              :placeholder="$t('m_tableKey_description')"
              class="model-input search-input c-dark"
            />
          </Tooltip>
        </div>
        <div class="form-footer" v-if="isTableChangeFormShow && isEditState">
          <Button @click="onDrawerClose" :disabled="!isEditState" class="mr-4">{{$t('m_button_cancel')}}</Button>
          <Button @click="onFormSave" :disabled="!isEditState" type="primary">{{$t('m_button_save')}}</Button>
        </div>
      </div>
    </Drawer>
  </div>
</template>

<script>
import axios from 'axios'
import cloneDeep from 'lodash/cloneDeep'
import isEmpty from 'lodash/isEmpty'
import hasIn from 'lodash/hasIn'
import Vue from 'vue'
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import {baseURL_config} from '@/assets/js/baseURL'
import {priorityList} from '@/assets/config/common-config.js'

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

const initNotify = {
  guid: '',
  endpoint_group: '',
  service_group: '',
  alarm_strategy: '',
  alarm_action: 'firing',
  alarm_priority: '',
  notify_num: 1,
  proc_callback_name: '',
  proc_callback_key: '',
  callback_url: '',
  callback_param: '',
  notify_roles: [],
  proc_callback_mode: '',
  description: ''
}

const statusButtonLabelMap = {
  1: 'm_alarm_open',
  0: 'm_alarm_close'
}

const initFormData = {
  guid: '',
  service_group: '',
  name: '',
  keyword: '',
  regulative: '',
  query_sql: '',
  monitor_type: '',
  endpoint_rel: [],
  priority: 'low',
  notify_enable: 1, // 通知,默认启用
  notify_delay_second: 0,
  active_window: ['00:00','23:59'], // 告警时间段
  content: '', // 通知内容
  notify: {
    alarm_action: 'firing',
    proc_callback_key: '',
    notify_roles: [],
    proc_callback_mode: '',
    description: ''
  }
}

export default {
  name: '',
  props: {
    keywordType: {
      type: String,
      default: 'service' // 为枚举，service代表层级对象，endpoint代表对象
    }
  },
  data() {
    return {
      token: null,
      MODALHEIGHT: 300,
      targetId: '',
      addAndEditModal: {
        isShow: false,
        isAdd: false,
        dataConfig: {
          service_group: '',
          log_path: '',
          monitor_type: '',
          endpoint_rel: []
        },
        pathOptions: [],
      },
      sourceEndpoints: [],
      targetEndpoints: [],
      monitorTypeOptions: [
        {
          label: 'process',
          value: 'process'
        },
        {
          label: 'java',
          value: 'java'
        },
        {
          label: 'nginx',
          value: 'nginx'
        },
        {
          label: 'http',
          value: 'http'
        },
        {
          label: 'mysql',
          value: 'mysql'
        }
      ],

      keywordCollapseValue: ['0'],
      keywordCollapseData: [],
      keywordTableColumns: [ // 日志文件table
        {
          title: this.$t('m_alarmName'),
          width: 150,
          key: 'name'
        },
        {
          title: this.$t('m_alarmPriority'),
          key: 'priority',
          width: 100,
          render: (h, params) => (
            <Tooltip placement="right" max-width="400">
              <Tag color={alarmLevelMap[params.row.priority].buttonType}>{this.$t(alarmLevelMap[params.row.priority].label)}</Tag>
            </Tooltip>
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
          key: 'active_window',
          render: (h, params) => <div>{params.row.active_window ? params.row.active_window : '-' }</div>
        },
        {
          title: this.$t('firing'),
          key: 'firing',
          width: 150,
          align: 'left',
          render: (h, params) => {
            const { showBtn, result } = this.mgmtConfigDetail(params.row.notify)
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
          title: this.$t('m_field_log'), // 更新人
          key: 'keyword',
          render: (h, params) => <span>{params.row.keyword || '-'}</span>
        },
        {
          title: this.$t('m_regular'), // 更新人
          key: 'regulative',
          render: (h, params) => {
            const regulative_enable = params.row.regulative === 0
            return (
              <Tag color={regulative_enable ? 'default' : 'green'}>{this.$t(statusButtonLabelMap[params.row.regulative])}</Tag>
            )
          }
        },
        {
          title: this.$t('m_table_action'),
          key: 'index',
          width: 160,
          render: (h, params) => this.isEditState ? (
            <div>
              <Button size="small" class="mr-1" type='primary' on-click={() => this.editCustomMetricItem(params.row)}>
                <Icon type="md-create" size="16" />
              </Button>
              <Poptip
                confirm
                title={this.$t('m_delConfirm_tip')}
                placement="left-end"
                on-on-ok={() => {
                  this.delCustomMericsItem(params.row)
                }}>
                <Button size="small" type="error">
                  <Icon type="md-trash" size="16" />
                </Button>
              </Poptip>
            </div>
          ) : (
            <Button size="small" class="mr-1" type='info' on-click={() => this.editCustomMetricItem(params.row)}>
              <Icon type="md-eye" size="16" />
            </Button>
          )
        }
      ],
      dataBaseTableColumns: [ // 数据库表结构
        {
          title: this.$t('m_alarmName'),
          width: 150,
          key: 'name'
        },
        {
          title: this.$t('m_alarmPriority'),
          key: 'priority',
          width: 100,
          render: (h, params) => (
            <Tooltip placement="right" max-width="400">
              <Tag color={alarmLevelMap[params.row.priority].buttonType}>{this.$t(alarmLevelMap[params.row.priority].label)}</Tag>
            </Tooltip>
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
          width: 200,
          key: 'active_window',
          render: (h, params) => <div>{params.row.active_window ? params.row.active_window : '-' }</div>
        },
        {
          title: this.$t('firing'),
          key: 'firing',
          width: 150,
          align: 'left',
          render: (h, params) => {
            const { showBtn, result } = this.mgmtConfigDetail(params.row.notify)
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
          title: this.$t('m_sql_script'), // 更新人
          key: 'query_sql',
          render: (h, params) => <span>{params.row.query_sql || '-'}</span>
        },
        {
          title: this.$t('m_table_action'),
          key: 'index',
          width: 173,
          render: (h, params) => this.isEditState ? (
            <div>
              <Button size="small" class="mr-1" type="primary" on-click={() => this.editDataBaseItem(params.row)}>
                <Icon type="md-create" size="16"></Icon>
              </Button>
              <Poptip
                confirm
                title={this.$t('m_delConfirm_tip')}
                placement="left-end"
                on-on-ok={() => this.deleteDataBaseItem(params.row)}>
                <Button size="small" type="error">
                  <Icon type="md-trash" size="16" />
                </Button>
              </Poptip>
            </div>
          ) : (
            <Button size="small" class="mr-1" type="info" on-click={() => this.editDataBaseItem(params.row)}>
              <Icon type="md-eye" size="16"></Icon>
            </Button>
          )
        }
      ],
      allRoles: [],
      allFlows: [],
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
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
      dataBaseTableData: [],
      isTableChangeFormShow: false,
      isAddState: true,
      formData: cloneDeep(initFormData),
      regularOptionList: [],
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
      ruleValidate: {
        name: [
          {
            type: 'string',
            required: true,
            message: '请输入',
            trigger: 'blur'
          }
        ],
        keyword: [
          {
            type: 'string',
            required: true,
            message: '请输入',
            trigger: 'blur'
          }
        ],
        regulative: [
          {
            type: 'number',
            required: true,
            message: '请输入',
            trigger: 'blur'
          }
        ],
        query_sql: [
          {
            type: 'string',
            required: true,
            message: '请输入',
            trigger: 'blur'
          }
        ],
        priority: [
          {
            type: 'string',
            required: true,
            message: '请输入',
            trigger: 'blur'
          }
        ],
        notify_enable: [
          {
            type: 'number',
            required: true,
            message: '请输入',
            trigger: 'blur'
          }
        ],
        active_window: [
          {
            type: 'array',
            required: true,
            message: '请输入',
            trigger: 'blur'
          }
        ]
      },
      sqlSourceEndpoints: [],
      sqlTargetEndpoints: [],
      currentEditType: 'logFile', // 为枚举值，logFile(日志文件新增和编辑)和database（数据库新增和编辑）
      service_group: ''
    }
  },
  computed: {
    uploadUrl() {
      return baseURL_config + `${this.$root.apiCenter.bussinessMonitorImport}?serviceGroup=${this.targetId}`
    },
    isEditState() {
      return this.keywordType === 'service'
    },
    isLogFile() {
      return this.currentEditType === 'logFile'
    }
  },
  mounted() {
    this.MODALHEIGHT = document.body.scrollHeight - 300
    this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
    this.getWorkFlows()
    this.getAllRoles()
  },
  methods: {
    getWorkFlows() {
      this.request('GET', '/monitor/api/v2/alarm/event/callback/list', '', responseData => {
        this.allFlows = responseData
      })
    },
    getAllRoles() {
      this.request('GET', '/monitor/api/v1/user/role/list?page=1&size=1000', '', responseData => {
        this.allRoles = responseData.data.map(_ => ({
          ..._,
          value: _.id
        }))
      })
    },
    mgmtConfigDetail(val) {
      let res = {
        showBtn: false,
        result: {
          role: '-',
          arrange: '-',
          description: '-'
        }
      }
      if (!val) {
        return res
      }
      res = {
        showBtn: true,
        result: {
          role: !val.notify_roles || val.notify_roles.length === 0 ? '-' : val.notify_roles.join(';'),
          arrange: (!val.proc_callback_name && !val.proc_callback_mode) ? '-' : val.proc_callback_name + (val.proc_callback_mode ? '(' + this.$t(find(this.callbackMode, {value: val.proc_callback_mode}).label) + ')' : ''),
          description: val.description || '-'
        }
      }
      if (val.notify_roles.length === 0 && val.proc_callback_key === '' && val.proc_callback_mode === '') {
        res.showBtn = false
      }
      return res
    },
    exportData() {
      const api = `${this.$root.apiCenter.bussinessMonitorExport}?serviceGroup=${this.targetId}`
      axios({
        method: 'GET',
        url: api,
        headers: {
          Authorization: this.token
        }
      }).then(response => {
        if (response.status < 400) {
          const content = JSON.stringify(response.data)
          const fileName = `${response.headers['content-disposition'].split(';')[1].trim().split('=')[1]}`
          const blob = new Blob([content])
          if ('msSaveOrOpenBlob' in navigator){
            // Microsoft Edge and Microsoft Internet Explorer 10-11
            window.navigator.msSaveOrOpenBlob(blob, fileName)
          }
          else {
            if ('download' in document.createElement('a')) { // 非IE下载
              const elink = document.createElement('a')
              elink.download = fileName
              elink.style.display = 'none'
              elink.href = URL.createObjectURL(blob)
              document.body.appendChild(elink)
              elink.click()
              URL.revokeObjectURL(elink.href) // 释放URL 对象
              document.body.removeChild(elink)
            }
            else { // IE10+下载
              navigator.msSaveOrOpenBlob(blob, fileName)
            }
          }
        }
      })
        .catch(() => {
          this.$Message.warning(this.$t('m_tips_failed'))
        })
    },
    uploadSucess() {
      this.$Message.success(this.$t('m_tips_success'))
      this.getDetail(this.targetId)
    },
    uploadFailed(file) {
      this.$Message.warning(file.message)
    },
    // other config
    editF(rowData) {
      this.service_group = rowData.service_group
      this.getEndpoint(rowData.monitor_type, 'host')
      this.cancelAddAndEdit()
      this.addAndEditModal.isAdd = false
      this.addAndEditModal.addRow = rowData
      this.addAndEditModal.dataConfig.guid = rowData.guid
      this.addAndEditModal.dataConfig.service_group = rowData.service_group
      this.addAndEditModal.dataConfig.monitor_type = rowData.monitor_type
      this.addAndEditModal.dataConfig.log_path = rowData.log_path
      this.addAndEditModal.dataConfig.endpoint_rel = rowData.endpoint_rel
      this.addAndEditModal.isShow = true
    },
    saveFileDetail(item) {
      const params = {
        log_keyword_monitor: item.guid,
        notify: item.notify
      }
      const api = '/monitor/api/v2/service/log_keyword/notify'
      this.request('POST', api, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
      })
    },
    addCustomMetric(item) {
      this.currentEditType = 'logFile'
      this.resetDrawerForm()
      this.formData.log_keyword_monitor = item.guid
      this.isAddState = true
      this.isTableChangeFormShow = true
    },
    editCustomMetricItem(rowData) {
      this.currentEditType = 'logFile'
      this.resetDrawerForm()
      this.fillingFormData(rowData)
      this.isAddState = false
      this.isTableChangeFormShow = true
    },
    fillingFormData(rowData) {
      for (const key in this.formData) {
        if (hasIn(rowData, key)) {
          Vue.set(this.formData, key, rowData[key])
        }
      }
      this.formData.active_window = rowData.active_window ? rowData.active_window.split('-') : ['00:00', '23:59']
      this.formData.notify = isEmpty(this.formData.notify) ? cloneDeep(initNotify) : this.formData.notify
    },
    delCustomMericsItem(rowData) {
      const api = '/monitor/api/v2/service/log_keyword/log_keyword_config'
      this.request('DELETE', api, {
        guid: rowData.guid
      }, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getDetail(this.targetId)
      })
    },
    deleteLogFileItem(item) {
      const api = '/monitor/api/v2/service/log_keyword/log_keyword_monitor' + '/' + item.guid
      this.request('DELETE', api, '', () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getDetail(this.targetId)
      })
    },
    okAddAndEdit() {
      const params = JSON.parse(JSON.stringify(this.addAndEditModal.dataConfig))
      const methodType = this.addAndEditModal.isAdd ? 'POST' : 'PUT'
      params.service_group = this.targetId
      if (this.addAndEditModal.isAdd) {
        params.log_path = this.addAndEditModal.pathOptions.map(p => p.path)
      }
      this.request(methodType, '/monitor/api/v2/service/log_keyword/log_keyword_monitor', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.addAndEditModal.isShow = false
        this.getDetail(this.targetId)
      }, {isNeedloading: false})
    },
    cancelAddAndEdit() {
      this.addAndEditModal.isShow = false
      this.addAndEditModal.pathOptions = []
      this.addAndEditModal.dataConfig = {
        service_group: '',
        log_path: [],
        monitor_type: '',
        endpoint_rel: []
      }
    },
    async getEndpoint(val, type) {
      this.addAndEditModal.dataConfig.endpoint_rel = []
      await this.getDefaultConfig(val, type)
      const sourceApi = this.$root.apiCenter.getEndpointsByType + '/' + (this.isEditState ? this.targetId : this.service_group) + '/endpoint/' + type
      this.request('GET', sourceApi, '', responseData => {
        this.sourceEndpoints = responseData
      }, {isNeedloading: false})
      const targetApi = this.$root.apiCenter.getEndpointsByType + '/' + (this.isEditState ? this.targetId : this.service_group) + '/endpoint/' + val

      this.request('GET', targetApi, '', responseData => {
        this.targetEndpoints = responseData
      }, {isNeedloading: false})
    },
    addEmptyItem(type) {
      switch (type) {
        case 'path': {
          const hasEmpty = this.addAndEditModal.pathOptions.every(item => item.path !== '')
          if (hasEmpty) {
            this.addAndEditModal.pathOptions.push(
              {path: ''}
            )
          }
          else {
            this.$Message.warning('Path Can Not Empty')
          }
          break
        }
        case 'relate': {
          const hasEmpty = this.addAndEditModal.dataConfig.endpoint_rel.every(item => item.source_endpoint !== '' && item.target_endpoint !== '')
          if (hasEmpty) {
            this.addAndEditModal.dataConfig.endpoint_rel.push(
              {
                source_endpoint: '',
                target_endpoint: ''
              }
            )
          }
          else {
            this.$Message.warning('Can Not Empty')
          }
          break
        }
      }
    },
    deleteItem(type, index) {
      switch (type) {
        case 'path': {
          this.addAndEditModal.pathOptions.splice(index, 1)
          break
        }
        case 'relate': {
          this.addAndEditModal.dataConfig.endpoint_rel.splice(index, 1)
          break
        }
      }
    },
    async addLogFileConfig() {
      this.cancelAddAndEdit()
      this.addAndEditModal.isAdd = true
      this.addAndEditModal.isShow = true
    },
    getDefaultConfig(val, type) {
      const api = `/monitor/api/v2/service/service_group/endpoint_rel?serviceGroup=${this.isEditState ? this.targetId : this.service_group}&sourceType=${type}&targetType=${val}`
      this.request('GET', api, '', responseData => {
        const tmp = responseData.map(r => ({
          source_endpoint: r.source_endpoint,
          target_endpoint: r.target_endpoint
        }))
        if (type === 'host') {
          tmp.forEach(t => {
            const find = this.addAndEditModal.dataConfig.endpoint_rel.find(rel => rel.source_endpoint === t.source_endpoint && rel.target_endpoint === t.target_endpoint)
            if (find === undefined) {
              this.addAndEditModal.dataConfig.endpoint_rel.push(t)
            }
          })
        }
      })
    },
    getDetail(targetId) {
      this.targetId = targetId
      this.getLogKeyWordDetail()
      this.getDataBaseDetail()
    },
    getLogKeyWordDetail() {
      this.request('GET', '/monitor/api/v2/service/log_keyword/list', {
        type: this.keywordType,
        guid: this.targetId
      }, res => {
        this.keywordCollapseData = isEmpty(res) ? [] : res[0].config
        !isEmpty(this.keywordCollapseData) && this.keywordCollapseData.forEach(item => {
          if (isEmpty(item.notify)) {
            item.notify = cloneDeep(initNotify)
          }
        })
      })
    },
    getDataBaseDetail() {
      this.request('GET', '/monitor/api/v2/service/db_keyword/list', {
        type: this.keywordType,
        guid: this.targetId
      }, res => {
        this.dataBaseTableData = isEmpty(res) ? [] : res[0].config
      })
    },
    addDataBase() {
      this.currentEditType = 'database'
      this.resetDrawerForm()
      this.formData.service_group = this.targetId
      this.isAddState = true
      this.isTableChangeFormShow = true
    },
    editDataBaseItem(rowData) {
      this.getSqlSourceOptions(rowData.monitor_type)
      this.currentEditType = 'database'
      this.resetDrawerForm()
      this.fillingFormData(rowData)
      this.isAddState = false
      this.isTableChangeFormShow = true
    },
    deleteDataBaseItem(rowData) {
      const api = '/monitor/api/v2/service/db_keyword/db_keyword_config'
      this.request('DELETE', api, {
        guid: rowData.guid
      }, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getDetail(this.targetId)
      })
    },
    onTableChangeFormClose() {
      this.isTableChangeFormShow = false
      this.resetDrawerForm()
    },
    addMetricConfig() {
      this.formData.endpoint_rel.push({
        target_endpoint: '',
        source_endpoint: ''
      })
    },
    getSqlSourceOptions(monitorType) {
      if (monitorType) {
        const publicPath = '/monitor/api/v2/service/service_group/'
        const sourceApi = publicPath + this.targetId + '/endpoint/mysql'
        this.request('GET', sourceApi, '', responseData => {
          this.sqlSourceEndpoints = responseData
        })
        const targetApi = publicPath + this.targetId + '/endpoint/' + monitorType
        this.request('GET', targetApi, '', responseData => {
          this.sqlTargetEndpoints = responseData
        })
      }
    },
    deleteDatabaseMapItem(index) {
      this.formData.endpoint_rel.splice(index, 1)
    },
    resetDrawerForm() {
      this.formData = cloneDeep(initFormData)
      this.$refs.formData.resetFields()
    },
    onDrawerClose() {
      this.resetDrawerForm()
      this.isTableChangeFormShow = false
    },
    async onFormSave() {
      const validResult = await this.$refs.formData.validate()
      if (validResult) {
        const params = cloneDeep(this.formData)
        params.active_window = params.active_window.length ? params.active_window.join('-') : ''
        let method = 'POST'
        if (this.isLogFile) {
          if (!this.isAddState) { // 日志编辑
            method = 'PUT'
          }
          this.request(method, '/monitor/api/v2/service/log_keyword/log_keyword_config', params, () => {
            this.$Message.success(this.$t('m_tips_success'))
            this.isTableChangeFormShow = false
            this.getLogKeyWordDetail()
          })
        }
        else {
          if (!this.isAddState) { // 数据库编辑
            method = 'PUT'
          }
          this.request(method, '/monitor/api/v2/service/db_keyword/db_keyword_config', params, () => {
            this.$Message.success(this.$t('m_tips_success'))
            this.isTableChangeFormShow = false
            this.getDataBaseDetail()
          })
        }
      }
    },
    procCallbackKeyChangeForm(proc_callback_key, item) {
      const findFlow = this.allFlows.find(f => f.procDefKey === proc_callback_key)
      if (findFlow) {
        item.proc_callback_name = `${findFlow.procDefName}[${findFlow.procDefVersion}]`
      }
      else {
        item.proc_callback_name = ''
      }
    }
  },
  components: {
    // extendTable
  },
}
</script>

<style lang='less'>
.ivu-table-wrapper {
  overflow: inherit;
}

.ivu-form-item {
  margin-bottom: 4px;
}

.file-log-form {
  .ivu-form-item {
    margin-bottom: 12px;
  }
}
.success-btn {
  color: #fff;
  background-color: #19be6b;
  border-color: #19be6b;
}

.ivu-collapse-header {
  display: flex;
  align-items: center;
}

.database-map-item {
  .ivu-tooltip {
    width: 48%;
  }
  .ivu-tooltip-rel {
    width: 100%
  }
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
      margin-top: -20px;
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

.section-content {
  position: relative;
  margin-top: 16px;
  .upload-content {
    display: flex;
    position: absolute;
    top: -45px;
    right: 20px;
    .btn-img {
      width: 16px;
      vertical-align: middle;
    }
  }

  .keyword-collapse-content {
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
    .use-underline-title {
      font-size: 14px;
    }
  }
}

.content-title {
  display: flex;
  width: 100%;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.config-drawer {
  position: relative;
  .file-log-form {
    width: 80%;
    .database-map-content {
      margin: 4px 0px;
      padding:8px 12px;
      border:1px solid #dcdee2;
      border-radius:4px;
      text-align: center;
      .database-map-item {
        display: flex;
        align-items: center;
      }
    }
  }
  .arrange-item {
    width: 80%;
    display: flex;
    justify-content: flex-end;
    margin-top: 20px;
    margin-bottom: 100px;

  }
  .form-footer {
    display: flex;
    justify-content: center;
    position: fixed;
    bottom: 20px;
    width: 65%;
    background-color: #fff;
    opacity: 1;
  }
}

</style>
