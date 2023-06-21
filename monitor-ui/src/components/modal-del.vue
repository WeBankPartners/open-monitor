<template>
  <Modal v-model="ModelDelConfig.deleteWarning" width="360" :mask-closable="false">
    <p slot="header" style="color:#f60;text-align:center">
      <Icon type="ios-information-circle"></Icon>
      <span>Delete confirmation ({{ModelDelConfig.msg}})</span>
    </p>
    <div style="text-align:center">
      <p>Will you delete it?</p>
    </div>
    <div slot="footer">
      <button class="btn-cancel-f" @click="ModelDelConfig.deleteWarning = false">Cancel</button>
      <button class="btn-delete-f" @click="deleteConfirm">Delete</button>
    </div>
  </Modal>
</template>

<script>
export default {
  name: '',
  data() {
    return {
    }
  },
  props: ['ModelDelConfig'],
  methods: {
    deleteConfirm () {
      this.ModelDelConfig.callback()
      this.ModelDelConfig.deleteWarning = false
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
</style>
