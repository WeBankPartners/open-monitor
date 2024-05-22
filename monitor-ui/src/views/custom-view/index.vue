<template>
  <div id="workbench">
    <div :style="benchStyle">
      <transition name="fade" mode="out-in">
        <router-view></router-view>
      </transition>
      <BenchMenu :menuList="menuList" @menuStatusChange="onMenuChange"></BenchMenu>
    </div>
  </div>
</template>

<script>
import BenchMenu from '@/components/bench-menu'
import Vue from 'vue'
Vue.prototype.$bus = new Vue()
export default {
  components: {
    BenchMenu
  },
  data () {
    return {
      expand: true,
      menuList: [
            {
                title: this.$t('m_screen'),
                icon: 'ios-albums',
                name: '1',
                children: [
                    { title: this.$t('m_create'), path: '/monitorConfigIndex/endpointManagement', name: '1-1' },
                    { title: this.$t('m_list'), path: '/viewConfigIndex/boardList', name: '1-2' }
                ]
            },
            {
                title: this.$t('m_chart_library'),
                icon: 'md-pulse',
                name: '2',
                children: [
                    { title: this.$t('m_list'), path: '/monitorConfigIndex/businessMonitor', name: '2-1' }
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
  methods: {
    onMenuChange(val) {
        this.expand = val
    }
  }
}
</script>
