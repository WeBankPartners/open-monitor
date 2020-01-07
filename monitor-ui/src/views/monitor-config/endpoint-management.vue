<template>
  <div class="main-content">
    <div v-if="showGroupMsg" style="padding-left:20px">
      <Tag type="border" closable color="primary" @on-close="closeTag">{{$t('field.group')}}:{{groupMsg.name}}</Tag>
    </div> 
    <PageTable :pageConfig="pageConfig"></PageTable>
    <ModalComponent :modelConfig="modelConfig">
      <div slot="advancedConfig" class="extentClass">   
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name lable-name-select">{{$t('field.endpoint')}}:</label>
          <Select v-model="modelConfig.slotConfig.resourceSelected" multiple filterable style="width:300px">
              <Option v-for="item in modelConfig.slotConfig.resourceOption" :value="item.id" :key="item.id">
                <Tag color="cyan" v-if="item.option_value.split(':')[1] == 'host'">host</Tag>
                <Tag color="blue" v-if="item.option_value.split(':')[1] == 'mysql'">mysql</Tag>
                <Tag color="geekblue" v-if="item.option_value.split(':')[1] == 'redis'">redis</Tag>
                <Tag color="purple" v-if="item.option_value.split(':')[1] == 'tomcat'">tomcat</Tag>
              {{ item.option_text }}</Option>
          </Select>
        </div>
      </div>
    </ModalComponent>
    <ModalComponent :modelConfig="historyAlarmModel">
      <div slot="historyAlarm">
        <tableTemp :table="historyAlarmPageConfig.table" :pageConfig="historyAlarmPageConfig"></tableTemp>
      </div>
    </ModalComponent>
    <ModalComponent :modelConfig="endpointRejectModel">
      <div slot="endpointReject">  
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name lable-name-select">{{$t('field.endpoint')}}:</label>
          <Select v-model="endpointRejectModel.addRow.type" style="width:338px">
              <Option v-for="item in endpointRejectModel.endpointType" :value="item.value" :key="item.value">
              {{item.label}}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each" v-if="showInstance">
          <label class="col-md-2 label-name lable-name-select">{{$t('field.instance')}}:</label>
          <input v-model="endpointRejectModel.addRow.instance" type="text" class="col-md-7 form-control model-input  c-dark">
          <label class="required-tip">*</label>
        </div>
      </div>
    </ModalComponent>
    <ModalComponent :modelConfig="processConfigModel">
      <div slot="processConfig">
        <div class="marginbottom params-each">
          <label class="col-md-1 label-name">{{$t('tableKey.condition')}}:</label>
          <div class="search-input-content">
            <input type="text" v-model="processConfigModel.processName" class="search-input c-dark" />
          </div>
          <button type="button" @click="addProcess" class="btn-cancle-f" style="vertical-align:middle">{{$t('button.confirm')}}</button>
        </div>
        <div class="marginbottom params-each row" style="">
          <div class="offset-md-1">
            <Tag
            v-for="(process, processIndex) in processConfigModel.addRow.processSet"
            color="primary"
            type="border"
            :key="processIndex"
            :name="processIndex"
            closable
            @on-close="delProcess(process)"
            >{{process|interceptParams}}
              <i class="fa fa-pencil" @click="editProcess(process)" aria-hidden="true"></i>
            </Tag>  
          </div>     
        </div>
      </div>
    </ModalComponent>
    <ModalComponent :modelConfig="businessConfigModel">
      <div slot="businessConfig">
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('tableKey.logPath')}}:</label>
          <div class="search-input-content">
            <input type="text" v-model="businessConfigModel.businessName" class="search-input c-dark" />
          </div>
          <button type="button" @click="addBusiness" class="btn-cancle-f" style="vertical-align:middle">{{$t('button.confirm')}}</button>
        </div>
        <div class="marginbottom params-each row" style="">
          <div class="offset-md-1">
            <Tag
            v-for="(business, businessIndex) in businessConfigModel.addRow.businessSet"
            color="primary"
            type="border"
            :key="businessIndex"
            :name="businessIndex"
            closable
            @on-close="delBusiness(business)"
            >{{business|interceptParams}}
              <i class="fa fa-pencil" @click="editBusiness(business)" aria-hidden="true"></i>
            </Tag>  
          </div>     
        </div>
      </div>
    </ModalComponent>
  </div>
