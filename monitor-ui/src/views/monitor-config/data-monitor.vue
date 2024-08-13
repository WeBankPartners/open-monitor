<template>
  <div class=" ">
    <section>
      <div style="text-align: end;">
        <Button type="primary" @click="addDbMonitor">{{$t('m_button_add')}}</Button>
      </div>
      <div style="height:500px;overflow-y:auto">
        <template v-for="(tableItem, tableIndex) in totalData">
          <section :key="tableIndex + 'f'">
            <div class="section-table-tip">
              <Tag color="blue" :key="tableIndex + 'a'">{{tableItem.sys_panel || 'Please Set Name'}}
                <span @click="editPanalName(tableItem.sys_panel_value)"><i class="fa fa-pencil" aria-hidden="true"></i></span>
              </Tag>
              <PageTable :pageConfig="tableItem.pageConfig"></PageTable>
            </div>
          </section>
        </template>
      </div>
    </section>
    <Modal
      v-model="db_add_Modal"
      :mask-closable="false"
      :title="$t('m_button_dataMonitoring')"
      @on-ok="addDbConfig"
      @on-cancel="cancelAddDbConfig"
    >
      <Form :model="activeSysPanal" :label-width="80">
        <FormItem :label="$t('m_field_resourceLevel')">
          <Select filterable clearable v-model="newPanalName" style="width:400px">
            <Option v-for="item in panalNameList" :value="item.option_value" :key="item.option_value">{{ item.option_text }}</Option>
          </Select>
        </FormItem>
        <FormItem :label="$t('m_tableKey_name')">
          <Input v-model="activeSysPanal.name" placeholder="" style="width:400px"></Input>
        </FormItem>
        <FormItem label="sql">
          <Input v-model="activeSysPanal.sql" type="textarea" style="width:400px"></Input>
        </FormItem>
      </Form>
    </Modal>
    <Modal
      v-model="db_edit_Modal"
      :mask-closable="false"
      :title="$t('m_button_dataMonitoring')"
      @on-ok="editDbConfig"
      @on-cancel="cancelEditDbConfig"
    >
      <Form :model="activeSysPanal" :label-width="80">
        <FormItem :label="$t('m_field_resourceLevel')">
          <Select filterable clearable v-model="newPanalName" style="width:400px">
            <Option v-for="item in panalNameList" :value="item.option_value" :key="item.option_value">{{ item.option_text }}</Option>
          </Select>
        </FormItem>
        <FormItem :label="$t('m_tableKey_name')">
          <Input v-model="activeSysPanal.name" placeholder=""></Input>
        </FormItem>
        <FormItem label="sql">
          <Input v-model="activeSysPanal.sql" type="textarea" placeholder=""></Input>
        </FormItem>
      </Form>
    </Modal>
    <Modal v-model="isShowWarning"
           :title="$t('m_delConfirm_title')"
           @on-ok="ok"
           @on-cancel="cancel"
    >
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('m_delConfirm_tip')}}</p>
        </div>
      </div>
    </Modal>
    <Modal
      v-model="isShowChangePanalName"
      :title="$t('m_button_edit')"
      @on-ok="changePanalName"
    >
      <Select filterable clearable v-model="newPanalName" style="width:400px">
        <Option v-for="item in panalNameList" :value="item.option_value" :key="item.option_value">{{ item.option_text }}</Option>
      </Select>
    </Modal>
  </div>
</template>

<script>
const tableEle = [{
  title: 'm_tableKey_endpoint',
  value: 'endpoint_guid',
  display: true
},
{
  title: 'm_tableKey_name',
  value: 'name',
  display: true
},
{
  title: 'sql',
  value: 'sql',
  display: true
},
{
  title: 'm_field_resourceLevel',
  value: 'sys_panel',
  display: true
}
]
const btn = [
  {
    btn_name: 'm_button_edit',
    btn_func: 'editDbMonitor'
  },
  {
    btn_name: 'm_button_remove',
    btn_func: 'removeDbmonitor',
    color: 'red'
  }
]
export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
      selectedData: null,

      isShowChangePanalName: false,
      activePanalName: '',
      newPanalName: '',
      panalNameList: [],

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
          sys_panel_value: item.sys_panel_value,
          pageConfig: {
            table: {
              tableData: item.data,
              tableEle,
              primaryKey: 'id',
              btn,
              handleFloat: true
            }
          }
        })
      })
    },
    addDbMonitor() {
      this.newPanalName = null
      this.activeSysPanal.name = ''
      this.activeSysPanal.sql = ''
      // this.activeSysPanal.sys_panel = sys_panel
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.endpointManagement.db.panalName, '', responseData => {
        this.panalNameList = responseData
        this.panalNameList.unshift({
          option_text: 'null',
          option_value: ''
        })
      })
      this.db_add_Modal = true
    },
    addDbConfig() {
      this.activeSysPanal.endpoint_id = Number(this.endpointId)
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.db.check, this.activeSysPanal, () => {
        this.addPost(this.activeSysPanal)
      })
    },
    addPost(params) {
      params.sys_panel = this.newPanalName
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
      const params = {
        id: val.id
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.db.delete, params, () => {
        this.getData()
      })
    },
    editDbMonitor(row) {
      this.activeId = row.id
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.endpointManagement.db.panalName, '', responseData => {
        this.panalNameList = responseData
        this.panalNameList.unshift({
          option_text: 'null',
          option_value: ''
        })
      })
      this.newPanalName = row.sys_panel
      this.activeSysPanal.name = row.name
      this.activeSysPanal.sql = row.sql
      this.db_edit_Modal = true
    },
    editDbConfig() {
      const params = Object.assign(this.activeSysPanal, {
        id: this.activeId,
        endpoint_id: Number(this.endpointId),
        sys_panel: this.newPanalName
      })
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
    editPanalName(panalName) {
      this.newPanalName = panalName
      this.panalName = panalName
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.endpointManagement.db.panalName, '', responseData => {
        this.panalNameList = responseData
        this.panalNameList.unshift({
          option_text: 'null',
          option_value: ''
        })
        this.isShowChangePanalName = true
      })
    },
    changePanalName() {
      const params = {
        old_name: this.panalName,
        new_name: this.newPanalName,
        endpoint_id: Number(this.endpointId)
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.db.updatePanalName, params, () => {
        this.$Message.success('Success!')
        this.getData()
      })
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
.modal-backdrop {
  display: none;
}
</style>
