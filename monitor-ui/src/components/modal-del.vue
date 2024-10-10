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
      <Button @click="ModelDelConfig.deleteWarning = false">{{ $t('m_button_cancel') }}</Button>
      <Button type="error" @click="deleteConfirm">{{ $t('m_delConfirm_button') }}</Button>
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
    deleteConfirm() {
      this.ModelDelConfig.callback()
      this.ModelDelConfig.deleteWarning = false
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
</style>
