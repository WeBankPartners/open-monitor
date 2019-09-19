<template>
  <div class="main-content">
    <PageTable :pageConfig="pageConfig"></PageTable>
  </div>
</template>
<script>
  let tableEle = [
    {title: 'guid', value: 'guid', display: true, editable: 'editOnline'},
    {title: 'groups', value: 'groups', display: true, frozen: true, sortable: true}
  ]
  const btn = [
    {btn_name: '告警配置', btn_func: 'xx'},
    {btn_name: '历史告警', btn_func: 'xx'},
  ]
  export default {
    name: '',
    data() {
      return {
        pageConfig: {
          CRUD: 'alarm/endpoint/list',
          researchConfig: {
            input_conditions: [
              {value: 'search', type: 'input', placeholder: '请输入', style: ''}],
            btn_group: [
              {btn_name: '搜索', btn_func: 'search', class: 'btn-confirm-f', btn_icon: 'fa fa-search'},
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
            primaryKey: 'guid',
            btn: btn,
            pagination: this.pagination,
            handleFloat:true,
          },
          pagination: { // [通用]-分页组件相关配置
            __orders: '-created_date',
            total: 0,
            page: 1,
            size: 2
          }
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
      }
    },
    components: {
    }
  }
</script>

<style lang="less" scoped>
</style>
