<template>
  <div id="workbench">
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
          title: this.$t('m_screen'),
          icon: 'ios-albums',
          name: '1',
          children: [
            {
              title: this.$t('m_create'),
              path: '/viewConfigIndex/boardList?isCreate=true',
              name: '1-1'
            },
            {
              title: this.$t('m_list'),
              path: '/viewConfigIndex/boardList',
              name: '1-2'
            }
          ]
        },
        {
          title: this.$t('m_chart_library'),
          icon: 'md-pulse',
          name: '2',
          children: [
            {
              title: this.$t('m_list'),
              path: '/viewConfigIndex/allChartList',
              name: '2-1'
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
