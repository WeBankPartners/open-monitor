<template>
  <div class="custom-view-index">
    <div>
      <Row>
        <Tabs v-model="searchMap.permission" @on-click="onFilterConditionChange()">
          <TabPane :label="$t('m_chart_all')" name="all"></TabPane>
          <TabPane :label="$t('m_can_edit')" name="mgmt"></TabPane>
        </Tabs>
      </Row>
      <div class="table-data-search">
        <div style="font-size: 14px;min-width: 110px">{{$t('m_show_user_created')}}</div>
        <i-switch v-model="searchMap.show" class='mr-3' style="width: 60px" @on-change="onFilterConditionChange" />
        <Input
          v-model="searchMap.name"
          class="mr-3"
          type="text"
          style="width: 15%"
          :placeholder="$t('m_name')"
          clearable
          @on-change="onFilterConditionChange"
        />
        <Input
          v-model="searchMap.id"
          class="mr-3"
          type="text"
          style="width: 15%"
          :placeholder="$t('m_id')"
          clearable
          @on-change="onFilterConditionChange"
        >
        </Input>
        <Select
          v-model="searchMap.useRoles"
          class="mr-3"
          style="width: 15%"
          clearable
          :max-tag-count="1"
          filterable
          multiple
          :placeholder="$t('m_use_role')"
          @on-change="onFilterConditionChange"
        >
          <Option v-for="item in userRolesOptions" :value="item.name" :key="item.name">{{ item.display_name }}</Option>
        </Select>
        <Select
          v-model="searchMap.mgmtRoles"
          class="mr-3"
          style="width: 15%"
          clearable
          filterable
          :max-tag-count="1"
          multiple
          :placeholder="$t('m_manage_role')"
          @on-change="onFilterConditionChange"
        >
          <Option v-for="item in userRolesOptions" :value="item.name" :key="item.name">{{ item.display_name }}</Option>
        </Select>
        <Input
          v-model="searchMap.updateUser"
          class="mr-3"
          style="width: 15%"
          type="text"
          :placeholder="$t('m_updatedBy')"
          clearable
          @on-change="onFilterConditionChange"
        >
        </Input>
        <Button class="mr-5" @click="handleReset" type="default">{{ $t('m_reset') }}</Button>
        <Upload
          action="/monitor/api/v2/dashboard/custom/import"
          name="file"
          style="display: none"
          with-credentials
          :headers="uploadHeaders"
          :on-success="uploadSucess"
          :on-error="uploadFailed"
          :data="importExtraData"
          accept=".json"
        >
          <Button />
        </Upload>
        <Button type="success" @click="addBoardItem">{{$t('m_button_add')}}</Button>
        <Dropdown
          placement="bottom-start"
          @on-click="importPanel"
        >
          <Button type="primary">
            {{$t('m_import')}}
          </Button>
          <template  slot='list'>
            <DropdownMenu>
              <DropdownItem v-for="(item, index) in importTypeOptions"
                            :name="item.value"
                            :key="index"
              >
                {{ $t(item.name) }}
              </DropdownItem>
            </DropdownMenu>
          </template>
        </Dropdown>
        <Button @click="setDashboard">{{$t('m_button_setDashboard')}}</Button>
      </div>
    </div>
    <div class="mt-3 all-card-content">
      <div v-if="!dataList || dataList.length === 0" class="no-card-tips">{{$t('m_noData')}}</div>
      <Row class="all-card-item" :gutter="15" v-else>
        <Col v-for="(item,index) in dataList" :span="8" :key="index" class="panal-list">
        <Card>
          <div slot="title" class="panal-title">
            <div class='panal-title-name'>
              <Tag v-if='item.logMetricGroup' style="width: 40px" color='green'>auto</Tag>
              <Tooltip :content="item.name" :max-width='400' theme="dark" transfer placement="top">
                <span class='panal-title-name-text'>{{ item.name }}</span>
              </Tooltip>
            </div>
            <span class="panal-title-update">
              <span>{{$t('m_updatedBy')}}: {{item.updateUser}}</span>
              <span class="mt-1">{{item.updateTime}}</span>
            </span>
          </div>
          <div class="all-card-item-content mb-1">
            <div class="card-content mb-1" v-for="(keyItem, index) in cardContentList" :key="index">
              <span style="min-width: 80px">{{$t(keyItem.label)}}: </span>
              <div v-if="keyItem.type === 'string'">
                {{item[keyItem.key]}}
              </div>
              <div v-if="keyItem.type === 'array'">
                <div v-if="item[keyItem.key].length" class="card-content-array">
                  <BaseScrollTag :list="item[keyItem.key]"></BaseScrollTag>
                </div>
                <div v-else>-</div>
              </div>
            </div>
          </div>
          <div class="card-divider"></div>
          <div class="card-content-footer">
            <Tooltip placement="top" transfer :content="$t('m_copy')">
              <Button size="small" class="mr-1" type="success" @click="onDashboardItemCopy(item)">
                <Icon type="md-document" size="16"></Icon>
              </Button>
            </Tooltip>
            <Tooltip placement="top" transfer :content="$t('m_preview')">
              <Button size="small" type="info" @click="goToPanal(item, 'view')">
                <Icon type="md-eye" />
              </Button>
            </Tooltip>
            <Tooltip placement="top" transfer :content="$t('m_export')">
              <Button size="small" type="info" class="export-button" @click.stop="exportPanel(item)">
                <Icon type="md-cloud-upload" />
              </Button>
            </Tooltip>
            <template v-if="item.permission === 'mgmt'">
              <Tooltip placement="top" transfer :content="$t('m_button_edit')">
                <Button size="small" type="primary" @click.stop="goToPanal(item, 'edit')">
                  <Icon type="md-create" />
                </Button>
              </Tooltip>
              <Tooltip placement="top" transfer :content="$t('m_role_drawer_title')">
                <Button size="small" type="warning" @click.stop="editBoardAuth(item)">
                  <Icon type="md-person" />
                </Button>
              </Tooltip>
              <Poptip
                confirm
                :title="$t('m_delConfirm_tip')"
                placement="left-end"
                @on-ok="onDeleteConfirm(item)"
              >
                <Button size="small" type="error">
                  <Icon type="md-trash" />
                </Button>
              </Poptip>
            </template>
          </div>
        </Card>
        </Col>
      </Row>
      <Page
        class="card-pagination"
        :total="pagination.totalRows"
        @on-change="(e) => {pagination.currentPage = e; this.getViewList()}"
        :current="pagination.currentPage"
        :page-size="pagination.pageSize"
        show-total
      />
      <ModalComponent :modelConfig="authorizationModel">
        <template slot='authorization'>
          <div style="margin: 4px 12px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
            <template v-for="(item, index) in authorizationModel.result">
              <p :key="index" style="margin:6px 0">
                <Button
                  @click="deleteAuth(index)"
                  size="small"
                  type="error"
                  icon="md-trash"
                ></Button>
                <Select v-model="item.role_id" filterable style="width:200px">
                  <Option v-for="item in userRolesOptions" :value="item.id" :key="item.id">
                    {{item.display_name}}</Option>
                </Select>
                <Select v-model="item.permission" filterable style="width:200px">
                  <Option v-for="permission in ['mgmt', 'use']" :value="permission" :key="permission">{{
                    $t(permission)
                  }}</Option>
                </Select>
              </p>
            </template>
            <Button
              @click="addEmptyAuth"
              type="success"
              size="small"
              long
            >{{ $t('m_button_add') }}</Button>
          </div>
        </template>
      </ModalComponent>
      <AuthDialog ref="authDialog" :useRolesRequired="true" @sendAuth="saveTemplate">
        <template slot='content-top'>
          <div v-if="isAuthModalNameShow" class="auth-dialog-content">
            <span class="mr-3">{{$t('m_name')}}:</span>
            <Input style="width:calc(100% - 60px);" :maxlength="30" show-word-limit v-model.trim="addViewName"></Input>
          </div>
        </template>
      </AuthDialog>

      <Modal v-model="isShowProcessConfigModel"
             :title="$t(processConfigModel.modalTitle)"
             :mask-closable="false"
             @on-ok="processConfigSave"
      >
        <div style="padding: 0 12px; max-height: 500px; overflow-y: auto">
          <div style="display: flex;">
            <div class="port-title">
              <span>{{$t('m_tableKey_role')}}:</span>
            </div>
            <div class="port-title">
              <span>{{$t('m_custom_dashboard')}}:</span>
            </div>
          </div>
          <div v-for="(pl, plIndex) in processConfigModel.dashboardConfig" :key="plIndex">
            <div class="port-config">
              <div style="width: 40%">
                <label>{{pl.display_role_name}}：</label>
              </div>
              <div style="width: 55%">
                <Select filterable clearable v-model="pl.main_page_id" class='dashboard-config-selection' :placeholder="$t('m_please_select') + $t('m_dashboard_for_role')">
                  <Option v-for="item in pl.options" :value="item.id" :key="item.id">{{ item.option_text }}</Option>
                </Select>
              </div>
            </div>
          </div>
        </div>
      </Modal>
      <ExportChartModal
        :isModalShow="isModalShow"
        :pannelId="pannelId"
        :panalName="panalName"
        @close="() => isModalShow = false"
      >
      </ExportChartModal>
    </div>
  </div>
