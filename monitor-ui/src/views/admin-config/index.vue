<template>
  <div id="admin-config">
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
      menuList: [
        {
          // 基础类型
          title: this.$t('m_basic_type'),
          icon: 'md-pulse',
          name: '1',
          children: [
            // 类型配置
            {
              title: this.$t('m_type_config'),
              path: '/adminConfig/typeConfig',
              name: '1-1'
            },
            // 看板配置
            {
              title: this.$t('m_menu_screenConfiguration'),
              path: '/adminConfig/groupBoard',
              name: '1-2'
            },
            // 指标配置
            {
              title: this.$t('m_menu_metricConfiguration'),
              path: '/adminConfig/adminMetric',
              name: '1-3'
            }
          ]
        },
        {
          // 其它
          title: this.$t('m_other'),
          icon: 'ios-more',
          name: '2',
          children: [
            {
              title: this.$t('m_field_exporter'),
              path: '/adminConfig/exporter',
              name: '2-1'
            },
            {
              title: this.$t('m_remote_sync'),
              path: '/adminConfig/remoteSync',
              name: '2-2'
            },
            {
              title: this.$t('m_prometheus_logs'),
              path: '/adminConfig/prometheusLogs',
              name: '2-3'
            }
          ]
        }
      ]
    }
  },
  computed: {
    benchStyle() {
      return {
        paddingLeft: this.expand ? '140px' : '0px'
      }
    }
  },
  mounted() {
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
