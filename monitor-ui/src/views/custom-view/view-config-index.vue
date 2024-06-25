<template>
  <div>
    <div>
      <Row>
        <Tabs v-model="searchMap.permission" @on-click="onFilterConditionChange()">
          <TabPane :label="$t('m_chart_all')" name="all"></TabPane>
          <TabPane :label="$t('m_can_edit')" name="mgmt"></TabPane>
        </Tabs>
      </Row>
      <div class="table-data-search">
        <Input
          v-model="searchMap.name"
          class="mr-3"
          type="text"
          :placeholder="$t('m_name')"
          clearable
          @on-change="onFilterConditionChange"
        />
        <Input
          v-model="searchMap.id"
          class="mr-3"
          type="text"
          :placeholder="$t('m_id')"
          clearable
          @on-change="onFilterConditionChange"
        >
        </Input>
        <Select
          v-model="searchMap.useRoles"
          class="mr-3"
          clearable
          :max-tag-count="1"
          filterable
          multiple
          :placeholder="$t('m_use_role')"
          style="width: 90%"
          @on-change="onFilterConditionChange"
        >
          <Option v-for="item in userRolesOptions" :value="item.name" :key="item.name">{{ item.display_name }}</Option>
        </Select>
        <Select
          v-model="searchMap.mgmtRoles"
          class="mr-3"
          clearable
          filterable
          :max-tag-count="1"
          multiple
          :placeholder="$t('m_manage_role')"
          style="width: 90%"
          @on-change="onFilterConditionChange"
        >
          <Option v-for="item in userRolesOptions" :value="item.name" :key="item.name">{{ item.display_name }}</Option>
        </Select>
        <Input
          v-model="searchMap.updateUser"
          class="mr-3"
          style="width: 90%"
          type="text"
          :placeholder="$t('m_updatedBy')"
          clearable
          @on-change="onFilterConditionChange"
        >
        </Input>
        <Button class="mr-5" @click="handleReset" type="default">{{ $t('m_reset') }}</Button>
        <Button type="success" @click="addBoardItem">{{$t('m_button_add')}}</Button>
        <button class="ml-2 btn btn-sm btn-cancel-f" @click="setDashboard">
          {{$t('m_button_setDashboard')}}
        </button>
      </div>
    </div>

    <div class="mt-3">
      <div v-if="!dataList || dataList.length === 0" class="no-card-tips">{{$t('m_noData')}}</div>
      <Row class="all-card-item" :gutter="15" v-else>
        <Col v-for="(item,index) in dataList" :span="8" :key="index" class="panal-list">
          <Card>
            <div slot="title" class="panal-title">
              <span class="panal-title-name">{{ item.name }}</span>
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
                      <ScrollTag :list="item[keyItem.key]"></ScrollTag>
                    </div>
                    <div v-else>-</div>
                  </div>
              </div>
            </div>
            <div class="card-divider"></div>
            <div class="card-content-footer">
              <Button size="small" type="info" @click="goToPanal(item, 'view')">
                <Icon type="md-eye" />
              </Button>
              <template v-if="item.permission === 'mgmt'">
                <Button size="small" type="primary" @click.stop="goToPanal(item, 'edit')">
                  <Icon type="md-create" />
                </Button>
                <Button size="small" type="warning" @click.stop="editBoardAuth(item)">
                  <Icon type="md-person" />
                </Button>
                <Button size="small" type="error" @click.stop="deleteConfirmModal(item)">
                  <Icon type="md-trash" />
                </Button>
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
        <template #authorization>
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
              >{{ $t('m_button_add') }}</Button
            >
          </div>
        </template>
      </ModalComponent>

      <AuthDialog ref="authDialog" :useRolesRequired="true" @sendAuth="saveTemplate">
        <template #content-top>
          <div v-if="isAddViewType" class="auth-dialog-content">
            <span class="mr-3">{{$t('m_name')}}:</span>
            <Input style="width:calc(100% - 60px);" :maxlength="30" show-word-limit v-model.trim="addViewName"></Input>
          </div>
        </template>
      </AuthDialog>
      <ModalComponent :modelConfig="processConfigModel">
        <div slot="processConfig" style="max-height: 500px;overflow-y: scroll;padding: 0 12px;">
          <section>
            <div style="display: flex;">
              <div class="port-title">
                <span>{{$t('m_tableKey_role')}}:</span>
              </div>
              <div class="port-title">
                <span>{{$t('m_custom_dashboard')}}:</span>
              </div>
            </div>
          </section>
          <section v-for="(pl, plIndex) in processConfigModel.dashboardConfig" :key="plIndex">
            <div class="port-config">
              <div style="width: 40%">
                <label>{{pl.display_role_name}}：</label>
              </div>
              <div style="width: 55%">
                <Select filterable clearable v-model="pl.main_page_id" style="width:200px" :placeholder="$t('m_please_select') + $t('m_dashboard_for_role')">
                  <Option v-for="item in pl.options" :value="item.id" :key="item.id">{{ item.option_text }}</Option>
                </Select>
              </div>
            </div>
          </section>
        </div>
      </ModalComponent>
      <Modal
        v-model="isShowWarning"
        :title="$t('m_delConfirm_title')"
        @on-ok="onDeleteConfirm"
        @on-cancel="onCancelDelete">
        <div class="modal-body" style="padding:10px 20px;">
          <div style="text-align:center">
            <p style="color: red">{{$t('m_delConfirm_tip')}}</p>
          </div>
        </div>
      </Modal>
    </div>
  </div>
