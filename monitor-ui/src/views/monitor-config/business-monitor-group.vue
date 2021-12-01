/* eslint-disable vue/valid-v-model */
<template>
  <div class=" ">
    <section v-if="showManagement" style="margin-top: 16px;">
      <Tag color="blue">{{targetDetail.display_name}}</Tag>
      <button @click="add" type="button" class="btn btn-sm btn-cancel-f">
        <i class="fa fa-plus"></i>
        {{$t('button.add')}}
      </button>
      <PageTable :pageConfig="pageConfig">
        <!-- <div slot='tableExtend'>
          <div style="margin:8px;border:1px solid #2db7f5">
            <button @click="singleAddF(pageConfig.table.isExtend.parentData)" type="button" style="margin-top:8px" class="btn btn-sm btn-cancel-f">
              <i class="fa fa-plus"></i>
              {{$t('m_add_json_regular')}}
            </button>
            <extendTable :detailConfig="pageConfig.table.isExtend.detailConfig"></extendTable>
          </div>
          <div style="margin:8px;border:1px solid #19be6b">
            <button @click="addCustomMetric(pageConfig.table.isCustomMetricExtend.parentData)" type="button" style="margin-top:8px" class="btn btn-sm btn-cancel-f">
              <i class="fa fa-plus"></i>
              {{$t('m_add_metric_regular')}}
            </button>
            <extendTable :detailConfig="pageConfig.table.isCustomMetricExtend.detailConfig"></extendTable>
          </div>
        </div> -->
      </PageTable>
    </section>
    <Modal
      v-model="addAndEditModal.isShow"
      :title="$t('button.add')"
      
      >
      <!-- @on-ok="okAddAndEdit" -->
      <!-- @on-cancel="cancelAddAndEdit" -->
      <div>
        <div style="margin: 4px 12px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
          <template v-for="(item, index) in addAndEditModal.pathOptions">
            <p :key="index">
              <Button
                @click="deleteItem('path', index)"
                size="small"
                style="background-color: #ff9900;border-color: #ff9900;"
                type="error"
                icon="md-close"
              ></Button>
              <Input v-model="item.path" style="width: 400px" />
            </p>
          </template>
          <Button
            @click="addEmptyItem('path')"
            type="success"
            size="small"
            style="background-color: #0080FF;border-color: #0080FF;"
            long
            >{{ $t('addMetricConfig') }}</Button
          >
        </div>
      </div>
    </Modal>
  </div>
</template>

<script>
let tableEle = [
  {title: 'field.displayName', value: 'display_name', display: true},
  {title: 'field.serviceType', value: 'service_type', display: true}
]
const btn = [
  // {btn_name: 'button.add', btn_func: 'singleAddF'},
  {btn_name: 'button.edit', btn_func: 'editF'},
  {btn_name: 'button.remove', btn_func: 'deleteConfirmModal'},
]
export default {
  name: '',
  data () {
    return {
      showManagement: false,
      pageConfig: {
        table: {
          tableData: [],
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
                {title: 'tableKey.regular', value: 'regular', display: true},
                {title: 'tableKey.tags', value: 'tags', display: true},
                {title: 'table.action',btn:[
                  {btn_name: 'button.edit', btn_func: 'editRuleItem'},
                  {btn_name: 'button.remove', btn_func: 'delRuleconfirmModal'}
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
                {title: 'tableKey.regular', value: 'value_regular', display: true},
                {title: 'field.metric', value: 'metric', display: true},
                {title: 'field.aggType', value: 'agg_type', display: true},
                {title: 'table.action',btn:[
                  {btn_name: 'button.edit', btn_func: 'editCustomMetricItem'},
                  {btn_name: 'button.remove', btn_func: 'delCustomMetricConfirmModal'}
                ]}
              ],
              data: [1],
              scales: ['25%', '20%', '15%', '20%', '20%']
            }]
          }
        }
      },
      addAndEditModal: {
        isShow: false,
        isAdd: false,
        dataConfig: {
          service_group: '',
          log_path: [],
          monitor_type: '',
          endpoint_rel: []
        },
        pathOptions: [],
      },
      showAddAndEditModal: false,

    }
  },
  methods: {
    addEmptyItem (type) {
      if (type === 'path') {
        this.addAndEditModal.pathOptions.push(
          {path: ''}
        )
      }
    },
    deleteItem(type, index) {
      if (type === 'path') {
        this.addAndEditModal.pathOptions.splice(index, 1)
      }
    },
    add () {
      this.addAndEditModal.isAdd = true
      this.addAndEditModal.isShow = true
    },
    getDetail (targrtId) {
      const api = this.$root.apiCenter.getTargetDetail + '/group/' + targrtId
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.showManagement = true
        this.targetDetail = responseData
        this.pageConfig.table.tableData = [responseData]
      }, {isNeedloading:false})
    }
  },
  components: {},
}
</script>

<style scoped lang="scss">
</style>
