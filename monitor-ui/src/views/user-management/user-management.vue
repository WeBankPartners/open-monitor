<template>
  <div class=" ">
    <PageTable :pageConfig="pageConfig"></PageTable>
  </div>
</template>

<script>
let tableEle = [
  {title: 'tableKey.name', value: 'name', display: true},
  {title: 'tableKey.nickname', value: 'display_name', display: true},
  {title: 'tableKey.email', value: 'email', display: true},
  {title: 'tableKey.phone', value: 'phone', display: true}, 
  {title: 'tableKey.role', value: 'role', display: true},
  {title: 'tableKey.activeDate', value: 'created_string', display: true}
]
export default {
  name: '',
  data() {
    return {
      pageConfig: {
        CRUD: this.$root.apiCenter.setup.userManagement.get,
        researchConfig: {
          input_conditions: [
            {value: 'search', type: 'input', placeholder: 'placeholder.input', style: ''}],
          btn_group: [
            {btn_name: 'button.search', btn_func: 'search', class: 'btn-confirm-f', btn_icon: 'fa fa-search'}
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
          btn: '',
          pagination: this.pagination,
          handleFloat:true,
        },
        pagination: {
          __orders: '',
          total: 0,
          page: 1,
          size: 10
        }
      }
    }
  },
  mounted () {
    this.initData(this.pageConfig.CRUD, this.pageConfig)
  },
  methods: {
    initData (url= this.pageConfig.CRUD, params) {
      this.$root.$tableUtil.initTable(this, 'GET', url, params)
    },
  },
  components: {},
}
</script>

<style scoped lang="less">
</style>
