<template>
  <div>
    <section class="section-content">
      <div v-for="(single, i) in logAndDataBaseAllDetail" :key="i">
        <div v-if="!isEditState" class="content-header mb-2 mt-3">
          <div class="use-underline-title header-title mr-4">
            {{ !isEmpty(single) ? single.display_name : ''}}
            <span class="underline"></span>
          </div>
          <Tag color="blue">{{ $t('m_field_resourceLevel') }}</Tag>
        </div>

        <div class="content-title mb-3">
          <div class="use-underline-title">
            {{$t('m_log_file')}}
            <span class="underline" style="margin-top: -10px"></span>
          </div>
          <Button v-if="isEditState" type="success" class="mr-4" @click="addLogFileConfig">
            {{ $t('m_button_add') }}
          </Button>
        </div>

        <Collapse v-model="keywordCollapseValue" v-if="!isEmpty(single) && !isEmpty(single.config)">
          <Panel v-for="(item, index) in single.config"
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
                  <Button size="small" class="mr-1"  type="info" @click.stop="saveFileDetail(item)">
                    <Icon type="ios-cloud-done" size="16" />
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
                      <span style="margin-right: 8px;line-height: 32px;">{{$t('m_firing')}}</span>
                      </Col>
                      <Col span="6" style="">
                      <Select v-model="item.notify.notify_roles" :disabled="!isEditState" :max-tag-count="2" style="width: 99%;" multiple filterable :placeholder="$t('m_field_role')">
                        <Option v-for="role in allRoles" :value="role.name" :key="role.value">{{ role.display_name }}</Option>
                      </Select>
                      </Col>
                      <Col span="5">
                      <Select v-model="item.notify.proc_callback_key"
                              clearable
                              :disabled="!isEditState"
                              @on-change="procCallbackKeyChangeForm(item.notify.proc_callback_key, item.notify)"
                              style="width:99%;"
                              :placeholder="$t('m_proc_callback_key')"
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
                        v-model.trim="item.notify.description"
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
        <div v-else class='no-data-class'>{{$t('m_table_noDataTip')}}</div>

        <div class="mt-2 pb-5">
          <div>
            <div class="content-title mb-2 ">
              <div class="use-underline-title">
                {{$t('m_db')}}
                <span class="underline" style="margin-top: -10px"></span>
              </div>
              <Button v-if="isEditState" type="success" class="mr-4" @click="addDataBase" style="margin: 8px 0">
                {{ $t('m_button_add') }}
              </Button>
            </div>

            <Table
              v-if='!isEmpty(single) && !isEmpty(single.dbConfig)'
              size="small"
              :columns="dataBaseTableColumns"
              :data="single.dbConfig"
            />
            <div v-else class='no-data-class'>{{$t('m_table_noDataTip')}}</div>
          </div>
        </div>

      </div>
    </section>
    <Modal
      v-model="addAndEditModal.isShow"
      :title="addAndEditModal.isAdd ? $t('m_button_add') : $t('m_button_edit')"
      :mask-closable="false"
      :width="900"
    >
      <div :style="{'max-height': MODALHEIGHT + 'px', overflow: 'auto'}">
        <div>
          <span>{{$t('m_field_type')}}:</span>
          <Select
            v-model="addAndEditModal.dataConfig.monitor_type"
            :disabled="!isEditState"
            @on-change="getEndpoint(addAndEditModal.dataConfig.monitor_type, 'host')"
            style="width: 800px"
          >
            <Option v-for="type in monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
          </Select>
        </div>
        <div v-if="addAndEditModal.isAdd" style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;width:830px">
          <template v-for="(item, index) in addAndEditModal.pathOptions">
            <p :key="index + 5">
              <Tooltip :content="$t('m_tableKey_logPath')" :delay="1000">
                <Input
                  v-model.trim="item.path"
                  :disabled="!isEditState"
                  style="width: 772px"
                  :placeholder="$t('m_tableKey_logPath')"
                />
              </Tooltip>
              <Button
                v-if="addAndEditModal.isAdd"
                :disabled="!isEditState"
                @click="deleteItem('path', index)"
                size="small"
                type="error"
                icon="md-trash"
              ></Button>
            </p>
          </template>
          <Button
            @click="addEmptyItem('path')"
            :disabled="!isEditState"
            type="success"
            size="small"
            style="width:800px"
            long
          >{{ $t('m_button_add') }}{{$t('m_tableKey_logPath')}}</Button>
        </div>
        <div v-else style="margin: 8px 0">
          <span>{{$t('m_tableKey_path')}}:</span>
          <Input :disabled="!isEditState" style="width: 800px" v-model.trim="addAndEditModal.dataConfig.log_path" />
        </div>
        <span v-if="!this.addAndEditModal.isAdd && isSystemConfigurationTipsShow" style="color: red">{{$t('m_recommended_system_configuration_tips')}}</span>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;width:830px">
          <template v-for="(item, index) in addAndEditModal.dataConfig.endpoint_rel">
            <p :key="index + 'c'">
              <Tooltip :content="$t('m_business_object')" :delay="1000">
                <Select
                  v-model="item.target_endpoint"
                  :disabled="!isEditState"
                  style="width: 386px"
                  :placeholder="$t('m_business_object')"
                >
                  <Option v-for="type in targetEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('m_log_server')" :delay="1000">
                <Select
                  v-model="item.source_endpoint"
                  :disabled="!isEditState"
                  style="width: 386px"
                  :placeholder="$t('m_log_server')"
                >
                  <Option v-for="type in sourceEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select>
              </Tooltip>
              <Button
                :disabled="!isEditState"
                @click="deleteItem('relate', index)"
                size="small"
                type="error"
                icon="md-trash"
              ></Button>
            </p>
          </template>
          <Button
            @click="addEmptyItem('relate')"
            type="success"
            :disabled="!isEditState"
            size="small"
            style="width:800px"
            long
          >{{$t('m_addStringMap')}}</Button>
        </div>
      </div>
      <div slot="footer">
        <Button @click="cancelAddAndEdit" :disabled="!isEditState">{{$t('m_button_cancel')}}</Button>
        <Button @click="okAddAndEdit" :disabled="!isEditState" type="primary">{{$t('m_button_save') }}</Button>
      </div>
    </Modal>
    <BaseDrawer
      :title="isAddState ? $t('m_add') : $t('m_button_edit')"
      :visible.sync="isTableChangeFormShow"
      :realWidth="1000"
      :scrollable="true"
      :mask-closable="false"
      class="config-drawer"
    >
      <template slot="content">
        <div class="file-log-form">
          <Form ref="formData" :model="formData" :rules="ruleValidate" :label-width="130">
            <FormItem :label="$t('m_alarmName')" prop="name">
              <Input v-model.trim="formData.name" :disabled="!isEditState" :maxlength="50" show-word-limit />
            </FormItem>
            <FormItem v-if="isLogFile" :label="$t('m_field_log')" prop="keyword">
              <Input v-model.trim="formData.keyword" :disabled="!isEditState"></Input>
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
                v-model.trim="formData.query_sql"
                :maxlength="200"
              >
              </Input>
            </FormItem>
            <FormItem v-if="!isLogFile" :label="$t('m_type')" prop="monitor_type">
              <Select
                v-model="formData.monitor_type"
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
                >{{ $t('m_addMetricConfig') }}</Button>
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
            <FormItem :label="$t('m_notification')" prop="notify_enable">
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
            <FormItem v-if="!isLogFile" :label="$t('m_collection_interval')" prop="step">
              <Select
                filterable
                :disabled="!isEditState"
                v-model="formData.step"
                transfer
              >
                <Option
                  v-for="item in stepOptions"
                  :value="item.value"
                  :key="item.value"
                  :label="item.name"
                >
                  {{ item.name }}
                </Option>
              </Select>
            </FormItem>
            <FormItem :label="$t('m_active_window')" prop="active_window_list">
              <activeWindowTime
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
          <span class="mr-1 mt-1" style="font-size: 12px; min-width: 60px">{{$t('m_firing')}}</span>
          <Tooltip :content="$t('m_resourceLevel_role')" :delay="1000">
            <Select v-model="formData.notify.notify_roles" transfer :disabled="!isEditState" clearable :max-tag-count="2" style="width: 300px" multiple filterable :placeholder="$t('m_field_role')">
              <Option v-for="item in allRoles" :value="item.name" :key="item.value">{{ item.display_name }}</Option>
            </Select>
          </Tooltip>
          <Tooltip :content="$t('m_proc_callback_key')" :delay="1000">
            <Select v-model="formData.notify.proc_callback_key" transfer :disabled="!isEditState" clearable @on-change="procCallbackKeyChangeForm(formData.notify.proc_callback_key, formData.notify)" style="width: 160px" :placeholder="$t('m_proc_callback_key')">
              <Option v-for="(flow, flowIndex) in allFlows" :key="flowIndex" :value="flow.procDefKey" :label="flow.procDefName + ' [' + flow.procDefVersion + ']'"><span>{{ flow.procDefName }} [{{ flow.procDefVersion }}]</span></Option>
            </Select>
          </Tooltip>
          <Tooltip :content="$t('m_callback_mode')" :delay="1000">
            <Select v-model="formData.notify.proc_callback_mode" transfer :disabled="!isEditState" clearable style="width: 180px" :placeholder="$t('m_callback_mode')">
              <Option v-for="item in callbackMode" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
            </Select>
          </Tooltip>
          <Tooltip :content="$t('m_tableKey_description')" :delay="1000">
            <Input
              v-model.trim="formData.notify.description"
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
      </template>
      <template slot="footer">
        <Button @click="onDrawerClose" :disabled="!isEditState">{{$t('m_button_cancel')}}</Button>
        <Button @click="onFormSave" :disabled="!isEditState" type="primary">{{$t('m_button_save')}}</Button>
      </template>
    </BaseDrawer>
  </div>
