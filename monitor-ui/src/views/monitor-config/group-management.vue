<template>
  <div class="main-content">
    <PageTable :pageConfig="pageConfig" ref="child">
      <!-- <div slot="extraBtn">
        <button type="button" class="btn-cancel-f" @click="exportThreshold">{{$t("m_button_export")}}</button>
        <div style="display: inline-block;margin-bottom: 3px;vertical-align: bottom;"> 
          <Upload 
          :action="uploadUrl" 
          :show-upload-list="false"
          :max-size="1000"
          with-credentials
          :headers="{'Authorization': token}"
          :on-success="uploadSucess"
          :on-error="uploadFailed">
            <Button icon="ios-cloud-upload-outline">{{$t('m_button_upload')}}</Button>
          </Upload>
        </div>
      </div> -->
    </PageTable>
    <ModalComponent :modelConfig="modelConfig">
    </ModalComponent>
    <ModalComponent :modelConfig="authorizationModel">
      <div slot="authorization">  
        <div>
          <label class="col-md-2 label-name">{{$t('m_field_role')}}:</label>
          <Select v-model="authorizationModel.addRow.role" multiple filterable style="width:338px">
              <Option v-for="item in authorizationModel.roleList" :value="item.id" :key="item.id">
              {{item.display_name}}</Option>
          </Select>
        </div>
      </div>
    </ModalComponent>
    <ModalComponent :modelConfig="endpointModel">
      <div slot="endpointOperate">  
        <Transfer
        :data="endpointModel.endpointOptions"
        :target-keys="endpointModel.endpoint"
        :titles="endpointModel.titles"
        :list-style="endpointModel.listStyle"
        @on-change="handleChange"
        filterable>
        </Transfer>
      </div>
    </ModalComponent>
    <Modal
        v-model="isShowWarning"
        :title="$t('m_delConfirm_title')"
        @on-ok="ok"
        @on-cancel="cancel">
        <div class="modal-body" style="padding:30px">
          <div style="text-align:center">
            <p style="color: red">{{$t('m_delConfirm_tip')}}</p>
          </div>
        </div>
    </Modal>
  </div>
