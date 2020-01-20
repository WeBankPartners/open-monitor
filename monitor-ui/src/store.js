import Vue from "vue";
import vuex from "vuex";
Vue.use(vuex);

export default new vuex.Store({
  state: {
    ip: {},
    tableExtendActive: -1, //table组件扩展状态
    recursiveChartconfig: {}
  },
  mutations: {
    storeip(state, ip) {
      state.ip = ip;
    },
    changeTableExtendActive(state, index) {
      state.tableExtendActive = index;
    },
    setRecursiveChartconfig(state, config) {
      state.recursiveChartconfig = config;
    }
  }
});