</template>
<script>
  import tableTemp from '@/components/table-page/table'
  import {interceptParams} from '@/assets/js/utils'
  let tableEle = [
    {title: 'tableKey.endpoint', value: 'guid', display: true},
    {title: 'tableKey.group', value: 'groups_name', display: true, }
  ]
  let historyAlarmEle = [
    {title: 'tableKey.status',value: 'status', style: 'min-width:70px', display: true},
    {title: 'tableKey.s_metric',value: 's_metric', display: true},
    {title: 'tableKey.start_value',value: 'start_value', display: true},
    {title: 'tableKey.s_cond',value: 's_cond', style: 'min-width:70px', display: true},
    {title: 'tableKey.s_last',value: 's_last', style: 'min-width:65px', display: true},
    {title: 'tableKey.s_priority',value: 's_priority', display: true},
    {title: 'tableKey.start',value: 'start_string', style: 'min-width:200px', display: true},
    {title: 'tableKey.end',value: 'end_string', style: 'min-width:200px',display: true,
      'render': (item) => {
        if (item.end_string === '0001-01-01 00:00:00') {
          return '-'
        } else {
          return item.end_string
        }
      }
    }]
  const btn = [
    {btn_name: 'button.thresholdManagement', btn_func: 'thresholdConfig'},
    {btn_name: 'button.historicalAlert', btn_func: 'historyAlarm'},
    {btn_name: 'button.remove', btn_func: 'delF'},
    {btn_name: 'button.logConfiguration', btn_func: 'logManagement'},
    {btn_name: 'button.processConfiguration', btn_func: 'processManagement'},
    {btn_name: 'button.businessConfiguration', btn_func: 'businessManagement'},
  ]
  export default {
    name: '',
    data() {
      return {
        pageConfig: {
          CRUD: this.$root.apiCenter.endpointManagement.list.api,
          researchConfig: {
            input_conditions: [
              {value: 'search', type: 'input', placeholder: 'placeholder.input', style: ''}],
            btn_group: [
              {btn_name: 'button.search', btn_func: 'search', class: 'btn-confirm-f', btn_icon: 'fa fa-search'}
            ],
            filters: {
              search: ''
            }
          },
          table: {
            tableData: [],
            tableEle: tableEle,
            filterMoreBtn: 'filterMoreBtn',
            primaryKey: 'guid',
            btn: btn,
            pagination: this.pagination,
            handleFloat:true,
          },
          pagination: {
            __orders: '-created_date',
            total: 0,
            page: 1,
            size: 10
          }
        },
        historyAlarmPageConfig: {
          table: {
            tableData: [],
            tableEle: historyAlarmEle,
            btn: [],
          },
        },
        modelConfig: {
          modalId: 'add_object_Modal',
          modalTitle: 'button.add',
          isAdd: true,
          config: [
            {name:'advancedConfig',type:'slot'}
          ],
          addRow: { // [通用]-保存用户新增、编辑时数据
            name: null,
            description: null,
          },
          slotConfig: {
            resourceSelected: [],
            resourceOption: []
          }
        },
        historyAlarmModel: {
          modalId: 'history_alarm_Modal',
          modalTitle: 'button.historicalAlert',
          modalStyle: 'width:930px;max-width: none;',
          noBtn: true,
          isAdd: true,
          config: [
            {name:'historyAlarm',type:'slot'}
          ],
          pageConfig: {
            table: {
              tableData: [],
              tableEle: tableEle
            }
          },
        },
        endpointRejectModel: {
          modalId: 'endpoint_reject_model',
          modalTitle: 'title.endpointAdd',
          isAdd: true,
          saveFunc: 'endpointRejectSave',
          config: [
            {name:'endpointReject',type:'slot'},
            {label: 'field.ip', value: 'exporter_ip', placeholder: 'tips.required', v_validate: 'required:true|isIP', disabled: false, type: 'text'},
            {label: 'field.port', value: 'exporter_port', placeholder: 'tips.required', v_validate: 'required:true|isNumber', disabled: false, type: 'text'},
          ],
          addRow: {
            instance: '',
            type: 'host',
            exporter_ip: null,
            exporter_port: 9100,
          },
          endpointType: [
            {label:'host',value:'host'},
            {label:'mysql',value:'mysql'},
            {label:'redis',value:'redis'},
            {label:'tomcat',value:'tomcat'}
          ],
        }, 
        processConfigModel: {
          modalId: 'process_config_model',
          modalTitle: 'button.processConfiguration',
          isAdd: true,
          saveFunc: 'processConfigSave',
          config: [
            {name:'processConfig',type:'slot'}
          ],
          addRow:{
            processSet: [],
          },
          processName: ''
        },
        businessConfigModel: {
          modalId: 'business_config_model',
          modalTitle: 'button.businessConfiguration',
          isAdd: true,
          saveFunc: 'businessConfigSave',
          config: [
            {name:'businessConfig',type:'slot'}
          ],
          addRow:{
            businessSet: [],
          },
          businessName: ''
        },
        id: null,
        showGroupMsg: false,
        groupMsg: {}
      }
    },
    mounted() {
      if (this.$root.$validate.isEmpty_reset(this.$route.params)) {
        this.groupMsg = {}
        this.showGroupMsg = false
        this.pageConfig.researchConfig.btn_group.push({btn_name: 'button.add', btn_func: 'endpointReject', class: 'btn-cancle-f', btn_icon: 'fa fa-plus'})
      } else {
        this.$parent.activeTab = '/monitorConfigIndex/endpointManagement'
        if (this.$route.params.hasOwnProperty('group')) {
          this.groupMsg = this.$route.params.group
          this.showGroupMsg = true
          this.pageConfig.researchConfig.btn_group.push({btn_name: 'button.add', btn_func: 'add', class: 'btn-cancle-f', btn_icon: 'fa fa-plus'})
          this.pageConfig.researchConfig.filters.grp = this.groupMsg.id
        }
        if (this.$route.params.hasOwnProperty('search')) {
          this.pageConfig.researchConfig.filters.search = this.$route.params.search
        }
      }
      this.initData(this.pageConfig.CRUD, this.pageConfig)
    },
    watch: {
      'endpointRejectModel.addRow.type':function(val){
        const typeToPort = {
          host: 9100,
          mysql: 9104,
          redis: 9121,
          tomcat: 9151,
        }
        this.endpointRejectModel.addRow.exporter_port = typeToPort[val]
      }
    },
    computed:{
      showInstance: function(){
        return this.endpointRejectModel.addRow.type === 'host' ? false: true
      }
    },
    filters: {
      interceptParams (val) {
        return interceptParams(val,55)
      }
    },
    methods: {
      initData (url= this.pageConfig.CRUD, params) {
        this.$root.$tableUtil.initTable(this, 'GET', url, params)
      },
      filterMoreBtn (rowData) {
        let moreBtnGroup = ['thresholdConfig','historyAlarm','logManagement']
        if (rowData.type === 'host') {
          moreBtnGroup.push('processManagement', 'businessManagement')
        }
        if (this.showGroupMsg) {
          moreBtnGroup.push('delF')
        }
        return moreBtnGroup
      },
      add () {
        this.modelConfig.slotConfig.resourceOption = []
        this.modelConfig.slotConfig.resourceSelected = []
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.resourceSearch.api, {search: '.'}, responseData => {
          this.modelConfig.slotConfig.resourceOption = responseData
        })
        this.$root.JQ('#add_object_Modal').modal('show')
      },
      addPost() {
        let params = {
          grp: this.groupMsg.id,
          endpoints: this.modelConfig.slotConfig.resourceSelected,
          operation: 'add'
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.update.api, params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#add_object_Modal').modal('hide')
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      },
      delF (rowData) {
        this.$parent.$parent.delConfirm({name: rowData.guid}, () => {
          let endpoints = []
          this.pageConfig.table.tableData.forEach((item)=>{
            endpoints.push(item.guid.split(':')[0])
          })
          let params = {
            grp: this.groupMsg.id,
            endpoints: [parseInt(rowData.id)],
            operation: 'delete'
          }
          this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.update.api, params, () => {
            this.$Message.success(this.$t('tips.success'))
            this.initData(this.pageConfig.CRUD, this.pageConfig)
          })
        })
      },
      thresholdConfig (rowData) {
        this.$router.push({name: 'thresholdManagement', params: {id: rowData.id, type: 'endpoint'}})
      },
      logManagement (rowData) {
        this.$router.push({name: 'logManagement', params: {id: rowData.id, type: 'endpoint'}})
      },
      closeTag () {
        this.groupMsg = {}
        this.showGroupMsg = false
        this.pageConfig.researchConfig.filters.grp = ''
        this.pageConfig.table.btn.splice(this.pageConfig.table.btn.length-1, 1)
        this.pageConfig.researchConfig.btn_group.splice(this.pageConfig.researchConfig.btn_group.length-1, 1)
        this.pageConfig.researchConfig.btn_group.push({btn_name: 'button.add', btn_func: 'endpointReject', class: 'btn-cancle-f', btn_icon: 'fa fa-plus'})
        this.initData(this.pageConfig.CRUD, this.pageConfig)
      },
      historyAlarm (rowData) {
        let params = {id: rowData.id}
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.alarm.history, params, (responseData) => {
          this.historyAlarmPageConfig.table.tableData = responseData
        })
        this.$root.JQ('#history_alarm_Modal').modal('show')
      },
      endpointReject () {
        this.endpointRejectModel.addRow.type = 'host'
        this.$root.JQ('#endpoint_reject_model').modal('show')
      },
      endpointRejectSave () {
        this.endpointRejectModel.addRow.exporter_port += ''
        let params= this.$root.$validate.isEmptyReturn_JSON(this.endpointRejectModel.addRow)
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.register.api, params, () => {
          this.$root.$validate.emptyJson(this.endpointRejectModel.addRow)
          this.$root.JQ('#endpoint_reject_model').modal('hide')
          this.$Message.success(this.$t('tips.success'))
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      },
      processManagement (rowData) {
        this.id = rowData.id
        this.processConfigModel.addRow.processSet = []
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET','alarm/process/list', {id:this.id}, responseData=> {
          responseData.forEach((item)=>{
            this.processConfigModel.addRow.processSet.push(item.name)
          })
          this.$root.JQ('#process_config_model').modal('show')
        })
      },
      processConfigSave () {
        const params = {
          endpoint_id: +this.id,
          process_list: this.processConfigModel.addRow.processSet
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST','alarm/process/update', params, ()=> {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#process_config_model').modal('hide')
        })
      },
      addProcess () {
        if (!this.$root.$validate.isEmpty_reset(this.processConfigModel.processName.trim())) {
          this.processConfigModel.addRow.processSet.push(this.processConfigModel.processName.trim())
          this.processConfigModel.processName = ''
        }
      },
      delProcess (process) {
        const i = this.processConfigModel.addRow.processSet.findIndex((val)=>{
          return val === process
        })
        this.processConfigModel.addRow.processSet.splice(i,1)
      },
      editProcess (process) {
        this.delProcess(process)
        this.processConfigModel.processName = process
      },

      businessManagement (rowData) {
        this.id = rowData.id
        this.businessConfigModel.addRow.businessSet = []
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET','alarm/business/list', {id:this.id}, responseData=> {
          responseData.forEach((item)=>{
            this.businessConfigModel.addRow.businessSet.push(item.path)
          })
          this.$root.JQ('#business_config_model').modal('show')
        })
      },
      businessConfigSave () {
        const params = {
          endpoint_id: +this.id,
          path_list: this.businessConfigModel.addRow.businessSet
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST','alarm/business/update', params, ()=> {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#business_config_model').modal('hide')
        })
      },
      addBusiness () {
        if (!this.$root.$validate.isEmpty_reset(this.businessConfigModel.businessName.trim())) {
          this.businessConfigModel.addRow.businessSet.push(this.businessConfigModel.businessName.trim())
          this.businessConfigModel.businessName = ''
        }
      },
      delBusiness (business) {
        const i = this.businessConfigModel.addRow.businessSet.findIndex((val)=>{
          return val === business
        })
        this.businessConfigModel.addRow.businessSet.splice(i,1)
      },
      editBusiness (business) {
        this.delBusiness(business)
        this.businessConfigModel.businessName = business
      }

    },
    components: {
      tableTemp
    }
  }
</script>

<style lang="less" scoped>
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
    cursor: text;
    
    width: 310px;
  }

  .search-input:focus {
    outline: 0;
    border-color: #57a3f3;
  }

  .search-input-content {
    display: inline-block;
    vertical-align: middle; 
  }
</style>

