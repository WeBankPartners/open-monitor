<template>
  <div class=" ">
    <section>
      <ul class="search-ul">
        <li class="search-li">
          <Select v-model="type" style="width:100px">
            <Option v-for="item in typeList" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
        </li>
        <li class="search-li">
          <Searchinput :parentConfig="searchInputConfig"></Searchinput> 
        </li>
        <li class="search-li">
          <button type="button" class="btn btn-sm btn-confirm-f"
          @click="search">
            <i class="fa fa-search" ></i>
            搜索
          </button>
        </li>
      </ul> 
    </section>
    <section>
      <template v-for="(tableItem, tableIndex) in totalPageConfig">
        <div :key="tableIndex + 'f'" class="section-table-tip">
          <Tag color="blue" :key="tableIndex + 'a'" v-if="tableItem.obj_name">{{tableItem.obj_name}}</Tag>
          <button @click="add(tableItem.obj_type)" type="button" v-if="tableItem.operation" class="btn btn-sm btn-cancle-f" :key="tableIndex + 'b'">
            <i class="fa fa-plus"></i>
            新增
          </button>
        </div>
        <PageTable :pageConfig="tableItem" :key="tableIndex + 'c'"></PageTable>
      </template>
      <ModalComponent :modelConfig="modelConfig">
        <div slot="thresholdConfig" class="extentClass">  
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name lable-name-select">条件:</label>
            <Select v-model="modelConfig.cond" style="width:100px">
              <Option v-for="item in modelConfig.condList" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
            <div class="search-input-content" style="margin-left: 8px">
              <input v-model="modelConfig.condValue" type="text" class="search-input" />
            </div>
          </div>
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name lable-name-select">时间范围:</label>
            <div class="search-input-content" style="margin-right: 8px">
              <input v-model="modelConfig.lastValue" type="text" class="search-input" />
            </div>
            <Select v-model="modelConfig.last" style="width:100px">
              <Option v-for="item in modelConfig.lastList" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </div>
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name lable-name-select">优先级:</label>
            <Select v-model="modelConfig.priority" style="width:100px">
              <Option v-for="item in modelConfig.priorityList" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </div>
        </div>
      </ModalComponent>
    </section>
  </div>
</template>

<script>
import {thresholdList, lastList, priorityList} from '@/assets/config/common-config.js'
import Searchinput from '@/components/components/Search-input'
let tableEle = [
  {title: '路径', value: 'path', display: true}
]
const btn = [
  {btn_name: '编辑', btn_func: 'editF'},
  {btn_name: '删除', btn_func: 'delF'},
]

