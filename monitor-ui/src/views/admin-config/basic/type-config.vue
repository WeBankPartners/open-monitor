<!--类型配置-->
<template>
  <div class="admin-type-config">
    <Input placeholder="请输入类型名" style="width:250px;margin-bottom:10px;" v-model="searchParams.name" />
    <Table
      size="small"
      :columns="tableColumns"
      :data="tableData"
      :loading="loading"
    ></Table>
    <Page
      style="float: right; margin-top: 16px"
      :total="pageable.total"
      @on-change="changPage"
      show-sizer
      :current="pageable.current"
      :page-size="pageable.pageSize"
      @on-page-size-change="changePageSize"
      show-total
    />
  </div>
</template>

<script>
export default {
  data() {
    return {
      searchParams: {
        name: ''
      },
      tableData: [],
      loading: false,
      pageable: {
        pageSize: 10,
        startIndex: 1,
        current: 1,
        total: 0
      },
      tableColumns: [
        {
          title: '类型名',
          minWidth: 200,
          key: 'name'
        },
        {
          title: '对象数',
          key: 'number',
          minWidth: 100
        },
        {
          title: '创建人',
          key: 'createdBy',
          minWidth: 120
        },
        {
          title: '更新人',
          key: 'updatedBy',
          minWidth: 120
        },
        {
          title: '更新时间',
          key: 'updatedTime',
          minWidth: 160
        },
        {
          title: this.$t('m_table_action'),
          key: 'action',
          width: 130,
          align: 'center',
          fixed: 'right',
          render: (h, params) => (
            <div style="display:flex;justify-content:center;">
              <Tooltip content={this.$t('be_details')} placement="top">
                <Button
                  size="small"
                  type="error"
                  onClick={() => {
                    this.handleDelete(params.row) // 查看
                  }}
                  style="margin-right:5px;"
                >
                  <Icon type="md-trash" size="16"></Icon>
                </Button>
              </Tooltip>
            </div>
          )
        }
      ]
    }
  },
  methods: {
    handleDelete() {},
    changPage() {},
    changePageSize() {}
  }
}
</script>

<style scoped lang="less">
.admin-type-config {
  width: 100%;
}
</style>
