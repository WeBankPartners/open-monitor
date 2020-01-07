<template>
  <div class=" ">
    <PageTable :pageConfig="pageConfig"></PageTable>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>

<script>
let tableEle = [
  {title: 'tableKey.name', value: 'name', display: true},
  {title: 'tableKey.nickname', value: 'display_name', display: true},
  {title: 'tableKey.email', value: 'email', display: true},
  {title: 'tableKey.activeDate', value: 'created', display: true}
]
const btn = [
    {btn_name: 'button.edit', btn_func: 'editF'},
    {btn_name: 'button.remove', btn_func: 'delF'}
  ]
export default {
  name: '',
  data() {
    return {
      pageConfig: {
        CRUD: this.$root.apiCenter.setup.role.get,
        researchConfig: {
          input_conditions: [
            {value: 'search', type: 'input', placeholder: 'placeholder.input', style: ''}],
          btn_group: [
            {btn_name: 'button.search', btn_func: 'search', class: 'btn-confirm-f', btn_icon: 'fa fa-search'},
            {btn_name: 'button.add', btn_func: 'addRole', class: 'btn-cancle-f', btn_icon: 'fa fa-plus'}
          ],
          filters: {
            search: ''
          }
        },
        table: {
          tableData: [],
          tableEle: tableEle,
          // filterMoreBtn: 'filterMoreBtn',
          // primaryKey: 'guid',
          btn: btn,
          pagination: this.pagination,
          handleFloat:true,
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
        modalTitle: 'button.add',
        isAdd: true,
        config: [
          {label: 'tableKey.name', value: 'name', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'},
          {label: 'tableKey.nickname', value: 'display_name', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'},
          {label: 'tableKey.email', value: 'email', placeholder: 'tips.required', v_validate: 'required:true|noEmail', disabled: false, type: 'text'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          name: null,
          display_name: null,
          email: null
        }
      },
      id: null
    }
  },
  mounted () {
    this.initData(this.pageConfig.CRUD, this.pageConfig)
  },
  methods: {
    initData (url= this.pageConfig.CRUD, params) {
      this.$root.$tableUtil.initTable(this, 'GET', url, params)
    },
    addRole () {
      this.modelConfig.isAdd = true
      this.$root.JQ('#add_role_Modal').modal('show')
    },
    addPost () {
      let params = JSON.parse(JSON.stringify(this.modelConfig.addRow))
      params.operation = 'add'
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.setup.role.update, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#add_role_Modal').modal('hide')
        this.initData(this.pageConfig.CRUD, this.pageConfig)
      })
    },
    editF (rowData) {
      this.modelConfig.isAdd = false
      this.modelTip.value = rowData[this.modelTip.key]
      this.id = rowData.id
      this.modelConfig.addRow = this.$root.$tableUtil.manageEditParams(this.modelConfig.addRow, rowData)
      this.$root.JQ('#add_role_Modal').modal('show')
    },
    editPost () {
      let params = JSON.parse(JSON.stringify(this.modelConfig.addRow))
      params.operation = 'update'
      params.role_id = this.id
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.setup.role.update, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#add_role_Modal').modal('hide')
        this.initData(this.pageConfig.CRUD, this.pageConfig)
      })
    },
    delF (rowData) {
      this.$parent.$parent.delConfirm({name: rowData.name}, () => {
        let params = {role_id: rowData.id, operation: 'delete' }
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.setup.role.update, params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      })
    },
  },
  components: {},
}
</script>

<style scoped lang="less">
</style>
