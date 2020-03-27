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
            <Tag color="cyan" class="tag-width" v-if="option.option_value.split(':')[1] == 'host'">host</Tag>
            <Tag color="blue" class="tag-width" v-if="option.option_value.split(':')[1] == 'mysql'">mysql </Tag>
            <Tag color="geekblue" class="tag-width" v-if="option.option_value.split(':')[1] == 'redis'">redis </Tag>
            <Tag color="purple" class="tag-width" v-if="option.option_value.split(':')[1] == 'tomcat'">tomcat</Tag>{{option.option_text}}</Option>
            <!-- <Option v-for="(option, index) in endpointOptions" :value="option.id" :key="index">{{option.option_text}}</Option> -->
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
          <button @click="add" type="button" v-if="tableItem.operation" class="btn btn-sm btn-cancle-f" :key="tableIndex + 'b'">
            <i class="fa fa-plus"></i>
            {{$t('button.add')}}
          </button>
        </div>
        <PageTable :pageConfig="tableItem" :key="tableIndex + 'c'"></PageTable>
      </template>
      <ModalComponent :modelConfig="modelConfig">
        <div slot="metricSelect" class="extentClass">  
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name lable-name-select">{{$t('tableKey.name')}}:</label>
            <Select v-model="modelConfig.addRow.expr" filterable style="width:340px"
            :label-in-value="true" @on-change="selectMetric">
              <Option v-for="item in modelConfig.metricList" :value="item.prom_ql" :key="item.prom_ql+item.metric">{{ item.metric }}</Option>
            </Select>
          </div> 
        </div>
        <div slot="thresholdConfig" class="extentClass">  
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name lable-name-select">{{$t('field.threshold')}}:</label>
            <Select v-model="modelConfig.threshold" style="width:100px">
              <Option v-for="item in modelConfig.thresholdList" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
            <div class="search-input-content" style="margin-left: 8px">
              <input v-model="modelConfig.thresholdValue" type="text" class="search-input" />
            </div>
          </div>
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name lable-name-select">{{$t('tableKey.s_last')}}:</label>
            <div class="search-input-content" style="margin-right: 8px">
              <input v-model="modelConfig.lastValue" type="text" class="search-input" />
            </div>
            <Select v-model="modelConfig.last" style="width:100px">
              <Option v-for="item in modelConfig.lastList" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </div>
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name lable-name-select">{{$t('tableKey.s_priority')}}:</label>
            <Select v-model="modelConfig.priority" style="width:100px">
              <Option v-for="item in modelConfig.priorityList" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </div>
        </div>
      </ModalComponent>
      <ModalDel :ModelDelConfig="ModelDelConfig"></ModalDel>
    </section>
  </div>
</template>

<script>
import {thresholdList, lastList, priorityList} from '@/assets/config/common-config.js'
let tableEle = [
  {title: 'ID', value: 'id', display: false},
  {title: 'tableKey.name', value: 'metric', display: true},
  {title: 'tableKey.expr', value: 'expr', display: true},
  {title: 'tableKey.s_cond', value: 'cond', display: true},
  {title: 'tableKey.s_last', value: 'last', display: true},
  {title: 'tableKey.s_priority', value: 'priority', display: true}
]
const btn = [
  {btn_name: 'button.edit', btn_func: 'editF'},
  {btn_name: 'button.remove', btn_func: 'deleteConfirm'},
]

