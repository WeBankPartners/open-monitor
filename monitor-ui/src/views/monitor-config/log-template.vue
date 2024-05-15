<template>
  <div>
    <div>
      <Input
        v-model="searchParams.name"
        :placeholder="$t('m_template_name')"
        class="search-item"
        clearable
        @on-change="getTemplateList"
      ></Input>
      <Input
        v-model="searchParams.update_user"
        :placeholder="$t('m_updatedBy')"
        class="search-item"
        clearable
        @on-change="getTemplateList"
      ></Input>
      <span style="margin-top: 8px;margin-left: 24px;">
        <Button @click="getTemplateList" type="primary" style="background-color: #2d8cf0;">{{ $t('button.search') }}</Button>
        <Button @click="handleReset" style="margin-left: 5px">{{ $t('m_reset_condition') }}</Button>
      </span>
    </div>
    <div class="table-zone">
      <Spin v-if="spinShow" size="large">
        <Icon type="ios-loading" size="36"></Icon>
      </Spin>
      <template v-else>
        <template v-if="data.length > 0">
          <div v-for="(item, itemIndex) in data" :key="itemIndex">
            <Card>
              <div class="w-header" slot="title">
                <div class="title">
                  {{ item.name }}
                  <!-- {{ roleData.manageRoleDisplay }} -->
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
              <Button slot="extra" type="primary" size="small" @click.prevent="addTemplate(item.log_type)">{{ $t('button.add') }}</Button>
              <div v-show="!hideRegex.includes(itemIndex)">
                <Table
                  size="small"
                  :max-height="300"
                  :columns="tableColumn"
                  :data="item.tableData"
                  width="100%"
                ></Table>
                <!-- @on-select-all="selection => onSelectAll(selection, itemIndex)"
                  @on-select-all-cancel="selection => onSelectAllCancel(selection, itemIndex)"
                  @on-select="(selection, row) => onSelect(selection, row, itemIndex)"
                  @on-select-cancel="(selection, row) => cancelSelect(selection, row, itemIndex)" -->
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
    <!-- 删除组 -->
    <Modal
      v-model="isShowDeleteWarning"
      :title="$t('delConfirm.title')"
      @on-ok="confirmDeleteTemplate"
      @on-cancel="isShowDeleteWarning = false">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('m_delete_tip')}}: {{ toBeDeleted }}</p>
        </div>
      </div>
    </Modal>
    <!-- 查看管理层级对象 -->
    <Modal
      v-model="showServiceGroup"
      :fullscreen="isfullscreen"
      :title="$t('field.resourceLevel')">
      <div slot="header" class="custom-modal-header">
        <span>
          {{$t('field.resourceLevel')}}
        </span>
        <Icon v-if="isfullscreen" @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-expand" />
      </div>
      <Input v-model="filterServiceGroup" :placeholder="$t('m_filtering_info')" style="margin-bottom: 12px;"></Input>
      <div  :class="isfullscreen? 'modal-container-fullscreen':'modal-container-normal'">
        <template v-if="serviceGroup.length > 0">
          <Tag size="large" v-for="(item, index) in serviceGroup.filter(data => data.display_name.includes(filterServiceGroup))" :key="index">{{ item.display_name }}</Tag>
        </template>
        <template v-else>
          <Alert type="warning">{{ $t('m_noData') }}</Alert>
        </template>
      </div>
      <div slot="footer">
        <Button type="primary" @click="showServiceGroup=false">{{ $t('close') }}</Button>
      </div>
    </Modal>
  </div>
</template>

<script>
import JsonRegex from './log-template-config/json-regex.vue'
import StandardRegex from './log-template-config/standard-regex.vue'

export default {
  name: "log-template",
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
          title: this.$t('title.updateTime'),
          key: 'update_time',
        },
        {
          title: this.$t('table.action'),
          key: 'action',
          width: 200,
          align: 'left',
          fixed: 'right',
          render: (h, params) => {
            return (
              <div style="text-align: left; cursor: pointer;display: inline-flex;">
              <Tooltip content={this.$t('button.edit')} placement="top" transfer={true}>
                  <Button
                    size="small"
                    type="success"
                    onClick={() => this.editAction(params.row)}
                    style="margin-right:5px;"
                  >
                    <Icon type="ios-create-outline" size="16"></Icon>
                  </Button>
                </Tooltip>
                <Tooltip content={this.$t('m_view_associated')} placement="top" transfer={true}>
                  <Button
                    size="small"
                    type="primary"
                    onClick={() => this.viewAction(params.row)}
                    style="margin-right:5px;"
                  >
                    <Icon type="md-eye" size="16"></Icon>
                  </Button>
                </Tooltip>
                <Tooltip content={this.$t('button.remove')} placement="top" transfer={true}>
                  <Button
                    size="small"
                    type="error"
                    onClick={() => this.removeAction(params.row)}
                    style="margin-right:5px;"
                  >
                    <Icon type="ios-trash-outline" size="16"></Icon>
                  </Button>
                </Tooltip>
              </div>
            )
          }
        }
      ],
      isShowDeleteWarning: false,
      toBeDeleted: '', // 将被删除的模版名
      toBeDeletedGuid: '', // 待删除数据
      showServiceGroup: false,
      filterServiceGroup: '',
      serviceGroup: [] // 层级对象
    }
  },
  mounted () {
    this.getTemplateList()
  },
  methods: {
    getTemplateList () {
      this.spinShow = true
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.logTemplateTableData, this.searchParams, (resp) => {
        this.data[0].tableData = resp.json_list
        this.data[1].tableData = resp.regular_list
      })
      this.spinShow = false
    },
    handleReset () {
      this.searchParams = {
        name: '',
        update_user: ''
      }
      this.getTemplateList()
    },
    changeRegexTableStatus (index, type) {
      if (type === 'in') {
        this.hideRegex.push(index)
      } else if (type === 'out') {
        const findIndex = this.hideRegex.findIndex(rIndex => rIndex === index)
        this.hideRegex.splice(findIndex, 1)
      }
    },
    // 添加模版
    addTemplate (log_type) {
      if (log_type === 'json') {
        this.$refs.jsonRegexRef.loadPage()
      } else {
        this.$refs.standardRegexRef.loadPage()
      }
    },
    // 编辑模版
    editAction (row) {
      if (row.log_type === 'json') {
        this.$refs.jsonRegexRef.loadPage(row.guid)
      } else {
        this.$refs.standardRegexRef.loadPage(row.guid)
      }
    },
    // 查看关联层级对象
    viewAction (row) {
      this.filterServiceGroup = ''
      const api = this.$root.apiCenter.getAffectServiceGroupByGuid + row.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, {}, (resp) => {
        this.serviceGroup = resp || []
        this.showServiceGroup = true
      })
    },
    // 删除模版
    removeAction (row) {
      this.toBeDeleted = row.name
      this.toBeDeletedGuid = row.guid
      this.isShowDeleteWarning = true
    },
    confirmDeleteTemplate () {
      let api = this.$root.apiCenter.deleteLogTemplate + this.toBeDeletedGuid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, {}, () => {
        this.$Message.success(this.$t('tips.success'))
        this.getTemplateList()
      })
    }
  },
  components: {
    JsonRegex,
    StandardRegex
  }
}
</script>
<style lang="less" scoped>
.search-item {
  width: 200px;
  margin-right: 6px;
  margin: 8px 6px 8px 0;
}
.table-zone {
  overflow: auto;
  // height: calc(100vh - 270px);
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
</style>