<template>
<div class=" ">
  <section>
    <template v-for="(tableItem, tableIndex) in dbMonitorData">
      <section :key="tableIndex + 'f'">
        <div class="section-table-tip">
          <Tag color="blue" :key="tableIndex + 'a'">{{tableItem.sys_panel || '请设置名称'}}</Tag>
          <button type="button" class="btn btn-sm btn-cancel-f" :key="tableIndex + 'b'">
            <i class="fa fa-plus"></i>
            {{$t('button.add')}}
          </button>
        </div>
        <PageTable :pageConfig="pageConfig"></PageTable>
      </section>
    </template>
  </section>
</div>
</template>

<script>
let tableEle = [{
  title: 'tableKey.endpoint',
  value: 'id',
  display: true
}]

export default {
  name: '',
  data() {
    return {
      pageConfig: {
        CRUD: '',
        researchConfig: {
          input_conditions: [{
            value: 'search',
            type: 'input',
            placeholder: 'placeholder.input',
            style: ''
          }],
          btn_group: [{
            btn_name: 'button.search',
            btn_func: 'search',
            class: 'btn-confirm-f',
            btn_icon: 'fa fa-search'
          }],
          filters: {
            search: ''
          }
        },
        table: {
          tableData: [{
            id: 1
          }],
          tableEle: tableEle,
          primaryKey: 'id',
          btn: [],
        }
      },
    }
  },
  props: ['endpointId', 'dbMonitorData'],
  mounted() {
    console.log(this.dbMonitorData, this.endpointId)
  },
  methods: {},
  components: {},
}
</script>

<style scoped lang="less">
</style>