</template>
<script>
import debounce from 'lodash/debounce'
import cloneDeep from 'lodash/cloneDeep'
import isEmpty from 'lodash/isEmpty'
import AuthDialog from '@/components/auth.vue'
import ExportChartModal from './export-chart-modal.vue'
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
export default {
  name: '',
  computed: {
    isAuthModalNameShow() {
      return this.authViewType === 'add'
    }
  },
  data() {
    return {
      dataList: [],
      processConfigModel: {
        modalId: 'set_dashboard_modal',
        modalTitle: 'm_button_setDashboard',
        isAdd: true,
        saveFunc: 'processConfigSave',
        config: [{
          name: 'processConfig',
          type: 'slot'
        }],
        addRow: {
          businessSet: [],
        },
        dashboardConfig: [],
      },
      authorizationModel: {
        modalId: 'authorization_model',
        modalTitle: 'm_button_authorization',
        isAdd: true,
        saveFunc: 'authorizationSave',
        config: [
          {
            name: 'authorization',
            type: 'slot'
          }
        ],
        addRow: {
          role: []
        },
        roleList: [],
        result: []
      },
      dashboard_id: '',
      pathMap: {},
      searchMap: {
        show: false,
        permission: '',
        name: '',
        id: '',
        useRoles: [],
        mgmtRoles: [],
        updateUser: ''
      },
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      cardContentList: [
        {
          key: 'id',
          label: 'm_id',
          type: 'string'
        },
        {
          key: 'displayMgmtRoles',
          label: 'm_manage_role',
          type: 'array'
        },
        {
          key: 'displayUseRoles',
          label: 'm_use_role',
          type: 'array'
        },
        {
          key: 'mainPage',
          label: 'm_home_application',
          type: 'array'
        }
      ],
      pagination: {
        totalRows: 100,
        currentPage: 1,
        pageSize: 18
      },
      mgmtRoles: [],
      userRoles: [],
      mgmtRolesOptions: [],
      userRolesOptions: [],
      addViewName: '',
      authViewType: 'add', // 枚举值为add(面板新增)，edit(面板编辑), import(面板导入)
      boardId: null,
      importTypeOptions: [
        {
          name: 'm_samename_coverage',
          value: 'cover'
        },
        {
          name: 'm_samename_add',
          value: 'insert'
        }
      ],
      importPanelType: '', // 枚举值为cover 同名覆盖,insert 同名新增
      importExtraData: {},
      uploadHeaders: {
        'X-Auth-Token': getToken() || null,
        Authorization: this.getAuthorization()
      },
      isModalShow: false,
      pannelId: null,
      panalName: '',
      isShowProcessConfigModel: false
    }
  },
  mounted() {
    if (this.$route.query.needCache === 'yes') {
      // 读取列表搜索参数
      const storage = window.sessionStorage.getItem('search_custom_view') || ''
      if (storage) {
        const { searchParams } = JSON.parse(storage)
        this.searchMap = searchParams
      }
    }
    this.initData()
  },
  beforeDestroy() {
    // 缓存列表搜索条件
    const storage = {
      searchParams: this.searchMap
    }
    window.sessionStorage.setItem('search_custom_view', JSON.stringify(storage))
  },
  methods: {
    initData() {
      this.pathMap = this.$root.apiCenter.template
      this.pagination.pageSize = 18
      this.pagination.currentPage = 1
      this.getViewList()
      this.getAllRoles()
      if (this.$route.query.isCreate) {
        setTimeout(() => {
          this.addBoardItem()
        }, 500)
      }
    },
    handleReset() {
      const resetObj = {
        name: '',
        id: '',
        useRoles: [],
        mgmtRoles: [],
        updateUser: '',
        show: false
      }
      this.searchMap = Object.assign({}, this.searchMap, resetObj)
      this.pagination.currentPage = 1
      this.getViewList()
    },
    getAuthorization() {
      if (localStorage.getItem('monitor-accessToken')) {
        return 'Bearer ' + localStorage.getItem('monitor-accessToken')
      }
      return (window.request ? 'Bearer ' + getPlatFormToken() : getToken()) || null

    },
    deleteAuth(index) {
      this.authorizationModel.result.splice(index, 1)
    },
    addEmptyAuth() {
      this.authorizationModel.result.push({
        role_id: '',
        permission: 'use'
      })
    },
    processConfigSave() {
      const params = []
      this.processConfigModel.dashboardConfig.forEach(item => {
        params.push({
          role_name: item.role_name,
          main_page_id: item.main_page_id,
        })
      })
      this.request('POST','/monitor/api/v1/dashboard/custom/main/set', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.isShowProcessConfigModel = false
        this.getViewList()
      })
    },
    authorization(item) {
      this.dashboard_id = item.id
      const params = {
        dashboard_id: item.id
      }
      this.request('GET','/monitor/api/v1/dashboard/custom/role/get', params, res => {
        this.authorizationModel.result = res
        this.$root.JQ('#authorization_model').modal('show')
      })
    },
    authorizationSave() {
      const params = {
        dashboard_id: this.dashboard_id,
        permission_list: this.authorizationModel.result
      }
      this.request('POST', '/monitor/api/v1/dashboard/custom/role/save', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#authorization_model').modal('hide')
      })
    },
    getAllRoles() {
      const params = {
        page: 1,
        size: 1000
      }
      this.request('GET','/monitor/api/v1/user/role/list', params, res => {
        this.userRolesOptions = this.processRolesList(res.data)
      })
      this.request('GET', '/monitor/api/v1/user/manage_role/list', {}, res => {
        this.mgmtRolesOptions = this.processRolesList(res)
      })
    },
    processRolesList(list = []) {
      if (isEmpty(list)) {
        return []
      }
      const resArr = cloneDeep(list).map(item => ({
        ...item,
        key: item.name,
        label: item.displayName || item.display_name
      }))
      return resArr
    },
    addBoardItem() {
      this.authViewType = 'add'
      this.addViewName = ''
      this.$refs.authDialog.startAuth([], [], this.mgmtRolesOptions, this.userRolesOptions)
    },
    editBoardAuth(item) {
      this.authViewType = 'edit'
      this.boardId = item.id
      this.$refs.authDialog.startAuth(item.mgmtRoles, item.useRoles, this.mgmtRolesOptions, this.userRolesOptions)
    },
    onDeleteConfirm(item) {
      this.removeTemplate(item)
    },
    removeTemplate(item) {
      const params = {id: item.id}
      this.request('DELETE',this.pathMap.deleteV2, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getViewList()
      })
    },
    getViewList() {
      const cloneSearchMap = cloneDeep(this.searchMap)
      cloneSearchMap.show = cloneSearchMap.show === true ? 'me' : ''
      const params = Object.assign(cloneSearchMap, {
        pageSize: this.pagination.pageSize,
        startIndex: this.pagination.pageSize * (this.pagination.currentPage - 1)
      })
      if (!params.id || isNaN(Number(params.id))) {
        params.id = 0
      } else {
        params.id = Number(params.id)
      }
      if (params.permission === 'all') {
        params.permission = ''
      }
      this.request('POST',this.pathMap.listV2, params, responseData => {
        this.dataList = responseData.contents
        this.pagination.totalRows = responseData.pageInfo.totalRows
      })
    },
    setDashboard() {
      this.request('GET',this.pathMap.portalConfig, '', responseData => {
        this.processConfigModel.dashboardConfig = responseData
        this.isShowProcessConfigModel = true
      })
    },
    goToPanal(panalItem, type) {
      const params = {
        permission: type,
        panalItem,
        pannelId: panalItem.id
      }
      this.$router.push({
        name: 'viewConfig',
        params
      })
    },
    onFilterConditionChange: debounce(function () {
      this.pagination.currentPage = 1
      this.getViewList()
    }, 300),
    saveTemplate(mgmtRoles, useRoles) {
      if (this.isAuthModalNameShow && !this.addViewName) {
        this.$nextTick(() => {
          this.$Message.warning(this.$t('m_name') + this.$t('m_cannot_be_empty'))
          this.$refs.authDialog.flowRoleManageModal = true
        })
        return
      }
      if (this.authViewType === 'import') {
        this.$nextTick(() => {
          this.$refs.authDialog.flowRoleManageModal = true
          document.querySelector('.ivu-upload-input').click()
        })
      }
      this.mgmtRoles = mgmtRoles
      this.userRoles = useRoles
      this.submitData()
    },
    submitData() {
      const params = {
        mgmtRoles: this.mgmtRoles,
        useRoles: this.userRoles
      }
      let path = ''
      if (this.authViewType === 'import') {
        this.importExtraData = {
          rule: this.importPanelType,
          useRoles: this.userRoles,
          mgmtRoles: this.mgmtRoles[0]
        }
        return
      }
      if (this.authViewType === 'add') {
        params.name = this.addViewName
        path = '/monitor/api/v2/dashboard/custom'
      } else if (this.authViewType === 'edit') {
        // 修改自定义看板权限
        params.id = this.boardId
        path = '/monitor/api/v2/dashboard/custom/permission'
      } else if (this.authViewType === 'copyDashboard') {
        params.dashboardId = this.pannelId
        path = '/monitor/api/v2/dashboard/custom/copy'
        params.mgmtRole = params.mgmtRoles[0]
      }
      this.request('POST', path, params, val => {
        this.getViewList()
        if (this.authViewType === 'add') {
          this.goToPanal(val, 'edit')
        }
      })
    },
    importPanel(type) {
      this.importPanelType = type
      this.authViewType = 'import'
      this.$refs.authDialog.startAuth([], [], this.mgmtRolesOptions, this.userRolesOptions)
    },
    uploadSucess(res) {
      if (!isEmpty(res.data)) {
        let content = ''
        for (const key in res.data) {
          content += key + '(图表名)不存在以下指标: ' + res.data[key].join(';') + '。 '
        }
        this.$Message.warning({
          content,
          duration: 5
        })
      } else {
        this.$Message.success(this.$t('m_tips_success'))
      }
      this.$refs.authDialog.flowRoleManageModal = false
      this.getViewList()
    },
    uploadFailed(file) {
      this.$Message.warning(file.message)
    },
    exportPanel(item) {
      this.panalName = item.name
      this.pannelId = item.id
      this.isModalShow = true
    },
    onDashboardItemCopy(item) {
      this.pannelId = item.id
      this.authViewType = 'copyDashboard'
      this.$refs.authDialog.startAuth([], [], this.mgmtRolesOptions, this.userRolesOptions)
    }
  },
  components: {
    AuthDialog,
    ExportChartModal
  }
}
</script>

