<template>
  <div>
    <div class='log-template-header'>
      <div class='log-template-header-left'>
        <Input
          v-model.trim="searchParams.name"
          style='width: 450px; margin-right: 20px'
          :placeholder="$t('m_template_name')"
          class="search-item"
          clearable
          @on-change="onFilterChange"
        ></Input>
        <Select
          v-model="searchParams.update_user"
          filterable
          clearable
          :placeholder="$t('m_updatedBy')"
          @on-change="onFilterChange"
        >
          <Option v-for="(user, index) in updateUserOptions" :value="user" :key="index">{{
            user
          }}</Option>
        </Select>
        <!-- <span style="margin-top: 8px;margin-left: 24px;">
          <Button @click="getTemplateList" type="primary" style="background-color: #5384FF;">{{ $t('m_button_search') }}</Button>
          <Button @click="handleReset" style="margin-left: 5px">{{ $t('m_reset_condition') }}</Button>
        </span> -->
      </div>
      <div>
        <ExportImport
          :isShowExportBtn="true"
          exportUrl="/monitor/api/v2/service/log_metric/log_monitor_template/export"
          exportMethod="POST"
          :exportData="exportData"
          :validateExportDataEmpty="true"
          :isShowImportBtn="true"
          uploadUrl="/monitor/api/v2/service/log_metric/log_monitor_template/import"
          @successCallBack="getTemplateList"
        ></ExportImport>
      </div>
    </div>
    <div class="table-zone">
      <Spin v-if="spinShow" size="large">
        <Icon type="ios-loading" size="36"></Icon>
      </Spin>
      <template v-else>
        <template v-if="data.length > 0">
          <div v-for="(item, itemIndex) in data" :key="itemIndex" style="margin-bottom: 16px;">
            <Card>
              <div class="w-header" slot="title">
                <div class="title">
                  {{ item.name }}
                  <span class="underline"></span>
                </div>
                <Icon
                  v-if="!hideRegex.includes(itemIndex)"
                  size="26"
                  @click="changeRegexTableStatus(itemIndex, 'in')"
                  type="md-arrow-dropdown"
                  style="cursor: pointer"
                />
                <Icon
                  v-else
                  size="26"
                  @click="changeRegexTableStatus(itemIndex, 'out')"
                  type="md-arrow-dropright"
                  style="cursor: pointer"
                />
              </div>
              <Button slot="extra" type="success" @click.prevent="addTemplate(item.log_type)">{{ $t('m_button_add') }}</Button>
              <div v-show="!hideRegex.includes(itemIndex)">
                <Table
                  size="small"
                  :columns="tableColumn"
                  :data="item.tableData"
                  @on-select-all="selection => onSelectAll(selection, itemIndex)"
                  @on-select-all-cancel="selection => onSelectAllCancel(selection, itemIndex)"
                  @on-select="(selection, row) => onSelect(selection, row, itemIndex)"
                  @on-select-cancel="(selection, row) => cancelSelect(selection, row, itemIndex)"
                  width="100%"
                ></Table>
              </div>
            </Card>
          </div>
        </template>
        <template v-else>
          <div style="text-align: center; margin-top: 16px">
            {{ $t('m_noData') }}
          </div>
        </template>
      </template>
    </div>
    <JsonRegex ref="jsonRegexRef" @refreshData="getTemplateList"></JsonRegex>
    <StandardRegex ref="standardRegexRef" @refreshData="getTemplateList"></StandardRegex>
    <CustomRegex ref="customRegexRef" @reloadMetricData="getTemplateList"></CustomRegex>
    <!-- 查看管理层级对象 -->
    <Modal
      v-model="showServiceGroup"
      :fullscreen="isfullscreen"
      footer-hide
      :title="$t('m_field_resourceLevel')"
    >
      <div slot="header" class="custom-modal-header">
        <span>
          {{$t('m_field_resourceLevel')}}
        </span>
        <Icon v-if="isfullscreen" @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-expand" />
      </div>
      <Input v-model="filterServiceGroup" :placeholder="$t('m_field_resourceLevel') + $t('m_tableKey_name')" style="margin-bottom: 12px;"></Input>
      <div  :class="isfullscreen ? 'modal-container-fullscreen' : 'modal-container-normal'">
        <template v-if="serviceGroup.length > 0">
          <Tag size="large" v-for="(item, index) in serviceGroup.filter(data => data.display_name.includes(filterServiceGroup))" :key="index">{{ item.display_name }}</Tag>
        </template>
        <template v-else>
          <Alert type="warning">{{ $t('m_noData') }}</Alert>
        </template>
      </div>
    </Modal>
  </div>
