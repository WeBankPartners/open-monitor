<template>
  <div class="">
    <button class="btn-confirm-f btn-small" @click="addPanel">{{$t('resourceLevel.addPanel')}}</button>
    <i class="fa fa-refresh" aria-hidden="true" @click="getAllResource"></i>
    <recursive :recursiveViewConfig="resourceRecursive"></recursive>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>

<script>
import recursive from '@/views/monitor-config/resource-recursive'
export default {
  name: '',
  data() {
    return {
      resourceRecursive: null,
      modelConfig: {
        modalId: 'add_panel_Modal',
        modalTitle: 'button.add',
        isAdd: true,
        config: [
          {label: 'field.guid', value: 'guid', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'},
          {label: 'field.displayName', value: 'display_name', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'},
          {label: 'field.type', value: 'type', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          guid: null,
          display_name: null
        }
      },
      id: null
    }
  },
  created () {
    this.$root.$eventBus.$on('updateResource', () => {
      this.getAllResource()
    })
  },
  mounted () {
    this.getAllResource()
  },
  methods: {
    getAllResource () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceLevel.getAll, '', (responseData) => {
        this.resourceRecursive = responseData
      })
    },
    addPanel () {
      this.modelConfig.isAdd = true
      this.$root.JQ('#add_panel_Modal').modal('show')
    },
    addPost () {
      const params = this.modelConfig.addRow
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', 'alarm/org/panel/add', params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#add_panel_Modal').modal('hide')
        this.getAllResource()
      })
    }
  },
  components: {
    recursive
  },
}
</script>

<style scoped lang="less">
 .fa {
   margin-left: 4px;
   font-size: 14px;
   padding: 4px 6px;
   border: 1px solid #2d8cf0;
   border-radius: 4px;
 }
</style>