<style lang='less'>
.screen-config-menu.ivu-menu-horizontal.ivu-menu-light:after {
  height: 0px;
}

.panal-list {
  .ivu-card-body {
    padding-bottom: 5px
  }
  .ivu-card-head {
    padding: 12px;
  }
}

.export-button.ivu-btn-info {
  background-color: #aa8aea;
  border-color: #aa8aea;
}

body::-webkit-scrollbar {
    // display: none;
}

.dashboard-config-selection {
  width:200px;
  .ivu-select-selection {
    height: 34px !important;
  }
}

</style>

<style scoped lang="less">
.all-card-content {
  max-height: ~'calc(100vh - 210px)';
  overflow-x: hidden;
  overflow-y: auto;
}
.screen-config-menu {
  width: 150px;
}
.port-title {
  width: 40%;
  font-size: 14px;
}
.port-config {
  display: flex;
  margin-top: 4px;
}
li {
  list-style: none;
}
.operational-zone {
  margin: 0 16px 16px 16px;
}
.all-card-item {
  display: flex;
  flex-wrap: wrap;
  padding-bottom: 50px;
  .all-card-item-content {
    min-height: 160px;
    height: 160px;
    .detail-eye {
      position: absolute;
      right: 20px;
      top: 60px;
      cursor: pointer;
      color: #2d8cf0;
    }
  }
}

