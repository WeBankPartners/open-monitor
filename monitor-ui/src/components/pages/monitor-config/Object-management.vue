<!-- 租户管理 -->
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
          titleConfig: {
            name: '租户管理',
            backUrl: ''
          },
          researchConfig: {
            input_conditions: [
              {value: 'name__icontains', type: 'input', placeholder: '名称', style: ''},
              {value: 'cmdb_tenant_id__icontains', type: 'input', placeholder: 'CMDB关联ID', style: ''}],
            btn_group: [
              {btn_name: '搜索', btn_func: 'search', class: 'btn-confirm-f', btn_icon: 'fa fa-search'},
              {btn_name: '新增', btn_func: 'add', class: 'btn-cancle-f', btn_icon: 'fa fa-plus'},
            ],
            filters: { // [通用]-搜索条件
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
