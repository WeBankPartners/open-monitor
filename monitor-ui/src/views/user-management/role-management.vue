<template>
  <div class=" ">
    <PageTable :pageConfig="pageConfig"></PageTable>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
    <ModalComponent :modelConfig="authorizationModel">
      <div slot="authorization">
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('m_field_endpoint')}}:</label>
          <Select v-model="authorizationModel.addRow.user" multiple filterable style="width:338px">
            <Option v-for="item in authorizationModel.userList" :value="item.id" :key="item.name">
              {{item.display_name}}({{item.name}})</Option>
          </Select>
        </div>
      </div>
    </ModalComponent>
    <Modal
      v-model="isShowWarning"
      :title="$t('m_delConfirm_title')"
      @on-ok="ok"
      @on-cancel="cancel"
    >
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('m_delConfirm_tip')}}</p>
        </div>
      </div>
    </Modal>
  </div>
</template>

<script>
const tableEle = [
  {
    title: 'm_tableKey_name',
    value: 'name',
    display: true
  },
  {
    title: 'm_tableKey_nickname',
    value: 'display_name',
    display: true
  },
  {
    title: 'm_tableKey_email',
    value: 'email',
    display: true
  },
  {
    title: 'm_tableKey_activeDate',
    value: 'created_string',
    display: true
  }
]
const btn = [
  {
    btn_name: 'm_button_edit',
    btn_func: 'editF'
  },
  {
    btn_name: 'm_button_authorization',
    btn_func: 'authorizationF'
  },
  {
    btn_name: 'm_button_remove',
    btn_func: 'deleteConfirmModal',
    color: 'red'
  }
]

export const custom_api_enum = [
  {
    'setup.role.get': 'get'
  }
]

export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
      selectedData: null,
      pageConfig: {
        CRUD: this.$root.apiCenter.setup.role.get,
        researchConfig: {
          input_conditions: [
            {
              value: 'search',
              type: 'input',
              placeholder: 'm_placeholder_input',
              style: ''
            }],
          btn_group: [
            {
              btn_name: 'm_button_search',
              btn_func: 'search',
              class: 'btn-confirm-f',
              btn_icon: 'fa fa-search'
            },
            {
              btn_name: 'm_button_add',
              btn_func: 'addRole',
              class: 'btn-cancel-f',
              btn_icon: 'fa fa-plus'
            }
          ],
          filters: {
            search: ''
          }
        },
        table: {
          tableData: [],
          tableEle,
          // filterMoreBtn: 'filterMoreBtn',
          // primaryKey: 'guid',
          btn,
          pagination: this.pagination,
          handleFloat: true,
        },
        pagination: {
          __orders: '',
          total: 0,
          page: 1,
          size: 10
        }
      },
      modelTip: {
        key: 'name',
        value: null
      },
      modelConfig: {
        modalId: 'add_role_Modal',
        modalTitle: 'm_button_add',
        isAdd: true,
        config: [
          {
            label: 'm_tableKey_name',
            value: 'name',
            placeholder: 'm_tips_required',
            v_validate: 'required:true',
            disabled: false,
            type: 'text'
          },
          {
            label: 'm_tableKey_nickname',
            value: 'display_name',
            placeholder: 'm_tips_required',
            v_validate: 'required:true',
            disabled: false,
            type: 'text'
          },
          {
            label: 'm_tableKey_email',
            value: 'email',
            placeholder: 'm_tips_required',
            v_validate: 'required:true|noEmail',
            disabled: false,
            type: 'text'
          }
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          name: null,
          display_name: null,
          email: null
        }
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
          user: []
        },
        userList: []
      },
      id: null,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  mounted() {
    this.initData(this.pageConfig.CRUD, this.pageConfig)
  },
  methods: {
    initData(url= this.pageConfig.CRUD, params) {
      this.$root.$tableUtil.initTable(this, 'GET', url, params)
    },
    addRole() {
      this.modelConfig.isAdd = true
      this.$root.JQ('#add_role_Modal').modal('show')
    },
    addPost() {
      const params = JSON.parse(JSON.stringify(this.modelConfig.addRow))
      params.operation = 'add'
      this.request('POST', this.apiCenter.setup.role.update, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#add_role_Modal').modal('hide')
        this.initData(this.pageConfig.CRUD, this.pageConfig)
      })
    },
    editF(rowData) {
      this.modelConfig.isAdd = false
      this.modelTip.value = rowData[this.modelTip.key]
      this.id = rowData.id
      this.modelConfig.addRow = this.$root.$tableUtil.manageEditParams(this.modelConfig.addRow, rowData)
      this.$root.JQ('#add_role_Modal').modal('show')
    },
    editPost() {
      const params = JSON.parse(JSON.stringify(this.modelConfig.addRow))
      params.operation = 'update'
      params.role_id = this.id
      this.request('POST', this.apiCenter.setup.role.update, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#add_role_Modal').modal('hide')
        this.initData(this.pageConfig.CRUD, this.pageConfig)
      })
    },
    deleteConfirmModal(rowData) {
      this.selectedData = rowData
      this.isShowWarning = true
    },
    ok() {
      this.delF(this.selectedData)
    },
    cancel() {
      this.isShowWarning = false
    },
    deleteConfirm(rowData) {
      this.$delConfirm({
        msg: rowData.name,
        callback: () => {
          this.delF(rowData)
        }
      })
    },
    delF(rowData) {
      const params = {
        role_id: rowData.id,
        operation: 'delete'
      }
      this.request('POST', this.apiCenter.setup.role.update, params, () => {
        // this.$root.$eventBus.$emit('hideConfirmModal')
        this.$Message.success(this.$t('m_tips_success'))
        this.initData(this.pageConfig.CRUD, this.pageConfig)
      })
    },
    authorizationF(rowData) {
      this.id = rowData.id
      this.request('GET', this.apiCenter.setup.userManagement.get, '', responseData => {
        this.authorizationModel.userList = responseData.data
        this.existUser()
      })
    },
    existUser() {
      this.request('GET', this.apiCenter.setup.userManagement.get, {role: this.id}, responseData => {
        responseData.data.forEach(item => {
          this.authorizationModel.addRow.user.push(item.id)
        })
        this.$root.JQ('#authorization_model').modal('show')
      })
    },
    authorizationSave() {
      const params = {
        role_id: this.id,
        user_id: this.authorizationModel.addRow.user
      }
      this.request('POST', this.apiCenter.setup.role.authorization, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#authorization_model').modal('hide')
      })
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
</style>
