<template>
  <div class=" ">
    <PageTable :pageConfig="pageConfig"></PageTable>
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
    title: 'm_tableKey_phone',
    value: 'phone',
    display: true
  },
  {
    title: 'm_tableKey_role',
    display: true,
    tags: {style: 'width: 300px;'},
    render: item => {
      if (item.role) {
        const res = item.role.split(',').map(i => ({
          label: i,
          value: i
        }))
        return res
      }
    }
  },
  {
    title: 'm_tableKey_activeDate',
    value: 'created_string',
    display: true
  }
]
export default {
  name: '',
  data() {
    return {
      pageConfig: {
        CRUD: this.$root.apiCenter.setup.userManagement.get,
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
          btn: '',
          pagination: this.pagination,
          handleFloat: true,
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
  mounted() {
    this.initData(this.pageConfig.CRUD, this.pageConfig)
  },
  methods: {
    initData(url= this.pageConfig.CRUD, params) {
      this.$root.$tableUtil.initTable(this, 'GET', url, params)
    },
  },
  components: {},
}
</script>

<style scoped lang="less">
</style>