</template>
<script>
  import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
  import {baseURL_config} from '@/assets/js/baseURL'
  let tableEle = [
    {title: 'm_tableKey_name', value: 'display_name', display: true},
    {title: 'm_tableKey_description', value: 'description', display: true},
    {title: 'm_tableKey_endpoint_type', value: 'monitor_type', display: true}
  ]
  const btn = [
    {btn_name: 'm_field_endpoint', btn_func: 'editEndpoints'},
    // {btn_name: 'm_field_threshold', btn_func: 'thresholdConfig'},
    {btn_name: 'm_button_edit', btn_func: 'editF'},
    {btn_name: 'm_button_remove', btn_func: 'deleteConfirmModal', color: 'red'},
    // {btn_name: 'm_field_log', btn_func: 'logManagement'},
    {btn_name: 'm_button_authorize', btn_func: 'authorizeF'}
  ]
  import axios from 'axios'
  export default {
    name: '',
    data() {
      return {
        isShowWarning: false,
        token: null,
        uploadUrl: '',
        pageConfig: {
          CRUD: '',
          researchConfig: {
            input_conditions: [
              {value: 'search', type: 'input', placeholder: 'm_placeholder_input', style: ''}],
            btn_group: [
              {btn_name: 'm_button_search', btn_func: 'search', class: 'btn-confirm-f', btn_icon: 'fa fa-search'},
              {btn_name: 'm_button_add', btn_func: 'add', class: 'btn-cancel-f', btn_icon: 'fa fa-plus'}
            ],
            filters: {
              name__icontains: '',
              cmdb_tenant_id__icontains: ''
            }
          },
          table: {
            selection: true,
            tableData: [],
            tableEle: tableEle,
            // filterMoreBtn: 'filterMoreBtn',
            primaryKey: 'id',
            btn: btn,
            pagination: this.pagination,
            handleFloat:true,
          },
          pagination: { // [通用]-分页组件相关配置
            __orders: '-created_date',
            total: 0,
            page: 1,
            size: 10
          }
        },
        modelTip: {
          key: 'display_name',
          value: null
        },
        modelConfig: {
          modalId: 'add_edit_Modal',
          modalTitle: 'm_field_group',
          isAdd: true,
          config: [
            {label: 'm_guid', value: 'guid', placeholder: 'm_tips_inputRequired', v_validate: 'required:true|min:2|max:60', disabled: false, type: 'text'},
            {label: 'm_tableKey_description', value: 'description', placeholder: '', disabled: false, type: 'text'},
            //{name: 'slotEndpointType', type:'slot'}
            {label: 'm_tableKey_endpoint_type', value: 'monitor_type', option: 'monitor_type',v_validate: 'required:true', disabled: false, type: 'select'},
            {label: 'm_field_resourceLevel', value: 'service_group', option: 'service_group', disabled: false, type: 'select'}
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
            {name:'endpointOperate',type:'slot'}
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
          saveFunc: 'authorizationSave',
          config: [
            {name:'authorization',type:'slot'}
          ],
          addRow: {
            role: []
          },
          roleList: []
        }, 
        id: null, // [通用]-待编辑数据id
        selectedData: '', // 存放选中数据
      }
    },
    mounted() {
      this.pageConfig.CRUD = this.$root.apiCenter.groupManagement.list.api
      this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
      this.initData(this.pageConfig.CRUD, this.pageConfig)
      this.uploadUrl =  baseURL_config + this.$root.apiCenter.groupManagement.upload.api
      this.$refs.child.clearSelectedData()
    },
    methods: {
      initData (url= this.pageConfig.CRUD, params) {
        this.$root.$tableUtil.initTable(this, 'GET', url, params)
      },
      async add () {
        this.modelConfig.isAdd = true
        let params = {
          page: 1,
          size: 10000,
        }
        await this.getServeGroup()
        await this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getEndpointType, params, res => {
          this.modelConfig.v_select_configs.monitor_type = res.map(item => {
            return {
              label: item,
              value: item
            }
          })
        })
        this.$root.JQ('#add_edit_Modal').modal('show')
      },
      async getServeGroup () {
        await this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v2/service_endpoint/search/group', '', res => {
          this.modelConfig.v_select_configs.service_group = res.map(item => {
            return {
              label: item.display_name,
              value: item.guid
            }
          })
        })
      },
      addPost () {
        let params= this.$root.$validate.isEmptyReturn_JSON(this.modelConfig.addRow)
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.groupManagement.add.api, params, () => {
          this.$root.$validate.emptyJson(this.modelConfig.addRow)
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.$Message.success(this.$t('m_tips_success'))
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      },
      editPost () {
        let params= this.$root.$validate.isEmptyReturn_JSON(this.modelConfig.addRow)
        params.id = this.id
        this.$root.$httpRequestEntrance.httpRequestEntrance('PUT', this.$root.apiCenter.groupManagement.update.api, params, () => {
          this.$root.$validate.emptyJson(this.modelConfig.addRow)
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.$Message.success(this.$t('m_tips_success'))
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      },
      async editF (rowData) {
        this.modelConfig.isAdd = false
        this.modelTip.value = rowData[this.modelTip.key]
        this.id = rowData.guid
        let params = {
          page: 1,
          size: 10000,
        }
        await this.getServeGroup()
        await this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getEndpointType, params, res => {
          this.modelConfig.v_select_configs.monitor_type = res.map(item => {
            return {
              label: item,
              value: item
            }
          })
        })
        this.modelConfig.addRow = this.$root.$tableUtil.manageEditParams(this.modelConfig.addRow, rowData)
        this.$root.JQ('#add_edit_Modal').modal('show')
      },
      async editEndpoints (rowData) {
        this.id = rowData.guid
        await this.$root.$httpRequestEntrance.httpRequestEntrance('GET', `/monitor/api/v2/monitor/endpoint/query?monitorType=${rowData.monitor_type}`, '', res => {
          this.endpointModel.endpointOptions = res.data.map(item => {
            return {
              label: item.guid,
              key: item.guid
            }
          })
        })
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET', `/monitor/api/v2/alarm/endpoint_group/${rowData.guid}/endpoint/list`, '', res => {
          this.endpointModel.endpoint = res.map(item => item.endpoint)
          this.$root.JQ('#endpoint_Modal').modal('show')
        })
      },
      handleChange (newTargetKeys) {
        this.endpointModel.endpoint = newTargetKeys
      },
      managementEndpoint() {
        let params = {
          group_guid: this.id,
          endpoint_guid_list: this.endpointModel.endpoint,
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v2/alarm/endpoint_group/{groupGuid}/endpoint/update', params, () => {
          this.$Message.success(this.$t('m_tips_success'))
          this.$root.JQ('#endpoint_Modal').modal('hide')
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      },
      deleteConfirmModal (rowData) {
        this.selectedData = rowData
        this.isShowWarning = true
      },
      ok () {
        this.delF(this.selectedData)
      },
      cancel () {
        this.isShowWarning = false
      },
      deleteConfirm (rowData) {
        this.$delConfirm({
          msg: rowData.display_name,
          callback: () => {
            this.delF(rowData)
          }
        })
      },
      delF (rowData) {
        const api = this.$root.apiCenter.groupManagement.delete.api + '/' + rowData.guid
        this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
          // this.$root.$eventBus.$emit('hideConfirmModal')
          this.$Message.success(this.$t('m_tips_success'))
          this.initData(this.pageConfig.CRUD, this.pageConfig)
          this.isShowWarning = false
        })
      },
      thresholdConfig (rowData) {
        this.$router.push({name: 'thresholdManagement', params: {id: rowData.guid, type: 'grp'}})
      },
      logManagement (rowData) {
        this.$router.push({name: 'logManagement', params: {id: rowData.guid, type: 'grp'}})
      },
      exportThreshold () {
        if (!this.$validate.isEmpty(this.selectedData.checkedIds)) {
          this.$Message.warning(this.$t('m_tips_selectData'))
          return
        }
        const api = this.$root.apiCenter.groupManagement.export.api + '?id=' + this.selectedData.checkedIds.join(',')
        axios({
          method: 'GET',
          url: api,
          headers: {
            'Authorization': this.token
          }
        }).then((response) => {
          if (response.status < 400) {
           let content = JSON.stringify(response.data)
          let fileName = `grp_strategy_tpl_${new Date().format('yyyyMMddhhmmss')}.json`
          let blob = new Blob([content])
          if('msSaveOrOpenBlob' in navigator){
            // Microsoft Edge and Microsoft Internet Explorer 10-11
            window.navigator.msSaveOrOpenBlob(blob, fileName)
          } else {
            if ('download' in document.createElement('a')) { // 非IE下载
              let elink = document.createElement('a')
              elink.download = fileName
              elink.style.display = 'none'
              elink.href = URL.createObjectURL(blob)  
              document.body.appendChild(elink)
              elink.click()
              URL.revokeObjectURL(elink.href) // 释放URL 对象
              document.body.removeChild(elink)
            } else { // IE10+下载
              navigator.msSaveOrOpenBlob(blob, fileName)
            }
          }
          }
        })
        .catch(() => {
          this.$Message.warning(this.$t('m_tips_failed'))
        });

      },
      uploadSucess () {
        this.$Message.success(this.$t('m_tips_success'))
        this.initData(this.pageConfig.CRUD, this.pageConfig)
      },
      uploadFailed () {
        this.$Message.warning(this.$t('m_tips_failed'))
      },
      authorizeF (rowData) {
        this.id = rowData.guid
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.groupManagement.allRoles.api, '', (responseData) => {
          this.authorizationModel.roleList = responseData.data
          this.existRole()
        })
      },
      existRole () {
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.groupManagement.exitRoles.api, {grp_id: this.id}, (responseData) => {
          responseData.forEach((item) => {
            this.authorizationModel.addRow.role.push(item.id)
          })
          this.$root.JQ('#authorization_model').modal('show')
        })
      },
      authorizationSave () {
        let params = {
          grp_id: this.id,
          role_id: this.authorizationModel.addRow.role
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.groupManagement.updateRoles.api, params, () => {
          this.$Message.success(this.$t('m_tips_success'))
          this.$root.JQ('#authorization_model').modal('hide')
        })
      } 
    },
    components: {
    }
  }
</script>

<style lang="less" scoped>
</style>
