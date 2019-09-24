<template>
  <div class="main-content">
    <PageTable :pageConfig="pageConfig"></PageTable>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>
<script>
  let tableEle = [
    {title: '名称', value: 'name', display: true},
    {title: '描述', value: 'description', display: true}
  ]
  const btn = [
    {btn_name: '成员', btn_func: 'checkMember'},
    {btn_name: '阀值配置', btn_func: 'thresholdConfig'},
    {btn_name: '编辑', btn_func: 'editF'},
    {btn_name: '删除', btn_func: 'delF'},
  ]
  export default {
    name: '',
    data() {
      return {
        pageConfig: {
          CRUD: 'alarm/grp/list',
          researchConfig: {
            input_conditions: [
              {value: 'search', type: 'input', placeholder: '请输入', style: ''}],
            btn_group: [
              {btn_name: '搜索', btn_func: 'search', class: 'btn-confirm-f', btn_icon: 'fa fa-search'},
              {btn_name: '新增', btn_func: 'add', class: 'btn-cancle-f', btn_icon: 'fa fa-plus'},
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
          modalTitle: '组管理',
          isAdd: true,
          config: [
            {label: '名称', value: 'name', placeholder: '必填,2-60字符', v_validate: 'required:true|min:2|max:60', disabled: true, hide: 'edit', type: 'text'},
            {label: '备注描述', value: 'description', placeholder: '', disabled: false, type: 'text'},
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
        this.$tableUtil.initTable(this, 'GET', url, params)
      },
      add () {
        this.JQ('#add_edit_Modal').modal('show')
      },
      addPost () {
        let params= this.$validate.isEmptyReturn_JSON(this.modelConfig.addRow)
        this.$httpRequestEntrance.httpRequestEntrance('POST', 'alarm/grp/add', params, () => {
          this.$validate.emptyJson(this.modelConfig.addRow)
          this.JQ('#add_edit_Modal').modal('hide')
          this.$Message.success('新增成功 !')
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      },
      editPost () {
        let params= this.$validate.isEmptyReturn_JSON(this.modelConfig.addRow)
        params.id = this.id
        this.$httpRequestEntrance.httpRequestEntrance('POST', 'alarm/grp/update', params, () => {
          this.$validate.emptyJson(this.modelConfig.addRow)
          this.JQ('#add_edit_Modal').modal('hide')
          this.$Message.success('修改成功 !')
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      },
      editF (rowData) {
        this.modelConfig.isAdd = false
        this.modelTip.value = rowData[this.modelTip.key]
        this.id = rowData.id
        this.JQ('#add_edit_Modal').modal('show')
      },
      checkMember (rowData) {
        this.$router.push({name: 'objectManagement', params: {group: rowData}})
      },
      delF (rowData) {
        let params = {id: rowData.id}
        this.$httpRequestEntrance.httpRequestEntrance('GET', 'alarm/grp/delete', params, () => {
          this.$Message.success('删除成功 !')
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      },
      thresholdConfig (rowData) {
        this.$router.push({name: 'thresholdManagement', params: {id: rowData.id, type: 'grp'}})
      },
    },
    components: {
    }
  }
</script>

<style lang="less" scoped>
</style>
