<template>
  <div class="main-content">
    <PageTable :pageConfig="pageConfig" ref="child">
      <div slot="extraBtn">
        <button type="button" class="btn-cancle-f" @click="exportThreshold">{{$t("button.export")}}</button>
        <div style="display: inline-block;margin-bottom: 3px;vertical-align: bottom;"> 
          <Upload 
          :action="uploadUrl" 
          :show-upload-list="false"
          :max-size="1000"
          :on-success="uploadSucess"
          :on-error="uploadFailed">
            <Button icon="ios-cloud-upload-outline">{{$t('button.upload')}}</Button>
          </Upload>
        </div>
      </div>
    </PageTable>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>
<script>
  import {baseURL_config} from '@/assets/js/baseURL'
  let tableEle = [
    {title: 'tableKey.name', value: 'name', display: true},
    {title: 'tableKey.description', value: 'description', display: true}
  ]
  const btn = [
    {btn_name: 'field.endpoint', btn_func: 'checkMember'},
    {btn_name: 'field.threshold', btn_func: 'thresholdConfig'},
    {btn_name: 'button.edit', btn_func: 'editF'},
    {btn_name: 'button.remove', btn_func: 'delF'},
    {btn_name: 'field.log', btn_func: 'logManagement'}
  ]
  export default {
    name: '',
    data() {
      return {
        uploadUrl: '',
        pageConfig: {
          CRUD: this.$root.apiCenter.groupManagement.list.api,
          researchConfig: {
            input_conditions: [
              {value: 'search', type: 'input', placeholder: 'placeholder.input', style: ''}],
            btn_group: [
              {btn_name: 'button.search', btn_func: 'search', class: 'btn-confirm-f', btn_icon: 'fa fa-search'},
              {btn_name: 'button.add', btn_func: 'add', class: 'btn-cancle-f', btn_icon: 'fa fa-plus'}
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
          key: 'name',
          value: null
        },
        modelConfig: {
          modalId: 'add_edit_Modal',
          modalTitle: 'title.groupAdd',
          isAdd: true,
          config: [
            {label: 'tableKey.name', value: 'name', placeholder: 'tips.inputRequired', v_validate: 'required:true|min:2|max:60', disabled: false, type: 'text'},
            {label: 'tableKey.description', value: 'description', placeholder: '', disabled: false, type: 'text'},
          ],
          addRow: { // [通用]-保存用户新增、编辑时数据
            name: null,
            description: null,
          },
        },
        id: null, // [通用]-待编辑数据id
        selectedData: '', // 存放选中数据
      }
    },
    mounted() {
      this.initData(this.pageConfig.CRUD, this.pageConfig)
      this.uploadUrl =  baseURL_config + this.$root.apiCenter.groupManagement.upload.api
      this.$refs.child.clearSelectedData()
    },
    methods: {
      initData (url= this.pageConfig.CRUD, params) {
        this.$root.$tableUtil.initTable(this, 'GET', url, params)
      },
      add () {
        this.$root.JQ('#add_edit_Modal').modal('show')
      },
      addPost () {
        let params= this.$root.$validate.isEmptyReturn_JSON(this.modelConfig.addRow)
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.groupManagement.add.api, params, () => {
          this.$root.$validate.emptyJson(this.modelConfig.addRow)
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.$Message.success(this.$t('tips.success'))
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      },
      editPost () {
        let params= this.$root.$validate.isEmptyReturn_JSON(this.modelConfig.addRow)
        params.id = this.id
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.groupManagement.update.api, params, () => {
          this.$root.$validate.emptyJson(this.modelConfig.addRow)
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.$Message.success(this.$t('tips.success'))
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      },
      editF (rowData) {
        this.modelConfig.isAdd = false
        this.modelTip.value = rowData[this.modelTip.key]
        this.id = rowData.id
        this.modelConfig.addRow = this.$root.$tableUtil.manageEditParams(this.modelConfig.addRow, rowData)
        this.$root.JQ('#add_edit_Modal').modal('show')
      },
      checkMember (rowData) {
        this.$router.push({name: 'endpointManagement', params: {group: rowData}})
      },
      delF (rowData) {
        this.$parent.$parent.delConfirm({name: rowData.name}, () => {
          let params = {id: rowData.id}
          this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.groupManagement.delete.api, params, () => {
            this.$Message.success(this.$t('tips.success'))
            this.initData(this.pageConfig.CRUD, this.pageConfig)
          })
        })
      },
      thresholdConfig (rowData) {
        this.$router.push({name: 'thresholdManagement', params: {id: rowData.id, type: 'grp'}})
      },
      logManagement (rowData) {
        this.$router.push({name: 'logManagement', params: {id: rowData.id, type: 'grp'}})
      },
      exportThreshold () {
        if (!this.$validate.isEmpty(this.selectedData.checkedIds)) {
          this.$Message.warning(this.$t('tips.selectData'))
          return
        }
        window.location.href= baseURL_config + this.$root.apiCenter.groupManagement.export.api + '?id=' + this.selectedData.checkedIds.join(',')
      },
      uploadSucess () {
        this.$Message.success(this.$t('tips.success'))
        this.initData(this.pageConfig.CRUD, this.pageConfig)
      },
      uploadFailed () {
        this.$Message.warning(this.$t('tips.failed'))
      }
    },
    components: {
    }
  }
</script>

<style lang="less" scoped>
</style>
