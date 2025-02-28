<template>
  <div class="main-content">
    <div class="content-seatch">
      <Input
        v-model="searchForm.search"
        style="width: 15%"
        clearable
        :placeholder="$t('m_enter_name_tips')"
        @on-change="onFilterConditionChange"
      />
      <Select
        v-model="searchForm.monitor_type"
        multiple
        filterable
        clearable
        style="width: 25%"
        :placeholder="$t('m_please_select') + $t('m_basic_type')"
        @on-change="onFilterConditionChange"
      >
        <Option v-for="name in objectTypeList" :value="name" :key="name">
          {{name}}
        </Option>
      </Select>
      <Button class="add-content-item" @click="onAddButtonClick"  type="success" >{{$t('m_add')}}</Button>
    </div>
    <div class="content-table">
      <Table
        size="small"
        :columns="objectTableColumns"
        :data="objectTableData"
      />
    </div>
    <Page
      class="table-pagination"
      :total="pagination.total"
      @on-change="(e) => {pagination.page = e; this.getTableList()}"
      @on-page-size-change="(e) => {pagination.size = e; this.getTableList()}"
      :current="pagination.page"
      :page-size="pagination.size"
      show-total
      show-sizer
    />

    <ModalComponent :modelConfig="modelConfig"/>

    <Modal
      v-model="isAuthorizationModelShow"
      :title="$t('m_button_authorization')"
      @on-ok="onAuthorizationSave"
      @on-cancel="onAuthorizationConfirmReset"
    >
      <label class="col-md-2 label-name">{{$t('m_field_role')}}:</label>
      <Select v-model="authorizationModel.addRow.role" multiple filterable style="width:338px">
        <Option v-for="item in authorizationModel.roleList" :value="item.id" :key="item.id">
          {{item.display_name}}</Option>
      </Select>
    </Modal>

    <ModalComponent :modelConfig="endpointModel">
      <div slot="endpointOperate">
        <Transfer
          :data="endpointModel.endpointOptions"
          :target-keys="endpointModel.endpoint"
          :titles="endpointModel.titles"
          :list-style="endpointModel.listStyle"
          @on-change="handleChange"
          filterable
        />
      </div>
    </ModalComponent>

  </div>
</template>
<script>
import debounce from 'lodash/debounce'
import isEmpty from 'lodash/isEmpty'
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import {showPoptipOnTable} from '@/assets/js/utils.js'

