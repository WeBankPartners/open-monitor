<template>
  <div class=" ">
    <section>
      <ul class="search-ul">
        <li class="search-li">
          <Select v-model="type" style="width:100px" @on-change="typeChange">
            <Option v-for="item in typeList" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
          </Select>
        </li>
        <li class="search-li">
          <Select
            style="width:300px"
            v-model="endpointID"
            filterable
            remote
            clearable
            :remote-method="getEndpointList"
            @on-clear="clearEndpoint"
            >
            <Option v-for="(option, index) in endpointOptions" :value="option.id" :key="index">
            <Tag color="cyan" class="tag-width" v-if="option.type == 'host'">host</Tag>
            <Tag color="blue" class="tag-width" v-if="option.type == 'mysql'">mysql </Tag>
            <Tag color="geekblue" class="tag-width" v-if="option.type == 'redis'">redis </Tag>
            <Tag color="purple" class="tag-width" v-if="option.type == 'tomcat'">tomcat</Tag>{{option.option_text}}</Option>
          </Select>
        </li>
        <li class="search-li">
          <button type="button" class="btn btn-sm btn-confirm-f"
          @click="search">
            <i class="fa fa-search" ></i>
            {{$t('button.search')}}
          </button>
        </li>
      </ul> 
    </section>
    <section>
      <template v-for="(tableItem, tableIndex) in totalPageConfig">
        <div :key="tableIndex + 'f'" class="section-table-tip">
          <Tag color="blue" :key="tableIndex + 'a'" v-if="tableItem.obj_name">{{tableItem.obj_name}}</Tag>
          <button @click="add(tableItem.obj_type)" type="button" v-if="tableItem.operation" class="btn btn-sm btn-cancel-f" :key="tableIndex + 'b'">
            <i class="fa fa-plus"></i>
            {{$t('button.add')}}
          </button>
        </div>
        <PageTable :pageConfig="tableItem" :key="tableIndex + 'c'">
          <div slot='tableExtend'>
            <extendTable :detailConfig="pageConfig.table.isExtend.detailConfig"></extendTable>
          </div>
        </PageTable>
      </template>
      <ModalComponent :modelConfig="pathModelConfig">
      </ModalComponent>
      <ModalComponent :modelConfig="modelConfig">
        <div slot="thresholdConfig" class="extentClass">  
          <!-- <div class="marginbottom params-each">
            <label class="col-md-2 label-name">{{$t('tableKey.condition')}}:</label>
            <Select v-model="modelConfig.cond" style="width:100px">
              <Option v-for="item in modelConfig.condList" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
            <div class="search-input-content" style="margin-left: 8px">
              <input 
                v-validate="'required|isNumber'" 
                v-model="modelConfig.condValue" 
                name="condValue"
                :class="{ 'red-border': veeErrors.has('condValue') }"
                type="text" 
                class="form-control model-input search-input c-dark"/>
              <label class="required-tip">*</label>
            </div>
            <div style="margin-left:120px">
              <label v-show="veeErrors.has('condValue')" class="is-danger">{{ veeErrors.first('condValue')}}</label>
            </div>
          </div>
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">{{$t('tableKey.s_last')}}:</label>
            <div class="search-input-content" style="margin-right: 8px">
              <input 
                v-validate="'required|isNumber'" 
                v-model="modelConfig.lastValue" 
                name="lastValue"
                :class="{ 'red-border': veeErrors.has('lastValue') }"
                type="text" 
                class="form-control model-input search-input c-dark"/>
              <label class="required-tip">*</label>
            </div>
            <Select v-model="modelConfig.last" style="width:100px">
              <Option v-for="item in modelConfig.lastList" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
            <div style="margin-left:10px">
              <label v-show="veeErrors.has('lastValue')" class="is-danger">{{ veeErrors.first('lastValue')}}</label>
            </div>
          </div> -->
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">{{$t('tableKey.s_priority')}}:</label>
            <Select v-model="modelConfig.priority" style="width:100px">
              <Option v-for="item in modelConfig.priorityList" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
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
import {thresholdList, lastList, priorityList} from '@/assets/config/common-config.js'
import extendTable from '@/components/table-page/extend-table'
let tableEle = [
  {title: 'tableKey.path', value: 'path', display: true}
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
                // {title: 'tableKey.condition', value: 'cond', display: true},
                {title: 'tableKey.keyword', value: 'keyword', display: true},
                // {title: 'tableKey.s_last', value: 'last', display: true},
                {title: 'tableKey.s_priority', value: 'priority', display: true},
                {title: 'table.action',btn:[
                  {btn_name: 'button.edit', btn_func: 'editPathItem'},
                  {btn_name: 'button.remove', btn_func: 'delPathconfirmModal'}
                ]}
              ],
              data: [],
              scales: ['25%', '20%', '15%', '20%', '20%']
            }]
          }
        }
      },
      modelTip: {
        key: '',
        value: 'metric'
      },
      pathModelConfig: {
        modalId: 'path_Modal',
        modalTitle: 'title.logAdd',
        saveFunc: 'savePath',
        config: [
          {label: 'tableKey.path', value: 'path', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          path: null,
        }
      },
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: 'title.logAdd',
        isAdd: true,
        config: [
          {label: 'tableKey.path', value: 'path', placeholder: 'tips.required', v_validate: 'required:true',hide: 'edit', disabled: false, type: 'text'},
          {label: 'tableKey.keyword', value: 'keyword', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'},
          {name:'thresholdConfig',type:'slot'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          path: null,
          keyword: null,
        },
        metricName: '',
        metricList: [],
        cond: '>',
        condList: thresholdList,
        condValue: '',
        last: 's',
        lastList: lastList,
        lastValue: '',
        priority: 'low',
        priorityList: priorityList,
        slotConfig: {
          resourceSelected: [],
          resourceOption: []
        }
      },
      id: null,
      singeAddId: '',
      activeData: null,
      extendData: null,
    }
  },
  mounted () {
    if (!this.$root.$validate.isEmpty_reset(this.$route.params)) {
      this.$parent.activeTab = '/monitorConfigIndex/logManagement'
      this.type = this.$route.params.type
      this.typeValue = this.$route.params.id
      this.requestData(this.type, this.typeValue)
    } else {
      this.type = 'endpoint'
      this.typeValue = ''
    }
    this.getEndpointList('.')
    this.$root.JQ('#add_edit_Modal').on('hidden.bs.modal', () => {
      this.modelConfig.thresholdValue = ''
      this.modelConfig.lastValue = ''
      this.modelConfig.condValue = ''
      this.singeAddId = ''
    })
  },
  methods: {
    search () {
      if (this.endpointID === null) {
        return
      }
      this.typeValue = this.endpointID
      this.requestData(this.type, this.endpointID)
    },
    typeChange() {
      this.totalPageConfig = []
      this.getEndpointList('.')
    },
    clearEndpoint () {
      this.totalPageConfig = []
      this.getEndpointList('.')
    },
    getEndpointList (query) {
      const params = {type: this.type,search: query}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceSearch.strategyApi, params, (responseData) => {
        this.endpointOptions = this.type === 'endpoint'? responseData.filter(item => item.type === 'host'):responseData
      })
    },
    requestData (type, id) {
      let params = {}
      params.type = type
      params.id = id
      this.totalPageConfig = []
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.logManagement.list.api, params, (responseData) => {
        responseData.forEach((item)=>{
          let config = this.$root.$validate.deepCopy(this.pageConfig.table)
          if (!item.operation) {
            config.btn = []
          }
          config.tableData = item.log_monitor
          this.totalPageConfig.push({table:config, obj_type: item.obj_type, obj_name: item.obj_name, operation:item.operation})
        })
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
      let params = {id: rowData.id}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.logManagement.delList.api, params, () => {
        // this.$root.$eventBus.$emit('hideConfirmModal')
        this.$Message.success(this.$t('tips.success'))
        this.requestData(this.type, this.typeValue)
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
      this.modelConfig.addRow.path = rowData.path
      this.singeAddId = rowData.id
      this.modelConfig.isAdd = false
      this.$root.JQ('#add_edit_Modal').modal('show')
    },
    add () {
      this.modelConfig.isAdd = true
      this.$root.JQ('#add_edit_Modal').modal('show')
    },
    addPost () {
      this.$validator.validate().then(result => {
        if (!result) return
        let params = this.paramsPrepare()
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.logManagement.add.api, params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.requestData(this.type, this.typeValue)
        })
      })
    },
    editF (rowData) {
      this.pathModelConfig.isAdd = false
      this.activeData = rowData
      this.pathModelConfig.addRow.path = rowData.path
      this.modelTip.value = rowData.path
      this.$root.JQ('#path_Modal').modal('show')
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
        this.requestData(this.type, this.typeValue)
      })
    },
    getExtendInfo(item){
      item.strategy.forEach((i)=>{
        i.tpl_id = item.tpl_id
        i.path = item.path
      })
      this.pageConfig.table.isExtend.detailConfig[0].data = item.strategy
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
        this.requestData(this.type, this.typeValue)
      })
    },
    editPost () {
      this.$validator.validate().then(result => {
        if (!result) return
        let params = this.paramsPrepare()
        let url = ''
        if (!this.$root.$validate.isEmpty_reset(this.singeAddId)) {
          params.id = this.singeAddId
          url = this.$root.apiCenter.logManagement.add.api
        } else {
          params.tpl_id = this.extendData.tpl_id
          params.strategy[0].id = this.extendData.id
          url = this.$root.apiCenter.logManagement.update.api
        }

        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', url, params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.requestData(this.type, this.typeValue)
          this.$root.$store.commit('changeTableExtendActive', -1)
        })
      })
    },
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
