<template>
  <div class="modal fade" id="exampleModal" tabindex="-1" role="dialog" aria-labelledby="exampleModalLabel" aria-hidden="true">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" id="exampleModalLabel">Modal title</h5>
          <button type="button" class="close" data-dismiss="modal" aria-label="Close">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          ...
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button>
          <button type="button" class="btn btn-primary">Save changes</button>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
export default {
    data() {
        return {
          isDisplay: false,
          notice: null
        }
    },
    created() {
      this.$root.$eventBus.$on('hideConfirmModal', () => {
        this.cancle()
      })
    },
    methods: {
      add(notice = {}) {
        this.JQ('#exampleModal').modal('show')
        this.notice = notice
      },
      cancle() {
        this.JQ('#exampleModal').modal('hide')
      },
      exect() {
        this.notice.callback()
      }
    }
}

</script>
<style lang="less">
.wrap {
    position: fixed;
    top: 50px;
    left: 50%;
    display: flex;
    flex-direction: column;
    align-items: center;
    transform: translateX(-50%);
}

.message {
    --borderWidth: 3px;
    min-width: 240px;
    max-width: 500px;
    margin-bottom: 10px;
    border-radius: 3px;
    box-shadow: 0 0 8px #ddd;
    overflow: hidden;
}

.content {
    padding: 8px;
    line-height: 1.3;
}

.message.info {
    border-left: var(--borderWidth) solid #909399;
    background: #F4F4F5;
}

.message.success {
    border-left: var(--borderWidth) solid #67C23A;
    background: #F0F9EB;
}

.message.error {
    border-left: var(--borderWidth) solid #F56C6C;
    background: #FEF0F0;
}

.message.warning {
    border-left: var(--borderWidth) solid #E6A23C;
    background: #FDF6EC;
}
</style>