export default {
  name: '',
  data() {
    return {
      type: '',
      typeValue: 'endpoint',
      typeList: [
        {label: '主机', value: 'endpoint'},
        {label: '组', value: 'grp'}
      ],
      searchInputConfig: {
        poptipWidth: 500,
        placeholder: '模糊匹配',
        inputStyle: "width:500px;",
        api: this.apiCenter.resourceSearch.strategyApi,
        params: {
          type: null
        }
      },
      inputValue: '',
      totalPageConfig: [],
      pageConfig: {
        table: {
          tableData: [],
          tableEle: tableEle,
          // filterMoreBtn: 'filterMoreBtn',
          primaryKey: 'id',
          btn: btn,
          handleFloat:true,
          isExtend:{
            func:'getExtendInfo',
            data:{},
            slot:'tableExtend',
            detailConfig:[{
              isExtendF:true,
              title: '磁盘列表',
              config:[
              {title: '磁盘名', value: 'name', display: true},
              {title: '磁盘类型', value: 'volume_type', display: true},
              {title: '设备', value: 'device', display: true},
              {title: '大小(GB)', value: 'size_gb', display: true},
              {title: '状态', value: 'status_state', display: true},
              {title: '创建时间', value: 'created_date', display: true}
              ],
              data:[],
              scales: ['25%', '20%', '10%', '15%', '10%', '20%']
            }]
          }
        }
      },
      modelTip: {
        key: '',
        value: 'metric'
      },
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: '阀值管理',
        isAdd: true,
        config: [
          {label: '路径', value: 'path', placeholder: '必填', v_validate: 'required:true', disabled: false, type: 'text'},
          {label: '关键字', value: 'keyword', placeholder: '必填', v_validate: 'required:true', disabled: false, type: 'text'},
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
    }
  },
  watch: {
    type: function (val) {
      this.searchInputConfig.params.type = val
    }
  },
  mounted () {
    if (!this.$validate.isEmpty_reset(this.$route.params)) {
      this.$parent.activeTab = '/monitorConfigIndex/logManagement'
      this.type = this.$route.params.type
      this.searchInputConfig.params.type = this.$route.params.type
      this.typeValue = this.$route.params.id
      this.requestData(this.type, this.typeValue)
    } else {
      this.type = 'endpoint'
      this.typeValue = ''
      this.searchInputConfig.params.type = 'endpoint'
    }
    this.JQ('#add_edit_Modal').on('hidden.bs.modal', () => {
      this.modelConfig.thresholdValue = ''
      this.modelConfig.lastValue = ''
    })
  },
  methods: {
    search () {
      this.typeValue = this.$store.state.ip.id
      this.requestData(this.searchInputConfig.params.type, this.$store.state.ip.id)
    },
    requestData (type, id) {
      let params = {}
      params.type = type
      params.id = id
      this.totalPageConfig = []
      this.$httpRequestEntrance.httpRequestEntrance('GET', this.apiCenter.logManagement.list.api, params, (responseData) => {
        responseData.forEach((item)=>{
          let config = this.$validate.deepCopy(this.pageConfig.table)
          if (!item.operation) {
            config.btn = []
          }
          config.tableData = item.log_monitor
          this.totalPageConfig.push({table:config, obj_type: item.obj_type, obj_name: item.obj_name, operation:item.operation})
        })
      })
    },
    delF (rowData) {
      let params = {id: rowData.id}
      this.$httpRequestEntrance.httpRequestEntrance('GET', this.apiCenter.logManagement.delete.api, params, () => {
        this.$Message.success('删除成功 !')
        this.requestData(this.type, this.typeValue)
      })
    },
    formValidate () {
      if (this.$validate.isEmpty_reset(this.modelConfig.condValue)) {
        this.$Message.warning('条件不能为空！')
        return false 
      }
      if (this.$validate.isEmpty_reset(this.modelConfig.lastValue)) {
        this.$Message.warning('持续时间不能为空！')
        return false 
      }
      return true
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
    add () {
      this.modelConfig.isAdd = true
      this.JQ('#add_edit_Modal').modal('show')
    },
    addPost () {
      if (!this.formValidate()) {
        return
      }
      let params = this.paramsPrepare()
      this.$httpRequestEntrance.httpRequestEntrance('POST', this.apiCenter.logManagement.add.api, params, () => {
        this.$Message.success('新增成功 !')
        this.JQ('#add_edit_Modal').modal('hide')
        this.requestData(this.type, this.typeValue)
      })
    },
    editF (rowData) {
      this.modelConfig.isAdd = false
      this.id = rowData.id
      this.modelTip.value = rowData.metric
      this.modelConfig.addRow = this.$tableUtil.manageEditParams(this.modelConfig.addRow, rowData)
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
      this.JQ('#add_edit_Modal').modal('show')
    },
    editPost () {
      if (!this.formValidate()) {
        return
      }
      let params = this.paramsPrepare()
      params.strategy_id = this.id
      this.$httpRequestEntrance.httpRequestEntrance('POST', this.apiCenter.logManagement.update.api, params, () => {
        this.$Message.success('编辑成功 !')
        this.JQ('#add_edit_Modal').modal('hide')
        this.requestData(this.type, this.typeValue)
      })
    },
    selectMetric (option) {
      if (option) {
        this.modelConfig.addRow.metric = option.label
      }
    },
    getExtendInfo(item){
        this.pageConfig.table.isExtend.detailConfig[0].data = []
        this.pageConfig.table.isExtend.detailConfig[1].data = []
        this.instance = item
        this.$httpRequestEntrance.httpRequestEntrance('GET', this.apiCenter.manage.ECS.ecs_manage.CRUD + '/' + item.id, '', res => {
          if(res){
            let data = res.volumes
            let state_mapping = {available: '空闲', using: '已挂载', error: '错误'}
            if(res.volumes.length>0){
              for(let i = 0, len = res.volumes.length; i < len; i++){
                let item = res.volumes[i]
                data[i].name = item.volume.name
                data[i].size_gb = item.volume.size_gb
                data[i].status_state = state_mapping[item.volume.status]
                data[i].detailId = item.volume.id
              }
            }
            this.pageConfig.table.isExtend.detailConfig[0].data =  data
            if(item.enis){
              this.pageConfig.table.isExtend.detailConfig[1].data= item.enis.map(item =>{
                item.subnetName = item.subnet.name
                item.userNickname = item.user.nickname
                return item
              })
            }
          }
        })
      }
  },
  components: {
    Searchinput
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
</style>