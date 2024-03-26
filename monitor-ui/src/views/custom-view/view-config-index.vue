<template>
  <div class="">
    <Title :title="$t('menu.customViews')"></Title>
    <div class="operational-zone">
      <button class="btn btn-sm btn-confirm-f" @click="addView">
        <i class="fa fa-plus"></i>{{$t('button.addViewTemplate')}}
      </button>
      <button class="btn btn-sm btn-cancel-f" @click="setDashboard">
        {{$t('button.setDashboard')}}
      </button>
    </div>
    <section>
        <template v-for="(panalItem,panalIndex) in dataList">
          <div :key="panalIndex" class="panal-list">
            <Card class="c-dark">
              <p slot="title" class="panal-title">
                <span>{{ $t('tableKey.name') }}:{{ panalItem.name }}</span>
                <span style="float: right;">{{ $t('field.id') }}:{{ panalItem.id }}</span>
              </p>
              <ul class="panal-content">
                <li>
                  <Button class="btn-cancel-f" style="height:24px" size="small" @click="goToPanal(panalItem, 'view')">{{$t('button.view')}}</Button>
                  <template v-if="panalItem.permission === 'mgmt'">
                    <Button type="primary" class="btn-confirm-f" style="height:24px" size="small" @click="goToPanal(panalItem, 'edit')">{{$t('m_configuration')}}</Button>
                    <Button class="btn-cancel-f" style="height:24px" size="small" @click="authorization(panalItem)">{{$t('button.authorization')}}</Button>
                    <Button class="btn-cancel-f" style="height:24px" size="small" @click="deleteConfirmModal(panalItem)">{{$t('button.remove')}}</Button>
                  </template>
                </li>
                <Divider />
                
                
                <li style="margin-top:8px">
                  <i class="fa fa-star" style="margin-right:16px;" v-if="panalItem.main === 1" aria-hidden="true"></i>
                  <template v-for="(role,roleIndex) in panalItem.main_page">
                    <Tag color="blue" :key="roleIndex">{{role}}</Tag>
                  </template>
                  <Tag style="visibility: hidden;"></Tag>
                </li>
              </ul>
            </Card>
          </div>
        </template>
      <!-- </ul> -->
    </section>
    <!-- <ModalComponent :modelConfig="authorizationModel">
      <div slot="authorization">  
        <div>
          <label class="col-md-2 label-name">{{$t('field.role')}}:</label>
          <Select v-model="authorizationModel.addRow.role" multiple filterable style="width:338px">
              <Option v-for="item in authorizationModel.roleList" :value="item.id" :key="item.id">
              {{item.display_name}}</Option>
          </Select>
        </div>
      </div>
    </ModalComponent> -->

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
                <Option v-for="item in authorizationModel.roleList" :value="item.id" :key="item.id">
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

    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
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
      @on-ok="ok"
      @on-cancel="cancel">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('delConfirm.tip')}}</p>
        </div>
      </div>
    </Modal>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
      dataList: [
      ],
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: 'title.viewTemplate',
        isAdd: true,
        config: [
          {label: 'tableKey.name', value: 'name', placeholder: 'tips.required', v_validate: 'required:true|min:2|max:60', disabled: false, type: 'text'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          name: null,
        },
      },
      setDashboardModel: {
        modalId: 'set_dashboard_modal',
        modalTitle: 'button.setDashboard',
        isAdd: true,
        saveFunc: 'setDashboardSave',
        config: [
          {name:'setDashboard',type:'slot'}
        ],
        addRow: {
          templateSelect: null
        },
        templateList: []
      },
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
      // permissionModel: {
      //   modalId: 'maintenance_window_model',
      //   modalTitle: 'm_button_maintenanceWindow',
      //   isAdd: true,
      //   saveFunc: 'maintenanceWindowSave',
      //   config: [{
      //     name: 'maintenanceWindow',
      //     type: 'slot'
      //   }],
      //   addRow: {
      //     // businessSet: [],
      //   },
      //   result: []
      // },
    }
  },
  mounted(){
    this.viewList()
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
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST','/monitor/api/v1/dashboard/custom/main/set', params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#set_dashboard_modal').modal('hide')
      })
    },
    authorization (panalItem) {
      this.dashboard_id = panalItem.id
      const params = {
        dashboard_id: panalItem.id
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET','/monitor/api/v1/dashboard/custom/role/get', params, (res) => {
        this.authorizationModel.result = res
        this.$root.JQ('#authorization_model').modal('show')
      })
    },
    authorizationSave () {
      let params = {
        dashboard_id: this.dashboard_id,
        permission_list: this.authorizationModel.result
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/dashboard/custom/role/save', params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#authorization_model').modal('hide')
      })
    },
    getAllRoles () {
      const params = {
        page: 1,
        size: 1000
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET','/monitor/api/v1/user/role/list?page=1&size=1000', params, (responseData) => {
        this.authorizationModel.roleList = responseData.data
      })
    },
    addPost () {
      this.$root.JQ('#add_edit_Modal').modal('hide')
      let params = {
        name: this.modelConfig.addRow.name,
        cfg: '[]'
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.template.save, params, () => {
        this.viewList()
      })
    },
    addView () {
      this.modelConfig.isAdd = true
      this.$root.JQ('#add_edit_Modal').modal('show')
    },
    deleteConfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarning = true
    },
    ok () {
      this.removeTemplate(this.selectedData)
    },
    cancel () {
      this.isShowWarning = false
    },
    deleteConfirm (item) {
      this.$delConfirm({
        msg: item.name,
        callback: () => {
          this.removeTemplate(item)
        }
      })
    },
    removeTemplate (item) {
      let params = {id: item.id}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.template.delete, params, () => {
        // this.$root.$eventBus.$emit('hideConfirmModal')
        this.$Message.success(this.$t('tips.success'))
        this.viewList()
      })
    },
    viewList () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.template.list, '', responseData => {
        this.setDashboardModel.templateList = []
        this.setDashboardModel.addRow.templateSelect = null
        this.dataList = responseData
        responseData.forEach((item) => {
          this.setDashboardModel.templateList.push({
            label: item.name,
            value: item.id
          })
          if (item.main === 1) {
            this.setDashboardModel.addRow.templateSelect = item.id
          }
        })
      })
    },
    setDashboard () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.template.portalConfig, '', responseData => {
        this.processConfigModel.dashboardConfig = responseData
        this.$root.JQ('#set_dashboard_modal').modal('show')
      })
    },
    setDashboardSave () {
      let params = {id: this.setDashboardModel.addRow.templateSelect}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.template.templateSet, params, () => {
        this.$root.JQ('#set_dashboard_modal').modal('hide')
        this.$Message.success(this.$t('tips.success'))
        this.viewList()
      })
    },
    goToPanal(panalItem, type) {
      const params = {
        permission: type,
        panalItem: panalItem
      }
      this.$router.push({name:'viewConfig',params:params})
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
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
.panal-list {
  margin: 8px;
  width: 390px;
  display: inline-block;
}
.panal-title {
  color: @blue-2;
  height: 26px;
}
.panal-content {
  font-size: 12px;
}
.fa-star {
  color: @color-orange-F;
}
</style>
