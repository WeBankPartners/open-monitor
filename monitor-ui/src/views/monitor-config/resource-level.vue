<template>
  <div class="">
    <!-- <i class="fa fa-plus" aria-hidden="true" @click="addPanel"> </i> -->
    <button class="btn-confirm-f btn-small" @click="addPanel">新增资源层级</button>
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
          {label: 'field.displayName', value: 'display_name', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          guid: null,
          display_name: null
        }
      },
      id: null
    }
  },
  mounted () {
    this.getAllResource()
  },
  methods: {
    updateDate () {
      console.log(222)
    },
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
   font-size: 16px;
    padding: 6px;
 }
</style>
