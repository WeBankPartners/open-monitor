<template>
<div class=" ">
  <section>
    <template v-for="(tableItem, tableIndex) in totalData">
      <section :key="tableIndex + 'f'">
        <div class="section-table-tip">
          <Tag color="blue" :key="tableIndex + 'a'">{{tableItem.sys_panel || '请设置名称'}}</Tag>
          <button type="button" @click="addDbMonitor(tableItem.sys_panel)" class="btn btn-sm btn-cancel-f" :key="tableIndex + 'b'">
            <i class="fa fa-plus"></i>
            {{$t('button.add')}}
          </button>
          <PageTable :pageConfig="tableItem.pageConfig"></PageTable>
        </div>
      </section>
    </template>
  </section>
  <Modal
   v-model="db_add_Modal"
    mask-closable
    :title="$t('button.dataMonitoring')" 
    @on-ok="addDbConfig"
    @on-cancel="cancelAddDbConfig"
  >
    <Form :model="activeSysPanal" :label-width="80">
      <FormItem :label="$t('tableKey.name')">
        <Input v-model="activeSysPanal.name" placeholder=""></Input>
      </FormItem>
      <FormItem label="sql">
        <Input v-model="activeSysPanal.sql" type="textarea" placeholder=""></Input>
      </FormItem>
    </Form>
  </Modal>
  <Modal
   v-model="db_edit_Modal"
    mask-closable
    :title="$t('button.dataMonitoring')" 
    @on-ok="editDbConfig"
    @on-cancel="cancelEditDbConfig"
  >
    <Form :model="activeSysPanal" :label-width="80">
      <FormItem :label="$t('tableKey.name')">
        <Input v-model="activeSysPanal.name" placeholder=""></Input>
      </FormItem>
      <FormItem label="sql">
        <Input v-model="activeSysPanal.sql" type="textarea" placeholder=""></Input>
      </FormItem>
    </Form>
  </Modal>
  <Modal v-model="isShowWarning" title="Delete confirmation" @on-ok="ok" @on-cancel="cancel">
    <div class="modal-body" style="padding:30px">
      <div style="text-align:center">
        <p style="color: red">Will you delete it?</p>
      </div>
    </div>
  </Modal>
</div>
</template>

<script>
let tableEle = [{
    title: 'tableKey.endpoint',
    value: 'endpoint_guid',
    display: true
  },
  {
    title: 'tableKey.name',
    value: 'name',
    display: true
  },
  {
    title: 'sql',
    value: 'sql',
    display: true
  },
  {
    title: 'field.resourceLevel',
    value: 'sys_panel',
    display: true
  }
]
const btn = [
  {
    btn_name: 'button.edit',
    btn_func: 'editDbMonitor'
  },
  {
    btn_name: 'button.remove',
    btn_func: 'removeDbmonitor'
  }
]
export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
      selectedData: null,

      db_add_Modal: false,
      db_edit_Modal: false,
      isAdd: true,
      totalData: [],

      activeSysPanal: {
        sys_panel: '',
        name: '',
        sql: ''
      },
      activeId: ''
    }
  },
  props: ['endpointId'],
  mounted() {},
  methods: {
    managementData(dbMonitorData) {
      this.totalData = []
      dbMonitorData.forEach(item => {
        this.totalData.push({
          sys_panel: item.sys_panel,
          pageConfig: {
            table: {
              tableData: item.data,
              tableEle: tableEle,
              primaryKey: 'id',
              btn: btn,
              handleFloat: true
            }
          }
        })
      })
    },
    addDbMonitor(sys_panel) {
      this.activeSysPanal.name = ''
      this.activeSysPanal.sql = ''
      this.activeSysPanal.sys_panel = sys_panel
      this.db_add_Modal = true
    },
    addDbConfig() {
      this.activeSysPanal.endpoint_id = Number(this.endpointId)
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.db.check, this.activeSysPanal, () => {
        this.addPost(this.activeSysPanal)
      })
    },
    addPost(params) {
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.db.add, params, () => {
        this.getData()
      })
    },
    getData() {
      const params = {
        endpoint_id: this.endpointId
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.endpointManagement.db.dbMonitor, params, responseData => {
        this.managementData(responseData)
      })
    },
    cancelAddDbConfig() {},
    removeDbmonitor(val) {
      this.selectedData = val
      this.isShowWarning = true
    },
    exectRemoveDbmonitor(val) {
      let params = {
        id: val.id
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.db.delete, params, () => {
        this.getData()
      })
    },
    editDbMonitor(row) {
      this.activeId = row.id
      // this.activeSysPanal.id = row.id
      this.activeSysPanal.sys_panel = row.sys_panel
      this.activeSysPanal.name = row.name
      this.activeSysPanal.sql = row.sql
      this.db_edit_Modal = true
    },
    editDbConfig() {
      let params = Object.assign(this.activeSysPanal, {id: this.activeId,endpoint_id: Number(this.endpointId)})
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.db.check, params, () => {
        this.editPost(params)
      })
    },
    editPost(params) {
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.db.update, params, () => {
        this.getData()
      })
    },
    cancelEditDbConfig() {},

    ok() {
      this.exectRemoveDbmonitor(this.selectedData)
    },
    cancel() {
      this.isShowWarning = false
    },
  },
  components: {},
}
</script>

<style scoped lang="less">
.modal-backdrop {
  display: none;
}
</style>
