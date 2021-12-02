<template>
  <div class=" ">
    <section v-if="showManagement" style="margin-top: 16px;">
      <div v-for="table in totalData" :key="table.guid">
        <Tag color="blue">{{table.tableData.display_name}}</Tag>
        <PageTable :pageConfig="table.tableConfig">
          <div slot='tableExtend'>
            <div style="margin:8px;border:1px solid #2db7f5">
              <Tag type="border" color="primary">{{$t('m_json_regular')}}</Tag>
              <extendTable :detailConfig="table.tableConfig.table.isExtend.detailConfig"></extendTable>
            </div>
            <div style="margin:8px;border:1px solid #19be6b">
              <Tag type="border" color="success">{{$t('m_metric_regular')}}</Tag>
              <extendTable :detailConfig="table.tableConfig.table.isCustomMetricExtend.detailConfig"></extendTable>
            </div>
          </div>
        </PageTable>
      </div>
    </section>
    <Modal
      v-model="ruleModelConfig.isShow"
      :title="$t('m_json_regular')"
      width="840"
      >
      <div :style="{ 'max-height': MODALHEIGHT + 'px', overflow: 'auto' }">
        <Form :label-width="100">
          <FormItem :label="$t('tableKey.regular')">
            <Input disabled v-model="ruleModelConfig.addRow.json_regular" style="width:100%"/>
          </FormItem>
          <FormItem :label="$t('tableKey.tags')">
            <Input disabled v-model="ruleModelConfig.addRow.tags" style="width:100%" />
          </FormItem>
        </Form>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
          <template v-for="(item, index) in ruleModelConfig.addRow.metric_list">
            <p :key="index + '3'">
              <Input disabled v-model="item.json_key" style="width: 190px" :placeholder="$t('m_key') + ' e.g:[.*][.*]'" />
              <Input disabled v-model="item.metric" style="width: 190px" :placeholder="$t('field.metric') + ' , e.g:code'" />
              <Select disabled v-model="item.agg_type" filterable clearable style="width:190px">
                <Option v-for="agg in ruleModelConfig.aggOption" :value="agg" :key="agg">{{
                  agg
                }}</Option>
              </Select>
              <Input disabled v-model="item.display_name" style="width: 160px" :placeholder="$t('tableKey.description')" />
            </p>
            <div v-if="item.string_map.length > 0" :key="index + 1" style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;">
              <template v-for="(stringMapItem, stringMapIndex) in item.string_map">
                <p :key="stringMapIndex + 2">
                  <Select disabled v-model="stringMapItem.regulative" filterable clearable style="width:230px">
                    <Option v-for="regulation in regulationOption" :value="regulation.value" :key="regulation.value">{{
                      regulation.label
                    }}</Option>
                  </Select>
                  <Input disabled v-model="stringMapItem.source_value" style="width: 230px" :placeholder="$t('m_log_server')" />
                  <Input disabled v-model="stringMapItem.target_value" style="width: 230px" :placeholder="$t('m_business_object')" />
                </p>
              </template>
            </div>
            <Divider :key="index + 'Q'" />
          </template>
        </div>
      </div>
      <div slot="footer">
        <Button @click="ruleModelConfig.isShow=false">{{$t('button.cancel')}}</Button>
      </div>
    </Modal>
    <ModalComponent :modelConfig="customMetricsModelConfig">
      <div slot="ruleConfig" class="extentClass">
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('field.aggType')}}:</label>
          <Select disabled v-model="customMetricsModelConfig.addRow.agg_type" filterable clearable style="width:375px">
            <Option v-for="agg in customMetricsModelConfig.slotConfig.aggOption" :value="agg" :key="agg">{{
              agg
            }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <div style="margin: 4px 12px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
            <template v-for="(item, index) in customMetricsModelConfig.addRow.string_map">
              <p :key="index">
                <Select disabled v-model="item.regulative" filterable clearable style="width:150px">
                  <Option v-for="regulation in regulationOption" :value="regulation.value" :key="regulation.value">{{
                    regulation.label
                  }}</Option>
                </Select>
                <Input disabled v-model="item.source_value" style="width: 150px" :placeholder="$t('m_log_server')" />
                <Input disabled v-model="item.target_value" style="width: 150px" :placeholder="$t('m_business_object')" />
              </p>
            </template>
          </div>
        </div>
        <Button style="float:right" @click="cancelModal">{{$t('button.cancel')}}</Button>
      </div>
    </ModalComponent>
  </div>
</template>

<script>
import extendTable from '@/components/table-page/extend-table'
let tableEle = [
  {title: 'tableKey.path', value: 'log_path', display: true},
  {title: 'field.type', value: 'monitor_type', display: true},
]
const btn = [
  // {btn_name: 'button.remove', btn_func: 'deleteConfirmModal'}
]
export default {
  name: '',
  data () {
    return {
       MODALHEIGHT: 300,
      showManagement: false,
      totalData: [],
      regulationOption: [
        {label: this.$t('m_regular_match'), value: 1},
        {label: this.$t('m_irregular_matching'), value: 0}
      ],
      ruleModelConfig: {
        isShow: false,
        isAdd: true,
        addRow: {
          log_metric_monitor: null,
          json_regular: null,
          tags: null,
          metric_list: []
        },
        aggOption: ['sum', 'avg', 'count', 'max', 'min']
      },
      customMetricsModelConfig: {
        modalId: 'custom_metrics',
        isAdd: true,
        modalStyle: 'min-width:550px',
        modalTitle: 'm_metric_regular',
        noBtn: true,
        config: [
          {label: 'field.metric', value: 'metric', placeholder: '', disabled: true, type: 'text'},
          {label: 'tableKey.description', value: 'display_name', placeholder: '', disabled: true, type: 'text'},
          {label: 'tableKey.regular', value: 'regular', placeholder: 'tips.required', v_validate: 'required:true', disabled: true, type: 'text'},
          {name:'ruleConfig',type:'slot'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          log_metric_monitor: '',
          display_name: '',
          agg_type: 'min',
          metric: null,
          regular: '',
          string_map: []
        },
        slotConfig: {
          aggOption: ['sum', 'avg', 'count', 'max', 'min'],
          regulationOption: [
            {label: this.$t('m_regular_match'), value: 1},
            {label: this.$t('m_irregular_matching'), value: 0}
          ]
        }
      },
      modelTip: {
        key: '',
        value: 'metric'
      },
    }
  },
  methods: {
    cancelModal () {
      this.$root.JQ('#custom_metrics').modal('hide')
    },
    editCustomMetricItem (rowData) {
      this.customMetricsModelConfig.isAdd = false
      this.modelTip.value = rowData.display_name
      this.customMetricsModelConfig.addRow = rowData
      this.$root.JQ('#custom_metrics').modal('show')
    },
    editRuleItem (rowData) {
      this.ruleModelConfig.isAdd = false
      this.ruleModelConfig.addRow = rowData
      this.ruleModelConfig.isShow = true
    },
    getExtendInfo(item){
      const guid = item.guid
      // eslint-disable-next-line no-redeclare
      let index = null
      this.totalData.forEach((td, tdIndex) => {
        const res = td.tableData.config.findIndex(x => x.guid === guid)
        if (res !== -1) {
          index = tdIndex
        }
      })
      item.json_config_list.forEach(xx => xx.pId = item.guid)
      this.totalData[index].tableConfig.table.isExtend.detailConfig[0].data = item.json_config_list
      this.totalData[index].tableConfig.table.isExtend.parentData = item

      item.metric_config_list.forEach(xx => xx.pId = item.guid)
      this.totalData[index].tableConfig.table.isCustomMetricExtend.detailConfig[0].data = item.metric_config_list
      this.totalData[index].tableConfig.table.isCustomMetricExtend.parentData = item
    },
    getDetail (targrtId) {
      this.targrtId = targrtId
      const api = this.$root.apiCenter.getTargetDetail + '/endpoint/' + targrtId
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.totalData = responseData.map(d => {
          let tmp = {}
          tmp.tableData = d
          tmp.tableConfig = {
            table: {
              tableData: d.config,
              tableEle: tableEle,
              // filterMoreBtn: 'filterMoreBtn',
              primaryKey: 'id',
              btn: btn,
              handleFloat:true,
              isExtend: {
                parentData: null,
                func: 'getExtendInfo',
                data: {},
                slot: 'tableExtend',
                detailConfig: [{
                  isExtendF: true,
                  title: '',
                  config: [
                    {title: 'tableKey.regular', value: 'json_regular', display: true},
                    {title: 'tableKey.tags', value: 'tags', display: true},
                    {title: 'table.action',btn:[
                      {btn_name: 'button.view', btn_func: 'editRuleItem'}
                    ]}
                  ],
                  data: [1],
                  scales: ['25%', '20%', '15%', '20%', '20%']
                }]
              },
              isCustomMetricExtend: {
                parentData: null,
                func: 'getExtendInfo',
                data: {},
                slot: 'rulesTableExtend',
                detailConfig: [{
                  isExtendF: true,
                  title: '',
                  config: [
                    {title: 'tableKey.regular', value: 'regular', display: true},
                    {title: 'field.metric', value: 'metric', display: true},
                    {title: 'field.aggType', value: 'agg_type', display: true},
                    {title: 'table.action',btn:[
                      {btn_name: 'button.view', btn_func: 'editCustomMetricItem'}
                    ]}
                  ],
                  data: [1],
                  scales: ['25%', '20%', '15%', '20%', '20%']
                }]
              }
            }
          }
          return tmp
        })
        this.showManagement = true
        this.$root.$store.commit('changeTableExtendActive', -1)
      }, {isNeedloading:false})
    }
  },
  mounted () {
    this.MODALHEIGHT = document.body.scrollHeight - 300
  },
  components: {
    extendTable
  }
}
</script>

<style scoped lang="scss">
</style>
