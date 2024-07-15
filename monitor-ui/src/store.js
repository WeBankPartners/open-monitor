import Vue from 'vue'
import vuex from 'vuex'
Vue.use(vuex)

export default new vuex.Store({
  state: {
    ip: {},
    tableExtendActive: -1, // table组件扩展状态
    cacheTagColor: {} // 递归视图缓存tag颜色指标
  },
  mutations: {
    storeip(state, ip) {
      state.ip = ip
    },
    changeTableExtendActive(state, index) {
      state.tableExtendActive = index
    },
    cacheTagColor(state, tagColor) {
      state.cacheTagColor = tagColor
    }
  }
})
