<template>
  <div class="modal fade" id="deleteConfirmId" tabindex="-1" role="dialog" aria-hidden="true">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header-f">
          <Icon style="font-size:16px;color:#ff6600;" type="ios-information-circle"></Icon>
          <span class="modal-title-">确认删除 ({{confirmInfo.msg}})</span>
          <button type="button" class="close" data-dismiss="modal" aria-label="Close">
            <Icon style="font-size: 26px" type="ios-close" />
          </button>
        </div>
        <div class="modal-body" style="padding:30px">
          <div style="text-align:center">
            <p>即将被删除?</p>
          </div>
        </div>
        <div class="modal-footer">
          <Button @click="cancel">{{ $t('m_button_cancel') }}</Button>
          <Button type="error" @click="exect">{{ $t('m_button_confirm') }}</Button>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  data() {
    return {
      confirmInfo: {
        msg: '',
        callback: null
      }
    }
  },
  created() {
    this.$root.$eventBus.$on('hideConfirmModal', () => {
      this.cancel()
    })
  },
  methods: {
    add(confirmInfo = {}) {
      this.$root.JQ('#deleteConfirmId').modal('show')
      this.confirmInfo = confirmInfo
    },
    cancel() {
      this.$root.JQ('#deleteConfirmId').modal('hide')
    },
    exect() {
      this.confirmInfo.callback()
      this.$root.JQ('#deleteConfirmId').modal('hide')
    }
  }
}

</script>
<style lang="less">
.ivu-icon:focus {
  outline: none;
}
.modal-dialog {
  top: 15%;
  // width: 400px;
  border-radius: 8px;
}
.modal-content {
  border-radius: 0.5rem;
}
.modal-header-f {
  padding: 12px;
  border-bottom: 1px solid #e8eaec;
  text-align: center;
  span {
    font-size: 16px;
    color: #ff6600;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