.panal-list {
  margin-bottom: 15px;
  min-height: 240px;
  display: inline-block;
  .panal-title {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    color: @blue-2;
    height: 30px;
    &-name {
      display: flex;
      font-size: 15px;
      flex-direction: row;
      align-items: center;
      justify-content: flex-start;
      .panal-title-name-text {
        line-height: 18px;
        display: inline-block;
        max-width: 200px;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
    }
    &-update {
      display: flex;
      flex-direction: column;
      width: 130px;
      font-size: 13px;
    }
  }
  .card-divider {
    height: 1px;
    width: 100%;
    background-color: #e8eaec;
  }
  .card-content-footer {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    height: 40px;
  }
  .card-content-array {
    position: relative;
    display: flex;
    flex-wrap: wrap;
    width: 100%;
    max-height: 52px;
    margin-top: 2px;
  }
  .exceed-content::after {
    content: '...';
    font-size: 20px;
    position: absolute;
    bottom: 0px;
    right: 0px;
  }
}

.fa-star {
  color: @color-orange-F;
}

.card-content {
  display: flex;
  align-items: center;
}

.auth-dialog-content {
  display: flex;
  align-items: center;
  margin-bottom: 10px
}
.auth-dialog-content::before {
    content: '*';
    display: inline-block;
    margin-right: 4px;
    line-height: 1;
    font-family: SimSun;
    font-size: 14px;
    color: #ed4014;
}

.card-pagination {
  position: fixed;
  right: 20px;
  bottom: 20px;
}

.no-card-tips {
  display: flex;
  justify-content: center;
  height: 40%;
  margin-top: 200px;
  font-size: 14px;
  color: #515a6e;
}
.table-data-search {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
