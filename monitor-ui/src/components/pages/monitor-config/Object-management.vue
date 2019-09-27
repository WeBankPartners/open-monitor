<template>
  <div class="main-content">
    <div v-if="showGroupMsg" style="padding-left:20px">
      <Tag type="border" closable color="primary" @on-close="closeTag">当前组:{{groupMsg.name}}</Tag>
    </div>
    <PageTable :pageConfig="pageConfig"></PageTable>
    <ModalComponent :modelConfig="modelConfig">
      <div slot="advancedConfig" class="extentClass">   
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name lable-name-select">对象名:</label>
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
  </div>
</template>
<script>
  import tableTemp from '@/components/components/table-page/table'
  let tableEle = [
    {title: 'guid', value: 'guid', display: true},
    {title: 'groups', value: 'groups_name', display: true, }
  ]
  let historyAlarmEle = [
    {title: '状态',value: 'status', style: 'min-width:70px', display: true},
    {title: '指标',value: 's_metric', display: true},
    {title: '异常值',value: 'start_value', display: true},
    {title: '阀值',value: 's_cond', style: 'min-width:70px', display: true},
    {title: '持续时间',value: 's_last', style: 'min-width:65px', display: true},
    {title: '级别',value: 's_priority', display: true},
    {title: '开始时间',value: 'start', style: 'min-width:200px', display: true},
    {title: '结束时间',value: 'end', style: 'min-width:200px',display: true,
    'render': (item) => {
      if (item.end.indexOf('0001')>-1) {
        return '-'
      } else {
        return item.end
      }
    }
    }]
  const btn = [
    {btn_name: '阀值配置', btn_func: 'thresholdConfig'},
    {btn_name: '历史告警', btn_func: 'historyAlarm'},
    {btn_name: '删除', btn_func: 'delF'}
  ]
  export default {
    name: '',
    data() {
      return {
        model10: [],
        pageConfig: {
          CRUD: this.apiCenter.objectManagement.list.api,
          researchConfig: {
            input_conditions: [
              {value: 'search', type: 'input', placeholder: '请输入', style: ''}],
            btn_group: [
              {btn_name: '搜索', btn_func: 'search', class: 'btn-confirm-f', btn_icon: 'fa fa-search'},
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
          modalTitle: '对象管理',
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
          modalTitle: '对象管理',
          modalStyle: 'width:930px;max-width: none;',
          noBtn: true,
          isAdd: true,
          config: [
            {name:'historyAlarm',type:'slot'}
          ],
          columns1: [
            {
                title: '状态',
                key: 'status'
            },
            {
                title: '指标',
                key: 's_metric'
            },
            {
                title: '异常值',
                key: 'start_value'
            },
            {
                title: '阀值',
                key: 's_cond'
            },
            {
                title: '持续时间',
                key: 's_last'
            },
            {
                title: '级别',
                key: 's_priority'
            },
            {
                title: '开始时间',
                key: 'start'
            },
            {
                title: '结束时间',
                key: 'end'
            }
          ],
          data2: [],
          pageConfig: {
            table: {
              tableData: [],
              tableEle: tableEle
            }
          },
        },
        id: null,
        showGroupMsg: false,
        groupMsg: {}
      }
    },
    mounted() {
      if (this.$validate.isEmpty_reset(this.$route.params)) {
        this.groupMsg = {}
        this.showGroupMsg = false
      } else {
        this.$parent.activeTab = '/monitorConfigIndex/objectManagement'
        if (this.$route.params.hasOwnProperty('group')) {
          this.groupMsg = this.$route.params.group
          this.showGroupMsg = true
          this.pageConfig.researchConfig.btn_group.push({btn_name: '新增', btn_func: 'add', class: 'btn-cancle-f', btn_icon: 'fa fa-plus'})
          this.pageConfig.researchConfig.filters.grp = this.groupMsg.id
        }
        if (this.$route.params.hasOwnProperty('search')) {
          this.pageConfig.researchConfig.filters.search = this.$route.params.search
        }
      }
      this.initData(this.pageConfig.CRUD, this.pageConfig)
    },
    methods: {
      initData (url= this.pageConfig.CRUD, params) {
        this.$tableUtil.initTable(this, 'GET', url, params)
      },
      filterMoreBtn () {
        let moreBtnGroup = ['thresholdConfig','historyAlarm']
        if (this.showGroupMsg) {
          moreBtnGroup.push('delF')
        }
        return moreBtnGroup
      },
      add () {
        this.modelConfig.slotConfig.resourceOption = []
        this.modelConfig.slotConfig.resourceSelected = []
        this.$httpRequestEntrance.httpRequestEntrance('GET',this.apiCenter.resourceSearch.api, {search: '.'}, responseData => {
          this.modelConfig.slotConfig.resourceOption = responseData
        })
        this.JQ('#add_object_Modal').modal('show')
      },
      addPost() {
        if (this.$validate.isEmpty_reset(this.modelConfig.slotConfig.resourceSelected)) {
          this.$Message.warning('请先选择要新增的对象 !')
        }
        let params = {
          grp: this.groupMsg.id,
          endpoints: this.modelConfig.slotConfig.resourceSelected,
          operation: 'add'
        }
        this.$httpRequestEntrance.httpRequestEntrance('POST', this.apiCenter.objectManagement.update.api, params, () => {
          this.$Message.success('新增成功 !')
          this.JQ('#add_object_Modal').modal('hide')
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      },
      delF (rowData) {
        let endpoints = []
        this.pageConfig.table.tableData.forEach((item)=>{
           endpoints.push(item.guid.split(':')[0])
        })
        let params = {
          grp: this.groupMsg.id,
          endpoints: [parseInt(rowData.id)],
          operation: 'delete'
        }
        this.$httpRequestEntrance.httpRequestEntrance('POST', this.apiCenter.objectManagement.update.api, params, () => {
          this.$Message.success('删除成功 !')
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      },
      thresholdConfig (rowData) {
        this.$router.push({name: 'thresholdManagement', params: {id: rowData.id, type: 'endpoint'}})
      },
      closeTag () {
        this.groupMsg = {}
        this.showGroupMsg = false
        this.pageConfig.researchConfig.filters.grp = ''
        this.pageConfig.table.btn.splice(this.pageConfig.table.btn.length-1, 1)
        this.pageConfig.researchConfig.btn_group.splice(this.pageConfig.researchConfig.btn_group.length-1, 1)
        this.initData(this.pageConfig.CRUD, this.pageConfig)
      },
      historyAlarm (rowData) {
        let params = {id: rowData.id}
        this.$httpRequestEntrance.httpRequestEntrance('GET', 'alarm/history', params, (responseData) => {
          this.historyAlarmPageConfig.table.tableData = responseData
        })
        this.JQ('#history_alarm_Modal').modal('show')
      }
    },
    components: {
      tableTemp
    }
  }
</script>

<style lang="less" scoped>
</style>