export default {
  name: '',
  data() {
    return {
      token: null,
      modelTip: {
        key: 'display_name',
        value: null
      },
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: 'm_field_group',
        isAdd: true,
        config: [
          {
            label: 'm_guid',
            value: 'guid',
            placeholder: 'm_tips_inputRequired',
            v_validate: 'required:true|min:2|max:60',
            disabled: false,
            type: 'text'
          },
          {
            label: 'm_tableKey_description',
            value: 'description',
            placeholder: '',
            disabled: false,
            type: 'text'
          },
          {
            label: 'm_tableKey_endpoint_type',
            value: 'monitor_type',
            option: 'monitor_type',
            v_validate: 'required:true',
            disabled: false,
            type: 'select'
          }
          // {
          //   label: 'm_field_resourceLevel',
          //   value: 'service_group',
          //   option: 'service_group',
          //   disabled: false,
          //   type: 'select'
          // }
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          display_name: null,
          guid: '',
          description: null,
          monitor_type: null,
          service_group: null
        },
        v_select_configs: {
          monitor_type: [],
          service_group: [],
        }
      },
      endpointModel: {
        modalId: 'endpoint_Modal',
        modalTitle: 'm_tableKey_endpoint',
        saveFunc: 'managementEndpoint',
        modalStyle: 'min-width:900px',
        isAdd: true,
        config: [
          {
            name: 'endpointOperate',
            type: 'slot'
          }
        ],
        endpoint: [],
        endpointOptions: [],
        titles: [this.$t('m_value_to_be_selected'), this.$t('m_selected_value')],
        listStyle: {
          width: '400px',
          height: '400px'
        }
      },
      authorizationModel: {
        modalId: 'authorization_model',
        modalTitle: 'm_button_authorization',
        isAdd: true,
        config: [
          {
            name: 'authorization',
            type: 'slot'
          }
        ],
        addRow: {
          role: []
        },
        roleList: []
      },
      id: null, // [通用]-待编辑数据id
      searchForm: {
        search: '',
        monitor_type: []
      },
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      pagination: {
        __orders: '-created_date',
        total: 0,
        page: 1,
        size: 10
      },
      objectTableData: [],
      objectTableColumns: [
        {
          title: this.$t('m_tableKey_name'),
          width: 250,
          key: 'display_name'
        },
        {
          title: this.$t('m_tableKey_description'),
          width: 250,
          key: 'description',
          render: (h, params) => (
            params.row.description ? (<span>{params.row.description}</span>) : <div>-</div>
          )
        },
        {
          title: this.$t('m_basic_type'),
          width: 150,
          key: 'type',
          render: (h, params) => (
            <div>
              {
                params.row.monitor_type ? (<TagShow tagName={params.row.monitor_type}></TagShow>) : <div>-</div>
              }
            </div>
          )
        },
        {
          title: this.$t('m_creator'),
          key: 'create_user',
          width: 150,
          render: (h, params) => (
            <span>{params.row.create_user ? params.row.create_user : '-'}</span>
          )
        },
        {
          title: this.$t('m_updatedBy'),
          key: 'update_user',
          width: 150,
          render: (h, params) => (
            <span>{params.row.update_user ? params.row.update_user : '-'}</span>
          )
        },
        {
          title: this.$t('m_update_time'),
          minWidth: 150,
          key: 'update_time'
        },
        {
          title: this.$t('m_table_action'),
          key: 'action',
          fixed: 'right',
          width: 180,
          render: (h, params) => (
            <div style="display: flex">
              <Tooltip placement="top" max-width="400" transfer content={this.$t('m_endpoint')}>
                <Button class="mr-1" size="small" type="success" on-click={() => {
                  this.editEndpoints(params.row)
                }}>
                  <Icon type="ios-cube" />
                </Button>
              </Tooltip>
              <Tooltip placement="top" max-width="400" transfer content={this.$t('m_button_edit')}>
                <Button class="mr-1" size="small" type="primary" on-click={() => {
                  this.editF(params.row)
                }}>
                  <Icon type="md-create" />
                </Button>
              </Tooltip>

              <Tooltip placement="top" max-width="400" transfer content={this.$t('m_permissions')}>
                <Button class="mr-1" size="small" on-click={() => {
                  this.authorizeF(params.row)
                }} type="warning">
                  <Icon type="md-person" />
                </Button>
              </Tooltip>
              <Poptip
                confirm
                transfer
                title={this.$t('m_delConfirm_tip')}
                placement="left-end"
                on-on-ok={() => {
                  this.deleteTableItem(params.row)
                }}>
                <Button class="mr-1" size="small" type="error" on-click={() => {
                  showPoptipOnTable()
                }}>
                  <Icon type="md-trash" />
                </Button>
              </Poptip>
            </div>
          )
        }
      ],
      objectTypeList: [],
      isAuthorizationModelShow: false
    }
  },
  mounted() {
    this.getTableList()
    this.getAllOptions()
    this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
  },
  methods: {
    getAllOptions() {
      const path = '/monitor/api/v2/alarm/endpoint_group/options'
      this.request('GET', path, {}, res => {
        this.objectTypeList = res
      })
    },
    async onAddButtonClick() {
      this.modelConfig.isAdd = true
      const params = {
        page: 1,
        size: 10000,
      }
      await this.getServeGroup()
      await this.request('GET', this.$root.apiCenter.getEndpointType, params, res => {
        this.modelConfig.v_select_configs.monitor_type = res.map(item => ({
          label: item,
          value: item
        }))
      })
      this.$root.JQ('#add_edit_Modal').modal('show')
    },
    async getServeGroup() {
      await this.request('GET', '/monitor/api/v2/service_endpoint/search/group', '', res => {
        this.modelConfig.v_select_configs.service_group = res.map(item => ({
          label: item.display_name,
          value: item.guid
        }))
      })
    },
    addPost() {
      const params= this.$root.$validate.isEmptyReturn_JSON(this.modelConfig.addRow)
      this.request('POST', this.$root.apiCenter.groupManagement.add.api, params, () => {
        this.$root.$validate.emptyJson(this.modelConfig.addRow)
        this.$root.JQ('#add_edit_Modal').modal('hide')
        this.$Message.success(this.$t('m_tips_success'))
        this.getTableList()
      })
    },
    editPost() {
      const params= this.$root.$validate.isEmptyReturn_JSON(this.modelConfig.addRow)
      params.id = this.id
      this.request('PUT', this.$root.apiCenter.groupManagement.update.api, params, () => {
        this.$root.$validate.emptyJson(this.modelConfig.addRow)
        this.$root.JQ('#add_edit_Modal').modal('hide')
        this.$Message.success(this.$t('m_tips_success'))
        this.getTableList()
      })
    },
    async editF(rowData) {
      this.modelConfig.isAdd = false
      this.modelTip.value = rowData[this.modelTip.key]
      this.id = rowData.guid
      const params = {
        page: 1,
        size: 10000,
      }
      await this.getServeGroup()
      await this.request('GET', this.$root.apiCenter.getEndpointType, params, res => {
        this.modelConfig.v_select_configs.monitor_type = res.map(item => ({
          label: item,
          value: item
        }))
      })
      this.modelConfig.addRow = this.$root.$tableUtil.manageEditParams(this.modelConfig.addRow, rowData)
      this.$root.JQ('#add_edit_Modal').modal('show')
    },
    async editEndpoints(rowData) {
      this.id = rowData.guid
      await this.request('GET', `/monitor/api/v2/monitor/endpoint/query?monitorType=${rowData.monitor_type}`, '', res => {
        this.endpointModel.endpointOptions = res.data.map(item => ({
          label: item.guid,
          key: item.guid
        }))
      })
      this.request('GET', `/monitor/api/v2/alarm/endpoint_group/${rowData.guid}/endpoint/list`, '', res => {
        this.endpointModel.endpoint = res.map(item => item.endpoint)
        this.$root.JQ('#endpoint_Modal').modal('show')
      })
    },
    handleChange(newTargetKeys) {
      this.endpointModel.endpoint = newTargetKeys
    },
    managementEndpoint() {
      const params = {
        group_guid: this.id,
        endpoint_guid_list: this.endpointModel.endpoint,
      }
      this.request('POST', '/monitor/api/v2/alarm/endpoint_group/{groupGuid}/endpoint/update', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#endpoint_Modal').modal('hide')
        this.getTableList()
      })
    },
    deleteTableItem(rowData) {
      const api = this.$root.apiCenter.groupManagement.delete.api + '/' + rowData.guid
      this.request('DELETE', api, '', () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getTableList()
      })
    },
    thresholdConfig(rowData) {
      this.$router.push({
        name: 'thresholdManagement',
        params: {
          id: rowData.guid,
          type: 'grp'
        }
      })
    },
    logManagement(rowData) {
      this.$router.push({
        name: 'logManagement',
        params: {
          id: rowData.guid,
          type: 'grp'
        }
      })
    },
    authorizeF(rowData) {
      this.id = rowData.guid
      this.request('GET', this.$root.apiCenter.groupManagement.allRoles.api, '', responseData => {
        this.authorizationModel.roleList = responseData.data
        this.existRole()
      })
    },
    existRole() {
      this.request('GET', this.$root.apiCenter.groupManagement.exitRoles.api, {grp_id: this.id}, responseData => {
        responseData.forEach(item => {
          this.authorizationModel.addRow.role.push(item.id)
        })
        this.isAuthorizationModelShow = true
      })
    },
    onAuthorizationSave() {
      const params = {
        grp_id: this.id,
        role_id: this.authorizationModel.addRow.role
      }
      this.request('POST', this.$root.apiCenter.groupManagement.updateRoles.api, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.authorizationModel.addRow.role = []
        this.isAuthorizationModelShow = false
      })
    },
    onAuthorizationConfirmReset() {
      this.authorizationModel.addRow.role = []
      this.isAuthorizationModelShow = false
    },
    onFilterConditionChange: debounce(function () {
      this.pagination.page = 1
      this.pagination.size = 10
      this.getTableList()
    }, 300),
    getTableList() {
      const params = Object.assign({}, this.searchForm, this.pagination)
      params.monitor_type = params.monitor_type.join(',')
      const path = '/monitor/api/v2/alarm/endpoint_group/query'
      this.request('GET', path, params, res => {
        this.pagination.total = res.num
        this.pagination.page = parseInt(params.page)
        this.objectTableData = isEmpty(res.data) ? [] : res.data
      })
    }
  },
}
</script>

<style lang="less" scoped>
.content-seatch {
  display: flex;
  flex-direction: row;
  justify-content: flex-start;
  align-items: center;
  margin: 15px 0;
  .add-content-item {
    margin-left: auto;
  }
}
.content-seatch > * {
  margin-right: 20px;
}
.table-pagination {
  position: fixed;
  right: 20px;
  bottom: 20px;
}

.main-content {
  .content-table {
    max-height: ~'calc(100vh - 200px)';
    overflow-y: auto;
  }
}

</style>

<style lang='less'>
.ivu-table-wrapper {
  overflow: inherit;
}
</style>
