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
        <Tag color="blue" :key="tableIndex + 'a'" v-if="tableItem.obj_name">{{tableItem.obj_name}}</Tag>
        <button @click="add()" type="button" v-if="tableItem.operation" class="btn btn-sm btn-cancle-f" :key="tableIndex + 'b'">
          <i class="fa fa-plus"></i>
          新增
        </button>
        <PageTable :pageConfig="tableItem" :key="tableIndex + 'c'"></PageTable>
      </template>
      <ModalComponent :modelConfig="modelConfig">
        <div slot="thresholdConfig" class="extentClass">   
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name lable-name-select">阀值:</label>
            <Select v-model="modelConfig.threshold" style="width:100px">
              <Option v-for="item in modelConfig.thresholdList" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
            <div class="search-input-content">
              <input v-model="modelConfig.thresholdValue" type="text" class="search-input" />
            </div>
          </div>
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name lable-name-select">持续时间:</label>
            <div class="search-input-content">
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
import Searchinput from '@/components/components/Search-input'
let tableEle = [
  {title: 'ID', value: 'id', display: false},
  {title: '名称', value: 'metric', display: true},
  {title: '表达式', value: 'expr', display: true},
  {title: '阀值', value: 'cond', display: true},
  {title: '持续时长', value: 'last', display: true},
  {title: '优先级', value: 'priority', display: true}
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
        api: '/alarm/strategy/search',
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
        }
      },
      modelTip: {
        key: '',
        value: ''
      },
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: '阀值管理',
        isAdd: true,
        config: [
          {label: '名称', value: 'metric', placeholder: '必填,2-60字符', v_validate: 'required:true|min:2|max:60', disabled: false, type: 'text'},
          {label: '表达式', value: 'expr', placeholder: '必填,2-60字符', v_validate: 'required:true|min:2|max:60', disabled: false, type: 'text'},
          {label: '通知内容', value: 'content', placeholder: '必填,2-60字符', v_validate: 'required:true|min:2|max:60', disabled: false, type: 'textarea'},
          {name:'thresholdConfig',type:'slot'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          metric: null,
          expr: null,
          content: null,
        },
        threshold: '>',
        thresholdList: [
          {label: '>', value: '>'},
          {label: '>=', value: '>='},
          {label: '<', value: '<'},
          {label: '<=', value: '<='},
          {label: '==', value: '=='},
          {label: '!=', value: '!='}
        ],
        thresholdValue: '',
        last: 's',
        lastList: [
          {label: 'sec', value: 's'},
          {label: 'min', value: 'm'},
          {label: 'hour', value: 'h'}
        ],
        lastValue: '',
        priority: 'low',
        priorityList: [
          {label: 'high', value: 'high'},
          {label: 'medium', value: 'medium'},
          {label: 'low', value: 'low'}
        ],
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
      this.$parent.activeTab = '/monitorConfigIndex/thresholdManagement'
      this.type = this.$route.params.type
      console.log(this.type)
      this.searchInputConfig.params.type = this.$route.params.type
      this.typeValue = this.$route.params.id
      this.requestData(this.type, this.typeValue)
    } else {
      this.type = 'endpoint'
      this.searchInputConfig.params.type = 'endpoint'
    }
  },
  methods: {
    search () {
      this.requestData(this.searchInputConfig.params.type, this.$store.state.ip.value)
    },
    requestData (type, id) {
      let params = {}
      params.type = type
      params.id = id
      this.totalPageConfig = []
      this.$httpRequestEntrance.httpRequestEntrance('GET', 'alarm/strategy/list', params, (responseData) => {
        responseData.forEach((item)=>{
          let config = this.$validate.deepCopy(this.pageConfig.table)
          if (!item.operation) {
            config.btn = []
          }
          config.tableData = item.strategy
          // this.$nextTick(() => {
          this.totalPageConfig.push({table:config, obj_name: item.obj_name, operation:item.operation})
          // }) 
        })
      })
    },
    delF (rowData) {
      let params = {id: rowData.id}
      this.$httpRequestEntrance.httpRequestEntrance('GET', 'alarm/strategy/delete', params, () => {
        this.$Message.success('删除成功 !')
        this.requestData(this.type, this.typeValue)
      })
    },
    add () {
      this.modelConfig.isAdd = true
      this.JQ('#add_edit_Modal').modal('show')
    },
    addPost () {
      if (this.$validate.isEmpty_reset(this.modelConfig.thresholdValue)) {
        this.$Message.warning('阀值不能为空')
        return
      }
      if (this.$validate.isEmpty_reset(this.modelConfig.lastValue)) {
        this.$Message.warning('持续时间不能为空')
        return
      }
      if (this.$validate.isEmpty_reset(this.modelConfig.addRow.content)) {
        this.$Message.warning('通知内容不能为空')
        return
      }
      let ss = {
        cond: this.modelConfig.threshold + this.modelConfig.thresholdValue,
        last: this.modelConfig.lastValue + this.modelConfig.last,
        priority: this.modelConfig.priority,        
      }
      if (this.type === 'grp') {
        ss.grp_id = this.typeValue
        ss.endpoint_id = 0
      } else {
        ss.endpoint_id = this.typeValue
        ss.grp_id = 0
      }
      let params = Object.assign(ss, this.modelConfig.addRow)
      this.$httpRequestEntrance.httpRequestEntrance('POST', 'alarm/strategy/add', params, () => {
        this.$Message.success('新增成功 !')
        this.JQ('#add_edit_Modal').modal('hide')
        this.requestData(this.type, this.typeValue)
      })
    },
    editF (rowData) {
      this.id = rowData.id
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
      this.modelTip.value = rowData.name
      this.modelConfig.isAdd = false
      this.JQ('#add_edit_Modal').modal('show')
    },
    editPost () {
      if (this.$validate.isEmpty_reset(this.modelConfig.thresholdValue)) {
        this.$Message.warning('阀值不能为空')
        return
      }
      if (this.$validate.isEmpty_reset(this.modelConfig.lastValue)) {
        this.$Message.warning('持续时间不能为空')
        return
      }
      if (this.$validate.isEmpty_reset(this.modelConfig.addRow.content)) {
        this.$Message.warning('通知内容不能为空')
        return
      }
      let ss = {
        cond: this.modelConfig.threshold + this.modelConfig.thresholdValue,
        last: this.modelConfig.lastValue + this.modelConfig.last,
        priority: this.modelConfig.priority,   
        strategy_id: this.id  
      }
      if (this.type === 'grp') {
        ss.grp_id = this.typeValue
        ss.endpoint_id = 0
      } else {
        ss.endpoint_id = this.typeValue
        ss.grp_id = 0
      }
      let params = Object.assign(ss, this.modelConfig.addRow)
      this.$httpRequestEntrance.httpRequestEntrance('POST', 'alarm/strategy/update', params, () => {
        this.$Message.success('编辑成功 !')
        this.JQ('#add_edit_Modal').modal('hide')
        this.requestData(this.type, this.typeValue)
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
  .search-input:focus {
    outline: 0;
    border-color: #57a3f3;
  }

  .search-input-content {
    display: inline-block;
    vertical-align: middle; 
  }
</style>

