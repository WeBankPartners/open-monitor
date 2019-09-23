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
          <Select v-model="modelConfig.slotConfig.resourceSelected" multiple style="width:260px">
              <Option v-for="item in modelConfig.slotConfig.resourceOption" :value="item.id" :key="item.id">{{ item.option_text }}</Option>
          </Select>
        </div>
      </div>
    </ModalComponent>
  </div>
</template>
<script>
  let tableEle = [
    {title: 'guid', value: 'guid', display: true, editable: 'editOnline'},
    {title: 'groups', value: 'groups', display: true, frozen: true, sortable: true}
  ]
  const btn = [
    {btn_name: '阀值配置', btn_func: 'thresholdConfig'},
    {btn_name: '历史告警', btn_func: 'xx'},
  ]
  export default {
    name: '',
    data() {
      return {
        model10: [],
        pageConfig: {
          CRUD: 'alarm/endpoint/list',
          researchConfig: {
            input_conditions: [
              {value: 'search', type: 'input', placeholder: '请输入', style: ''}],
            btn_group: [
              {btn_name: '搜索', btn_func: 'search', class: 'btn-confirm-f', btn_icon: 'fa fa-search'},
            ],
            filters: {
              name__icontains: '',
              cmdb_tenant_id__icontains: ''
            }
          },
          table: {
            tableData: [],
            tableEle: tableEle,
            // filterMoreBtn: 'filterMoreBtn',
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
        this.groupMsg = this.$route.params.group
        this.showGroupMsg = true
        this.pageConfig.table.btn.push({btn_name: '删除', btn_func: 'delF'})
        this.pageConfig.researchConfig.btn_group.push({btn_name: '新增', btn_func: 'add', class: 'btn-cancle-f', btn_icon: 'fa fa-plus'})
        this.pageConfig.researchConfig.filters.grp = this.groupMsg.id
      }
      this.initData(this.pageConfig.CRUD, this.pageConfig)
    },
    methods: {
      initData (url= this.pageConfig.CRUD, params) {
        this.$tableUtil.initTable(this, 'GET', url, params)
      },
      add () {
        this.modelConfig.slotConfig.resourceOption = []
        this.modelConfig.slotConfig.resourceSelected = []
        this.$httpRequestEntrance.httpRequestEntrance('GET','/dashboard/search', {search: '.'}, responseData => {
          console.log(responseData)
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
        this.$httpRequestEntrance.httpRequestEntrance('POST', 'alarm/endpoint/update', params, () => {
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
        this.$httpRequestEntrance.httpRequestEntrance('POST', 'alarm/endpoint/update', params, () => {
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
      }
    },
    components: {
    }
  }
</script>

<style lang="less" scoped>
</style>
