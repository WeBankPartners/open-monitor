<template>
  <div id="monitor">
    <div :style="benchStyle">
      <transition name="fade" mode="out-in">
        <router-view></router-view>
      </transition>
      <BaseMenu :menuList="menuList" />
    </div>
  </div>
</template>

<script>
import Vue from 'vue'
Vue.prototype.$bus = new Vue()
export default {
  data () {
    return {
      expand: true,
      menuList: [
            {
                title: this.$t('m_object_design'),
                icon: 'md-cube',
                name: '1',
                children: [
                    { title: this.$t('m_endpoint'), path: '/monitorConfigIndex/endpointManagement', name: '1-1' },
                    { title: this.$t('m_object_group'), path: '/monitorConfigIndex/groupManagement', name: '1-2' },
                    { title: this.$t('m_field_resourceLevel'), path: '/monitorConfigIndex/resourceLevel', name: '1-3' }
                ]
            },
            {
                title: this.$t('m_metric'),
                icon: 'md-trending-up',
                name: '2',
                children: [
                    { title: this.$t('m_business_configuration'), path: '/monitorConfigIndex/businessMonitor', name: '2-1' },
                    { title: this.$t('m_business_log_template'), path: '/monitorConfigIndex/logTemplate', name: '2-2' },
                    { title: this.$t('m_metric_list'), path: '/monitorConfigIndex/metricConfig', name: '2-3' }
                ]
            },
            {
                title: this.$t('m_menu_alert'),
                icon: 'md-warning',
                name: '3',
                children: [
                    { title: this.$t('m_metric_threshold'), path: '/monitorConfigIndex/thresholdManagement', name: '3-1' },
                    { title: this.$t('m_field_log'), path: '/monitorConfigIndex/logManagement', name: '3-2' }
                ]
            }
        ]
    }
  },
  computed: {
    benchStyle () {
      return {
        paddingLeft: this.expand ? '140px' : '0px'
      }
    }
  },
  mounted () {
    if (this.$eventBusP) {
      this.$eventBusP.$on('expand-menu', val => {
        this.expand = val
      })
    } else {
      this.$bus.$on('expand-menu', val => {
        this.expand = val
      })
    }
  },
  methods: {}
}
</script>