export default {
  name: '',
  data() {
    return {
      ModelDelConfig: {
        deleteWarning: false,
        msg: '',
        callback: null
      },
      type: '',
      typeValue: 'endpoint',
      typeList: [
        {label: 'field.endpoint', value: 'endpoint'},
        {label: 'field.group', value: 'grp'}
      ],
      paramsType: null, // For get thresholdList
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
        }
      },
      modelTip: {
        key: '',
        value: 'metric'
      },
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: 'field.threshold',
        isAdd: true,
        config: [
          {name:'metricSelect',type:'slot'},
          {label: 'tableKey.expr', value: 'expr', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'textarea'},
          {label: 'tableKey.content', value: 'content', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'textarea'},
          {name:'thresholdConfig',type:'slot'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          metric: null,
          expr: null,
          content: null,
        },
        metricName: '',
        metricList: [],
        threshold: '>',
        thresholdList: thresholdList,
        thresholdValue: '',
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
    }
  },
  watch: {
    endpointID: function (id) {
      this.paramsType = null
      if (this.type === 'endpoint' && id) {
        const selectedEndpoint = this.endpointOptions.find((endpoint) => {
          return endpoint.id === id
        })
        this.paramsType = selectedEndpoint.option_value.split(':')[1]
      }
    }
  },
  mounted () {
    if (!this.$root.$validate.isEmpty_reset(this.$route.params)) {
      this.$parent.activeTab = '/monitorConfigIndex/thresholdManagement'
      this.type = this.$route.params.type
      this.typeValue = this.$route.params.id
      this.paramsType = this.$route.params.paramsType
      this.requestData(this.type, this.typeValue)
    } else {
      this.type = 'endpoint'
      this.typeValue = ''
    }
    this.getEndpointList('.')
    this.$root.JQ('#add_edit_Modal').on('hidden.bs.modal', () => {
      this.modelConfig.thresholdValue = ''
      this.modelConfig.lastValue = ''
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
        this.endpointOptions = responseData
      })
    },
    requestData (type, id) {
      let params = {}
      params.type = type
      params.id = id
      this.totalPageConfig = []
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.thresholdManagement.list.api, params, (responseData) => {
        responseData.forEach((item)=>{
          let config = this.$root.$validate.deepCopy(this.pageConfig.table)
          if (!item.operation) {
            config.btn = []
          }
          item.strategy.forEach(rowData => {
            rowData.type = item.obj_type
          })
          config.tableData = item.strategy
          this.totalPageConfig.push({table:config, obj_type: item.obj_type, obj_name: item.obj_name, operation:item.operation})
        })
      })
    },
    deleteConfirm (rowData) {
      this.ModelDelConfig =  {
        deleteWarning: true,
        msg: rowData.name,
        callback: () => {
          this.delF(rowData)
        }
      }
    },
    delF (rowData) {
      let params = {id: rowData.id}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.thresholdManagement.delete.api, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.requestData(this.type, this.typeValue)
      })
    },
    formValidate () {
      if (this.$root.$validate.isEmpty_reset(this.modelConfig.thresholdValue)) {
        this.$Message.warning(this.$t('tableKey.threshold')+this.$t('tips.required'))
        return false 
      }
      if (this.$root.$validate.isEmpty_reset(this.modelConfig.lastValue)) {
        this.$Message.warning(this.$t('tableKey.s_last')+this.$t('tips.required'))
        return false 
      }
      if (this.$root.$validate.isEmpty_reset(this.modelConfig.addRow.content)) {
        this.$Message.warning(this.$t('tableKey.content')+this.$t('tips.required'))
        return false
      }
      return true
    },
    paramsPrepare() {
      let modelParams = {
        cond: this.modelConfig.threshold + this.modelConfig.thresholdValue,
        last: this.modelConfig.lastValue + this.modelConfig.last,
        priority: this.modelConfig.priority,        
      }
      if (this.type === 'grp') {
        modelParams.grp_id = this.typeValue
        modelParams.endpoint_id = 0
      } else {
        modelParams.endpoint_id = parseInt(this.typeValue)
        modelParams.grp_id = 0
      }
      return Object.assign(modelParams, this.modelConfig.addRow)
    },
    add () {
      var params = {}
      if (this.type === 'endpoint') {
        params = {type: this.paramsType}
      } 
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.metricList.api, params, (responseData) => {
        this.modelConfig.metricList = responseData
      })
      this.modelConfig.isAdd = true
      this.$root.JQ('#add_edit_Modal').modal('show')
    },
    addPost () {
      if (!this.formValidate()) {
        return
      }
      let params = this.paramsPrepare()
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.thresholdManagement.add.api, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#add_edit_Modal').modal('hide')
        this.requestData(this.type, this.typeValue)
      })
    },
    editF (rowData) {
      let params = {}
      if (this.type === 'endpoint') {
        params = {type: this.paramsType}
      } 
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.metricList.api, params, (responseData) => {
        this.modelConfig.metricList = responseData
      })

      this.modelConfig.isAdd = false
      this.id = rowData.id
      this.modelTip.value = rowData.metric
      this.modelConfig.addRow = this.$root.$tableUtil.manageEditParams(this.modelConfig.addRow, rowData)
      let cond = rowData.cond.split('')
      if (cond.indexOf('=') > 0) {
        this.modelConfig.threshold = cond.slice(0,2).join('')
        this.modelConfig.thresholdValue = cond.slice(2).join('')
      } else {
        this.modelConfig.threshold = cond.slice(0,1).join('')
        this.modelConfig.thresholdValue = cond.slice(1).join('')
      }
      let last = rowData.last
      this.modelConfig.last = last.substring(last.length-1)
      this.modelConfig.lastValue = last.substring(0,last.length-1)
      this.modelConfig.priority = rowData.priority
      this.$root.JQ('#add_edit_Modal').modal('show')
    },
    editPost () {
      if (!this.formValidate()) {
        return
      }
      let params = this.paramsPrepare()
      params.strategy_id = this.id
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.thresholdManagement.update.api, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#add_edit_Modal').modal('hide')
        this.requestData(this.type, this.typeValue)
      })
    },
    selectMetric (option) {
      if (option) {
        this.modelConfig.addRow.metric = option.label
      }
    }
  },
  components: {
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
  .search-input {
    display: inline-block;
    height: 32px;
    padding: 4px 7px;
    font-size: 12px;
    border: 1px solid #dcdee2;
    border-radius: 4px;
    color: #515a6e;
    background-color: #fff;
    background-image: none;
    position: relative;
    cursor: text

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
    width: 55px;
    text-align: center;
  }
</style>