</template>

<script>
import debounce from 'lodash/debounce';
import cloneDeep from 'lodash/cloneDeep'
import isEmpty from 'lodash/isEmpty'
import AuthDialog from '@/components/auth.vue'
import ScrollTag from '@/components/scroll-tag.vue'
export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
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
          {name:'authorization',type:'slot'}
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
        permission: "",
        name: "",
        id: "",
        useRoles: [],
        mgmtRoles: [],
        updateUser: ""
      },
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      cardContentList: [
        {
          key: "id",
          label: "m_id",
          type: "string"
        }, 
        {
          key: "displayMgmtRoles",
          label: "m_manage_role",
          type: "array"
        },
        {
          key: "displayUseRoles",
          label: "m_use_role",
          type: "array"
        },
        {
          key: "mainPage",
          label: "m_home_application",
          type: "array"
        }
      ],
      pagination: {
        totalRows: 100,
        currentPage: 2,
        pageSize: 18
      },
      mgmtRoles: [],
      userRoles: [],
      mgmtRolesOptions: [], 
      userRolesOptions: [],
      addViewName: '',
      isAddViewType: true,
      boardId: null
    }
  },
  mounted(){
    this.pathMap = this.$root.apiCenter.template;
    this.pagination.pageSize = 18;
    this.pagination.currentPage = 1;
    this.getViewList()
    this.getAllRoles()
    if (this.$route.query.isCreate) {
      setTimeout(() => {
        this.addBoardItem()
      }, 500)
    }
  },
  methods: {
    handleReset () {
      const resetObj = {
        name: "",
        id: "",
        useRoles: [],
        mgmtRoles: [],
        updateUser: ""
      }
      this.searchMap = Object.assign({}, this.searchMap, resetObj)
      this.pagination.currentPage = 1;
      this.getViewList()
    },
    deleteAuth (index) {
      this.authorizationModel.result.splice(index, 1)
    },
    addEmptyAuth () {
      this.authorizationModel.result.push({role_id: '', permission: 'use'})
    },

    processConfigSave () {
      let params = []
      this.processConfigModel.dashboardConfig.forEach(item => {
        params.push({
          role_name: item.role_name,
          main_page_id: item.main_page_id,
        })
      })
      this.request('POST','/monitor/api/v1/dashboard/custom/main/set', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#set_dashboard_modal').modal('hide')
        this.getViewList()
      })
    },
    authorization (item) {
      this.dashboard_id = item.id
      const params = {
        dashboard_id: item.id
      }
      this.request('GET','/monitor/api/v1/dashboard/custom/role/get', params, (res) => {
        this.authorizationModel.result = res
        this.$root.JQ('#authorization_model').modal('show')
      })
    },
    authorizationSave () {
      let params = {
        dashboard_id: this.dashboard_id,
        permission_list: this.authorizationModel.result
      }
      this.request('POST', '/monitor/api/v1/dashboard/custom/role/save', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#authorization_model').modal('hide')
      })
    },
    getAllRoles () {
      const params = {
        page: 1,
        size: 1000
      }
      this.request('GET','/monitor/api/v1/user/role/list', params, res => {
        this.userRolesOptions = this.processRolesList(res.data)
      })
      this.request('GET', '/monitor/api/v1/user/manage_role/list', {}, res => {
        this.mgmtRolesOptions = this.processRolesList(res);
      })
    },
    processRolesList(list = []) {
      if (isEmpty(list)) return [];
      const resArr = cloneDeep(list).map(item => {
        return {
          ...item,
          key: item.name,
          label: item.displayName || item.display_name
        }
      })
      return resArr
    },
    addBoardItem () {
      this.isAddViewType = true;
      this.addViewName = '';
      this.$refs.authDialog.startAuth([], [], this.mgmtRolesOptions, this.userRolesOptions);
    },
    editBoardAuth(item) {
      this.isAddViewType = false;
      this.boardId = item.id;
      this.$refs.authDialog.startAuth(item.mgmtRoles, item.useRoles, this.mgmtRolesOptions, this.userRolesOptions);
    },
    
    deleteConfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarning = true
    },
    onDeleteConfirm () {
      this.removeTemplate(this.selectedData)
    },
    onCancelDelete () {
      this.isShowWarning = false
    },
    removeTemplate (item) {
      let params = {id: item.id}
      this.request('DELETE',this.pathMap.deleteV2, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getViewList()
      })
    },
    getViewList () {
      const params = Object.assign(cloneDeep(this.searchMap), {
        pageSize: this.pagination.pageSize,
        startIndex: this.pagination.pageSize * (this.pagination.currentPage - 1)
      });
      if (!params.id || isNaN(Number(params.id))) {
        params.id = 0
      } else {
        params.id = Number(params.id)
      }
      if (params.permission === 'all') {
        params.permission = ''
      }
      this.request('POST',this.pathMap.listV2, params, responseData => {
        this.dataList = responseData.contents;
        this.pagination.totalRows = responseData.pageInfo.totalRows;
      })
    },
    setDashboard () {
      this.request('GET',this.pathMap.portalConfig, '', responseData => {
        this.processConfigModel.dashboardConfig = responseData
        this.$root.JQ('#set_dashboard_modal').modal('show')
      })
    },
    goToPanal(panalItem, type) {
      const params = {
        permission: type,
        panalItem,
        pannelId: panalItem.id
      }
      this.$router.push({name:'viewConfig',params})
    },

    onFilterConditionChange: debounce(function () {
      this.getViewList()
    }, 300),

    saveTemplate(mgmtRoles, useRoles) {
      if (this.isAddViewType && !this.addViewName ) {
        this.$nextTick(() => {
          this.$Message.warning(this.$t('m_name') + this.$t('m_cannot_be_empty'))
          this.$refs.authDialog.flowRoleManageModal = true
        })
        return
      }
      this.mgmtRoles = mgmtRoles;
      this.userRoles = useRoles;
      this.submitData()
    },
    submitData() {
      const params = {
        mgmtRoles: this.mgmtRoles,
        useRoles: this.userRoles
      }
      let path = '';
      if (this.isAddViewType) {
        params.name = this.addViewName;
        path = '/monitor/api/v2/dashboard/custom';
      } else {
        // 修改自定义看板权限
        params.id = this.boardId;
        path = '/monitor/api/v2/dashboard/custom/permission'
      }
      this.request('POST', path, params, (val) => {
        this.getViewList();
        if (this.isAddViewType) {
          this.goToPanal(val, 'edit')
        }
      })
    }
  },
  components: {
    AuthDialog,
    ScrollTag
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

</style>

<style scoped lang="less">
.screen-config-menu {
  width: 150px;
}
.port-title {
  width: 40%;
  font-size: 14px;
  // padding: 2px 0 2px 4px;
  // border: 1px solid @blue-2;
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
      font-size: 15px;
      flex: 1;
      display: -webkit-box;
      -webkit-box-orient: vertical;
      -webkit-line-clamp: 2;
      overflow: hidden;
      text-overflow: ellipsis;
      word-break: break-all;
      padding-right: 5px;
    }
    &-update {
      display: flex;
      flex-direction: column;
      width: 130px;
      font-size: 13px;
    }
  }
  .panal-title > div {
    display: flex;
    flex-direction: column;
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
}
</style>