</template>

<script>
import {
  cloneDeep, isEmpty, hasIn, find
} from 'lodash'
import Vue from 'vue'
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import {priorityList} from '@/assets/config/common-config.js'
import activeWindowTime from '../../components/active-window-time.vue'
import {showPoptipOnTable} from '@/assets/js/utils.js'

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
  regulative: 1,
  query_sql: '',
  monitor_type: '',
  endpoint_rel: [],
  priority: 'low',
  notify_enable: 1, // 通知,默认启用
  step: 10,
  active_window_list: ['00:00-23:59'], // 告警时间段
  content: '', // 通知内容
  notify: {
    alarm_action: 'firing',
    proc_callback_key: '',
    notify_roles: [],
    proc_callback_mode: '',
    description: ''
  }
}

export const custom_api_enum = [
  {
    logKeywordMonitorById: 'delete'
  },
  {
    getEndpointsByTypeByType: 'get'
  },
  {
    serviceGroupEendpoint: 'get'
  },
  {
    logKeywordConfig: 'put'
  },
  {
    logKeywordConfig: 'post'
  },
  {
    serviceDbKeywordConfig: 'post',
  },
  {
    serviceDbKeywordConfig: 'put',
  }
]

export default {
  name: '',
  props: {
    keywordType: {
      type: String,
      default: 'service' // 为枚举，service代表层级对象，endpoint代表对象
    }
  },
  components: {
    activeWindowTime
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
          monitor_type: 'process',
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
      allLogFileData: [],
      allDataBaseDataInfo: [],
      keywordCollapseValue: ['0'],
      keywordTableColumns: [ // 日志文件table
        {
          title: this.$t('m_alarmName'),
          width: 150,
          key: 'name',
          render: (h, params) => (<Tooltip class='table-key-word-name' placement="right" max-width="400" content={params.row.name}>
            {params.row.name || '-'}
          </Tooltip>)
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
            return (<Tooltip class='table-active-time' placement="right" max-width="400" content={text}>
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
          title: this.$t('m_field_log'),
          key: 'keyword',
          minWidth: 150,
          render: (h, params) => (<Tooltip class='table-key-word-name' placement="right" max-width="400" content={params.row.keyword || '-'}>
            {params.row.keyword || '-'}
          </Tooltip>)
        },
        {
          title: this.$t('m_regular'), // 更新人
          key: 'regulative',
          width: 100,
          render: (h, params) => {
            const regulative_enable = params.row.regulative === 0
            return (
              <Tag color={regulative_enable ? 'default' : 'green'}>{this.$t(statusButtonLabelMap[params.row.regulative])}</Tag>
            )
          }
        },
        {
          title: this.$t('m_updatedBy'),
          key: 'update_user',
          minWidth: 100,
          render: (h, params) => <span>{params.row.update_user || '-'}</span>
        },
        {
          title: this.$t('m_update_time'),
          key: 'update_time',
          width: 150,
          render: (h, params) => <span>{params.row.update_time || '-'}</span>
        },
        {
          title: this.$t('m_table_action'),
          key: 'index',
          fixed: 'right',
          width: 150,
          render: (h, params) => this.isEditState ? (
            <div style='display: flex'>
              <Tooltip max-width={400} placement="top" transfer content={this.$t('m_copy')}>
                <Button size="small" class="mr-1" type="success" on-click={() => this.copyCustomMetricItem(params.row)}>
                  <Icon type="md-document" size="16"></Icon>
                </Button>
              </Tooltip>
              <Tooltip max-width={400} placement="top" transfer content={this.$t('m_button_edit')}>
                <Button size="small" class="mr-1" type='primary' on-click={() => this.editCustomMetricItem(params.row)}>
                  <Icon type="md-create" size="16" />
                </Button>
              </Tooltip>
              <Poptip
                confirm
                transfer
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
          key: 'name',
          render: (h, params) => (<Tooltip class='table-key-word-name' placement="right" max-width="400" content={params.row.name}>
            {params.row.name}
          </Tooltip>)
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
          width: 200,
          key: 'active_window_list',
          render: (h, params) => {
            const text = !isEmpty(params.row.active_window_list) ? params.row.active_window_list.join(';') : '-'
            return (<Tooltip class='table-active-time' placement="right" max-width="400" content={text}>
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
          minWidth: 250,
          render: (h, params) => (<Tooltip class='table-sql-script' placement="top" max-width="400" content={params.row.query_sql || '-'}>
            {params.row.query_sql || '-'}
          </Tooltip>)
        },
        {
          title: this.$t('m_updatedBy'),
          key: 'update_user',
          minWidth: 100,
          render: (h, params) => <span>{params.row.update_user || '-'}</span>
        },
        {
          title: this.$t('m_update_time'),
          key: 'update_time',
          width: 150,
          render: (h, params) => <span>{params.row.update_time || '-'}</span>
        },
        {
          title: this.$t('m_table_action'),
          key: 'index',
          fixed: 'right',
          width: 150,
          render: (h, params) => this.isEditState ? (
            <div style='display: flex'>
              <Tooltip placement="top" transfer content={this.$t('m_copy')}>
                <Button size="small" class="mr-1" type="success" on-click={() => this.copyDataBaseItem(params.row)}>
                  <Icon type="md-document" size="16"></Icon>
                </Button>
              </Tooltip>
              <Tooltip placement="top" transfer content={this.$t('m_button_edit')}>
                <Button size="small" class="mr-1" type="primary" on-click={() => this.editDataBaseItem(params.row)}>
                  <Icon type="md-create" size="16"></Icon>
                </Button>
              </Tooltip>
              <Poptip
                confirm
                transfer
                title={this.$t('m_delConfirm_tip')}
                placement="left-end"
                on-on-ok={() => this.deleteDataBaseItem(params.row)}>
                <Button size="small" type="error" on-click={() => {
                  showPoptipOnTable()
                }}>
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
      stepOptions: [
        {
          name: '10s',
          value: 10
        },
        {
          name: '30s',
          value: 30
        },
        {
          name: '1min',
          value: 60
        },
        {
          name: '5min',
          value: 300
        },
        {
          name: '10min',
          value: 600
        },
        {
          name: '1h',
          value: 3600
        },
        {
          name: '2h',
          value: 7200
        },
        {
          name: '12h',
          value: 43200
        },
        {
          name: '24h',
          value: 86400
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
        active_window_list: [
          {
            type: 'array',
            required: true,
            message: '请输入',
            trigger: 'blur'
          }
        ],
        endpoint_rel: [
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
      service_group: '',
      isEmpty,
      dataBaseGuid: '',
      logAndDataBaseAllDetail: [],
      alarmName: '',
      actionType: '',
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
      isSystemConfigurationTipsShow: false
    }
  },
  computed: {
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
  },
  methods: {
    getFlowsAndRolesOptions() {
      if (isEmpty(this.allFlows) && isEmpty(this.allRoles)) {
        this.request('GET', this.apiCenter.eventCallbackList, '', responseData => {
          this.allFlows = responseData
        })
        this.request('GET', this.apiCenter.userRoleList, '', responseData => {
          this.allRoles = responseData.data.map(_ => ({
            ..._,
            value: _.id
          }))
        })
      }
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
    // other config
    editF(rowData) {
      this.service_group = rowData.service_group
      this.cancelAddAndEdit()
      this.addAndEditModal.isAdd = false
      this.addAndEditModal.addRow = rowData
      this.addAndEditModal.dataConfig.guid = rowData.guid
      this.addAndEditModal.dataConfig.service_group = rowData.service_group
      this.addAndEditModal.dataConfig.monitor_type = rowData.monitor_type
      this.addAndEditModal.dataConfig.log_path = rowData.log_path
      this.addAndEditModal.dataConfig.endpoint_rel = cloneDeep(rowData.endpoint_rel)
      this.getEndpoint(rowData.monitor_type, 'host')
      this.addAndEditModal.isShow = true
    },
    saveFileDetail(item) {
      const params = {
        log_keyword_monitor: item.guid,
        notify: item.notify
      }
      // const api = '/monitor/api/v2/service/log_keyword/notify'
      this.request('POST', this.apiCenter.logKeywordNotify, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
      })
    },
    addCustomMetric(item) {
      this.getFlowsAndRolesOptions()
      this.currentEditType = 'logFile'
      this.resetDrawerForm()
      this.formData.name = this.$t('m_alert') + new Date().getTime()
      this.formData.log_keyword_monitor = item.guid
      this.isAddState = true
      this.isTableChangeFormShow = true
    },
    editCustomMetricItem(rowData) {
      this.getFlowsAndRolesOptions()
      this.currentEditType = 'logFile'
      this.resetDrawerForm()
      this.fillingFormData(rowData)
      this.isAddState = false
      this.isTableChangeFormShow = true
    },
    copyCustomMetricItem(rowData) {
      this.getFlowsAndRolesOptions()
      this.currentEditType = 'logFile'
      this.resetDrawerForm()
      this.fillingFormData(rowData)
      this.formData.name += '1'
      delete this.formData.guid
      const needUseNotifyKeys = ['alarm_action', 'proc_callback_key', 'notify_roles', 'proc_callback_mode', 'description']
      if (!isEmpty(this.formData.notify)) {
        for (const key in this.formData.notify) {
          if (!needUseNotifyKeys.includes(key)) {
            delete this.formData.notify[key]
          }
        }
      }
      this.formData.guid = ''
      this.formData.log_keyword_monitor = rowData.log_keyword_monitor
      this.formData.keyword += '1'
      this.isAddState = true
      this.isTableChangeFormShow = true
    },

    fillingFormData(rowData) {
      for (const key in this.formData) {
        if (hasIn(rowData, key)) {
          Vue.set(this.formData, key, cloneDeep(rowData[key]))
        }
      }
      this.formData.notify = isEmpty(this.formData.notify) ? cloneDeep(initNotify) : this.formData.notify
    },
    delCustomMericsItem(rowData) {
      // const api = '/monitor/api/v2/service/log_keyword/log_keyword_config'
      this.request('DELETE', this.apiCenter.logKeywordConfig, { guid: rowData.guid}, () => {
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
    async okAddAndEdit() {
      if (!this.addAndEditModal.dataConfig.monitor_type) {
        return this.$Message.error('类型不能为空')
      }
      if (this.addAndEditModal.isAdd) {
        const pathFlag = this.addAndEditModal.pathOptions.every(item => item.path !== '')
        if (!pathFlag || this.addAndEditModal.pathOptions.length === 0) {
          return this.$Message.error('日志路径不能为空')
        }
      } else {
        if (!this.addAndEditModal.dataConfig.log_path) {
          return this.$Message.error('日志路径不能为空')
        }
      }
      const endpointRelFlag = this.addAndEditModal.dataConfig.endpoint_rel.every(item => item.source_endpoint !== '' && item.target_endpoint !== '')
      if (!endpointRelFlag || this.addAndEditModal.dataConfig.endpoint_rel.length === 0) {
        return this.$Message.error('映射不能为空')
      }
      const params = JSON.parse(JSON.stringify(this.addAndEditModal.dataConfig))
      // const methodType = this.addAndEditModal.isAdd ? 'POST' : 'PUT'
      params.service_group = this.targetId
      if (this.addAndEditModal.isAdd) {
        params.log_path = this.addAndEditModal.pathOptions.map(p => p.path)
      }
      if (this.addAndEditModal.isAdd) {
        await this.request('POST', this.apiCenter.serviceLogKeywordMonitor, params)
      } else {
        await this.request('PUT', this.apiCenter.serviceLogKeywordMonitor, params)
      }
      this.$Message.success(this.$t('m_tips_success'))
      this.addAndEditModal.isShow = false
      this.getDetail(this.targetId)
    },
    cancelAddAndEdit() {
      this.addAndEditModal.isShow = false
      this.addAndEditModal.pathOptions = []
      this.addAndEditModal.dataConfig = {
        service_group: '',
        log_path: [],
        monitor_type: 'process',
        endpoint_rel: []
      }
    },
    async getEndpoint(val, type) {
      if (isEmpty(this.addAndEditModal.dataConfig.endpoint_rel)) {
        this.addAndEditModal.dataConfig.endpoint_rel = []
        this.isSystemConfigurationTipsShow = true
        await this.getDefaultConfig(val, type)
      } else {
        this.isSystemConfigurationTipsShow = false
      }
      const sourceApi = this.apiCenter.getEndpointsByType + '/' + (this.isEditState ? this.targetId : this.service_group) + '/endpoint/' + type
      this.request('GET', sourceApi, '', responseData => {
        this.sourceEndpoints = responseData
      }, {isNeedloading: false})
      const targetApi = this.apiCenter.getEndpointsByType + '/' + (this.isEditState ? this.targetId : this.service_group) + '/endpoint/' + val

      this.request('GET', targetApi, '', responseData => {
        this.targetEndpoints = responseData
      }, {isNeedloading: false})
    },
    addEmptyItem(type) {
      switch (type) {
        case 'path': {
          this.addAndEditModal.pathOptions.push(
            {path: ''}
          )
          break
        }
        case 'relate': {
          this.addAndEditModal.dataConfig.endpoint_rel.push(
            {
              source_endpoint: '',
              target_endpoint: ''
            }
          )
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
      this.getEndpoint(this.addAndEditModal.dataConfig.monitor_type, 'host')
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
    async getDetail(targetId, alarmName = this.alarmName) {
      this.getFlowsAndRolesOptions()
      if (targetId) {
        if (this.alarmName !== alarmName) {
          this.keywordCollapseValue = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '10']
          this.alarmName = alarmName
        } else {
          this.keywordCollapseValue = ['0']
        }
        this.targetId = targetId
        await this.getLogKeyWordDetail()
        this.processAllInfo()
      } else {
        this.logAndDataBaseAllDetail = []
      }
    },
    processAllInfo() {
      this.logAndDataBaseAllDetail = []
      this.logAndDataBaseAllDetail = cloneDeep(this.allLogFileData)
      this.$emit('feedbackInfo', this.logAndDataBaseAllDetail)
    },
    getLogKeyWordDetail() {
      return new Promise(resolve => {
        this.request('GET', this.apiCenter.getLogKeywordList, {
          type: this.keywordType,
          guid: this.targetId,
          alarmName: this.alarmName
        }, res => {
          this.allLogFileData = isEmpty(res) ? [] : res
          this.allLogFileData.forEach(logFile => {
            !isEmpty(logFile.config) && logFile.config.forEach(item => {
              if (isEmpty(item.notify)) {
                item.notify = cloneDeep(initNotify)
              }
            })
          })
          resolve(this.allLogFileData)
        })
      })
    },
    addDataBase() {
      this.getFlowsAndRolesOptions()
      this.currentEditType = 'database'
      this.isAddState = true
      this.actionType = ''
      this.resetDrawerForm()
      this.formData.service_group = this.targetId
      this.formData.monitor_type = 'process'
      this.formData.name = this.$t('m_alert') + new Date().getTime()
      this.dataBaseGuid = 'process'
      this.getSqlSourceOptions(this.formData.monitor_type)
      this.isTableChangeFormShow = true
    },
    editDataBaseItem(rowData) {
      this.actionType = 'edit'
      this.getFlowsAndRolesOptions()
      this.isAddState = false
      this.currentEditType = 'database'
      this.resetDrawerForm()
      this.fillingFormData(rowData)
      this.dataBaseGuid = rowData.dataBaseGuid
      this.getSqlSourceOptions(rowData.monitor_type)
      this.isTableChangeFormShow = true
    },
    copyDataBaseItem(rowData) {
      this.actionType = 'copy'
      this.getFlowsAndRolesOptions()
      this.isAddState = true
      this.currentEditType = 'database'
      this.resetDrawerForm()
      this.fillingFormData(rowData)
      this.dataBaseGuid = rowData.dataBaseGuid
      this.getSqlSourceOptions(rowData.monitor_type)
      this.formData.name += '1'
      this.formData.guid = ''
      this.isTableChangeFormShow = true
    },
    deleteDataBaseItem(rowData) {
      // const api = '/monitor/api/v2/service/db_keyword/db_keyword_config'
      this.request('DELETE', this.apiCenter.serviceDbKeywordConfig, {
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
      this.formData.endpoint_rel = this.formData.endpoint_rel || []
      this.formData.endpoint_rel.push({
        target_endpoint: '',
        source_endpoint: ''
      })
    },
    getSqlSourceOptions(monitorType) {
      if (monitorType) {
        const publicPath = '/monitor/api/v2/service/service_group/'
        const sourceApi = publicPath + (this.keywordType === 'service' ? this.targetId : this.dataBaseGuid) + '/endpoint/mysql'
        this.request('GET', sourceApi, '', responseData => {
          this.sqlSourceEndpoints = responseData
        })
        const targetApi = publicPath + (this.keywordType === 'service' ? this.targetId : this.dataBaseGuid) + '/endpoint/' + monitorType
        this.request('GET', targetApi, '', responseData => {
          this.sqlTargetEndpoints = responseData
        })
        if (this.isAddState && this.actionType !== 'copy') {
          this.formData.endpoint_rel = this.formData.endpoint_rel || []
          const api = `/monitor/api/v2/service/service_group/endpoint_rel?serviceGroup=${this.targetId}&sourceType=mysql&targetType=${monitorType}`
          this.request('GET', api, '', responseData => {
            if (!isEmpty(responseData)) {
              responseData.forEach(item => {
                this.formData.endpoint_rel.push({
                  target_endpoint: item.target_endpoint,
                  source_endpoint: item.source_endpoint
                })
              })
            }
          })
        }
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
        let method = 'POST'
        if (this.isLogFile) {
          if (!this.isAddState) { // 日志编辑
            method = 'PUT'
          }
          this.request(method, this.apiCenter.logKeywordConfig, params, async () => {
            this.$Message.success(this.$t('m_tips_success'))
            this.isTableChangeFormShow = false
            this.getDetail(this.targetId)
          })
        } else {
          if (!this.isAddState) { // 数据库编辑
            method = 'PUT'
          }
          if (isEmpty(params.endpoint_rel) || isEmpty(params.endpoint_rel[0].target_endpoint) || isEmpty(params.endpoint_rel[0].source_endpoint)) {
            return this.$Message.error(this.$t('m_database_map') + this.$t('m_cannot_be_empty'))
          }
          this.request(method, this.apiCenter.serviceDbKeywordConfig, params, async () => {
            this.$Message.success(this.$t('m_tips_success'))
            this.isTableChangeFormShow = false
            this.getDetail(this.targetId)
          })
        }
      }
    },
    procCallbackKeyChangeForm(proc_callback_key, item) {
      const findFlow = this.allFlows.find(f => f.procDefKey === proc_callback_key)
      if (findFlow) {
        item.proc_callback_name = `${findFlow.procDefName}[${findFlow.procDefVersion}]`
      } else {
        item.proc_callback_name = ''
      }
    },
    onActiveTimeChange(time) {
      this.formData.active_window_list = time
    }
  }
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
    margin-bottom: 20px;
  }
}
.success-btn {
  color: #fff;
  background-color:  #00CB91;
  border-color:  #00CB91;
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

.table-active-time {
  .ivu-tooltip-rel {
    width: 120px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.table-key-word-name {
  .ivu-tooltip-rel {
    width: 140px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.table-sql-script {
  .ivu-tooltip-rel {
    width: 240px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
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

.use-underline-title.header-title {
  .underline {
    margin-top: -14px;
  }
}

.section-content {
  position: relative;
  margin-top: 16px;

  .keyword-collapse-content {
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
    .use-underline-title {
      font-size: 14px;
    }
  }

  .no-data-class {
    display: flex;
    justify-content: center;
  }

}

.content-title {
  display: flex;
  width: 100%;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.content-header {
  display: flex;
  justify-content: flex-start;
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
    width: 90%;
    display: flex;
    justify-content: flex-start;
    margin-left: 66px;
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

.database-item {
  margin-top: 16px;
  padding-bottom: 30px
}

</style>
