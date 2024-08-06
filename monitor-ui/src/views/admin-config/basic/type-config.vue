<!--类型配置-->
<template>
  <div class="admin-type-config">
    <div class="search">
      <Input placeholder="类型名" style="width:250px;" v-model="searchParams.displayName" clearable @on-change="handleQuery" />
      <Button type="success" @click="handleAdd">{{ $t('m_button_add') }}</Button>
    </div>
    <Table
      size="small"
      :columns="tableColumns"
      :data="tableData"
      :loading="loading"
      :maxHeight="maxHeight"
    ></Table>
    <Modal v-model="addVisible" :title="$t('m_button_add')">
      <Form ref="form" :model="form" :rules="ruleForm" label-position="left" :label-width="80">
        <FormItem label="类型名" prop="displayName">
          <Input v-model.trim="form.displayName" :maxlength="20" show-word-limit />
        </FormItem>
      </Form>
      <div slot="footer">
        <Button @click="addVisible = false">{{ $t('button.cancel') }}</Button>
        <Button type="primary" @click="saveAdd">{{ $t('button.save') }}</Button>
      </div>
    </Modal>
    <Modal
      class="delete-confirm-modal"
      v-model="deleteVisible"
      title="提示"
      @on-ok="handleDelete"
      @on-cancel="deleteVisible = false"
    >
      <div class="confirm-body">
        <Icon type="md-alert" color="#ff9900" size="28" />
        {{ $t('m_delConfirm_title') }}
      </div>
    </Modal>
  </div>
</template>

<script>
import debounce from 'lodash/debounce'
export default {
  data() {
    return {
      searchParams: {
        displayName: ''
      },
      form: {
        displayName: ''
      },
      tableData: [],
      maxHeight: 500,
      loading: false,
      addVisible: false,
      deleteVisible: false,
      currentRow: {},
      ruleForm: {
        displayName: [
          {
            validator: (rule, value, callback) => {
              if (value === '') {
                callback(new Error('请输入类型名'))
              }
              else if (!/^[^\u4e00-\u9fa5]*$/.test(value)) {
                callback(new Error('不能输入中文字符'))
              }
              else {
                callback()
              }
            },
            trigger: 'blur'
          }
        ]
      },
      tableColumns: [
        {
          title: '类型名',
          minWidth: 200,
          render: (h, params) => (
            <div>
              {
                params.row.displayName ? (<TagShow tagName={params.row.displayName}></TagShow>) : <div>-</div>
              }
            </div>
          )
        },
        {
          title: '对象数',
          key: 'objectCount',
          minWidth: 100
        },
        {
          title: '创建人',
          key: 'createUser',
          minWidth: 120,
          render: (h, params) => <span>{params.row.createUser || '-'}</span>
        },
        {
          title: '创建时间',
          key: 'createTime',
          minWidth: 160,
          render: (h, params) => <span>{params.row.createTime || '-'}</span>
        },
        {
          title: this.$t('m_table_action'),
          key: 'action',
          width: 100,
          align: 'center',
          fixed: 'right',
          render: (h, params) => (
            <div style="display:flex;justify-content:center;">
              <Tooltip content={this.$t('button.remove')} placement="top">
                <Button
                  size="small"
                  type="error"
                  onClick={() => {
                    this.deleteVisible = true
                    this.currentRow = params.row
                  }}
                  disabled={params.row.systemType === 1}
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
  mounted() {
    this.getList()
    this.maxHeight = document.body.clientHeight - 140
  },
  methods: {
    handleQuery: debounce(function () {
      this.getList()
    }, 300),
    getList() {
      this.loading = true
      const params = {
        name: this.searchParams.displayName
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getTypeConfigList, params, data => {
        this.loading = false
        this.tableData = data || []
      }, {isNeedloading: false})
    },
    handleAdd() {
      this.addVisible = true
      this.$refs.form.resetFields()
    },
    saveAdd() {
      this.$refs.form.validate(valid => {
        if (valid) {
          this.$root.$httpRequestEntrance.httpRequestEntrance('post', this.$root.apiCenter.createTypeConfig, this.form, () => {
            this.addVisible = false
            this.getList()
            this.$Message.success(this.$t('m_tips_success'))
          }, {isNeedloading: false})
        }
      })
    },
    handleDelete() {
      const params = {
        id: this.currentRow.guid
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', this.$root.apiCenter.createTypeConfig, params, () => {
        this.deleteVisible = false
        this.getList()
        this.$Message.success(this.$t('m_tips_success'))
      }, {isNeedloading: false})
    }
  }
}
</script>

<style lang="less">
.delete-confirm-modal{
  .ivu-modal-header {
    border-bottom: none;
  }
  .ivu-modal-footer {
    border-top: none;
  }
  .confirm-body {
    display: flex;
    align-items: center;
  }
}
</style>
<style scoped lang="less">
.admin-type-config {
  width: 100%;
  .search {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom:10px;
  }
}
</style>
