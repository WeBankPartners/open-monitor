<template>
  <div>
    <div>
      <Row>
        <Tabs v-model="searchMap.permission" @on-click="onFilterConditionChange()">
          <TabPane :label="$t('m_chart_all')" name="ALL"></TabPane>
          <TabPane :label="$t('m_can_edit')" name="MGMT"></TabPane>
        </Tabs>
      </Row>
      <Row>
        <Col span="3">
          <Input
            v-model="searchMap.name"
            style="width: 90%"
            type="text"
            :placeholder="$t('m_name')"
            clearable
            @on-change="onFilterConditionChange"
          />
        </Col>
        <Col span="3">
          <Input
            v-model="searchMap.id"
            style="width: 90%"
            type="text"
            :placeholder="$t('m_id')"
            clearable
            @on-change="onFilterConditionChange"
          >
          </Input>
        </Col>
        <Col span="4">
          <Select
            v-model="searchMap.useRoles"
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
        </Col>
        <Col span="4">
          <Select
            v-model="searchMap.mgmtRoles"
            clearable
            filterable
            :max-tag-count="1"
            multiple
            :placeholder="$t('m_manage_role')"
            style="width: 90%"
            @on-change="onFilterConditionChange"
          >
            <Option v-for="item in mgmtRolesOptions" :value="item.name" :key="item.name">{{ item.display_name }}</Option>
          </Select>
        </Col>
        <Col span="4">
          <Input
            v-model="searchMap.updateUser"
            style="width: 90%"
            type="text"
            :placeholder="$t('m_updatedBy')"
            clearable
            @on-change="onFilterConditionChange"
          >
          </Input>
        </Col>
        <Col span="3">
          <Button @click="getViewList" type="primary">{{ $t('m_search') }}</Button>
        </Col>
        <div style="display:flex;float:right;">
          <button class="btn btn-sm btn-confirm-f" @click="addBoardItem">
            <i class="fa fa-plus"></i>{{$t('button.add')}}
          </button>
          <button class="btn btn-sm btn-cancel-f" @click="setDashboard">
            {{$t('button.setDashboard')}}
          </button>
        </div>
      </Row>
    </div>

    <div class="mt-3">
      <section class="all-card-item">
          <template v-for="(item,index) in dataList">
            <div :key="index" class="panal-list" @click="goToPanal(item, 'view')">
              <Card>
                <div slot="title" class="panal-title">
                  <h5>{{ item.name }}</h5>
                  <div>
                    <span>{{$t('m_updatedBy')}}: {{item.updateUser}}</span>
                    <span>{{$t('m_update_time')}}: {{item.updateTime}}</span>
                  </div>
                </div>
                <div>
                  <div class="card-content mb-1" v-for="(keyItem, index) in cardContentList" :key="index"> 
                      <span style="min-width: 80px">{{$t(keyItem.label)}}: </span>
                      <div v-if="keyItem.type === 'string'">
                        {{item[keyItem.key]}}
                      </div>
                      <div class="card-content-array" v-if="keyItem.type === 'array'">
                          <Tag 
                            v-for="(tag, tagIndex) in item[keyItem.key]"
                            :key="tagIndex"
                            color="blue">
                            {{tag}}
                          </Tag>
                      </div>
                  </div>
                </div>
                <div class="card-divider"></div>
                <div class="card-content-footer">
                  <template v-if="item.permission === 'mgmt'">
                    <Button size="small"  type="primary" @click.stop="goToPanal(item, 'edit')">
                      <Icon type="md-create" />
                    </Button>
                    <Button size="small"  type="warning" @click.stop="editBoardAuth(item)">
                      <Icon type="md-person" />
                    </Button>
                    <Button size="small" type="error" @click.stop="deleteConfirmModal(item)">
                      <Icon type="md-trash" />
                    </Button>
                  </template>
                </div>
              </Card>
            </div>
          </template>
      </section>
      <Page
        style="float: right"
        :total="pagination.totalRows"
        @on-change="(e) => {pagination.currentPage = e; this.getViewList(e)}"
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
                  style="background-color: #ff9900;border-color: #ff9900;"
                  type="error"
                  icon="md-close"
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
              style="background-color: #0080FF;border-color: #0080FF;"
              long
              >{{ $t('button.add') }}</Button
            >
          </div>
        </template>
      </ModalComponent>

      <AuthDialog ref="authDialog" :useRolesRequired="true" @sendAuth="saveTemplate">
        <template #content-top>
          <div v-if="isAddViewType" class="auth-dialog-content">
            <span class="mr-3">{{$t('m_name')}}:</span>
            <Input style="width: 350px" v-model="addViewName"></Input>
          </div>
        </template>
      </AuthDialog>
      <ModalComponent :modelConfig="processConfigModel">
        <div slot="processConfig">
          <section>
            <div style="display: flex;">
              <div class="port-title">
                <span>{{$t('tableKey.role')}}:</span>
              </div>
              <div class="port-title">
                <span>{{$t('menu.customViews')}}:</span>
              </div>
            </div>
          </section>
          <section v-for="(pl, plIndex) in processConfigModel.dashboardConfig" :key="plIndex">
            <div class="port-config">
              <div style="width: 40%">
                <label>{{pl.role_name}}：</label>
              </div>
              <div style="width: 55%">
                <Select filterable clearable v-model="pl.main_page_id" style="width:200px" :placeholder="$t('placeholder.refresh')">
                  <Option v-for="item in pl.options" :value="item.id" :key="item.id">{{ item.option_text }}</Option>
                </Select>
              </div>
            </div>
          </section>
        </div>
      </ModalComponent>
      <Modal
        v-model="isShowWarning"
        :title="$t('delConfirm.title')"
        @on-ok="onDeleteConfirm"
        @on-cancel="onCancelDelete">
        <div class="modal-body" style="padding:30px">
          <div style="text-align:center">
            <p style="color: red">{{$t('delConfirm.tip')}}</p>
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
import { nextTick } from 'vue';
export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
      dataList: [
      ],
      processConfigModel: {
        modalId: 'set_dashboard_modal',
        modalTitle: 'button.setDashboard',
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
        modalTitle: 'button.authorization',
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
          key: "mgmtRoles",
          label: "m_manage_role",
          type: "array"
        },
        {
          key: "useRoles",
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
        pageSize: 6
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
    this.getViewList()
    this.getAllRoles()
  },
  methods: {
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
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#set_dashboard_modal').modal('hide')
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
        this.$Message.success(this.$t('tips.success'))
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
        this.$Message.success(this.$t('tips.success'))
        this.getViewList()
      })
    },
    getViewList (page=1, pageSize=6) {
      const params = Object.assign(cloneDeep(this.searchMap), {
        pageSize,
        startIndex: pageSize * (page - 1)
      });

      if (!params.id || isNaN(Number(params.id))) {
        params.id = 0
      } else {
        params.id = Number(params.id)
      }
      if (params.permission === 'ALL') {
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
        panalItem
      }
      this.$router.push({name:'viewConfig',params:params})
    },

    onFilterConditionChange: debounce(function () {
      this.getViewList()
    }, 300),

    saveTemplate(mgmtRoles, useRoles) {
      this.mgmtRoles = mgmtRoles;
      this.userRoles = useRoles;
      this.submitData()
      if (this.isAddViewType && !this.addViewName ) {
        nextTick(() => {
          this.$refs.authDialog.flowRoleManageModal = true
        })
      }
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
      this.request('POST', path, params, () => {
        this.getViewList();
      })
    }
  },
  components: {
    AuthDialog
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
}

</style>

<style scoped lang="less">
.screen-config-menu {
  width: 150px;
}
.port-title {
  width: 40%;
  font-size: 14px;
  padding: 2px 0 2px 4px;
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
}
.panal-list {
  margin: 8px;
  width: 390px;
  min-height: 240px;
  display: inline-block;
  .panal-title {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    color: @blue-2;
    height: 26px;
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
    display: inline-block;
    overflow: scroll;
    white-space: nowrap
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
</style>
