<template>
  <div id="k8s-config">
    <div :style="benchStyle">
      <transition name="fade" mode="out-in">
        <router-view :key="$route.fullPath"></router-view>
      </transition>
      <BaseMenu :menuList="menuList"></BaseMenu>
    </div>
  </div>
</template>

<script>
import Vue from 'vue'
Vue.prototype.$bus = new Vue()
export default {
  data() {
    return {
      expand: true,
      menuList: [{
        title: this.$t('m_k8s_config'),
        icon: 'md-cube',
        name: '1',
        children: [
          {title: this.$t('m_k8s_sources'), path: '/k8sConfig/k8sSources', name: '1-1'},
          {title: this.$t('m_k8s_objects'), path: '/k8sConfig/k8sObjects', name: '1-2'},
          {title: this.$t('m_k8s_alert_rules'), path: '/k8sConfig/k8sAlertRules', name: '1-3'}
        ]
      }]
    }
  },
  computed: {
    benchStyle() {
      return {paddingLeft: this.expand ? '140px' : '0px'}
    }
  },
  mounted() {
    this.menuBus = this.$eventBusP || this.$bus
    this.onExpand = val => { this.expand = val }
    this.menuBus.$on('expand-menu', this.onExpand)
  },
  beforeDestroy() {
    this.menuBus.$off('expand-menu', this.onExpand)
  }
}
</script>
