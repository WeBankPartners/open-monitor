<template>
  <div class="main-content">
    <PageTable :pageConfig="pageConfig"></PageTable>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>
<script>
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
        pageConfig: {
          CRUD: this.$root.apiCenter.groupManagement.list.api,
          researchConfig: {
            input_conditions: [
              {value: 'search', type: 'input', placeholder: 'placeholder.input', style: ''}],
            btn_group: [
              {btn_name: 'button.search', btn_func: 'search', class: 'btn-confirm-f', btn_icon: 'fa fa-search'},
              {btn_name: 'button.add', btn_func: 'add', class: 'btn-cancle-f', btn_icon: 'fa fa-plus'},
            ],
            filters: {
              name__icontains: '',
              cmdb_tenant_id__icontains: ''
            }
          },
          table: {
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
      }
    },
    mounted() {
      this.initData(this.pageConfig.CRUD, this.pageConfig)
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
    },
    components: {
    }
  }
</script>

<style lang="less" scoped>
</style>
