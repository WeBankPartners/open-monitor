<template>
  <div class=" ">
    <section>
      <ul class="search-ul">
        <li class="search-li">
          <Select
            style="width:300px"
            v-model="endpointID"
            filterable
            remote
            clearable
            :remote-method="getEndpointList"
            @on-change="changeEndpoint"
            @on-clear="clearEndpoint"
            >
            <Option v-for="(option, index) in endpointOptions" :value="option.id" :key="index">
             <Tag color="cyan" class="tag-width" v-if="option.type == 'host'">host</Tag>{{option.option_text}}</Option>
          </Select>
        </li>
      </ul>
    </section>
    <section v-if="!!endpointID" style="margin-top: 16px;">
      <Tag color="blue">{{endpointGuid}}</Tag>
      <button @click="add" type="button" class="btn btn-sm btn-cancel-f">
        <i class="fa fa-plus"></i>
        {{$t('button.add')}}
      </button>
      <PageTable :pageConfig="pageConfig">
        <div slot='tableExtend'>
          <extendTable :detailConfig="pageConfig.table.isExtend.detailConfig"></extendTable>
        </div>
      </PageTable>
      <ModalComponent :modelConfig="modelConfig">
        <div slot="thresholdConfig" class="extentClass">  
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">{{$t('field.endpoint')}}:</label>
            <Select v-model="modelConfig.addRow.owner_endpoint" style="width:338px">
              <Option v-for="item in modelConfig.slotConfig.endpointOption" :value="item.guid" :key="item.guid">{{ item.groups_name }}</Option>
            </Select>
          </div>
        </div>
      </ModalComponent>
      <ModalComponent :modelConfig="ruleModelConfig">
        <div slot="ruleConfig" class="extentClass">
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">metric_config:</label>
           <div style="margin: 4px 12px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
              <template v-for="(item, index) in ruleModelConfig.addRow.metric_config">
                <p :key="index">
                  <Button
                    @click="deleterule(index)"
                    size="small"
                    style="background-color: #ff9900;border-color: #ff9900;"
                    type="error"
                    icon="md-close"
                  ></Button>
                  <Input v-model="item.key" style="width: 146px" placeholder="e.g:[.*][.*]" />
                  <Input v-model="item.metric" style="width: 146px" placeholder="e.g:code" />
                  <Select v-model="item.agg_type" filterable style="width:140px">
                    <Option v-for="agg in ruleModelConfig.slotConfig.aggOption" :value="agg" :key="agg">{{
                      agg
                    }}</Option>
                  </Select>
                </p>
              </template>
              <Button
                @click="addEmpty('metric_config')"
                type="success"
                size="small"
                style="background-color: #0080FF;border-color: #0080FF;"
                long
                >{{ $t('hr_add_rule') }}</Button
              >
            </div>
          </div>
        </div>
      </ModalComponent>
      <Modal
        v-model="isShowWarning"
        title="Delete confirmation"
        @on-ok="ok"
        @on-cancel="cancel">
        <div class="modal-body" style="padding:30px">
          <div style="text-align:center">
            <p style="color: red">Will you delete it?</p>
          </div>
        </div>
      </Modal>
      <Modal
        v-model="isShowWarningDelete"
        title="Delete confirmation"
        @on-ok="okDelRow"
        @on-cancel="cancleDelRow">
        <div class="modal-body" style="padding:30px">
          <div style="text-align:center">
            <p style="color: red">Will you delete it?</p>
          </div>
        </div>
      </Modal>
    </section>
  </div>
</template>

<script>
import extendTable from '@/components/table-page/extend-table'
let tableEle = [
  {title: 'tableKey.path', value: 'path', display: true},
  {title: 'tableKey.endpoint', value: 'owner_endpoint', display: true}
]
const btn = [
  {btn_name: 'button.add', btn_func: 'singeAddF'},
  {btn_name: 'button.edit', btn_func: 'editF'},
  {btn_name: 'button.remove', btn_func: 'deleteConfirmModal'},
]

