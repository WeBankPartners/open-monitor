<template>
  <div>
    <Modal v-model="flowRoleManageModal" width="800" :title="$t('m_role_drawer_title')" :mask-closable="false">
      <div style="width: 100%; overflow-x: auto">
        <div style="min-width: 760px" class="content">
          <div>
            <slot name="content-top"></slot>
            <div class="role-transfer-title">{{ $t('m_mgmt_role') }}</div>
            <div class="role-transfer-content">
              <Transfer
                :titles="transferTitles"
                :list-style="transferStyle"
                :data="currentUserRoles"
                :target-keys="mgmtRolesKeyToFlow"
                :render-format="renderRoleNameForTransfer"
                @on-change="handleMgmtRoleTransferChange"
                filterable
              ></Transfer>
            </div>
          </div>
          <div style="margin-top: 30px">
            <div class="role-transfer-title">{{ $t('m_use_role') }}</div>
            <div class="role-transfer-content">
              <Transfer
                :titles="transferTitles"
                :list-style="transferStyle"
                :data="allRoles"
                :target-keys="useRolesKeyToFlow"
                :render-format="renderRoleNameForTransfer"
                @on-change="handleUseRoleTransferChange"
                filterable
              ></Transfer>
            </div>
          </div>
        </div>
      </div>
      <div slot="footer">
        <Button type="primary" :disabled="disabled" @click="confirmRole">{{ $t('m_button_confirm') }}</Button>
      </div>
    </Modal>
  </div>
</template>
<script>
export default {
  props: {
    useRolesRequired: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      isAdd: false, // 标记编排状态
      flowRoleManageModal: false, // 权限弹窗控制
      transferTitles: [this.$t('m_unselected_role'), this.$t('m_selected_role')],
      transferStyle: { width: '300px' },
      allRoles: [],
      currentUserRoles: [],
      mgmtRolesKeyToFlow: [], // 管理角色
      useRolesKeyToFlow: [] // 使用角色
    }
  },
  computed: {
    disabled() {
      if (this.useRolesRequired) {
        return this.mgmtRolesKeyToFlow.length === 0 || this.useRolesKeyToFlow.length === 0
      }

      return this.mgmtRolesKeyToFlow.length === 0

    }
  },
  methods: {
    renderRoleNameForTransfer(item) {
      return item.label
    },
    handleMgmtRoleTransferChange(newTargetKeys) {
      if (newTargetKeys.length > 1) {
        this.$Message.warning(this.$t('m_choose_one'))
      } else {
        this.mgmtRolesKeyToFlow = newTargetKeys
      }
    },
    handleUseRoleTransferChange(newTargetKeys) {
      this.useRolesKeyToFlow = newTargetKeys
    },
    async confirmRole() {
      this.$emit('sendAuth', this.mgmtRolesKeyToFlow, this.useRolesKeyToFlow)
      this.flowRoleManageModal = false
    },
    // 启动入口
    async startAuth(mgmtRoles, userRoles, mgmtRolesOptions, userRolesOptions) {
      this.mgmtRolesKeyToFlow = mgmtRoles
      this.useRolesKeyToFlow = userRoles
      this.allRoles = userRolesOptions
      this.currentUserRoles = mgmtRolesOptions
      this.flowRoleManageModal = true
    }
  }
}
</script>

<style lang="less" scoped>
.role-transfer-content {
  display: flex;
  justify-content: center
}
</style>