</template>

<script>
import {debounce, isEmpty} from 'lodash'
import ExportImport from '@/components/export-import.vue'
import JsonRegex from './log-template-config/json-regex.vue'
import StandardRegex from './log-template-config/standard-regex.vue'
import {showPoptipOnTable} from '@/assets/js/utils.js'
import CustomRegex from '@/views/monitor-config/log-template-config/custom-regex.vue'

export const custom_api_enum = [
  {
    getAffectServiceGroup: 'get'
  },
  {
    getConfigDetailById: 'delete'
  },
  {
    logMonitorTemplateExport: 'post'
  },
  {
    logMonitorTemplateImport: 'post'
  }
]

export default {
  name: 'log-template',
  data() {
    return {
      spinShow: false,
      isfullscreen: false,
      searchParams: {
        name: '',
        update_user: ''
      },
      hideRegex: [],
      data: [
        {
          name: this.$t('m_standard_json'),
          log_type: 'json',
          tableData: []
        },
        {
          name: this.$t('m_standard_regex'),
          log_type: 'regular',
          tableData: []
        },
        {
          name: this.$t('m_custom_templates'),
          log_type: 'custom',
          tableData: []
        }
      ],
      tableColumn: [
        {
          type: 'selection',
          width: 60,
          align: 'center'
        },
        {
          title: this.$t('m_template_name'),
          key: 'name',
        },
        {
          title: this.$t('m_updatedBy'),
          key: 'update_user',
        },
        {
          title: this.$t('m_title_updateTime'),
          key: 'update_time',
        },
        {
          title: this.$t('m_table_action'),
          key: 'action',
          width: 200,
          align: 'left',
          fixed: 'right',
          render: (h, params) => (
            <div style="text-align: left; cursor: pointer;display: inline-flex;">
              <Tooltip max-width={400} placement="top" transfer content={this.$t('m_copy')}>
                <Button size="small" class="mr-1" type="success" on-click={() => this.copySingleItem(params.row)}>
                  <Icon type="md-document" size="16"></Icon>
                </Button>
              </Tooltip>
              <Tooltip content={this.$t('m_button_edit')} placement="top" transfer={true}>
                <Button
                  size="small"
                  type="primary"
                  onClick={() => this.editAction(params.row)}
                  style="margin-right:5px;"
                >
                  <Icon type="md-create" size="16"></Icon>
                </Button>
              </Tooltip>
              <Tooltip content={this.$t('m_view_associated')} placement="top" transfer={true}>
                <Button
                  size="small"
                  type="warning"
                  onClick={() => this.viewAction(params.row)}
                  style="margin-right:5px;"
                >
                  <Icon type="md-cube" size="16"></Icon>
                </Button>
              </Tooltip>
              <Tooltip content={this.$t('m_button_remove')} placement="top" transfer={true}>
                <Poptip
                  confirm
                  transfer
                  title={this.$t('m_delConfirm_tip')}
                  placement="left-end"
                  on-on-ok={() => {
                    this.confirmDeleteTemplate(params.row)
                  }}>
                  <Button size="small" type="error" class="mr-2" on-click={() => {
                    showPoptipOnTable()
                  }}>
                    <Icon type="md-trash" size="16" />
                  </Button>
                </Poptip>
              </Tooltip>
            </div>
          )
        }
      ],
      showServiceGroup: false,
      filterServiceGroup: '',
      serviceGroup: [], // 层级对象
      selectedParams: [], // 待导出数据
      updateUserOptions: [],
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  computed: {
    exportData() {
      return this.selectedParams.map(p => p.guid)
    }
  },
  mounted() {
    this.getTemplateList()
    this.getUpdateUserOptions()
  },
  methods: {
    getTemplateList() {
      this.spinShow = true
      this.selectedParams = []
      this.request('POST', this.apiCenter.logTemplateTableData, this.searchParams, resp => {
        this.data[0].tableData = resp.json_list
        this.data[1].tableData = resp.regular_list
        this.data[2].tableData = !isEmpty(resp.custom_list) ? resp.custom_list : []
      })
      this.spinShow = false
    },
    onFilterChange: debounce(function () {
      this.getTemplateList()
    }, 300),
    handleReset() {
      this.searchParams = {
        name: '',
        update_user: ''
      }
      this.getTemplateList()
    },
    changeRegexTableStatus(index, type) {
      if (type === 'in') {
        this.hideRegex.push(index)
      } else if (type === 'out') {
        const findIndex = this.hideRegex.findIndex(rIndex => rIndex === index)
        this.hideRegex.splice(findIndex, 1)
      }
    },
    // 添加模版
    addTemplate(log_type) {
      if (log_type === 'json') {
        this.$refs.jsonRegexRef.loadPage()
      } else if (log_type === 'regular') {
        this.$refs.standardRegexRef.loadPage()
      } else {
        this.$refs.customRegexRef.loadPage('add', '', '', '', true)
      }
    },
    // 编辑模版
    editAction(row) {
      if (row.log_type === 'json') {
        this.$refs.jsonRegexRef.loadPage(row.guid)
      } else if (row.log_type === 'regular') {
        this.$refs.standardRegexRef.loadPage(row.guid)
      } else {
        this.$refs.customRegexRef.loadPage('edit', '', '', row.guid, true)
        // this.$refs.customRegexRef.loadPage('add', '', '', '', true)
      }
    },
    copySingleItem(row) {
      if (row.log_type === 'json') {
        this.$refs.jsonRegexRef.loadPage(row.guid, 'copy')
      } else if (row.log_type === 'regular') {
        this.$refs.standardRegexRef.loadPage(row.guid, 'copy')
      } else {
        this.$refs.customRegexRef.loadPage('copy', '', '', row.guid, true)
      }
    },
    // 查看关联层级对象
    viewAction(row) {
      this.filterServiceGroup = ''
      const api = this.apiCenter.getAffectServiceGroupByGuid + row.guid
      this.request('GET', api, {}, resp => {
        this.serviceGroup = resp || []
        this.showServiceGroup = true
      })
    },
    confirmDeleteTemplate(row) {
      const api = this.apiCenter.deleteLogTemplate + row.guid
      this.request('DELETE', api, {}, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getTemplateList()
      })
    },

    // #region 导入导出
    onSelectAll(selection, tableIndex) {
      selection.forEach(se => {
        const findIndex = this.selectedParams.findIndex(
          param => param.guid === se.guid && param.tableIndex === tableIndex
        )
        if (findIndex === -1) {
          this.selectedParams.push({
            guid: se.guid,
            tableIndex
          })
        }
      })
    },
    onSelectAllCancel(selection, tableIndex) {
      this.selectedParams = this.selectedParams.filter(param => param.tableIndex !== tableIndex)
    },
    onSelect(selection, row, tableIndex) {
      this.selectedParams.push({
        guid: row.guid,
        tableIndex
      })
    },
    cancelSelect(selection, row, tableIndex) {
      const findIndex = this.selectedParams.findIndex(
        param => param.guid === row.guid && param.tableIndex === tableIndex
      )
      this.selectedParams.splice(findIndex, 1)
    },
    getUpdateUserOptions() {
      // const api = '/monitor/api/v2/service/log_metric/log_monitor_template/options'
      this.request('GET', this.apiCenter.logMonitorTemplateOptions, '', res => {
        this.updateUserOptions = res
      })
    }
  },
  components: {
    JsonRegex,
    ExportImport,
    StandardRegex,
    CustomRegex
  }
}
</script>
<style lang="less" scoped>
.log-template-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  .log-template-header-left {
    display: flex;
    align-items: center;
  }
}

.search-item {
  width: 200px;
  margin-right: 6px;
  // margin: 8px 6px 8px 0;
}
.table-zone {
  overflow: auto;
  height: ~"calc(100vh - 180px)";
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

.modal-container-normal {
  max-height: ~"calc(100vh - 400px)";
  overflow: auto;
}
.modal-container-fullscreen {
  max-height: ~"calc(100vh - 200px)";
  overflow: auto;
}

/deep/ .ivu-card-extra {
  top: 6px;
}
</style>