export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
      requestParams: null,
      isShowWarningDelete: false,
      type: '',
      typeValue: 'endpoint',
      typeList: [
        {label: 'field.endpoint', value: 'endpoint'},
        {label: 'field.group', value: 'grp'}
      ],

      endpointID: null,
      endpointGuid: null,
      endpointOptions: [],

      totalPageConfig: [],
      pageConfig: {
        table: {
          tableData: [],
          tableEle: tableEle,
          // filterMoreBtn: 'filterMoreBtn',
          primaryKey: 'id',
          btn: btn,
          handleFloat:true,
          isExtend: {
            func: 'getExtendInfo',
            data: {},
            slot: 'tableExtend',
            detailConfig: [{
              isExtendF: true,
              title: '',
              config: [
                {title: 'tableKey.regular', value: 'regular', display: true},
                {title: 'tableKey.tags', value: 'tags', display: true},
                {title: 'table.action',btn:[
                  {btn_name: 'button.edit', btn_func: 'editPathItem'},
                  {btn_name: 'button.remove', btn_func: 'delPathconfirmModal'}
                ]}
              ],
              data: [1],
              scales: ['25%', '20%', '15%', '20%', '20%']
            }]
          }
        }
      },
      modelTip: {
        key: '',
        value: 'metric'
      },
      ruleModelConfig: {
        modalId: 'rule_Modal',
        isAdd: true,
        modalStyle: 'min-width:53px',
        modalTitle: 'rule',
        saveFunc: 'saveRule',
        config: [
          {label: 'regular', value: 'regular', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'},
          {label: 'tags', value: 'tags', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'},
          {name:'ruleConfig',type:'slot'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          regular: null,
          tags: null,
          metric_config: []
        },
        slotConfig: {
          aggOption: ['sum', 'avg', 'count']
        }
      },
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: 'field.businessMonitor',
        isAdd: true,
        config: [
          {label: 'tableKey.path', value: 'path', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'},
          {name:'thresholdConfig',type:'slot'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          path: null,
          owner_endpoint: ''
        },
        slotConfig: {
          endpointOption: []
        }
      },
      id: null,
      activeData: null,
      extendData: null,
    }
  },
  
  mounted () {
    this.getEndpointList('.')
  },
  methods: {
    /*********/
    addEmpty (type) {
      if (type === 'metric_config') {
        this.ruleModelConfig.addRow.metric_config.push({
          key: '',
          metric: '',
          agg_type: 'avg'
        })
      }
    },
    /*********/
    typeChange() {
      this.totalPageConfig = []
      this.getEndpointList('.')
    },
    clearEndpoint () {
      this.totalPageConfig = []
      this.getEndpointList('.')
    },
    changeEndpoint (val) {
      if (val) {
        this.endpointGuid = this.endpointOptions.find(item => item.id === val).option_text
        this.requestData(this.endpointID)
      } else {
        this.endpointGuid = ''
      }
    },
    getEndpointList (query) {
      const params = {type: 'endpoint', search: query}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceSearch.strategyApi, params, (responseData) => {
        this.endpointOptions = responseData.filter(item => item.type === 'host')
      })
    },
    requestData (id) {
      let params = {id}
      this.totalPageConfig = []
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', 'monitor/api/v1/alarm/business/list', params, (responseData) => {
        this.pageConfig.table.tableData = responseData.path_list
      })
    },
    deleteConfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarning = true
    },
    ok () {
      this.delF(this.selectedData)
    },
    cancel () {
      this.isShowWarning = false
    },
    deleteConfirm (rowData) {
      this.$delConfirm({
        msg: rowData.path,
        callback: () => {
          this.delF(rowData)
        }
      })
    },
    delF (rowData) {
      const index = this.pageConfig.table.tableData.findIndex(item => item.id === rowData.id)
      let tableData = JSON.parse(JSON.stringify(this.pageConfig.table.tableData))
      tableData.splice(index,1)
      let params = {
        endpoint_id: this.endpointID,
        path_list: tableData
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', 'monitor/api/v1/alarm/business/update', params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.requestData(this.endpointID)
      })
    },
    paramsPrepare() {
      let modelParams = {
        path: this.modelConfig.addRow.path,
        strategy: [{
          keyword: this.modelConfig.addRow.keyword,
          cond: this.modelConfig.cond + this.modelConfig.condValue,
          last: this.modelConfig.lastValue + this.modelConfig.last,
          priority: this.modelConfig.priority
        }]
      }
      if (this.type === 'grp') {
        modelParams.grp_id = this.typeValue
        modelParams.endpoint_id = 0
      } else {
        modelParams.endpoint_id = parseInt(this.typeValue)
        modelParams.grp_id = 0
      }
              
      return modelParams
    },
    singeAddF (rowData) {
      this.activeData = rowData
      // this.ruleModelConfig.addRow.path = rowData.path
      this.ruleModelConfig.isAdd = true
      this.$root.JQ('#rule_Modal').modal('show')
    },
    add () {
      this.modelConfig.isAdd = true
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', 'monitor/api/v1/alarm/endpoint/list?page=1&size=1000', '', responseData => {
        this.modelConfig.slotConfig.endpointOption = responseData.data
        this.$root.JQ('#add_edit_Modal').modal('show')
      })
    },
    addPost () {
      this.$validator.validate().then(result => {
        if (!result) return
        let params = {
          endpoint_id: this.endpointID,
          path_list: [
            {
              id: 0,
              ...this.modelConfig.addRow,
              rules: []
            }
          ]
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/business/add', params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.requestData(this.endpointID)
        })
      })
    },
    editF (rowData) {
      this.modelConfig.isAdd = false
      this.activeData = rowData
      this.modelConfig.addRow.path = rowData.path
      this.modelConfig.addRow.owner_endpoint = rowData.owner_endpoint
      this.modelTip.value = rowData.path
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', 'monitor/api/v1/alarm/endpoint/list?page=1&size=1000', '', responseData => {
        this.modelConfig.slotConfig.endpointOption = responseData.data
        this.$root.JQ('#add_edit_Modal').modal('show')
      })
    },
    editPost () {
      this.$validator.validate().then(result => {
        if (!result) return
        const index = this.pageConfig.table.tableData.findIndex(item => item.id === this.activeData.id)
        let tableData = JSON.parse(JSON.stringify(this.pageConfig.table.tableData))
        tableData[index].path = this.modelConfig.addRow.path
        tableData[index].owner_endpoint = this.modelConfig.addRow.owner_endpoint
        let params = {
          endpoint_id: this.endpointID,
          path_list: tableData
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', 'monitor/api/v1/alarm/business/update', params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.requestData(this.endpointID)
        })
      })
    },
    savePath () {
      let params = {
        id: this.activeData.id,
        tpl_id: this.activeData.tpl_id,
        path: this.pathModelConfig.addRow.path
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.logManagement.editList.api, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#path_Modal').modal('hide')
        this.requestData(this.typeValue)
      })
    },
    getExtendInfo(item){
      console.log(item)
      console.log(this.pageConfig.table.isExtend.detailConfig)
      this.pageConfig.table.isExtend.detailConfig[0].data = item.rules
    },
    editPathItem (rowData) {
      this.modelConfig.isAdd = false
      this.id = rowData.id
      this.extendData = rowData
      this.modelTip.value = rowData.id
      this.modelConfig.addRow = this.$root.$tableUtil.manageEditParams(this.modelConfig.addRow, rowData)
      let cond = rowData.cond.split('')
      if (cond.indexOf('=') > 0) {
        this.modelConfig.cond = cond.slice(0,2).join('')
        this.modelConfig.condValue = cond.slice(2).join('')
      } else {
        this.modelConfig.cond = cond.slice(0,1).join('')
        this.modelConfig.condValue = cond.slice(1).join('')
      }
      let last = rowData.last
      this.modelConfig.last = last.substring(last.length-1)
      this.modelConfig.lastValue = last.substring(0,last.length-1)
      this.modelConfig.priority = rowData.priority
      this.$root.JQ('#add_edit_Modal').modal('show')
    },
    delPathconfirm (rowData) {
      this.$delConfirm({
        msg: rowData.keyword,
        callback: () => {
          this.delPathItem(rowData)
        }
      })
    },
    delPathconfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarningDelete = true
    },
    okDelRow () {
      this.delPathItem(this.selectedData)
    },
    cancleDelRow () {
      this.isShowWarningDelete = false
    },
    delPathItem (rowData) {
      let params = {id: rowData.id}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.logManagement.delete.api, params, () => {
        // this.$root.$eventBus.$emit('hideConfirmModal')
        this.$Message.success(this.$t('tips.success'))
        this.requestData(this.typeValue)
      })
    }
  },
  components: {
    extendTable
  },
}
</script>

<style scoped lang="less">
.search-li {
    display: inline-block;
  }
  .search-ul>li:not(:first-child) {
    padding-left: 10px;
  }
</style>
<style scoped lang="less">
  .is-danger {
    color: red;
    margin-bottom: 0px;
  }
  .search-input {
    height: 32px;
    padding: 4px 7px;
    font-size: 12px;
    border: 1px solid #dcdee2;
    border-radius: 4px;
    width: 230px;
  }

  .section-table-tip {
    margin: 24px 20px 0;
  }
  .search-input:focus {
    outline: 0;
    border-color: #57a3f3;
  }

  .search-input-content {
    display: inline-block;
    vertical-align: middle; 
  }
  .tag-width {
    cursor: auto;
    width: 55px;
    text-align: center;
  } 
</style>
